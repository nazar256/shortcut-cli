package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultBaseURL = "https://api.app.shortcut.com"
	DefaultTimeout = 30 * time.Second
)

type Config struct {
	APIToken string
	BaseURL  string
	Timeout  time.Duration
}

type LoadOptions struct {
	EnvFilePath string
	NoEnvFile   bool
}

// Load reads configuration from environment variables, optionally loading from a .env file first.
func Load(options LoadOptions) (*Config, error) {
	if err := loadEnvFile(options); err != nil {
		return nil, err
	}

	token := os.Getenv("SHORTCUT_API_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("SHORTCUT_API_TOKEN environment variable is required")
	}

	baseURL := os.Getenv("SHORTCUT_BASE_URL")
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	timeout := DefaultTimeout
	if timeoutStr := os.Getenv("SHORTCUT_TIMEOUT"); timeoutStr != "" {
		if d, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = d
		}
	}

	return &Config{
		APIToken: token,
		BaseURL:  baseURL,
		Timeout:  timeout,
	}, nil
}

// loadEnvFile attempts to load environment values from dotenv files.
// It does not override existing environment variables.
func loadEnvFile(options LoadOptions) error {
	if options.NoEnvFile {
		return nil
	}

	if options.EnvFilePath != "" {
		if err := loadEnv(options.EnvFilePath); err != nil {
			return fmt.Errorf("load env file %q: %w", options.EnvFilePath, err)
		}
		return nil
	}

	loaded, err := loadEnvIfExists(".env")
	if err != nil {
		return err
	}
	if loaded {
		return nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		return nil
	}

	_, err = loadEnvIfExists(filepath.Join(homeDir, ".env"))
	return err
}

func loadEnvIfExists(filename string) (bool, error) {
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("stat env file %q: %w", filename, err)
	}

	if err := loadEnv(filename); err != nil {
		return false, fmt.Errorf("load env file %q: %w", filename, err)
	}

	return true, nil
}

func loadEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// Remove surrounding quotes if present
		if len(val) >= 2 && (val[0] == '"' && val[len(val)-1] == '"' || val[0] == '\'' && val[len(val)-1] == '\'') {
			val = val[1 : len(val)-1]
		}

		// Only set if not already set
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, val)
		}
	}

	return scanner.Err()
}
