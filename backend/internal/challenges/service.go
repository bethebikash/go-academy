package challenges

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"go-academy/backend/internal/database"
	"go-academy/backend/internal/database/sqlc"
)

var (
	ErrNotFound     = errors.New("challenge not found")
	ErrSlugExists   = errors.New("slug already in use")
	ErrInvalidInput = errors.New("invalid input")
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

// GenerateSlug converts title to URL-friendly string.
func GenerateSlug(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "challenge-" + uuid.New().String()[:8]
	}
	return s
}

// ListPublishedChallenges lists published challenges for learners.
func (s *Service) ListPublishedChallenges(ctx context.Context, difficulty string) ([]ChallengeSummary, error) {
	rows, err := s.db.Queries.ListPublishedChallenges(ctx, difficulty)
	if err != nil {
		return nil, fmt.Errorf("failed to list published challenges: %w", err)
	}

	summaries := make([]ChallengeSummary, len(rows))
	for i, r := range rows {
		item := ChallengeSummary{
			ID:               uuid.UUID(r.ID.Bytes),
			Title:            r.Title,
			Slug:             r.Slug,
			Description:      r.Description,
			Difficulty:       r.Difficulty,
			EstimatedMinutes: int(r.EstimatedMinutes),
			StarterCode:      r.StarterCode,
			Status:           r.Status,
			SortOrder:        int(r.SortOrder),
			CreatedAt:        r.CreatedAt.Time,
			UpdatedAt:        r.UpdatedAt.Time,
		}
		if r.CourseID.Valid {
			id := uuid.UUID(r.CourseID.Bytes)
			item.CourseID = &id
		}
		if r.LessonID.Valid {
			id := uuid.UUID(r.LessonID.Bytes)
			item.LessonID = &id
		}
		if r.PublishedAt.Valid {
			t := r.PublishedAt.Time
			item.PublishedAt = &t
		}
		summaries[i] = item
	}

	return summaries, nil
}

// GetPublishedChallengeBySlug returns challenge details with visible sample test cases.
func (s *Service) GetPublishedChallengeBySlug(ctx context.Context, slug string) (*ChallengeDetail, error) {
	row, err := s.db.Queries.GetPublishedChallengeBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get published challenge: %w", err)
	}

	tcRows, err := s.db.Queries.ListSampleTestCasesByChallengeID(ctx, row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sample test cases: %w", err)
	}

	testCases := make([]TestCase, len(tcRows))
	for i, tc := range tcRows {
		testCases[i] = TestCase{
			ID:             uuid.UUID(tc.ID.Bytes),
			ChallengeID:    uuid.UUID(tc.ChallengeID.Bytes),
			Title:          tc.Title,
			Input:          tc.Input,
			ExpectedOutput: tc.ExpectedOutput,
			Hidden:         tc.Hidden,
			SortOrder:      int(tc.SortOrder),
			CreatedAt:      tc.CreatedAt.Time,
		}
	}

	detail := &ChallengeDetail{
		ChallengeSummary: ChallengeSummary{
			ID:               uuid.UUID(row.ID.Bytes),
			Title:            row.Title,
			Slug:             row.Slug,
			Description:      row.Description,
			Difficulty:       row.Difficulty,
			EstimatedMinutes: int(row.EstimatedMinutes),
			StarterCode:      row.StarterCode,
			Status:           row.Status,
			SortOrder:        int(row.SortOrder),
			CreatedAt:        row.CreatedAt.Time,
			UpdatedAt:        row.UpdatedAt.Time,
		},
		SampleTestCases: testCases,
	}

	if row.CourseID.Valid {
		id := uuid.UUID(row.CourseID.Bytes)
		detail.CourseID = &id
		course, err := s.db.Queries.GetCourseByID(ctx, row.CourseID)
		if err == nil {
			detail.CourseTitle = course.Title
			detail.CourseSlug = course.Slug
		}
	}
	if row.LessonID.Valid {
		id := uuid.UUID(row.LessonID.Bytes)
		detail.LessonID = &id
	}
	if row.PublishedAt.Valid {
		t := row.PublishedAt.Time
		detail.PublishedAt = &t
	}

	return detail, nil
}

// --- Admin Methods ---

// ListAllChallenges lists all challenges regardless of status.
func (s *Service) ListAllChallenges(ctx context.Context) ([]ChallengeSummary, error) {
	rows, err := s.db.Queries.ListAllChallenges(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list all challenges: %w", err)
	}

	res := make([]ChallengeSummary, len(rows))
	for i, r := range rows {
		res[i] = fromSQLCChallenge(r)
	}
	return res, nil
}

// GetChallengeByID fetches full challenge details including solution code and all test cases.
func (s *Service) GetChallengeByID(ctx context.Context, id uuid.UUID) (*AdminChallengeDetail, error) {
	row, err := s.db.Queries.GetChallengeByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get challenge: %w", err)
	}

	tcRows, err := s.db.Queries.ListAllTestCasesByChallengeID(ctx, row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list test cases: %w", err)
	}

	testCases := make([]TestCase, len(tcRows))
	for i, tc := range tcRows {
		testCases[i] = fromSQLCTestCase(tc)
	}

	summary := fromSQLCChallenge(row)
	return &AdminChallengeDetail{
		ChallengeSummary: summary,
		SolutionCode:     row.SolutionCode,
		TestCases:        testCases,
	}, nil
}

// CreateChallenge creates a challenge with test cases.
func (s *Service) CreateChallenge(ctx context.Context, req CreateChallengeRequest) (*AdminChallengeDetail, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = GenerateSlug(title)
	}

	difficulty := req.Difficulty
	if difficulty == "" {
		difficulty = DifficultyBeginner
	}

	estMin := req.EstimatedMinutes
	if estMin <= 0 {
		estMin = 15
	}

	status := req.Status
	if status == "" {
		status = StatusDraft
	}

	var pubAt *time.Time
	if status == StatusPublished {
		now := time.Now()
		pubAt = &now
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.db.Queries.WithTx(tx)

	created, err := qtx.CreateChallenge(ctx, sqlc.CreateChallengeParams{
		CourseID:         toOptionalPgUUID(req.CourseID),
		LessonID:         toOptionalPgUUID(req.LessonID),
		Title:            title,
		Slug:             slug,
		Description:      req.Description,
		Difficulty:       difficulty,
		EstimatedMinutes: int32(estMin),
		StarterCode:      req.StarterCode,
		SolutionCode:     req.SolutionCode,
		Status:           status,
		SortOrder:        int32(req.SortOrder),
		PublishedAt:      toPgTimestamptz(pubAt),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrSlugExists
		}
		return nil, fmt.Errorf("failed to create challenge: %w", err)
	}

	// Insert test cases
	testCases := make([]TestCase, 0, len(req.TestCases))
	for _, tcReq := range req.TestCases {
		tc, err := qtx.CreateTestCase(ctx, sqlc.CreateTestCaseParams{
			ChallengeID:    created.ID,
			Title:          tcReq.Title,
			Input:          tcReq.Input,
			ExpectedOutput: tcReq.ExpectedOutput,
			Hidden:         tcReq.Hidden,
			SortOrder:      int32(tcReq.SortOrder),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create test case %q: %w", tcReq.Title, err)
		}
		testCases = append(testCases, fromSQLCTestCase(tc))
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	summary := fromSQLCChallenge(created)
	return &AdminChallengeDetail{
		ChallengeSummary: summary,
		SolutionCode:     created.SolutionCode,
		TestCases:        testCases,
	}, nil
}

// UpdateChallenge updates challenge metadata and test cases.
func (s *Service) UpdateChallenge(ctx context.Context, id uuid.UUID, req UpdateChallengeRequest) (*AdminChallengeDetail, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = GenerateSlug(title)
	}

	difficulty := req.Difficulty
	if difficulty == "" {
		difficulty = DifficultyBeginner
	}

	estMin := req.EstimatedMinutes
	if estMin <= 0 {
		estMin = 15
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.db.Queries.WithTx(tx)

	updated, err := qtx.UpdateChallenge(ctx, sqlc.UpdateChallengeParams{
		ID:               toPgUUID(id),
		CourseID:         toOptionalPgUUID(req.CourseID),
		LessonID:         toOptionalPgUUID(req.LessonID),
		Title:            title,
		Slug:             slug,
		Description:      req.Description,
		Difficulty:       difficulty,
		EstimatedMinutes: int32(estMin),
		StarterCode:      req.StarterCode,
		SolutionCode:     req.SolutionCode,
		SortOrder:        int32(req.SortOrder),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrSlugExists
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to update challenge: %w", err)
	}

	// Update test cases if provided
	var testCases []TestCase
	if req.TestCases != nil {
		if err := qtx.ClearTestCases(ctx, toPgUUID(id)); err != nil {
			return nil, fmt.Errorf("failed to clear test cases: %w", err)
		}

		testCases = make([]TestCase, 0, len(req.TestCases))
		for _, tcReq := range req.TestCases {
			tc, err := qtx.CreateTestCase(ctx, sqlc.CreateTestCaseParams{
				ChallengeID:    toPgUUID(id),
				Title:          tcReq.Title,
				Input:          tcReq.Input,
				ExpectedOutput: tcReq.ExpectedOutput,
				Hidden:         tcReq.Hidden,
				SortOrder:      int32(tcReq.SortOrder),
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create test case %q: %w", tcReq.Title, err)
			}
			testCases = append(testCases, fromSQLCTestCase(tc))
		}
	} else {
		existing, err := qtx.ListAllTestCasesByChallengeID(ctx, toPgUUID(id))
		if err == nil {
			testCases = make([]TestCase, len(existing))
			for i, tc := range existing {
				testCases[i] = fromSQLCTestCase(tc)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	summary := fromSQLCChallenge(updated)
	return &AdminChallengeDetail{
		ChallengeSummary: summary,
		SolutionCode:     updated.SolutionCode,
		TestCases:        testCases,
	}, nil
}

// UpdateChallengeStatus toggles status of a challenge.
func (s *Service) UpdateChallengeStatus(ctx context.Context, id uuid.UUID, status string) (*ChallengeSummary, error) {
	if status != StatusDraft && status != StatusPublished && status != StatusArchived {
		return nil, fmt.Errorf("%w: status must be draft, published, or archived", ErrInvalidInput)
	}

	var pubAt *time.Time
	if status == StatusPublished {
		now := time.Now()
		pubAt = &now
	}

	updated, err := s.db.Queries.UpdateChallengeStatus(ctx, sqlc.UpdateChallengeStatusParams{
		ID:          toPgUUID(id),
		Status:      status,
		PublishedAt: toPgTimestamptz(pubAt),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to update challenge status: %w", err)
	}

	summary := fromSQLCChallenge(updated)
	return &summary, nil
}

// DeleteChallenge deletes a challenge by ID.
func (s *Service) DeleteChallenge(ctx context.Context, id uuid.UUID) error {
	err := s.db.Queries.DeleteChallenge(ctx, toPgUUID(id))
	if err != nil {
		return fmt.Errorf("failed to delete challenge: %w", err)
	}
	return nil
}
