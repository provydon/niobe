package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	"niobe/agent/config"
	"niobe/agent/handler"
	"niobe/agent/live"
	"niobe/agent/store"
)

func main() {
	_ = godotenv.Load("../niobe/.env", ".env")

	cfg := config.Load()
	connector := live.GoogleConnector{}

	// Register /health so Cloud Run can see the container is up; listen immediately on PORT
	http.HandleFunc("/health", handler.Health())

	// Listen on PORT immediately so Cloud Run sees the container as started (required: PORT=8080)
	go func() {
		log.Printf("AI waitress API listening on :%s", cfg.Port)
		if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
			log.Fatal(err)
		}
	}()

	// Connect to DB after listening; avoid blocking startup on slow Cloud SQL
	db, err := sql.Open("pgx", cfg.DatabaseDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("database ping: %v", err)
	}

	waitresses := store.NewWaitressRepository(db)
	http.HandleFunc("/live", handler.Live(connector, cfg, waitresses, db, &cfg))

	select {} // block forever; server runs in goroutine
}
