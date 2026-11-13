package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"go-academy/backend/internal/config"
	"go-academy/backend/internal/content"
	"go-academy/backend/internal/database"
	"go-academy/backend/internal/users"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	db := database.NewDB(pool)
	contentService := content.NewService(db)
	userService := users.NewService(db)

	// Fetch or use existing admin
	admin, err := userService.GetUserByEmail(ctx, "test_admin@example.com")
	var adminID uuid.UUID
	if err != nil {
		log.Printf("Admin not found, searching first available user...")
		usersList, _, listErr := userService.ListUsers(ctx, 1, 10)
		if listErr != nil || len(usersList) == 0 {
			log.Fatalf("no admin user available to seed content")
		}
		adminID = usersList[0].ID
	} else {
		adminID = admin.ID
	}

	log.Printf("Using author ID %s to seed curriculum...", adminID)

	// 1. Learning Path: Go Fundamentals & Idiomatic Syntax
	path1, err := contentService.CreateLearningPath(ctx, content.CreateLearningPathRequest{
		Title:       "Go Fundamentals & Idiomatic Syntax",
		Slug:        "go-fundamentals",
		Description: "Build an unshakeable foundation in the Go programming language. Covers syntax, typing, memory semantics, and core language idioms.",
		Status:      content.StatusPublished,
		SortOrder:   1,
	}, adminID)
	if err != nil && err != content.ErrSlugExists {
		log.Fatalf("failed to create path 1: %v", err)
	}
	if err == content.ErrSlugExists {
		log.Println("Curriculum already seeded. Skipping.")
		return
	}

	log.Printf("Created path: %s", path1.Title)

	// Course 1: Getting Started with Go
	course1, err := contentService.CreateCourse(ctx, content.CreateCourseRequest{
		LearningPathID: path1.ID,
		Title:          "Getting Started with Go",
		Slug:           "getting-started-with-go",
		Description:    "Your first deep dive into Go semantics, toolchain, packages, and primitive types.",
		Status:         content.StatusPublished,
		SortOrder:      1,
	})
	if err != nil {
		log.Fatalf("failed to create course 1: %v", err)
	}

	// Module 1: Language Basics
	mod1, err := contentService.CreateModule(ctx, content.CreateModuleRequest{
		CourseID:    course1.ID,
		Title:       "Language Basics",
		Slug:        "language-basics",
		Description: "Syntax, variables, types, and fundamental control flow.",
		SortOrder:   1,
	})
	if err != nil {
		log.Fatalf("failed to create module 1: %v", err)
	}

	// Lesson 1: Variables, Types, and Zero Values
	lesson1, err := contentService.CreateLesson(ctx, content.CreateLessonRequest{
		ModuleID: mod1.ID,
		Title:    "Variables, Types, and Zero Values",
		Slug:     "variables-types-zero-values",
		Summary:  "Explore Go's strong static type system and automatic zero-value initialization.",
		Content: `# Variables, Types, and Zero Values

In Go, variables are explicitly typed and strictly enforced at compile time. 

## Declaring Variables
There are two common idioms for declaring variables in Go:

### 1. Using ` + "`var`" + `
` + "```go" + `
package main

import "fmt"

func main() {
    var count int = 10
    var name string = "Go Academy"
    fmt.Printf("%s has %d lessons\n", name, count)
}
` + "```" + `

### 2. Short Variable Declaration (` + "`:=`" + `)
Inside functions, the short declaration syntax ` + "`:=`" + ` infers types automatically:

` + "```go" + `
message := "Hello from Gopher"
active := true
` + "```" + `

> [!NOTE]
> Variables in Go are never uninitialized. When declared without an initial value, they always receive their **Zero Value** (` + "`0`" + ` for numbers, ` + "`\"\"`" + ` for strings, ` + "`false`" + ` for booleans, and ` + "`nil`" + ` for pointers/slices/maps).
`,
		Difficulty:       content.DifficultyBeginner,
		EstimatedMinutes: 10,
		Status:           content.StatusPublished,
		SortOrder:        1,
	})
	if err != nil {
		log.Fatalf("failed to create lesson 1: %v", err)
	}

	// Lesson 2: Control Flow (Prerequisite: Lesson 1)
	lesson2, err := contentService.CreateLesson(ctx, content.CreateLessonRequest{
		ModuleID: mod1.ID,
		Title:    "Control Flow: Branches and Loops",
		Slug:     "control-flow-branches-loops",
		Summary:  "Master if/else, switch statements, and Go's single unified loop keyword: for.",
		Content: `# Control Flow: Branches and Loops

Go prioritizes simplicity and readability. Unlike languages with ` + "`while`" + ` or ` + "`do-while`" + `, Go has only one looping keyword: **` + "`for`" + `**.

## If-Else with Pre-Statements
Go supports declaring local scoped variables directly inside ` + "`if`" + ` conditions:

` + "```go" + `
if err := runTask(); err != nil {
    log.Fatalf("Task failed: %v", err)
}
` + "```" + `

## Looping with ` + "`for`" + `
Standard three-component loop:
` + "```go" + `
for i := 0; i < 5; i++ {
    fmt.Println(i)
}
` + "```" + `

While-style loop:
` + "```go" + `
n := 1
for n < 100 {
    n *= 2
}
` + "```" + `

Infinite loop:
` + "```go" + `
for {
    // runs until break or return
    break
}
` + "```" + `
`,
		Difficulty:       content.DifficultyBeginner,
		EstimatedMinutes: 12,
		Status:           content.StatusPublished,
		SortOrder:        2,
		PrerequisiteIDs:  []uuid.UUID{lesson1.ID},
	})
	if err != nil {
		log.Fatalf("failed to create lesson 2: %v", err)
	}

	// Lesson 3: Functions, Defer, and Error Handling (Prerequisite: Lesson 2)
	_, err = contentService.CreateLesson(ctx, content.CreateLessonRequest{
		ModuleID: mod1.ID,
		Title:    "Functions, Defer, and Multi-Value Returns",
		Slug:     "functions-defer-multi-value-returns",
		Summary:  "Understand first-class functions, deferred resource cleanup, and explicit error values.",
		Content: `# Functions, Defer, and Multi-Value Returns

Functions in Go are first-class citizens. They can return multiple values, commonly used for the ` + "`(result, error)`" + ` pattern.

## Multi-Value Returns & Error Handling
` + "```go" + `
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("division by zero")
    }
    return a / b, nil
}
` + "```" + `

## The ` + "`defer`" + ` Statement
A ` + "`defer`" + ` statement defers the execution of a function until the surrounding function returns. It is commonly used for resource cleanup (closing files, releasing locks):

` + "```go" + `
file, err := os.Open("data.txt")
if err != nil {
    return err
}
defer file.Close() // Guaranteed to run on exit
` + "```" + `
`,
		Difficulty:       content.DifficultyBeginner,
		EstimatedMinutes: 15,
		Status:           content.StatusPublished,
		SortOrder:        3,
		PrerequisiteIDs:  []uuid.UUID{lesson2.ID},
	})
	if err != nil {
		log.Fatalf("failed to create lesson 3: %v", err)
	}

	// 2. Learning Path: Concurrency & Systems Programming
	path2, err := contentService.CreateLearningPath(ctx, content.CreateLearningPathRequest{
		Title:       "High-Performance Concurrency in Go",
		Slug:        "high-performance-concurrency",
		Description: "Master CSP-style concurrency with Goroutines, buffered/unbuffered channels, select multiplexing, sync primitives, and context cancellation.",
		Status:      content.StatusPublished,
		SortOrder:   2,
	}, adminID)
	if err != nil {
		log.Fatalf("failed to create path 2: %v", err)
	}

	course2, err := contentService.CreateCourse(ctx, content.CreateCourseRequest{
		LearningPathID: path2.ID,
		Title:          "Goroutines and Channel Multiplexing",
		Slug:           "goroutines-channels-multiplexing",
		Description:    "Harness the power of Go's lightweight runtime scheduler and non-blocking channel communication.",
		Status:         content.StatusPublished,
		SortOrder:      1,
	})
	if err != nil {
		log.Fatalf("failed to create course 2: %v", err)
	}

	mod2, err := contentService.CreateModule(ctx, content.CreateModuleRequest{
		CourseID:    course2.ID,
		Title:       "Core Concurrency Primitives",
		Slug:        "core-concurrency-primitives",
		Description: "Goroutines, channels, and select.",
		SortOrder:   1,
	})
	if err != nil {
		log.Fatalf("failed to create module 2: %v", err)
	}

	_, err = contentService.CreateLesson(ctx, content.CreateLessonRequest{
		ModuleID: mod2.ID,
		Title:    "Goroutines: Lightweight Threading",
		Slug:     "goroutines-lightweight-threading",
		Summary:  "How Go multiplexes thousands of Goroutines onto OS threads with a tiny 2KB initial stack.",
		Content: `# Goroutines: Lightweight Threading

A Goroutine is an independently executing function managed by the Go runtime scheduler (the M:N scheduler).

` + "```go" + `
package main

import (
    "fmt"
    "time"
)

func say(s string) {
    for i := 0; i < 3; i++ {
        time.Sleep(100 * time.Millisecond)
        fmt.Println(s)
    }
}

func main() {
    go say("world") // launched asynchronously
    say("hello")
}
` + "```" + `

Unlike operating system threads (which usually have fixed 1-2MB stacks), Goroutines start with a dynamic **2KB** stack that grows and shrinks as needed.
`,
		Difficulty:       content.DifficultyIntermediate,
		EstimatedMinutes: 15,
		Status:           content.StatusPublished,
		SortOrder:        1,
	})
	if err != nil {
		log.Fatalf("failed to create concurrency lesson: %v", err)
	}

	fmt.Println("Curriculum seeding completed successfully!")
}
