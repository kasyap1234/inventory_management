package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// TallyConfig represents the complete configuration
type TallyConfig struct {
	Tally        TallyIntegration   `toml:"tally"`
	Queuing      QueuingConfig      `toml:"queuing"`
	ExportImport ExportImportConfig `toml:"export_import"`
}

// TallyIntegration contains API and database settings for Tally
type TallyIntegration struct {
	APIKey      string `toml:"api_key"`
	APISecret   string `toml:"api_secret"`
	APIEndpoint string `toml:"api_endpoint"`
	DatabaseURL string `toml:"database_url"`
}

// QueuingConfig contains Redis and concurrency settings
type QueuingConfig struct {
	RedisAddr       string         `toml:"redis_addr"`
	RedisPassword   string         `toml:"redis_password"`
	RedisDB         int            `toml:"redis_db"`
	Concurrency     int            `toml:"concurrency"`
	Queues          []string       `toml:"queues"`
	QueuePriorities map[string]int `toml:"queue_priorities"`
}

// ExportImportConfig contains timeouts and retry settings
type ExportImportConfig struct {
	Mode              string `toml:"mode"`
	TimeoutSeconds    int    `toml:"timeout_seconds"`
	MaxRetryAttempts  int    `toml:"max_retry_attempts"`
	RetryDelaySeconds int    `toml:"retry_delay_seconds"`
}

/*
LoadTallyConfig loads configuration from a TOML file
Note: Webhook test configuration is environment-driven (not from TOML).
*/
func LoadTallyConfig(filename string) (*TallyConfig, error) {
	config := &TallyConfig{}
	_, err := toml.DecodeFile(filename, config)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}
	return config, nil
}

// WebhookTestConfig holds environment-driven settings for webhook test endpoint
type WebhookTestConfig struct {
	TimeoutMs        int      // default 5000
	MaxResponseBytes int      // default 2048
	AllowedSchemes   []string // default ["https"]
	AllowHTTPInDev   bool     // default false
	RateLimitPerMin  int      // default 5
	MaxRedirects     int      // default 0
}

// LoadWebhookTestConfigFromEnv reads env vars and applies defaults.
// Enforces HTTPS-only in production (GO_ENV=production|prod).
func LoadWebhookTestConfigFromEnv() *WebhookTestConfig {
	goEnv := strings.ToLower(strings.TrimSpace(os.Getenv("GO_ENV")))
	isProd := goEnv == "production" || goEnv == "prod"

	timeoutMs := parseIntEnv("WEBHOOK_TEST_TIMEOUT_MS", 5000)
	maxResp := parseIntEnv("WEBHOOK_TEST_MAX_RESPONSE_BYTES", 2048)
	ratePerMin := parseIntEnv("WEBHOOK_TEST_RATE_LIMIT_PER_MIN", 5)
	maxRedirects := parseIntEnv("WEBHOOK_TEST_MAX_REDIRECTS", 0)

	allowHTTPInDev := strings.ToLower(strings.TrimSpace(os.Getenv("WEBHOOK_TEST_ALLOW_HTTP_IN_DEV"))) == "true"
	allowedSchemesEnv := strings.TrimSpace(os.Getenv("WEBHOOK_TEST_ALLOWED_SCHEMES"))
	var allowedSchemes []string
	if allowedSchemesEnv == "" {
		allowedSchemes = []string{"https"}
	} else {
		parts := strings.Split(allowedSchemesEnv, ",")
		for _, p := range parts {
			p = strings.TrimSpace(strings.ToLower(p))
			if p != "" {
				allowedSchemes = append(allowedSchemes, p)
			}
		}
		if len(allowedSchemes) == 0 {
			allowedSchemes = []string{"https"}
		}
	}

	// Enforce HTTPS-only in production
	if isProd {
		allowedSchemes = []string{"https"}
		allowHTTPInDev = false
	}

	return &WebhookTestConfig{
		TimeoutMs:        timeoutMs,
		MaxResponseBytes: maxResp,
		AllowedSchemes:   allowedSchemes,
		AllowHTTPInDev:   allowHTTPInDev,
		RateLimitPerMin:  ratePerMin,
		MaxRedirects:     maxRedirects,
	}
}

// parseIntEnv reads an int env with default fallback
func parseIntEnv(key string, def int) int {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return def
	}
	if n, err := strconv.Atoi(val); err == nil {
		return n
	}
	return def
}
