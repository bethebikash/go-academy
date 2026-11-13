package analytics

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"go-academy/backend/internal/database"
	"go-academy/backend/internal/database/sqlc"
)

var (
	ErrNotFound         = errors.New("record not found")
	ErrPathNotCompleted = errors.New("learning path is not yet 100% completed")
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

func toPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func toPgDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

// RecordActivity records a user activity, updates streaks, and triggers automatic achievements.
func (s *Service) RecordActivity(ctx context.Context, userID uuid.UUID, activityType string, metadata map[string]interface{}) error {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	metaJSON, _ := json.Marshal(metadata)

	// 1. Insert activity log
	_, err := s.db.Queries.RecordUserActivity(ctx, sqlc.RecordUserActivityParams{
		UserID:       toPgUUID(userID),
		ActivityType: activityType,
		ActivityDate: toPgDate(today),
		Metadata:     metaJSON,
	})
	if err != nil {
		return fmt.Errorf("failed to record user activity: %w", err)
	}

	// 2. Calculate and update streak
	currentStreak := 1
	longestStreak := 1

	streakRow, err := s.db.Queries.GetUserStreak(ctx, toPgUUID(userID))
	if err == nil {
		if streakRow.LastActivityDate.Valid {
			lastDate := streakRow.LastActivityDate.Time
			lastDay := time.Date(lastDate.Year(), lastDate.Month(), lastDate.Day(), 0, 0, 0, 0, time.UTC)
			diffDays := int(today.Sub(lastDay).Hours() / 24)

			if diffDays == 0 {
				// Already active today; preserve existing streak
				currentStreak = int(streakRow.CurrentStreak)
				longestStreak = int(streakRow.LongestStreak)
			} else if diffDays == 1 {
				// Consecutive day
				currentStreak = int(streakRow.CurrentStreak) + 1
				longestStreak = int(streakRow.LongestStreak)
				if currentStreak > longestStreak {
					longestStreak = currentStreak
				}
			} else {
				// Broken streak
				currentStreak = 1
				longestStreak = int(streakRow.LongestStreak)
			}
		}
	}

	_, err = s.db.Queries.UpsertUserStreak(ctx, sqlc.UpsertUserStreakParams{
		UserID:           toPgUUID(userID),
		CurrentStreak:    int32(currentStreak),
		LongestStreak:    int32(longestStreak),
		LastActivityDate: toPgDate(today),
	})
	if err != nil {
		return fmt.Errorf("failed to update user streak: %w", err)
	}

	// 3. Evaluate achievements
	switch activityType {
	case ActivityLessonCompleted:
		_ = s.AwardBadge(ctx, userID, BadgeFirstStep)
	case ActivityChallengePassed:
		_ = s.AwardBadge(ctx, userID, BadgeBugHunter)
	case ActivityNoteCreated:
		stats, err := s.db.Queries.GetUserStats(ctx, toPgUUID(userID))
		if err == nil && stats.TotalNotes >= 5 {
			_ = s.AwardBadge(ctx, userID, BadgeScholar)
		}
	}

	if currentStreak >= 7 {
		_ = s.AwardBadge(ctx, userID, BadgeStreak7)
	}

	return nil
}

// AwardBadge grants a badge if not already unlocked.
func (s *Service) AwardBadge(ctx context.Context, userID uuid.UUID, badgeKey string) error {
	_, err := s.db.Queries.AwardBadge(ctx, sqlc.AwardBadgeParams{
		UserID:   toPgUUID(userID),
		BadgeKey: badgeKey,
	})
	return err
}

// GetSummaryStats fetches top-level analytics, streak, and counters for the user.
func (s *Service) GetSummaryStats(ctx context.Context, userID uuid.UUID) (*UserSummaryStats, error) {
	streakInfo := StreakInfo{
		CurrentStreak: 0,
		LongestStreak: 0,
	}

	streakRow, err := s.db.Queries.GetUserStreak(ctx, toPgUUID(userID))
	if err == nil {
		streakInfo.CurrentStreak = int(streakRow.CurrentStreak)
		streakInfo.LongestStreak = int(streakRow.LongestStreak)
		if streakRow.LastActivityDate.Valid {
			t := streakRow.LastActivityDate.Time
			streakInfo.LastActivityDate = &t
		}
	}

	stats, err := s.db.Queries.GetUserStats(ctx, toPgUUID(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	badges, err := s.db.Queries.ListUserBadges(ctx, toPgUUID(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to get user badges: %w", err)
	}

	return &UserSummaryStats{
		Streak:            streakInfo,
		CompletedLessons:  int(stats.CompletedLessons),
		SolvedChallenges:  int(stats.SolvedChallenges),
		TotalNotes:        int(stats.TotalNotes),
		TotalCertificates: int(stats.TotalCertificates),
		EarnedBadgesCount: len(badges),
		TotalBadgesCount:  len(AvailableBadges),
	}, nil
}

// GetActivityHistory returns daily activity counts for heatmap visualization.
func (s *Service) GetActivityHistory(ctx context.Context, userID uuid.UUID, days int) ([]ActivityCount, error) {
	if days <= 0 {
		days = 90
	}
	since := time.Now().UTC().AddDate(0, 0, -days)

	rows, err := s.db.Queries.CountUserActivitiesByDate(ctx, sqlc.CountUserActivitiesByDateParams{
		UserID:       toPgUUID(userID),
		ActivityDate: toPgDate(since),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count activities: %w", err)
	}

	results := make([]ActivityCount, len(rows))
	for i, r := range rows {
		results[i] = ActivityCount{
			Date:  r.ActivityDate.Time.Format("2006-01-02"),
			Count: int(r.Count),
		}
	}

	return results, nil
}

// GetUserBadges returns all badges annotated with awarded status.
func (s *Service) GetUserBadges(ctx context.Context, userID uuid.UUID) ([]UserBadge, error) {
	earnedRows, err := s.db.Queries.ListUserBadges(ctx, toPgUUID(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to list user badges: %w", err)
	}

	earnedMap := make(map[string]time.Time, len(earnedRows))
	for _, r := range earnedRows {
		earnedMap[r.BadgeKey] = r.AwardedAt.Time
	}

	result := make([]UserBadge, len(AvailableBadges))
	for i, b := range AvailableBadges {
		ub := UserBadge{
			BadgeDefinition: b,
			Awarded:         false,
		}
		if awardedAt, ok := earnedMap[b.Key]; ok {
			ub.Awarded = true
			ub.AwardedAt = &awardedAt
		}
		result[i] = ub
	}

	return result, nil
}

// ClaimCertificate verifies path completion and generates an official certificate.
func (s *Service) ClaimCertificate(ctx context.Context, userID uuid.UUID, pathID uuid.UUID) (*CertificateInfo, error) {
	// 1. Check if already claimed
	existing, err := s.db.Queries.GetCertificateByUserAndPath(ctx, sqlc.GetCertificateByUserAndPathParams{
		UserID:         toPgUUID(userID),
		LearningPathID: toPgUUID(pathID),
	})
	if err == nil {
		return &CertificateInfo{
			ID:                uuid.UUID(existing.ID.Bytes),
			UserID:            uuid.UUID(existing.UserID.Bytes),
			LearningPathID:    uuid.UUID(existing.LearningPathID.Bytes),
			CertificateCode:   existing.CertificateCode,
			RecipientName:     existing.RecipientName,
			LearningPathTitle: existing.LearningPathTitle,
			IssuedAt:          existing.IssuedAt.Time,
		}, nil
	}

	// 2. Fetch learning path details
	pathRow, err := s.db.Queries.GetLearningPathByID(ctx, toPgUUID(pathID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get learning path: %w", err)
	}

	// 3. Verify user completed all lessons in this path
	courses, err := s.db.Queries.ListPublishedCoursesByPathID(ctx, pathRow.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get path courses: %w", err)
	}

	totalLessons := 0
	completedLessons := 0

	for _, c := range courses {
		modules, err := s.db.Queries.ListModulesByCourseID(ctx, c.ID)
		if err != nil {
			continue
		}
		for _, m := range modules {
			lessons, err := s.db.Queries.ListPublishedLessonsByModuleID(ctx, m.ID)
			if err != nil {
				continue
			}
			for _, l := range lessons {
				totalLessons++
				prog, err := s.db.Queries.GetLessonProgress(ctx, sqlc.GetLessonProgressParams{
					UserID:   toPgUUID(userID),
					LessonID: l.ID,
				})
				if err == nil && prog.Status == "completed" {
					completedLessons++
				}
			}
		}
	}

	if totalLessons == 0 || completedLessons < totalLessons {
		return nil, fmt.Errorf("%w: %d of %d lessons completed", ErrPathNotCompleted, completedLessons, totalLessons)
	}

	// 4. Fetch user display name
	userRow, err := s.db.Queries.GetUserByID(ctx, toPgUUID(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	recipientName := userRow.Username

	// 5. Generate unique certificate code (e.g. GA-2026-A1B2C3D4)
	bytes := make([]byte, 4)
	_, _ = rand.Read(bytes)
	code := fmt.Sprintf("GA-%d-%s", time.Now().Year(), strings.ToUpper(hex.EncodeToString(bytes)))

	certRow, err := s.db.Queries.CreateCertificate(ctx, sqlc.CreateCertificateParams{
		UserID:            toPgUUID(userID),
		LearningPathID:    pathRow.ID,
		CertificateCode:   code,
		RecipientName:     recipientName,
		LearningPathTitle: pathRow.Title,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to issue certificate: %w", err)
	}

	// Award Path Master badge
	_ = s.AwardBadge(ctx, userID, BadgePathMaster)

	return &CertificateInfo{
		ID:                uuid.UUID(certRow.ID.Bytes),
		UserID:            uuid.UUID(certRow.UserID.Bytes),
		LearningPathID:    uuid.UUID(certRow.LearningPathID.Bytes),
		CertificateCode:   certRow.CertificateCode,
		RecipientName:     certRow.RecipientName,
		LearningPathTitle: certRow.LearningPathTitle,
		IssuedAt:          certRow.IssuedAt.Time,
	}, nil
}

// GetCertificateByCode fetches certificate details for verification.
func (s *Service) GetCertificateByCode(ctx context.Context, code string) (*CertificateInfo, error) {
	row, err := s.db.Queries.GetCertificateByCode(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get certificate: %w", err)
	}

	return &CertificateInfo{
		ID:                uuid.UUID(row.ID.Bytes),
		UserID:            uuid.UUID(row.UserID.Bytes),
		LearningPathID:    uuid.UUID(row.LearningPathID.Bytes),
		CertificateCode:   row.CertificateCode,
		RecipientName:     row.RecipientName,
		LearningPathTitle: row.LearningPathTitle,
		IssuedAt:          row.IssuedAt.Time,
	}, nil
}

// ListUserCertificates lists all certificates awarded to a learner.
func (s *Service) ListUserCertificates(ctx context.Context, userID uuid.UUID) ([]CertificateInfo, error) {
	rows, err := s.db.Queries.ListCertificatesByUser(ctx, toPgUUID(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to list user certificates: %w", err)
	}

	res := make([]CertificateInfo, len(rows))
	for i, r := range rows {
		res[i] = CertificateInfo{
			ID:                uuid.UUID(r.ID.Bytes),
			UserID:            uuid.UUID(r.UserID.Bytes),
			LearningPathID:    uuid.UUID(r.LearningPathID.Bytes),
			CertificateCode:   r.CertificateCode,
			RecipientName:     r.RecipientName,
			LearningPathTitle: r.LearningPathTitle,
			IssuedAt:          r.IssuedAt.Time,
		}
	}

	return res, nil
}
