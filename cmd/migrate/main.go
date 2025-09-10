package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	host := envOr("PG_HOST", "localhost")
	port := envOr("PG_PORT", "5432")
	user := envOr("PG_USER", "order_user")
	pass := envOr("PG_PASSWORD", "order_pass")
	db := envOr("PG_DB", "order_db")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, port, db)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	log.Println("Running migrations...")
	sql, err := os.ReadFile("migrations/001_init.sql")
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	_, err = pool.Exec(context.Background(), string(sql))
	if err != nil {
		log.Fatalf("Failed to apply migration: %v", err)
	}

	log.Println("Migrations applied successfully!")
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}