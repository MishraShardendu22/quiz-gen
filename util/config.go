package util

import (
	"bufio"
	"os"
	"runtime"
	"strings"
	"time"
)

// ConfigStruct holds all application-wide configuration values.
// This is the SINGLE SOURCE OF TRUTH for application constants, endpoints, and settings.
type ConfigStruct struct {
	ModelName          string
	OpenRouterBaseURL  string
	OpenRouterAPIKey   string
	HTTPTimeout        time.Duration
	RetryCount         int
	RetryWaitTime      time.Duration
	RetryMaxWaitTime   time.Duration
	DefaultTokenBudget int
	MaxRequestedCount  int
	WorkerCount        int
	ServerPort         string
	DatabasePath       string
	ContentPackDir     string
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
}

// Config is the global configuration instance used across all packages.
// .env ONLY contains OPENROUTER_API_KEY; all other settings are hardcoded here as constants.
var Config = ConfigStruct{
	RetryCount:         3,
	MaxRequestedCount:  20,
	DefaultTokenBudget: 5000,
	WorkerCount:        runtime.NumCPU(),
	ServerPort:         ":9000",
	DatabasePath:       "./quiz.db",
	ContentPackDir:     "./content-pack",
	RetryWaitTime:      2 * time.Second,
	RetryMaxWaitTime:   10 * time.Second,
	HTTPTimeout:        60 * time.Second,
	ReadTimeout:        30 * time.Second,
	WriteTimeout:       60 * time.Second,
	IdleTimeout:        120 * time.Second,
	OpenRouterBaseURL:  "https://openrouter.ai/api/v1",
	OpenRouterAPIKey:   os.Getenv("OPENROUTER_API_KEY"),
	ModelName:          "cohere/north-mini-code:free",
}

func init() {
	if Config.OpenRouterAPIKey == "" {
		Config.OpenRouterAPIKey = loadEnvKey(".env", "OPENROUTER_API_KEY")
	}
}

// loadEnvKey parses a key from a simple .env file without external dependencies
func loadEnvKey(filePath, key string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == key {
			val := strings.TrimSpace(parts[1])
			return strings.Trim(val, `"'`)
		}
	}
	return ""
}

