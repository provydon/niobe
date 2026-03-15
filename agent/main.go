package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"

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
	db, err := sql.Open("pgx", cfg.DatabaseDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		log.Fatal(err)
	}

	waitresses := store.NewWaitressRepository(db)

	http.HandleFunc("/health", handler.Health())
	http.HandleFunc("/live", handler.Live(connector, cfg, waitresses, db, &cfg))

	log.Printf("AI waitress API listening on http://localhost:%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Fatal(err)
	}
}
