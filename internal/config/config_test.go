package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	t.Run("missing token", func(t *testing.T) {
		unsetEnv(t, "SHORTCUT_API_TOKEN", "SHORTCUT_BASE_URL", "SHORTCUT_TIMEOUT")

		_, err := Load(LoadOptions{})
		if err == nil {
			t.Fatal("expected error when SHORTCUT_API_TOKEN is missing")
		}
	})

	t.Run("defaults", func(t *testing.T) {
		t.Setenv("SHORTCUT_API_TOKEN", "test-token")
		unsetEnv(t, "SHORTCUT_BASE_URL", "SHORTCUT_TIMEOUT")

		cfg, err := Load(LoadOptions{})
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

		cfg, err := Load(LoadOptions{})
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

	t.Run("home dotenv loads when local dotenv is absent", func(t *testing.T) {
		workDir := t.TempDir()
		homeDir := t.TempDir()
		chdir(t, workDir)
		t.Setenv("HOME", homeDir)

		writeEnvFile(t, filepath.Join(homeDir, ".env"), "SHORTCUT_API_TOKEN=from-home\nSHORTCUT_BASE_URL=https://home.example\nSHORTCUT_TIMEOUT=11s\n")
		unsetEnv(t, "SHORTCUT_API_TOKEN", "SHORTCUT_BASE_URL", "SHORTCUT_TIMEOUT")

		cfg, err := Load(LoadOptions{})
		if err != nil {
			t.Fatalf("Load returned error: %v", err)
		}

		if cfg.APIToken != "from-home" {
			t.Fatalf("expected home .env token, got %q", cfg.APIToken)
		}
		if cfg.BaseURL != "https://home.example" {
			t.Fatalf("expected home .env base URL, got %q", cfg.BaseURL)
		}
		if cfg.Timeout != 11*time.Second {
			t.Fatalf("expected home .env timeout, got %v", cfg.Timeout)
		}
	})

	t.Run("local dotenv wins over home dotenv", func(t *testing.T) {
		workDir := t.TempDir()
		homeDir := t.TempDir()
		chdir(t, workDir)
		t.Setenv("HOME", homeDir)

		writeEnvFile(t, filepath.Join(workDir, ".env"), "SHORTCUT_API_TOKEN=from-local\nSHORTCUT_BASE_URL=https://local.example\nSHORTCUT_TIMEOUT=9s\n")
		writeEnvFile(t, filepath.Join(homeDir, ".env"), "SHORTCUT_API_TOKEN=from-home\nSHORTCUT_BASE_URL=https://home.example\nSHORTCUT_TIMEOUT=11s\n")
		unsetEnv(t, "SHORTCUT_API_TOKEN", "SHORTCUT_BASE_URL", "SHORTCUT_TIMEOUT")

		cfg, err := Load(LoadOptions{})
		if err != nil {
			t.Fatalf("Load returned error: %v", err)
		}

		if cfg.APIToken != "from-local" {
			t.Fatalf("expected local .env token, got %q", cfg.APIToken)
		}
		if cfg.BaseURL != "https://local.example" {
			t.Fatalf("expected local .env base URL, got %q", cfg.BaseURL)
		}
		if cfg.Timeout != 9*time.Second {
			t.Fatalf("expected local .env timeout, got %v", cfg.Timeout)
		}
	})

	t.Run("process env overrides home dotenv", func(t *testing.T) {
		workDir := t.TempDir()
		homeDir := t.TempDir()
		chdir(t, workDir)
		t.Setenv("HOME", homeDir)

		writeEnvFile(t, filepath.Join(homeDir, ".env"), "SHORTCUT_API_TOKEN=from-home\nSHORTCUT_BASE_URL=https://home.example\nSHORTCUT_TIMEOUT=7s\n")
		t.Setenv("SHORTCUT_API_TOKEN", "from-process")
		t.Setenv("SHORTCUT_BASE_URL", "")
		unsetEnv(t, "SHORTCUT_TIMEOUT")

		cfg, err := Load(LoadOptions{})
		if err != nil {
			t.Fatalf("Load returned error: %v", err)
		}

		if cfg.APIToken != "from-process" {
			t.Fatalf("expected process env token, got %q", cfg.APIToken)
		}
		if cfg.BaseURL != DefaultBaseURL {
			t.Fatalf("expected default base URL from empty process env, got %q", cfg.BaseURL)
		}
		if cfg.Timeout != 7*time.Second {
			t.Fatalf("expected home .env timeout for missing process env, got %v", cfg.Timeout)
		}
	})

	t.Run("no env file option ignores local and home dotenv", func(t *testing.T) {
		workDir := t.TempDir()
		homeDir := t.TempDir()
		chdir(t, workDir)
		t.Setenv("HOME", homeDir)

		writeEnvFile(t, filepath.Join(workDir, ".env"), "SHORTCUT_API_TOKEN=from-local\nSHORTCUT_BASE_URL=https://local.example\nSHORTCUT_TIMEOUT=9s\n")
		writeEnvFile(t, filepath.Join(homeDir, ".env"), "SHORTCUT_API_TOKEN=from-home\nSHORTCUT_BASE_URL=https://home.example\nSHORTCUT_TIMEOUT=11s\n")
		t.Setenv("SHORTCUT_API_TOKEN", "from-process")
		unsetEnv(t, "SHORTCUT_BASE_URL", "SHORTCUT_TIMEOUT")

		cfg, err := Load(LoadOptions{NoEnvFile: true})
		if err != nil {
			t.Fatalf("Load returned error: %v", err)
		}

		if cfg.APIToken != "from-process" {
			t.Fatalf("expected process env token, got %q", cfg.APIToken)
		}
		if cfg.BaseURL != DefaultBaseURL {
			t.Fatalf("expected default base URL, got %q", cfg.BaseURL)
		}
		if cfg.Timeout != DefaultTimeout {
			t.Fatalf("expected default timeout, got %v", cfg.Timeout)
		}
	})

	t.Run("missing explicit env file returns error and does not fallback", func(t *testing.T) {
		workDir := t.TempDir()
		homeDir := t.TempDir()
		chdir(t, workDir)
		t.Setenv("HOME", homeDir)

		writeEnvFile(t, filepath.Join(workDir, ".env"), "SHORTCUT_API_TOKEN=from-local\n")
		writeEnvFile(t, filepath.Join(homeDir, ".env"), "SHORTCUT_API_TOKEN=from-home\n")
		unsetEnv(t, "SHORTCUT_API_TOKEN", "SHORTCUT_BASE_URL", "SHORTCUT_TIMEOUT")

		explicitPath := filepath.Join(workDir, "missing.env")
		_, err := Load(LoadOptions{EnvFilePath: explicitPath})
		if err == nil {
			t.Fatal("expected error for missing explicit env file")
		}
		if !strings.Contains(err.Error(), explicitPath) {
			t.Fatalf("expected error to mention explicit path %q, got %v", explicitPath, err)
		}
		if strings.Contains(err.Error(), "SHORTCUT_API_TOKEN") {
			t.Fatalf("expected env-file load error before config validation, got %v", err)
		}
	})
}

func chdir(t *testing.T, dir string) {
	t.Helper()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %q: %v", dir, err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd %q: %v", cwd, err)
		}
	})
}

func writeEnvFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
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
