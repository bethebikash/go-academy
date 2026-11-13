package analytics

import (
	"time"

	"github.com/google/uuid"
)

// Activity Types
const (
	ActivityLessonCompleted = "lesson_completed"
	ActivityChallengePassed = "challenge_passed"
	ActivityNoteCreated     = "note_created"
)

// Badge Keys
const (
	BadgeFirstStep  = "first_step"
	BadgeBugHunter  = "bug_hunter"
	BadgeScholar    = "scholar"
	BadgeStreak7    = "streak_7"
	BadgePathMaster = "path_master"
)

type StreakInfo struct {
	CurrentStreak    int        `json:"current_streak"`
	LongestStreak    int        `json:"longest_streak"`
	LastActivityDate *time.Time `json:"last_activity_date,omitempty"`
}

type ActivityCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type BadgeDefinition struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Icon        string `json:"icon"`
}

type UserBadge struct {
	BadgeDefinition
	Awarded   bool       `json:"awarded"`
	AwardedAt *time.Time `json:"awarded_at,omitempty"`
}

type CertificateInfo struct {
	ID                uuid.UUID `json:"id"`
	UserID            uuid.UUID `json:"user_id"`
	LearningPathID    uuid.UUID `json:"learning_path_id"`
	CertificateCode   string    `json:"certificate_code"`
	RecipientName     string    `json:"recipient_name"`
	LearningPathTitle string    `json:"learning_path_title"`
	IssuedAt          time.Time `json:"issued_at"`
}

type UserSummaryStats struct {
	Streak             StreakInfo `json:"streak"`
	CompletedLessons   int        `json:"completed_lessons"`
	SolvedChallenges   int        `json:"solved_challenges"`
	TotalNotes         int        `json:"total_notes"`
	TotalCertificates  int        `json:"total_certificates"`
	EarnedBadgesCount  int        `json:"earned_badges_count"`
	TotalBadgesCount   int        `json:"total_badges_count"`
}

type ClaimCertificateRequest struct {
	LearningPathID uuid.UUID `json:"learning_path_id"`
}
