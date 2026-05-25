package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/domain/user"
	userrepo "github.com/liang21/heka/internal/infrastructure/persistence/postgres"
)

func bcryptHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func main() {
	email := flag.String("email", "", "Admin email (required)")
	password := flag.String("password", "", "Admin password (required)")
	name := flag.String("name", "Admin", "Admin display name")
	dbHost := flag.String("db-host", "localhost", "Database host")
	dbPort := flag.Int("db-port", 5432, "Database port")
	dbUser := flag.String("db-user", "heka", "Database user")
	dbPass := flag.String("db-password", "heka", "Database password")
	dbName := flag.String("db-name", "heka", "Database name")

	flag.Parse()

	if *email == "" || *password == "" {
		log.Fatal("--email and --password are required")
	}

	if len(*password) < 8 {
		log.Fatal("Password must be at least 8 characters")
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		*dbHost, *dbPort, *dbUser, *dbPass, *dbName,
	)

	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Use bcrypt directly for password hashing
	// In production, you should use the auth package hasher
	hashedPassword, err := bcryptHash(*password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	admin := &user.User{
		ID:           shared.NewID(),
		Name:         *name,
		Email:        *email,
		PasswordHash: hashedPassword,
	}

	repo := userrepo.NewUserRepository(db)

	ctx := context.Background()
	existing, _ := repo.FindByEmail(ctx, *email)
	if existing != nil {
		log.Printf("User with email %s already exists (ID: %s)", *email, existing.ID)
		return
	}

	if err := repo.Create(ctx, admin); err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	log.Printf("Admin user created successfully:")
	log.Printf("  ID: %s", admin.ID)
	log.Printf("  Name: %s", admin.Name)
	log.Printf("  Email: %s", admin.Email)
}
