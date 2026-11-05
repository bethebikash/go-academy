package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"

	"go-academy/backend/internal/config"
	"go-academy/backend/internal/database"
	"go-academy/backend/internal/users"
)

func printUsage() {
	fmt.Println(`Go Academy Admin CLI

Usage:
  admin bootstrap [flags]   Bootstrap the initial administrator account

Flags:
  --email string      Admin email address
  --username string   Admin username
  --password string   Admin password (prompts securely if omitted)

Examples:
  go run cmd/admin/main.go bootstrap --email="admin@example.com" --username="admin"
  go run cmd/admin/main.go bootstrap --email="admin@example.com" --username="admin" --password="SuperSecretPassword123!"`)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "bootstrap":
		runBootstrap(os.Args[2:])
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %q\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func runBootstrap(args []string) {
	fs := flag.NewFlagSet("bootstrap", flag.ExitOnError)
	email := fs.String("email", "", "Admin email address")
	username := fs.String("username", "", "Admin username")
	password := fs.String("password", "", "Admin password")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Interactive prompts if flags missing
	reader := bufio.NewReader(os.Stdin)

	if *email == "" {
		fmt.Print("Enter admin email: ")
		input, _ := reader.ReadString('\n')
		*email = strings.TrimSpace(input)
	}

	if *username == "" {
		fmt.Print("Enter admin username: ")
		input, _ := reader.ReadString('\n')
		*username = strings.TrimSpace(input)
	}

	if *password == "" {
		fmt.Print("Enter admin password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println() // Add newline after hidden input
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read password: %v\n", err)
			os.Exit(1)
		}
		*password = string(bytePassword)

		fmt.Print("Confirm admin password: ")
		byteConfirm, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read password confirmation: %v\n", err)
			os.Exit(1)
		}
		if *password != string(byteConfirm) {
			fmt.Fprintln(os.Stderr, "Error: passwords do not match")
			os.Exit(1)
		}
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Connect to database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database connection failed: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	db := database.NewDB(pool)
	userService := users.NewService(db)

	adminUser, err := userService.BootstrapAdmin(ctx, *email, *username, *password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Bootstrap failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nSuccess: Administrator account bootstrapped successfully!")
	fmt.Printf("User ID:  %s\n", adminUser.ID)
	fmt.Printf("Email:    %s\n", adminUser.Email)
	fmt.Printf("Username: %s\n", adminUser.Username)
	fmt.Printf("Role:     %s\n", adminUser.Role)
	fmt.Printf("Status:   %s\n", adminUser.Status)
}
