package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoEnvFileFlagIgnoresDotenvAndReturnsMissingToken(t *testing.T) {
	workDir := t.TempDir()
	homeDir := t.TempDir()
	chdirForCLITest(t, workDir)
	t.Setenv("HOME", homeDir)

	writeCLIEnvFile(t, filepath.Join(workDir, ".env"), "SHORTCUT_API_TOKEN=from-local\n")
	writeCLIEnvFile(t, filepath.Join(homeDir, ".env"), "SHORTCUT_API_TOKEN=from-home\n")
	unsetCLIEnv(t, "SHORTCUT_API_TOKEN", "SHORTCUT_BASE_URL", "SHORTCUT_TIMEOUT")

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--no-env-file", "me"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected missing token error")
	}
	if !strings.Contains(err.Error(), "SHORTCUT_API_TOKEN environment variable is required") {
		t.Fatalf("expected missing token error, got %v", err)
	}
}

func TestExplicitMissingEnvFileReturnsLoadErrorWithoutFallback(t *testing.T) {
	workDir := t.TempDir()
	homeDir := t.TempDir()
	chdirForCLITest(t, workDir)
	t.Setenv("HOME", homeDir)

	writeCLIEnvFile(t, filepath.Join(workDir, ".env"), "SHORTCUT_API_TOKEN=from-local\n")
	writeCLIEnvFile(t, filepath.Join(homeDir, ".env"), "SHORTCUT_API_TOKEN=from-home\n")
	unsetCLIEnv(t, "SHORTCUT_API_TOKEN", "SHORTCUT_BASE_URL", "SHORTCUT_TIMEOUT")

	missingEnvPath := filepath.Join(workDir, "missing.env")
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--env-file", missingEnvPath, "me"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected env-file load error")
	}
	if !strings.Contains(err.Error(), "load env file") || !strings.Contains(err.Error(), missingEnvPath) {
		t.Fatalf("expected explicit env-file load error for %q, got %v", missingEnvPath, err)
	}
	if strings.Contains(err.Error(), "SHORTCUT_API_TOKEN") {
		t.Fatalf("expected env-file load error before token validation, got %v", err)
	}
}

func chdirForCLITest(t *testing.T, dir string) {
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

func writeCLIEnvFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func unsetCLIEnv(t *testing.T, keys ...string) {
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
