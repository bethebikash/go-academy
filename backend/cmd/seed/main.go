package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"go-academy/backend/internal/challenges"
	"go-academy/backend/internal/config"
	"go-academy/backend/internal/content"
	"go-academy/backend/internal/database"
	"go-academy/backend/internal/runner"
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
	codeRunner := runner.NewProcessRunner(5 * time.Second)
	challengeService := challenges.NewService(db, codeRunner)
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

	log.Printf("Using author ID %s to seed curriculum & challenges...", adminID)

	// Seed Curriculum if not already present
	seedCurriculum(ctx, contentService, adminID)

	// Seed Coding Challenges
	seedChallenges(ctx, challengeService)

	fmt.Println("All seeding completed successfully!")
}

func seedCurriculum(ctx context.Context, contentService *content.Service, adminID uuid.UUID) {
	// 1. Learning Path: Go Fundamentals & Idiomatic Syntax
	path1, err := contentService.CreateLearningPath(ctx, content.CreateLearningPathRequest{
		Title:       "Go Fundamentals & Idiomatic Syntax",
		Slug:        "go-fundamentals",
		Description: "Build an unshakeable foundation in the Go programming language. Covers syntax, typing, memory semantics, and core language idioms.",
		Status:      content.StatusPublished,
		SortOrder:   1,
	}, adminID)
	if err != nil {
		if errors.Is(err, content.ErrSlugExists) {
			log.Println("Curriculum 'go-fundamentals' already seeded. Skipping curriculum.")
			return
		}
		log.Fatalf("failed to create path 1: %v", err)
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

In Go, every declared variable is allocated memory and initialized to its canonical **zero value** if no explicit initialization is provided.

` + "```go" + `
package main

import "fmt"

func main() {
    var count int       // 0
    var active bool     // false
    var message string  // ""
    var pointer *int    // nil

    fmt.Printf("count=%d, active=%t, message=%q, pointer=%v\n", count, active, message, pointer)
}
` + "```" + `

## Short Variable Declaration
Inside function bodies, the ` + "`:=`" + ` operator offers concise type inference:

` + "```go" + `
language := "Go" // inferred as string
score := 100     // inferred as int
` + "```" + `

> [!NOTE]
> Constants in Go are untyped by default until assigned to a variable or used in an expression that requires a typed value.
`,
		Difficulty:       content.DifficultyBeginner,
		EstimatedMinutes: 10,
		Status:           content.StatusPublished,
		SortOrder:        1,
	})
	if err != nil {
		log.Fatalf("failed to create lesson 1: %v", err)
	}

	// Lesson 2: Pointers and Memory Semantics
	_, err = contentService.CreateLesson(ctx, content.CreateLessonRequest{
		ModuleID: mod1.ID,
		Title:    "Pointers and Memory Semantics",
		Slug:     "pointers-memory-semantics",
		Summary:  "Understand pass-by-value vs pointers, stack vs heap allocation, and escape analysis.",
		Content: `# Pointers and Memory Semantics

Every assignment and function argument pass in Go is **strictly pass-by-value**. Pointers allow sharing memory across stack frames.

` + "```go" + `
package main

import "fmt"

func increment(val *int) {
    *val++
}

func main() {
    x := 42
    increment(&x)
    fmt.Println(x) // prints 43
}
` + "```" + `

## Escape Analysis
The Go compiler performs **escape analysis** during compilation. If a variable's address is referenced outside its originating stack frame, the compiler moves the allocation to the **heap**. Otherwise, it stays on the high-performance **stack**.
`,
		Difficulty:       content.DifficultyBeginner,
		EstimatedMinutes: 15,
		Status:           content.StatusPublished,
		SortOrder:        2,
		PrerequisiteIDs:  []uuid.UUID{lesson1.ID},
	})
	if err != nil {
		log.Fatalf("failed to create lesson 2: %v", err)
	}

	// 2. Learning Path: Concurrency & Systems Architecture
	path2, err := contentService.CreateLearningPath(ctx, content.CreateLearningPathRequest{
		Title:       "Concurrency & Systems Architecture",
		Slug:        "concurrency-systems-architecture",
		Description: "Master CSP (Communicating Sequential Processes) concurrency patterns, goroutines, channels, sync primitives, and race condition prevention.",
		Status:      content.StatusPublished,
		SortOrder:   2,
	}, adminID)
	if err != nil {
		log.Fatalf("failed to create path 2: %v", err)
	}

	course2, err := contentService.CreateCourse(ctx, content.CreateCourseRequest{
		LearningPathID: path2.ID,
		Title:          "Goroutines, Channels & Multiplexing",
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
}

func seedChallenges(ctx context.Context, challengeService *challenges.Service) {
	challengeSeeds := []challenges.CreateChallengeRequest{
		{
			Title:            "Two Sum in Go",
			Slug:             "two-sum-in-go",
			Description:      "Given an array of integers `nums` and an integer `target`, return indices of the two numbers such that they add up to `target`.\n\nYou may assume that each input has exactly one solution, and you may not use the same element twice. Return the indices as `[]int{i, j}` with `i < j`.\n\n### Complexity Goal\n- Time Complexity: **O(n)** using a hash map\n- Space Complexity: **O(n)**",
			Difficulty:       challenges.DifficultyBeginner,
			EstimatedMinutes: 15,
			StarterCode: `package main

// TwoSum returns the indices of the two numbers in nums that add up to target.
func TwoSum(nums []int, target int) []int {
	// TODO: Implement your O(n) hash map solution here
	return nil
}
`,
			SolutionCode: `package main

func TwoSum(nums []int, target int) []int {
	seen := make(map[int]int, len(nums))
	for i, num := range nums {
		diff := target - num
		if j, ok := seen[diff]; ok {
			return []int{j, i}
		}
		seen[num] = i
	}
	return nil
}
`,
			Status:    challenges.StatusPublished,
			SortOrder: 1,
			TestCases: []challenges.CreateTestCaseRequest{
				{
					Title:          "Sample 1: nums = [2, 7, 11, 15], target = 9",
					Input:          "nums = [2, 7, 11, 15], target = 9",
					ExpectedOutput: "[0, 1]",
					Hidden:         false,
					SortOrder:      1,
				},
				{
					Title:          "Sample 2: nums = [3, 2, 4], target = 6",
					Input:          "nums = [3, 2, 4], target = 6",
					ExpectedOutput: "[1, 2]",
					Hidden:         false,
					SortOrder:      2,
				},
				{
					Title:          "Hidden: Duplicate Elements [3, 3], target = 6",
					Input:          "nums = [3, 3], target = 6",
					ExpectedOutput: "[0, 1]",
					Hidden:         true,
					SortOrder:      3,
				},
				{
					Title:          "Hidden: Negative Numbers [-1, -2, -3, -4, -5], target = -8",
					Input:          "nums = [-1, -2, -3, -4, -5], target = -8",
					ExpectedOutput: "[2, 4]",
					Hidden:         true,
					SortOrder:      4,
				},
			},
		},
		{
			Title:            "Reverse a String with Runes",
			Slug:             "reverse-string-runes",
			Description:      "In Go, strings are immutable byte sequences encoded in UTF-8. A single character (like an accent or emoji) might span multiple bytes.\n\nImplement `ReverseString(s string) string` that correctly reverses a UTF-8 string by converting it to a rune slice.\n\n### Example\n- Input: `\"Go Academy 🚀\"`\n- Output: `\"🚀 ymedacA oG\"`",
			Difficulty:       challenges.DifficultyBeginner,
			EstimatedMinutes: 10,
			StarterCode: `package main

// ReverseString reverses a UTF-8 string rune-by-rune.
func ReverseString(s string) string {
	// TODO: convert string to []rune and reverse
	return ""
}
`,
			SolutionCode: `package main

func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
`,
			Status:    challenges.StatusPublished,
			SortOrder: 2,
			TestCases: []challenges.CreateTestCaseRequest{
				{
					Title:          "Sample 1: ASCII String \"hello\"",
					Input:          "\"hello\"",
					ExpectedOutput: "\"olleh\"",
					Hidden:         false,
					SortOrder:      1,
				},
				{
					Title:          "Sample 2: Unicode Emoji \"Go Academy 🚀\"",
					Input:          "\"Go Academy 🚀\"",
					ExpectedOutput: "\"🚀 ymedacA oG\"",
					Hidden:         false,
					SortOrder:      2,
				},
				{
					Title:          "Hidden: Multibyte Japanese characters \"日本語\"",
					Input:          "\"日本語\"",
					ExpectedOutput: "\"語本日\"",
					Hidden:         true,
					SortOrder:      3,
				},
				{
					Title:          "Hidden: Empty String",
					Input:          "\"\"",
					ExpectedOutput: "\"\"",
					Hidden:         true,
					SortOrder:      4,
				},
			},
		},
		{
			Title:            "Concurrent Worker Pool",
			Slug:             "concurrent-worker-pool",
			Description:      "Implement a concurrent worker pool that processes a slice of integer jobs across `numWorkers` goroutines.\n\nEach worker squares its input number (`v * v`). Results must be returned in the original ascending order of input indices.\n\nEnsure proper channel closing and waitgroup synchronization so no goroutines leak.",
			Difficulty:       challenges.DifficultyIntermediate,
			EstimatedMinutes: 25,
			StarterCode: `package main

// ProcessJobs coordinates numWorkers goroutines to square integers concurrently.
func ProcessJobs(jobs []int, numWorkers int) []int {
	// TODO: implement worker pool with channels and sync.WaitGroup
	return nil
}
`,
			SolutionCode: `package main

import (
	"sync"
)

type jobItem struct {
	index int
	val   int
}

type resultItem struct {
	index int
	val   int
}

func ProcessJobs(jobs []int, numWorkers int) []int {
	if len(jobs) == 0 {
		return []int{}
	}
	if numWorkers <= 0 {
		numWorkers = 1
	}

	jobCh := make(chan jobItem, len(jobs))
	resCh := make(chan resultItem, len(jobs))

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobCh {
				resCh <- resultItem{index: j.index, val: j.val * j.val}
			}
		}()
	}

	for i, v := range jobs {
		jobCh <- jobItem{index: i, val: v}
	}
	close(jobCh)

	wg.Wait()
	close(resCh)

	results := make([]int, len(jobs))
	for res := range resCh {
		results[res.index] = res.val
	}
	return results
}
`,
			Status:    challenges.StatusPublished,
			SortOrder: 3,
			TestCases: []challenges.CreateTestCaseRequest{
				{
					Title:          "Sample 1: jobs = [1, 2, 3, 4, 5], workers = 3",
					Input:          "jobs = [1, 2, 3, 4, 5], numWorkers = 3",
					ExpectedOutput: "[1, 4, 9, 16, 25]",
					Hidden:         false,
					SortOrder:      1,
				},
				{
					Title:          "Hidden: Empty Jobs Slice",
					Input:          "jobs = [], numWorkers = 2",
					ExpectedOutput: "[]",
					Hidden:         true,
					SortOrder:      2,
				},
				{
					Title:          "Hidden: Large Batch with High Concurrency",
					Input:          "jobs = [10, 20, 30, 40, 50], numWorkers = 5",
					ExpectedOutput: "[100, 400, 900, 1600, 2500]",
					Hidden:         true,
					SortOrder:      3,
				},
			},
		},
		{
			Title:            "Thread-Safe LRU Cache",
			Slug:             "thread-safe-lru-cache",
			Description:      "Design and implement a concurrent Least Recently Used (LRU) Cache supporting `Get(key int) (int, bool)` and `Put(key, value int)` in $O(1)$ time complexity.\n\nUse a doubly linked list and a hash map protected by a synchronization lock (`sync.RWMutex`).",
			Difficulty:       challenges.DifficultyAdvanced,
			EstimatedMinutes: 35,
			StarterCode: `package main

// LRUCache represents a thread-safe LRU cache.
type LRUCache struct {
	// TODO: Define fields
}

func Constructor(capacity int) LRUCache {
	// TODO: Initialize cache
	return LRUCache{}
}

func (c *LRUCache) Get(key int) (int, bool) {
	// TODO: Implement O(1) Get
	return 0, false
}

func (c *LRUCache) Put(key int, value int) {
	// TODO: Implement O(1) Put with eviction
}
`,
			SolutionCode: `package main

import (
	"container/list"
	"sync"
)

type cacheEntry struct {
	key   int
	value int
}

type LRUCache struct {
	mu       sync.RWMutex
	capacity int
	items    map[int]*list.Element
	evict    *list.List
}

func Constructor(capacity int) LRUCache {
	return LRUCache{
		capacity: capacity,
		items:    make(map[int]*list.Element, capacity),
		evict:    list.New(),
	}
}

func (c *LRUCache) Get(key int) (int, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.evict.MoveToFront(elem)
		return elem.Value.(*cacheEntry).value, true
	}
	return 0, false
}

func (c *LRUCache) Put(key int, value int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.evict.MoveToFront(elem)
		elem.Value.(*cacheEntry).value = value
		return
	}

	if c.evict.Len() >= c.capacity {
		oldest := c.evict.Back()
		if oldest != nil {
			c.evict.Remove(oldest)
			delete(c.items, oldest.Value.(*cacheEntry).key)
		}
	}

	elem := c.evict.PushFront(&cacheEntry{key: key, value: value})
	c.items[key] = elem
}
`,
			Status:    challenges.StatusPublished,
			SortOrder: 4,
			TestCases: []challenges.CreateTestCaseRequest{
				{
					Title:          "Sample 1: Basic Get and Put sequence",
					Input:          "cap = 2, put(1, 1), put(2, 2), get(1)",
					ExpectedOutput: "1, true",
					Hidden:         false,
					SortOrder:      1,
				},
				{
					Title:          "Hidden: Eviction of Least Recently Used",
					Input:          "cap = 2, put(1, 1), put(2, 2), get(1), put(3, 3), get(2)",
					ExpectedOutput: "0, false",
					Hidden:         true,
					SortOrder:      2,
				},
			},
		},
	}

	for _, ch := range challengeSeeds {
		created, err := challengeService.CreateChallenge(ctx, ch)
		if err != nil {
			if errors.Is(err, challenges.ErrSlugExists) {
				log.Printf("Challenge %q already exists. Skipping.", ch.Title)
				continue
			}
			log.Fatalf("failed to seed challenge %q: %v", ch.Title, err)
		}
		log.Printf("Seeded challenge: %s (ID: %s, %d test cases)", created.Title, created.ID, len(created.TestCases))
	}
}
