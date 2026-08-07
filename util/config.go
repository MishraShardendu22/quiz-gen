package util

import (
	"os"
	"runtime"
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

