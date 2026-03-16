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
	if cfg.DatabaseDSN() == "" {
		log.Fatal("database config missing: set DATABASE_URL or Laravel-style DB_HOST, DB_PORT, DB_DATABASE, DB_USERNAME, DB_PASSWORD")
	}
	connector := live.GoogleConnector{}

	// Register /health first so Cloud Run sees the container as up as soon as we listen
	http.HandleFunc("/health", handler.Health())

	// Listen on PORT immediately (Cloud Run requires listening within startup timeout)
	go func() {
		log.Printf("AI waitress API listening on :%s", cfg.Port)
		if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
			log.Fatal(err)
		}
	}()

	// Connect to DB in background; do not block or fatal so container stays up even if Cloud SQL is slow
	go func() {
		for {
			db, err := sql.Open("pgx", cfg.DatabaseDSN())
			if err != nil {
				log.Printf("database open: %v; retrying in 5s", err)
				time.Sleep(5 * time.Second)
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			err = db.PingContext(ctx)
			cancel()
			if err != nil {
				log.Printf("database ping: %v; retrying in 5s", err)
				_ = db.Close()
				time.Sleep(5 * time.Second)
				continue
			}
			waitresses := store.NewWaitressRepository(db)
			http.HandleFunc("/live", handler.Live(connector, cfg, waitresses, db, &cfg))
			log.Print("database connected, /live registered")
			return
		}
	}()

	select {} // block forever; server runs in goroutine
}
