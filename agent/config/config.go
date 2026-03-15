package config

import (
	"net/url"
	"os"
	"strings"
)

// Config holds application configuration loaded from the environment.
// Single responsibility: configuration only.
type Config struct {
	APIKey      string
	Port        string
	UseVertex   bool
	LaravelURL  string
	AgentSecret string
	DBHost      string
	DBPort      string
	DBName      string
	DBUser      string
	DBPassword  string
	DBSSLMode   string
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
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "127.0.0.1"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbName := os.Getenv("DB_DATABASE")
	if dbName == "" {
		dbName = "niobe"
	}
	dbUser := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbSSLMode := os.Getenv("DB_SSLMODE")
	if dbSSLMode == "" {
		dbSSLMode = "disable"
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
		DBHost:       dbHost,
		DBPort:       dbPort,
		DBName:       dbName,
		DBUser:       dbUser,
		DBPassword:   dbPassword,
		DBSSLMode:    dbSSLMode,
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

func (c Config) DatabaseDSN() string {
	return "host=" + c.DBHost +
		" port=" + c.DBPort +
		" dbname=" + c.DBName +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" sslmode=" + c.DBSSLMode
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
