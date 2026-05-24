package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.DB.Path != ".ave/data.avdb" {
		t.Errorf("expected default db path '.ave/data.avdb', got %q", cfg.DB.Path)
	}
	if cfg.Map.Path != ".ave/map.yaml" {
		t.Errorf("expected default map path '.ave/map.yaml', got %q", cfg.Map.Path)
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("expected default log level 'info', got %q", cfg.Logging.Level)
	}
	if cfg.Search.DefaultLimit != 10 {
		t.Errorf("expected default limit 10, got %d", cfg.Search.DefaultLimit)
	}
}

func TestLoadConfigFile(t *testing.T) {
	content := []byte(`
version: 1
db:
  path: "/custom/db.avdb"
map:
  path: "/custom/map.yaml"
logging:
  level: "debug"
search:
  default_limit: 5
`)
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	os.WriteFile(configPath, content, 0644)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.DB.Path != "/custom/db.avdb" {
		t.Errorf("expected '/custom/db.avdb', got %q", cfg.DB.Path)
	}
	if cfg.Map.Path != "/custom/map.yaml" {
		t.Errorf("expected '/custom/map.yaml', got %q", cfg.Map.Path)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("expected 'debug', got %q", cfg.Logging.Level)
	}
	if cfg.Search.DefaultLimit != 5 {
		t.Errorf("expected 5, got %d", cfg.Search.DefaultLimit)
	}
}

func TestLoadMissingConfigFallsBackToDefaults(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should fall back to defaults
	if cfg.DB.Path != ".ave/data.avdb" {
		t.Errorf("expected default db path, got %q", cfg.DB.Path)
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	cfg := DefaultConfig()

	os.Setenv("AVE_DB_PATH", "/env/db.avdb")
	os.Setenv("AVE_MAP_PATH", "/env/map.yaml")
	os.Setenv("AVE_LOG_LEVEL", "debug")
	os.Setenv("AVE_SEARCH_LIMIT", "20")
	defer func() {
		os.Unsetenv("AVE_DB_PATH")
		os.Unsetenv("AVE_MAP_PATH")
		os.Unsetenv("AVE_LOG_LEVEL")
		os.Unsetenv("AVE_SEARCH_LIMIT")
	}()

	cfg.ApplyEnvOverrides()

	if cfg.DB.Path != "/env/db.avdb" {
		t.Errorf("expected '/env/db.avdb', got %q", cfg.DB.Path)
	}
	if cfg.Map.Path != "/env/map.yaml" {
		t.Errorf("expected '/env/map.yaml', got %q", cfg.Map.Path)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("expected 'debug', got %q", cfg.Logging.Level)
	}
	if cfg.Search.DefaultLimit != 20 {
		t.Errorf("expected 20, got %d", cfg.Search.DefaultLimit)
	}
}

func TestApplyEnvOverridesPartial(t *testing.T) {
	cfg := DefaultConfig()

	os.Setenv("AVE_LOG_LEVEL", "warn")
	defer os.Unsetenv("AVE_LOG_LEVEL")

	cfg.ApplyEnvOverrides()

	// Only log level should change
	if cfg.DB.Path != ".ave/data.avdb" {
		t.Errorf("expected default db path, got %q", cfg.DB.Path)
	}
	if cfg.Logging.Level != "warn" {
		t.Errorf("expected 'warn', got %q", cfg.Logging.Level)
	}
}

func TestValidateLogLevel(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Logging.Level = "invalid"

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for invalid log level")
	}
}

func TestValidateLogLevelValid(t *testing.T) {
	cfg := DefaultConfig()
	for _, level := range []string{"debug", "info", "warn", "error"} {
		cfg.Logging.Level = level
		if err := cfg.Validate(); err != nil {
			t.Errorf("unexpected error for level %q: %v", level, err)
		}
	}
}

func TestResolveDBPath(t *testing.T) {
	cfg := DefaultConfig()

	// explicit path takes precedence
	if cfg.ResolveDBPath("/explicit/db.avdb") != "/explicit/db.avdb" {
		t.Errorf("expected explicit path")
	}

	// falls back to config
	cfg.DB.Path = "/config/db.avdb"
	if cfg.ResolveDBPath("") != "/config/db.avdb" {
		t.Errorf("expected config path")
	}

	// falls back to default
	if cfg.ResolveDBPath("") != "/config/db.avdb" {
		t.Errorf("expected config path")
	}
}

func TestResolveMapPath(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.ResolveMapPath("/explicit/map.yaml") != "/explicit/map.yaml" {
		t.Errorf("expected explicit path")
	}

	cfg.Map.Path = "/config/map.yaml"
	if cfg.ResolveMapPath("") != "/config/map.yaml" {
		t.Errorf("expected config path")
	}
}

func TestSetGlobalAndReset(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DB.Path = "/test/path"

	SetGlobal(cfg)
	if Global().DB.Path != "/test/path" {
		t.Errorf("expected global config to be set")
	}

	ResetGlobal()
	// After reset, Global() should reload (with defaults since no config file)
	if Global() == nil {
		t.Errorf("expected non-nil global after reset")
	}
}