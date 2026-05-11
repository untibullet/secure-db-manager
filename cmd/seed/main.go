// cmd/seed creates the first admin user in a fresh database.
// Run once after migrations: go run cmd/seed/main.go
// Env vars: DB_HOST, DB_PORT, DB_NAME, DB_SUPERUSER, DB_SUPERUSER_PASSWORD (same as the app).
// Flags: -username (default: admin) -password (default: admin123)
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	username := flag.String("username", "admin", "admin username")
	password := flag.String("password", "admin123", "admin password")
	flag.Parse()

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		getenv("DB_SUPERUSER", "postgres"),
		getenv("DB_SUPERUSER_PASSWORD", "postgres"),
		getenv("DB_HOST", "localhost"),
		getenv("DB_PORT", "5432"),
		getenv("DB_NAME", "security_db"),
	)

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close(ctx)

	hash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("bcrypt: %v", err)
	}

	// PG login role (idempotent via DO block).
	_, err = conn.Exec(ctx, fmt.Sprintf(`
		DO $$
		BEGIN
			IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = '%s') THEN
				CREATE ROLE %s WITH LOGIN PASSWORD '%s';
				GRANT db_admin TO %s;
			END IF;
		END
		$$`, *username, *username, *password, *username))
	if err != nil {
		log.Fatalf("create pg role: %v", err)
	}

	// users record.
	var userID int
	err = conn.QueryRow(ctx,
		`INSERT INTO users (username, password_hash, is_active)
		 VALUES ($1, $2, TRUE)
		 ON CONFLICT (username) DO UPDATE SET password_hash = EXCLUDED.password_hash
		 RETURNING user_id`,
		*username, string(hash),
	).Scan(&userID)
	if err != nil {
		log.Fatalf("insert user: %v", err)
	}

	// user_roles record.
	_, err = conn.Exec(ctx,
		`INSERT INTO user_roles (user_id, role_id)
		 SELECT $1, role_id FROM roles WHERE code = 'ADMIN'
		 ON CONFLICT DO NOTHING`,
		userID,
	)
	if err != nil {
		log.Fatalf("insert user_roles: %v", err)
	}

	fmt.Printf("OK: user %q created (id=%d), password=%q\n", *username, userID, *password)
	fmt.Println("Login at http://localhost:8080/login")
}
