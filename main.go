package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to load config: %v\n", err)
		os.Exit(1)
	}

	err = RunMigrations(cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to run migrations: %v\n", err)
		os.Exit(1)
	}

	dbpool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	repo := NewBookRepo(dbpool)
	bookHandler := NewBookHandler(repo)

	r := chi.NewRouter()
	bookHandler.RegisterRoutes(r)

	err = http.ListenAndServe(":"+cfg.Port, r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to start server: %v\n", err)
		os.Exit(1)
	}
}
