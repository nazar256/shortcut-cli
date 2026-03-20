package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	t.Run("missing token", func(t *testing.T) {
		unsetEnv(t, "SHORTCUT_API_TOKEN", "SHORTCUT_BASE_URL", "SHORTCUT_TIMEOUT")

		_, err := Load()
		if err == nil {
			t.Fatal("expected error when SHORTCUT_API_TOKEN is missing")
		}
	})

	t.Run("defaults", func(t *testing.T) {
		t.Setenv("SHORTCUT_API_TOKEN", "test-token")
		unsetEnv(t, "SHORTCUT_BASE_URL", "SHORTCUT_TIMEOUT")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load returned error: %v", err)
		}

		if cfg.APIToken != "test-token" {
			t.Fatalf("unexpected token: %q", cfg.APIToken)
		}
		if cfg.BaseURL != DefaultBaseURL {
			t.Fatalf("unexpected base URL: %q", cfg.BaseURL)
		}
		if cfg.Timeout != DefaultTimeout {
			t.Fatalf("unexpected timeout: %v", cfg.Timeout)
		}
	})

	t.Run("custom values", func(t *testing.T) {
		t.Setenv("SHORTCUT_API_TOKEN", "custom-token")
		t.Setenv("SHORTCUT_BASE_URL", "https://custom.shortcut.com")
		t.Setenv("SHORTCUT_TIMEOUT", "10s")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load returned error: %v", err)
		}

		if cfg.APIToken != "custom-token" {
			t.Fatalf("unexpected token: %q", cfg.APIToken)
		}
		if cfg.BaseURL != "https://custom.shortcut.com" {
			t.Fatalf("unexpected base URL: %q", cfg.BaseURL)
		}
		if cfg.Timeout != 10*time.Second {
			t.Fatalf("unexpected timeout: %v", cfg.Timeout)
		}
	})

	t.Run("dotenv fills missing values without overriding env", func(t *testing.T) {
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("getwd: %v", err)
		}

		tmpDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("SHORTCUT_API_TOKEN=from-dotenv\nSHORTCUT_BASE_URL=https://dotenv.example\nSHORTCUT_TIMEOUT=12s\n"), 0o644); err != nil {
			t.Fatalf("write .env: %v", err)
		}

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("chdir temp dir: %v", err)
		}
		defer os.Chdir(cwd)

		unsetEnv(t, "SHORTCUT_API_TOKEN", "SHORTCUT_BASE_URL", "SHORTCUT_TIMEOUT")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load returned error: %v", err)
		}

		if cfg.APIToken != "from-dotenv" {
			t.Fatalf("expected .env token, got %q", cfg.APIToken)
		}
		if cfg.BaseURL != "https://dotenv.example" {
			t.Fatalf("expected .env base URL, got %q", cfg.BaseURL)
		}
		if cfg.Timeout != 12*time.Second {
			t.Fatalf("expected .env timeout, got %v", cfg.Timeout)
		}
	})
}

func unsetEnv(t *testing.T, keys ...string) {
	t.Helper()
	for _, key := range keys {
		value, exists := os.LookupEnv(key)
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
		t.Cleanup(func() {
			if exists {
				_ = os.Setenv(key, value)
				return
			}
			_ = os.Unsetenv(key)
		})
	}
}
