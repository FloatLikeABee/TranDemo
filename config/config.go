package config

import (
	_ "embed"
	"os"
)

//go:embed form_sample.json
var FormSampleJSON string

type Config struct {
	Port             string
	GeminiAPIKey     string
	ModelName        string
	DBPath           string
	SQLFilesDir      string
	ResultsDir       string
	SitesDir         string
	VoiceSamplesDir  string
	ExternalAPIBase  string // Image reader, PDF reader, Gathering (e.g. http://localhost:8000)
	SQLServer        SQLServerConfig
}

type SQLServerConfig struct {
	Server   string
	Port     string
	Database string
	UserID   string
	Password string
	Encrypt  bool
}

func GetConfig() Config {
	return Config{
		Port:         getEnv("PORT", "9090"),
		// GeminiAPIKey: getEnv("GEMINI_API_KEY", "sk-c0587cfb940347c4b2a3c96f62360649"),
		GeminiAPIKey: "sk-c0587cfb940347c4b2a3c96f62360649",
		// ModelName:    getEnv("GEMINI_MODEL", "qwen3-coder"),
		ModelName:    "qwen3-max",
		DBPath:         getEnv("DB_PATH", "./data/badger"),
		SQLFilesDir:    getEnv("SQL_FILES_DIR", "./sql_files"),
		ResultsDir:     getEnv("RESULTS_DIR", "./results"),
		SitesDir:       getEnv("SITES_DIR", "./sites"),
		VoiceSamplesDir: getEnv("VOICE_SAMPLES_DIR", "./voice_samples"),
		ExternalAPIBase:  getEnv("EXTERNAL_API_BASE", "http://localhost:8000"),
		SQLServer: SQLServerConfig{
			Server:   getEnv("SQL_SERVER", "192.168.9.9"),
			Port:     getEnv("SQL_PORT", "1433"),
			Database: getEnv("SQL_DATABASE", "team2_ent"),
			UserID:   getEnv("SQL_USER", "tfuser"),
			Password: getEnv("SQL_PASSWORD", "$transfinder2006"),
			Encrypt:  getEnv("SQL_ENCRYPT", "true") == "true",
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

