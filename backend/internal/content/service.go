package content

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
	"github.com/jackc/pgx/v5/pgtype"

	"go-academy/backend/internal/database"
	"go-academy/backend/internal/database/sqlc"
)

var (
	ErrNotFound           = errors.New("resource not found")
	ErrSlugExists         = errors.New("slug already in use")
	ErrInvalidInput       = errors.New("invalid input")
	ErrPrerequisitesUnmet = errors.New("prerequisites not completed")
)

// Service provides learning content management and curriculum retrieval.
type Service struct {
	db *database.DB
}

// NewService constructs a new content Service.
func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

// GenerateSlug turns a title into a URL-friendly slug.
func GenerateSlug(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "item-" + uuid.New().String()[:8]
	}
	return s
}

// --- Learning Paths ---

// ListPublishedLearningPaths returns all published paths with learner progress stats.
func (s *Service) ListPublishedLearningPaths(ctx context.Context, userID *uuid.UUID) ([]LearningPathWithStats, error) {
	rows, err := s.db.Queries.ListPublishedLearningPaths(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list published learning paths: %w", err)
	}

	result := make([]LearningPathWithStats, 0, len(rows))
	for _, row := range rows {
		lp := fromSQLCLearningPath(row)
		statsItem := LearningPathWithStats{
			LearningPath: lp,
		}

		if userID != nil {
			stats, err := s.db.Queries.GetLearningPathProgressStats(ctx, sqlc.GetLearningPathProgressStatsParams{
				UserID:         toPgUUID(*userID),
				LearningPathID: row.ID,
			})
			if err == nil {
				statsItem.TotalLessons = int(stats.TotalLessons)
				statsItem.CompletedLessons = int(stats.CompletedLessons)
				if stats.TotalLessons > 0 {
					statsItem.ProgressPercent = int((float64(stats.CompletedLessons) / float64(stats.TotalLessons)) * 100)
				}
			}
		}

		result = append(result, statsItem)
	}

	return result, nil
}

// ListAllLearningPaths returns all paths for admin management.
func (s *Service) ListAllLearningPaths(ctx context.Context) ([]LearningPath, error) {
	rows, err := s.db.Queries.ListAllLearningPaths(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list all learning paths: %w", err)
	}

	result := make([]LearningPath, len(rows))
	for i, r := range rows {
		result[i] = fromSQLCLearningPath(r)
	}
	return result, nil
}

// GetLearningPathByID fetches a path by its UUID.
func (s *Service) GetLearningPathByID(ctx context.Context, id uuid.UUID) (*LearningPath, error) {
	row, err := s.db.Queries.GetLearningPathByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get learning path by id: %w", err)
	}

	lp := fromSQLCLearningPath(row)
	return &lp, nil
}

// GetLearningPathBySlug fetches a path by its slug along with its courses.
func (s *Service) GetLearningPathBySlug(ctx context.Context, slug string, userID *uuid.UUID) (*LearningPathWithStats, []CourseWithStats, error) {
	pathRow, err := s.db.Queries.GetLearningPathBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, fmt.Errorf("failed to get learning path: %w", err)
	}

	coursesRows, err := s.db.Queries.ListPublishedCoursesByPathID(ctx, pathRow.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list published courses: %w", err)
	}

	courses := make([]CourseWithStats, 0, len(coursesRows))
	for _, c := range coursesRows {
		course := fromSQLCCourse(c)
		cws := CourseWithStats{Course: course}
		if userID != nil {
			stats, err := s.db.Queries.GetCourseProgressStats(ctx, sqlc.GetCourseProgressStatsParams{
				UserID:   toPgUUID(*userID),
				CourseID: c.ID,
			})
			if err == nil {
				cws.TotalLessons = int(stats.TotalLessons)
				cws.CompletedLessons = int(stats.CompletedLessons)
				if stats.TotalLessons > 0 {
					cws.ProgressPercent = int((float64(stats.CompletedLessons) / float64(stats.TotalLessons)) * 100)
				}
			}
		}
		courses = append(courses, cws)
	}

	lp := fromSQLCLearningPath(pathRow)
	lpStats := LearningPathWithStats{LearningPath: lp}
	if userID != nil {
		stats, err := s.db.Queries.GetLearningPathProgressStats(ctx, sqlc.GetLearningPathProgressStatsParams{
			UserID:         toPgUUID(*userID),
			LearningPathID: pathRow.ID,
		})
		if err == nil {
			lpStats.TotalLessons = int(stats.TotalLessons)
			lpStats.CompletedLessons = int(stats.CompletedLessons)
			if stats.TotalLessons > 0 {
				lpStats.ProgressPercent = int((float64(stats.CompletedLessons) / float64(stats.TotalLessons)) * 100)
			}
		}
	}

	return &lpStats, courses, nil
}

// CreateLearningPath creates a new learning path.
func (s *Service) CreateLearningPath(ctx context.Context, req CreateLearningPathRequest, adminID uuid.UUID) (*LearningPath, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = GenerateSlug(title)
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

	created, err := s.db.Queries.CreateLearningPath(ctx, sqlc.CreateLearningPathParams{
		Title:       title,
		Slug:        slug,
		Description: toPgText(req.Description),
		Status:      status,
		SortOrder:   int32(req.SortOrder),
		CreatedBy:   toPgUUID(adminID),
		PublishedAt: toPgTimestamptz(pubAt),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrSlugExists
		}
		return nil, fmt.Errorf("failed to create learning path: %w", err)
	}

	res := fromSQLCLearningPath(created)
	return &res, nil
}

// UpdateLearningPath updates title, slug, description, and sort order.
func (s *Service) UpdateLearningPath(ctx context.Context, id uuid.UUID, req UpdateLearningPathRequest) (*LearningPath, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = GenerateSlug(title)
	}

	updated, err := s.db.Queries.UpdateLearningPath(ctx, sqlc.UpdateLearningPathParams{
		ID:          toPgUUID(id),
		Title:       title,
		Slug:        slug,
		Description: toPgText(req.Description),
		SortOrder:   int32(req.SortOrder),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrSlugExists
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to update learning path: %w", err)
	}

	res := fromSQLCLearningPath(updated)
	return &res, nil
}

// UpdateLearningPathStatus changes the status of a learning path.
func (s *Service) UpdateLearningPathStatus(ctx context.Context, id uuid.UUID, status string) (*LearningPath, error) {
	if status != StatusDraft && status != StatusPublished && status != StatusArchived {
		return nil, fmt.Errorf("%w: status must be draft, published, or archived", ErrInvalidInput)
	}

	var pubAt *time.Time
	if status == StatusPublished {
		now := time.Now()
		pubAt = &now
	}

	updated, err := s.db.Queries.UpdateLearningPathStatus(ctx, sqlc.UpdateLearningPathStatusParams{
		ID:          toPgUUID(id),
		Status:      status,
		PublishedAt: toPgTimestamptz(pubAt),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to update learning path status: %w", err)
	}

	res := fromSQLCLearningPath(updated)
	return &res, nil
}

// DeleteLearningPath deletes a learning path by ID.
func (s *Service) DeleteLearningPath(ctx context.Context, id uuid.UUID) error {
	err := s.db.Queries.DeleteLearningPath(ctx, toPgUUID(id))
	if err != nil {
		return fmt.Errorf("failed to delete learning path: %w", err)
	}
	return nil
}

// --- Courses ---

// ListCoursesByPath lists courses for a learning path.
func (s *Service) ListCoursesByPath(ctx context.Context, pathID uuid.UUID, onlyPublished bool) ([]Course, error) {
	var courses []sqlc.Course
	var err error

	if onlyPublished {
		courses, err = s.db.Queries.ListPublishedCoursesByPathID(ctx, toPgUUID(pathID))
	} else {
		courses, err = s.db.Queries.ListAllCoursesByPathID(ctx, toPgUUID(pathID))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list courses: %w", err)
	}

	res := make([]Course, len(courses))
	for i, c := range courses {
		res[i] = fromSQLCCourse(c)
	}
	return res, nil
}

// GetCourseByID fetches a course by its UUID.
func (s *Service) GetCourseByID(ctx context.Context, id uuid.UUID) (*Course, error) {
	row, err := s.db.Queries.GetCourseByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get course by id: %w", err)
	}

	c := fromSQLCCourse(row)
	return &c, nil
}

// GetCourseBySlug fetches full course syllabus with modules and lessons.
func (s *Service) GetCourseBySlug(ctx context.Context, slug string, userID *uuid.UUID) (*CourseDetail, error) {
	courseRow, err := s.db.Queries.GetCourseBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get course by slug: %w", err)
	}

	modulesRows, err := s.db.Queries.ListModulesByCourseID(ctx, courseRow.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list modules for course: %w", err)
	}

	// Map user completed lessons if user logged in
	completedLessons := make(map[uuid.UUID]bool)
	if userID != nil {
		progRows, err := s.db.Queries.ListUserProgressForCourse(ctx, sqlc.ListUserProgressForCourseParams{
			UserID:   toPgUUID(*userID),
			CourseID: courseRow.ID,
		})
		if err == nil {
			for _, p := range progRows {
				if p.Status == "completed" {
					completedLessons[uuid.UUID(p.LessonID.Bytes)] = true
				}
			}
		}
	}

	modulesWithLessons := make([]ModuleWithLessons, 0, len(modulesRows))
	for _, m := range modulesRows {
		mod := fromSQLCModule(m)
		// For syllabus, retrieve published lessons
		lessonsRows, err := s.db.Queries.ListPublishedLessonsByModuleID(ctx, m.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to list lessons for module %s: %w", mod.Title, err)
		}

		lessonsSummaries := make([]LessonSummary, 0, len(lessonsRows))
		for _, l := range lessonsRows {
			lessonID := uuid.UUID(l.ID.Bytes)
			isCompleted := completedLessons[lessonID]

			isLocked := false
			if userID != nil {
				uncompleted, err := s.db.Queries.GetUncompletedPrerequisitesForUser(ctx, sqlc.GetUncompletedPrerequisitesForUserParams{
					LessonID: l.ID,
					UserID:   toPgUUID(*userID),
				})
				if err == nil && len(uncompleted) > 0 {
					isLocked = true
				}
			}

			summaryText := ""
			if l.Summary.Valid {
				summaryText = l.Summary.String
			}

			lessonsSummaries = append(lessonsSummaries, LessonSummary{
				ID:               lessonID,
				Title:            l.Title,
				Slug:             l.Slug,
				Summary:          summaryText,
				Difficulty:       l.Difficulty,
				EstimatedMinutes: int(l.EstimatedMinutes),
				Status:           l.Status,
				SortOrder:        int(l.SortOrder),
				IsCompleted:      isCompleted,
				IsLocked:         isLocked,
			})
		}

		modulesWithLessons = append(modulesWithLessons, ModuleWithLessons{
			Module:  mod,
			Lessons: lessonsSummaries,
		})
	}

	c := fromSQLCCourse(courseRow)
	detail := &CourseDetail{
		Course:  c,
		Modules: modulesWithLessons,
	}

	if userID != nil {
		stats, err := s.db.Queries.GetCourseProgressStats(ctx, sqlc.GetCourseProgressStatsParams{
			UserID:   toPgUUID(*userID),
			CourseID: courseRow.ID,
		})
		if err == nil {
			detail.TotalLessons = int(stats.TotalLessons)
			detail.CompletedLessons = int(stats.CompletedLessons)
			if stats.TotalLessons > 0 {
				detail.ProgressPercent = int((float64(stats.CompletedLessons) / float64(stats.TotalLessons)) * 100)
			}
		}
	}

	return detail, nil
}

// CreateCourse creates a course under a learning path.
func (s *Service) CreateCourse(ctx context.Context, req CreateCourseRequest) (*Course, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = GenerateSlug(title)
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

	created, err := s.db.Queries.CreateCourse(ctx, sqlc.CreateCourseParams{
		LearningPathID: toPgUUID(req.LearningPathID),
		Title:          title,
		Slug:           slug,
		Description:    toPgText(req.Description),
		Status:         status,
		SortOrder:      int32(req.SortOrder),
		PublishedAt:    toPgTimestamptz(pubAt),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrSlugExists
		}
		return nil, fmt.Errorf("failed to create course: %w", err)
	}

	res := fromSQLCCourse(created)
	return &res, nil
}

// UpdateCourse updates course title, slug, description, sort order.
func (s *Service) UpdateCourse(ctx context.Context, id uuid.UUID, req UpdateCourseRequest) (*Course, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = GenerateSlug(title)
	}

	updated, err := s.db.Queries.UpdateCourse(ctx, sqlc.UpdateCourseParams{
		ID:          toPgUUID(id),
		Title:       title,
		Slug:        slug,
		Description: toPgText(req.Description),
		SortOrder:   int32(req.SortOrder),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrSlugExists
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to update course: %w", err)
	}

	res := fromSQLCCourse(updated)
	return &res, nil
}

// UpdateCourseStatus updates course status.
func (s *Service) UpdateCourseStatus(ctx context.Context, id uuid.UUID, status string) (*Course, error) {
	if status != StatusDraft && status != StatusPublished && status != StatusArchived {
		return nil, fmt.Errorf("%w: status must be draft, published, or archived", ErrInvalidInput)
	}

	var pubAt *time.Time
	if status == StatusPublished {
		now := time.Now()
		pubAt = &now
	}

	updated, err := s.db.Queries.UpdateCourseStatus(ctx, sqlc.UpdateCourseStatusParams{
		ID:          toPgUUID(id),
		Status:      status,
		PublishedAt: toPgTimestamptz(pubAt),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to update course status: %w", err)
	}

	res := fromSQLCCourse(updated)
	return &res, nil
}

// DeleteCourse deletes a course by ID.
func (s *Service) DeleteCourse(ctx context.Context, id uuid.UUID) error {
	err := s.db.Queries.DeleteCourse(ctx, toPgUUID(id))
	if err != nil {
		return fmt.Errorf("failed to delete course: %w", err)
	}
	return nil
}

// --- Modules ---

// ListModulesByCourse lists modules for a course.
func (s *Service) ListModulesByCourse(ctx context.Context, courseID uuid.UUID) ([]Module, error) {
	rows, err := s.db.Queries.ListModulesByCourseID(ctx, toPgUUID(courseID))
	if err != nil {
		return nil, fmt.Errorf("failed to list modules: %w", err)
	}

	res := make([]Module, len(rows))
	for i, m := range rows {
		res[i] = fromSQLCModule(m)
	}
	return res, nil
}

// GetModuleByID fetches a module by ID.
func (s *Service) GetModuleByID(ctx context.Context, id uuid.UUID) (*Module, error) {
	row, err := s.db.Queries.GetModuleByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get module by id: %w", err)
	}

	m := fromSQLCModule(row)
	return &m, nil
}

// CreateModule creates a new module.
func (s *Service) CreateModule(ctx context.Context, req CreateModuleRequest) (*Module, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = GenerateSlug(title)
	}

	created, err := s.db.Queries.CreateModule(ctx, sqlc.CreateModuleParams{
		CourseID:    toPgUUID(req.CourseID),
		Title:       title,
		Slug:        slug,
		Description: toPgText(req.Description),
		SortOrder:   int32(req.SortOrder),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrSlugExists
		}
		return nil, fmt.Errorf("failed to create module: %w", err)
	}

	res := fromSQLCModule(created)
	return &res, nil
}

// UpdateModule updates module title, slug, description, and sort order.
func (s *Service) UpdateModule(ctx context.Context, id uuid.UUID, req UpdateModuleRequest) (*Module, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = GenerateSlug(title)
	}

	updated, err := s.db.Queries.UpdateModule(ctx, sqlc.UpdateModuleParams{
		ID:          toPgUUID(id),
		Title:       title,
		Slug:        slug,
		Description: toPgText(req.Description),
		SortOrder:   int32(req.SortOrder),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrSlugExists
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to update module: %w", err)
	}

	res := fromSQLCModule(updated)
	return &res, nil
}

// DeleteModule deletes a module by ID.
func (s *Service) DeleteModule(ctx context.Context, id uuid.UUID) error {
	err := s.db.Queries.DeleteModule(ctx, toPgUUID(id))
	if err != nil {
		return fmt.Errorf("failed to delete module: %w", err)
	}
	return nil
}

// --- Lessons ---

// ListLessonsByModule lists lessons for a module.
func (s *Service) ListLessonsByModule(ctx context.Context, moduleID uuid.UUID, onlyPublished bool) ([]Lesson, error) {
	var rows []sqlc.Lesson
	var err error

	if onlyPublished {
		rows, err = s.db.Queries.ListPublishedLessonsByModuleID(ctx, toPgUUID(moduleID))
	} else {
		rows, err = s.db.Queries.ListAllLessonsByModuleID(ctx, toPgUUID(moduleID))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list lessons: %w", err)
	}

	res := make([]Lesson, len(rows))
	for i, l := range rows {
		res[i] = fromSQLCLesson(l)
	}
	return res, nil
}

// GetLessonByID fetches a lesson by its ID along with prerequisite IDs.
func (s *Service) GetLessonByID(ctx context.Context, id uuid.UUID) (*Lesson, []uuid.UUID, error) {
	row, err := s.db.Queries.GetLessonByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, fmt.Errorf("failed to get lesson by id: %w", err)
	}

	l := fromSQLCLesson(row)

	prereqRows, err := s.db.Queries.GetLessonPrerequisites(ctx, row.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get lesson prerequisites: %w", err)
	}

	prereqIDs := make([]uuid.UUID, len(prereqRows))
	for i, p := range prereqRows {
		prereqIDs[i] = uuid.UUID(p.ID.Bytes)
	}

	return &l, prereqIDs, nil
}

// GetLessonBySlug fetches lesson detail for reading, checking prerequisites and navigation links.
func (s *Service) GetLessonBySlug(ctx context.Context, slug string, userID *uuid.UUID) (*LessonDetail, error) {
	lessonRow, err := s.db.Queries.GetLessonBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get lesson by slug: %w", err)
	}

	// Fetch module
	moduleRow, err := s.db.Queries.GetModuleByID(ctx, lessonRow.ModuleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get module for lesson: %w", err)
	}

	// Fetch course
	courseRow, err := s.db.Queries.GetCourseByID(ctx, moduleRow.CourseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get course for lesson: %w", err)
	}

	// Fetch prerequisites
	prereqRows, err := s.db.Queries.GetLessonPrerequisites(ctx, lessonRow.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get prerequisites: %w", err)
	}

	prerequisites := make([]PrerequisiteInfo, len(prereqRows))
	for i, p := range prereqRows {
		prerequisites[i] = PrerequisiteInfo{
			ID:    uuid.UUID(p.ID.Bytes),
			Title: p.Title,
			Slug:  p.Slug,
		}
	}

	isLocked := false
	missingPrerequisites := []PrerequisiteInfo{}
	progressStatus := "not_started"
	isCompleted := false

	if userID != nil {
		// Check uncompleted prerequisites
		uncompleted, err := s.db.Queries.GetUncompletedPrerequisitesForUser(ctx, sqlc.GetUncompletedPrerequisitesForUserParams{
			LessonID: lessonRow.ID,
			UserID:   toPgUUID(*userID),
		})
		if err == nil && len(uncompleted) > 0 {
			isLocked = true
			for _, u := range uncompleted {
				missingPrerequisites = append(missingPrerequisites, PrerequisiteInfo{
					ID:    uuid.UUID(u.ID.Bytes),
					Title: u.Title,
					Slug:  u.Slug,
				})
			}
		}

		// Check current progress
		prog, err := s.db.Queries.GetLessonProgress(ctx, sqlc.GetLessonProgressParams{
			UserID:   toPgUUID(*userID),
			LessonID: lessonRow.ID,
		})
		if err == nil {
			progressStatus = prog.Status
			isCompleted = prog.Status == "completed"
		} else if !isLocked {
			// Auto mark as started if accessible
			_, _ = s.db.Queries.UpsertLessonProgress(ctx, sqlc.UpsertLessonProgressParams{
				UserID:   toPgUUID(*userID),
				LessonID: lessonRow.ID,
				Status:   "started",
			})
			progressStatus = "started"
		}
	}

	// Calculate Prev / Next lesson within course
	var prevSlug, nextSlug *string
	allCourseLessons, err := s.getOrderedCourseLessons(ctx, courseRow.ID)
	if err == nil {
		for i, item := range allCourseLessons {
			if item.ID == uuid.UUID(lessonRow.ID.Bytes) {
				if i > 0 {
					prev := allCourseLessons[i-1].Slug
					prevSlug = &prev
				}
				if i < len(allCourseLessons)-1 {
					next := allCourseLessons[i+1].Slug
					nextSlug = &next
				}
				break
			}
		}
	}

	lesson := fromSQLCLesson(lessonRow)
	return &LessonDetail{
		Lesson:               lesson,
		CourseTitle:          courseRow.Title,
		CourseSlug:           courseRow.Slug,
		ModuleTitle:          moduleRow.Title,
		ModuleSlug:           moduleRow.Slug,
		ProgressStatus:       progressStatus,
		IsCompleted:          isCompleted,
		IsLocked:             isLocked,
		MissingPrerequisites: missingPrerequisites,
		Prerequisites:        prerequisites,
		PrevLessonSlug:       prevSlug,
		NextLessonSlug:       nextSlug,
	}, nil
}

type courseLessonPointer struct {
	ID   uuid.UUID
	Slug string
}

func (s *Service) getOrderedCourseLessons(ctx context.Context, courseID pgtype.UUID) ([]courseLessonPointer, error) {
	mods, err := s.db.Queries.ListModulesByCourseID(ctx, courseID)
	if err != nil {
		return nil, err
	}

	var pointers []courseLessonPointer
	for _, m := range mods {
		lessons, err := s.db.Queries.ListPublishedLessonsByModuleID(ctx, m.ID)
		if err != nil {
			continue
		}
		for _, l := range lessons {
			pointers = append(pointers, courseLessonPointer{
				ID:   uuid.UUID(l.ID.Bytes),
				Slug: l.Slug,
			})
		}
	}
	return pointers, nil
}

// CreateLesson creates a new lesson with optional prerequisites.
func (s *Service) CreateLesson(ctx context.Context, req CreateLessonRequest) (*Lesson, error) {
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
		estMin = 10
	}

	status := req.Status
	if status == "" {
		status = StatusDraft
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.db.Queries.WithTx(tx)

	var pubAt *time.Time
	if status == StatusPublished {
		now := time.Now()
		pubAt = &now
	}

	created, err := qtx.CreateLesson(ctx, sqlc.CreateLessonParams{
		ModuleID:         toPgUUID(req.ModuleID),
		Title:            title,
		Slug:             slug,
		Summary:          toPgText(req.Summary),
		Content:          req.Content,
		Difficulty:       difficulty,
		EstimatedMinutes: int32(estMin),
		Status:           status,
		SortOrder:        int32(req.SortOrder),
		PublishedAt:      toPgTimestamptz(pubAt),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrSlugExists
		}
		return nil, fmt.Errorf("failed to create lesson: %w", err)
	}

	// Add prerequisites
	for _, prereqID := range req.PrerequisiteIDs {
		// Prevent self-prerequisite
		if prereqID == uuid.UUID(created.ID.Bytes) {
			continue
		}
		err = qtx.AddLessonPrerequisite(ctx, sqlc.AddLessonPrerequisiteParams{
			LessonID:             created.ID,
			PrerequisiteLessonID: toPgUUID(prereqID),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to add prerequisite: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	res := fromSQLCLesson(created)
	return &res, nil
}

// UpdateLesson updates a lesson and re-links prerequisites.
func (s *Service) UpdateLesson(ctx context.Context, id uuid.UUID, req UpdateLessonRequest) (*Lesson, error) {
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
		estMin = 10
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.db.Queries.WithTx(tx)

	updated, err := qtx.UpdateLesson(ctx, sqlc.UpdateLessonParams{
		ID:               toPgUUID(id),
		Title:            title,
		Slug:             slug,
		Summary:          toPgText(req.Summary),
		Content:          req.Content,
		Difficulty:       difficulty,
		EstimatedMinutes: int32(estMin),
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
		return nil, fmt.Errorf("failed to update lesson: %w", err)
	}

	// Clear and re-insert prerequisites
	if err := qtx.ClearLessonPrerequisites(ctx, toPgUUID(id)); err != nil {
		return nil, fmt.Errorf("failed to clear prerequisites: %w", err)
	}

	for _, prereqID := range req.PrerequisiteIDs {
		if prereqID == id {
			continue // Prevent self-prerequisite
		}
		err = qtx.AddLessonPrerequisite(ctx, sqlc.AddLessonPrerequisiteParams{
			LessonID:             toPgUUID(id),
			PrerequisiteLessonID: toPgUUID(prereqID),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to re-add prerequisite: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	res := fromSQLCLesson(updated)
	return &res, nil
}

// UpdateLessonStatus updates lesson publication status.
func (s *Service) UpdateLessonStatus(ctx context.Context, id uuid.UUID, status string) (*Lesson, error) {
	if status != StatusDraft && status != StatusPublished && status != StatusArchived {
		return nil, fmt.Errorf("%w: status must be draft, published, or archived", ErrInvalidInput)
	}

	var pubAt *time.Time
	if status == StatusPublished {
		now := time.Now()
		pubAt = &now
	}

	updated, err := s.db.Queries.UpdateLessonStatus(ctx, sqlc.UpdateLessonStatusParams{
		ID:          toPgUUID(id),
		Status:      status,
		PublishedAt: toPgTimestamptz(pubAt),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to update lesson status: %w", err)
	}

	res := fromSQLCLesson(updated)
	return &res, nil
}

// DeleteLesson deletes a lesson.
func (s *Service) DeleteLesson(ctx context.Context, id uuid.UUID) error {
	err := s.db.Queries.DeleteLesson(ctx, toPgUUID(id))
	if err != nil {
		return fmt.Errorf("failed to delete lesson: %w", err)
	}
	return nil
}
