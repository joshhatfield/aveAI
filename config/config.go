package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds all AVE configuration.
type Config struct {
	Version int            `yaml:"version"`
	DB      DBConfig       `yaml:"db"`
	Map     MapConfig      `yaml:"map"`
	Logging LogConfig      `yaml:"logging"`
	Search  SearchConfig   `yaml:"search"`
	Context ContextConfig  `yaml:"context"`
}

// DBConfig holds database path settings.
type DBConfig struct {
	Path string `yaml:"path"`
}

// MapConfig holds map.yaml path settings.
type MapConfig struct {
	Path string `yaml:"path"`
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level  string `yaml:"level"`  // debug | info | warn | error
	Format string `yaml:"format"` // text | json
}

// SearchConfig holds search defaults.
type SearchConfig struct {
	DefaultLimit int `yaml:"default_limit"`
}

// ContextConfig holds context/pull defaults.
type ContextConfig struct {
	MaxDepth    int  `yaml:"max_depth"`    // max hierarchy depth (0 = unlimited)
	MaxKeys     int  `yaml:"max_keys"`    // max total keys to show
	ShowCounts  bool `yaml:"show_counts"` // show item counts
	Summary     bool `yaml:"summary"`     // summary mode
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Version: 1,
		DB: DBConfig{
			Path: ".ave/data.avdb",
		},
		Map: MapConfig{
			Path: ".ave/map.yaml",
		},
		Logging: LogConfig{
			Level:  "info",
			Format: "text",
		},
		Search: SearchConfig{
			DefaultLimit: 10,
		},
		Context: ContextConfig{
			MaxDepth:   0,   // unlimited
			MaxKeys:    1000,
			ShowCounts: false,
			Summary:    false,
		},
	}
}

// Load reads config from the given project path, falling back to user-global config.
func Load(projectPath string) (*Config, error) {
	cfg := DefaultConfig()

	// Try project-local config
	projectConfigPath := filepath.Join(projectPath, "config.yaml")
	if data, err := os.ReadFile(projectConfigPath); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse %s: %w", projectConfigPath, err)
		}
	}

	// Try user-global config
	userConfigPath := filepath.Join(os.Getenv("HOME"), ".config", "ave", "config.yaml")
	if data, err := os.ReadFile(userConfigPath); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse %s: %w", userConfigPath, err)
		}
	}

	// Apply environment variable overrides
	cfg.ApplyEnvOverrides()

	return cfg, nil
}

// ApplyEnvOverrides checks environment variables and overrides config fields.
func (c *Config) ApplyEnvOverrides() {
	if v := os.Getenv("AVE_DB_PATH"); v != "" {
		c.DB.Path = v
	}
	if v := os.Getenv("AVE_MAP_PATH"); v != "" {
		c.Map.Path = v
	}
	if v := os.Getenv("AVE_LOG_LEVEL"); v != "" {
		c.Logging.Level = v
	}
	if v := os.Getenv("AVE_LOG_FORMAT"); v != "" {
		c.Logging.Format = v
	}
	if v := os.Getenv("AVE_SEARCH_LIMIT"); v != "" {
		var limit int
		if _, err := fmt.Sscanf(v, "%d", &limit); err == nil {
			c.Search.DefaultLimit = limit
		}
	}
}

// globalConfig is the process-global config instance.
var globalConfig *Config
var globalConfigPath string

// Global returns the process-global config, loading it if not already loaded.
func Global() *Config {
	if globalConfig == nil {
		cfg, err := Load(".ave")
		if err != nil {
			// Fall back to defaults if load fails
			cfg = DefaultConfig()
		}
		globalConfig = cfg
	}
	return globalConfig
}

// SetGlobal replaces the process-global config (used for testing).
func SetGlobal(cfg *Config) {
	globalConfig = cfg
}

// ResetGlobal clears the process-global config (forces reload on next Global()).
func ResetGlobal() {
	globalConfig = nil
	globalConfigPath = ""
}

// ResolveDBPath resolves the .avdb path using config precedence.
func (c *Config) ResolveDBPath(explicitPath string) string {
	if explicitPath != "" {
		return explicitPath
	}
	if c != nil && c.DB.Path != "" {
		return c.DB.Path
	}
	return ".ave/data.avdb"
}

// ResolveMapPath resolves the map.yaml path using config precedence.
func (c *Config) ResolveMapPath(explicitPath string) string {
	if explicitPath != "" {
		return explicitPath
	}
	if c != nil && c.Map.Path != "" {
		return c.Map.Path
	}
	return ".ave/map.yaml"
}

// Validate checks that the config is well-formed.
func (c *Config) Validate() error {
	if c.Logging.Level != "" {
		switch strings.ToLower(c.Logging.Level) {
		case "debug", "info", "warn", "error":
			// valid
		default:
			return fmt.Errorf("invalid log level: %q (want debug|info|warn|error)", c.Logging.Level)
		}
	}
	if c.Logging.Format != "" {
		switch strings.ToLower(c.Logging.Format) {
		case "text", "json":
			// valid
		default:
			return fmt.Errorf("invalid log format: %q (want text|json)", c.Logging.Format)
		}
	}
	return nil
}