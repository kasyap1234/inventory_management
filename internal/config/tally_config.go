package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// TallyConfig represents the complete configuration
type TallyConfig struct {
	Tally         TallyIntegration    `toml:"tally"`
	Queuing       QueuingConfig       `toml:"queuing"`
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
	RedisAddr       string            `toml:"redis_addr"`
	RedisPassword   string            `toml:"redis_password"`
	RedisDB         int               `toml:"redis_db"`
	Concurrency     int               `toml:"concurrency"`
	Queues          []string          `toml:"queues"`
	QueuePriorities map[string]int    `toml:"queue_priorities"`
}

// ExportImportConfig contains timeouts and retry settings
type ExportImportConfig struct {
	TimeoutSeconds   int `toml:"timeout_seconds"`
	MaxRetryAttempts int `toml:"max_retry_attempts"`
	RetryDelaySeconds int `toml:"retry_delay_seconds"`
}

// LoadTallyConfig loads configuration from a TOML file
func LoadTallyConfig(filename string) (*TallyConfig, error) {
	config := &TallyConfig{}
	_, err := toml.DecodeFile(filename, config)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}
	return config, nil
}