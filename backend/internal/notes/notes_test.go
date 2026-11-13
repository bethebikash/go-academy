package notes_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"go-academy/backend/internal/config"
	"go-academy/backend/internal/database"
	"go-academy/backend/internal/database/sqlc"
	"go-academy/backend/internal/notes"
)

func setupTestDB(t *testing.T) (*database.DB, func()) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		t.Skipf("skipping live database test: %v", err)
		return nil, nil
	}

	db := database.NewDB(pool)
	cleanup := func() {
		pool.Close()
	}

	return db, cleanup
}

func TestNotesLifecycleAndOptimisticLocking(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	noteService := notes.NewService(db)

	suffix := time.Now().UnixNano()
	userRow, err := db.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        fmt.Sprintf("notes_user_%d@example.com", suffix),
		Username:     fmt.Sprintf("notes_usr_%d", suffix),
		PasswordHash: "fakehash",
		Status:       "active",
		Timezone:     "UTC",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	userID := uuid.UUID(userRow.ID.Bytes)

	// 1. Create a note with tags
	note1, err := noteService.CreateNote(ctx, userID, notes.CreateNoteRequest{
		Title:   "Buffered Channels Behavior",
		Content: "Writing to a full buffered channel blocks until a receiver reads from it.",
		Pinned:  true,
		Tags:    []string{"concurrency", "channels", "go-idioms"},
	})
	if err != nil {
		t.Fatalf("failed to create note: %v", err)
	}
	if note1.Version != 1 || !note1.Pinned || len(note1.Tags) != 3 {
		t.Fatalf("unexpected note state: %+v", note1)
	}

	// 2. Create a second note
	note2, err := noteService.CreateNote(ctx, userID, notes.CreateNoteRequest{
		Title:   "Slices Memory Representation",
		Content: "A slice header contains pointer to backing array, length, and capacity.",
		Pinned:  false,
		Tags:    []string{"memory", "slices"},
	})
	if err != nil {
		t.Fatalf("failed to create note 2: %v", err)
	}

	// 3. List all notes (pinned should come first)
	allNotes, err := noteService.ListNotes(ctx, userID, "", nil, "")
	if err != nil {
		t.Fatalf("failed to list notes: %v", err)
	}
	if len(allNotes) != 2 || allNotes[0].ID != note1.ID {
		t.Fatalf("expected note1 (pinned) first, got %+v", allNotes)
	}

	// 4. Filter by tag
	concurrencyNotes, err := noteService.ListNotes(ctx, userID, "concurrency", nil, "")
	if err != nil {
		t.Fatalf("failed to list by tag: %v", err)
	}
	if len(concurrencyNotes) != 1 || concurrencyNotes[0].ID != note1.ID {
		t.Fatalf("expected 1 note for tag concurrency, got %d", len(concurrencyNotes))
	}

	// 5. Full-text search
	searchRes, err := noteService.ListNotes(ctx, userID, "", nil, "capacity")
	if err != nil {
		t.Fatalf("failed to search notes: %v", err)
	}
	if len(searchRes) != 1 || searchRes[0].ID != note2.ID {
		t.Fatalf("expected note2 for search 'capacity', got %d", len(searchRes))
	}

	// 6. Optimistic locking: valid update with version 1
	updatedNote, err := noteService.UpdateNote(ctx, userID, note1.ID, notes.UpdateNoteRequest{
		Title:   "Buffered Channels: Detailed Behavior",
		Content: "Updated content with examples.",
		Pinned:  true,
		Version: 1, // matches current version
		Tags:    []string{"concurrency", "channels"},
	})
	if err != nil {
		t.Fatalf("failed to update note with valid version: %v", err)
	}
	if updatedNote.Version != 2 || updatedNote.Title != "Buffered Channels: Detailed Behavior" {
		t.Fatalf("expected version 2, got %d", updatedNote.Version)
	}

	// 7. Optimistic locking: update with stale version 1 should fail with ErrVersionConflict
	_, err = noteService.UpdateNote(ctx, userID, note1.ID, notes.UpdateNoteRequest{
		Title:   "Stale Update Attempt",
		Content: "Should conflict",
		Version: 1, // stale! current version is 2
	})
	if err != notes.ErrVersionConflict {
		t.Fatalf("expected ErrVersionConflict, got %v", err)
	}

	// 8. List tags with note counts
	tags, err := noteService.ListTags(ctx, userID)
	if err != nil {
		t.Fatalf("failed to list tags: %v", err)
	}
	if len(tags) < 3 {
		t.Fatalf("expected at least 3 tags, got %d", len(tags))
	}

	// 9. Delete note
	err = noteService.DeleteNote(ctx, userID, note1.ID)
	if err != nil {
		t.Fatalf("failed to delete note: %v", err)
	}

	// Verify 404 after deletion
	_, err = noteService.GetNoteByID(ctx, userID, note1.ID)
	if err != notes.ErrNotFound {
		t.Fatalf("expected ErrNotFound after deletion, got %v", err)
	}
}
