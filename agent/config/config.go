package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// Config holds application configuration loaded from the environment.
// Single responsibility: configuration only.
type Config struct {
	APIKey       string
	Port         string
	UseVertex   bool
	LaravelURL  string
	AgentSecret string
	DatabaseURL string // PostgreSQL URL from DATABASE_URL/DB_URL or built from DB_* (same as Laravel)
	// SMTP for sending tool emails (e.g. order notifications)
	MailHost     string
	MailPort     string
	MailUser     string
	MailPassword string
	MailFrom     string
	MailFromName string
}

// Load reads configuration from the environment (and .env if present).
func Load() Config {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}
	laravelURL := os.Getenv("APP_URL")
	if laravelURL == "" {
		laravelURL = "http://localhost:8000"
	}
	laravelURL = normalizeLaravelURL(laravelURL)
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DB_URL")
	}
	if dbURL == "" {
		dbURL = buildDSNFromLaravelEnv()
	}
	useVertex := os.Getenv("GOOGLE_GENAI_USE_VERTEXAI") == "True" ||
		os.Getenv("GOOGLE_GENAI_USE_VERTEXAI") == "true"
	mailHost := os.Getenv("MAIL_HOST")
	if mailHost == "" {
		mailHost = "127.0.0.1"
	}
	mailPort := os.Getenv("MAIL_PORT")
	if mailPort == "" {
		mailPort = "1025"
	}
	return Config{
		APIKey:       apiKey,
		Port:         port,
		UseVertex:    useVertex,
		LaravelURL:   laravelURL,
		AgentSecret:  os.Getenv("AGENT_SHARED_SECRET"),
		DatabaseURL:  dbURL,
		MailHost:     mailHost,
		MailPort:     mailPort,
		MailUser:     os.Getenv("MAIL_USERNAME"),
		MailPassword: os.Getenv("MAIL_PASSWORD"),
		MailFrom:     os.Getenv("MAIL_FROM_ADDRESS"),
		MailFromName: os.Getenv("MAIL_FROM_NAME"),
	}
}

// GetAPIKey returns the API key so Config can satisfy live.Config.
func (c Config) GetAPIKey() string { return c.APIKey }

// GetUseVertex returns whether Vertex AI is used so Config can satisfy live.Config.
func (c Config) GetUseVertex() bool { return c.UseVertex }

// DatabaseDriver returns "sqlite" when DATABASE_URL is a file: URL or path ending in .sqlite; otherwise "pgx".
func (c Config) DatabaseDriver() string {
	u := strings.TrimSpace(c.DatabaseURL)
	if strings.HasPrefix(u, "file:") || (strings.Contains(u, ".sqlite") && !strings.HasPrefix(u, "postgres")) {
		return "sqlite"
	}
	return "pgx"
}

// DatabaseDSN returns the connection URL or path (from DATABASE_URL/DB_URL or built from Laravel DB_* vars).
func (c Config) DatabaseDSN() string {
	return c.DatabaseURL
}

// buildDSNFromLaravelEnv builds a PostgreSQL DSN from Laravel-style DB_* env vars (same as niobe .env).
// Supports Cloud SQL Unix socket (DB_HOST=/cloudsql/PROJECT:REGION:INSTANCE) and TCP (DB_HOST=127.0.0.1).
func buildDSNFromLaravelEnv() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	database := os.Getenv("DB_DATABASE")
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	if database == "" || username == "" {
		return ""
	}
	// Cloud SQL Unix socket: host is path like /cloudsql/niobe-489920:us-central1:niobe-db
	if strings.HasPrefix(host, "/cloudsql/") {
		// postgres://user:pass@/dbname?host=/cloudsql/CONN
		escPass := url.QueryEscape(password)
		return fmt.Sprintf("postgres://%s:%s@/%s?host=%s", username, escPass, database, url.QueryEscape(host))
	}
	// TCP
	escPass := url.QueryEscape(password)
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, escPass, host, port, database)
}

func normalizeLaravelURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	if parsed.Scheme == "http" && strings.HasSuffix(parsed.Hostname(), ".test") {
		parsed.Scheme = "https"
		return parsed.String()
	}

	return raw
}
