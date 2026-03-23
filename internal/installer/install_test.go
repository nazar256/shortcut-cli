package installer

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const fixtureVersion = "v-test"

func TestInstallScriptSkipsLanguageManagedPathEntriesAndUsesCanonicalAllowedPath(t *testing.T) {
	homeDir := t.TempDir()
	disallowedDir := filepath.Join(homeDir, ".nvm", "versions", "node", "v20.19.4", "bin")
	if err := os.MkdirAll(disallowedDir, 0o755); err != nil {
		t.Fatalf("mkdir disallowed dir: %v", err)
	}

	allowedRealDir := filepath.Join(t.TempDir(), "usr-local-bin-real")
	if err := os.MkdirAll(allowedRealDir, 0o755); err != nil {
		t.Fatalf("mkdir allowed dir: %v", err)
	}
	allowedLinkDir := filepath.Join(t.TempDir(), "usr-local-bin-link")
	if err := os.Symlink(allowedRealDir, allowedLinkDir); err != nil {
		t.Fatalf("symlink allowed dir: %v", err)
	}

	baseURL := releaseFixtureServer(t)
	output := runInstallScript(t, []string{"--version", fixtureVersion}, map[string]string{
		"SHORTCUT_INSTALL_BASE_URL": baseURL,
		"HOME":                      homeDir,
		"NVM_DIR":                   filepath.Join(homeDir, ".nvm"),
		"PATH":                      joinPath(disallowedDir, allowedLinkDir, requiredToolPath(t)),
	})

	installedPath := filepath.Join(allowedRealDir, "shortcut")
	if !strings.Contains(output, "Installed to "+installedPath) {
		t.Fatalf("expected installer output to mention canonical install path %q, got:\n%s", installedPath, output)
	}
	assertExecutableExists(t, installedPath)

	if _, err := os.Stat(filepath.Join(disallowedDir, "shortcut")); err == nil {
		t.Fatalf("expected installer to skip language-managed PATH dir %q", disallowedDir)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat disallowed install path: %v", err)
	}
	assertVersionOutput(t, installedPath)
	if strings.Contains(installedPath, ".nvm") || strings.Contains(installedPath, filepath.Base(allowedLinkDir)) {
		t.Fatalf("installer should use canonical non-language-managed dir, got %q", installedPath)
	}
	if !strings.Contains(installedPath, allowedRealDir) {
		t.Fatalf("expected install into canonical allowed dir %q, got %q", allowedRealDir, installedPath)
	}
}

func TestInstallScriptFallsBackToHomeBinWhenOnlyLanguageManagedPathEntriesExist(t *testing.T) {
	homeDir := t.TempDir()
	disallowedDir := filepath.Join(homeDir, ".cargo", "bin")
	if err := os.MkdirAll(disallowedDir, 0o755); err != nil {
		t.Fatalf("mkdir disallowed dir: %v", err)
	}

	baseURL := releaseFixtureServer(t)
	output := runInstallScript(t, []string{"--version", fixtureVersion}, map[string]string{
		"SHORTCUT_INSTALL_BASE_URL": baseURL,
		"HOME":                      homeDir,
		"PATH":                      joinPath(disallowedDir, requiredToolPath(t)),
	})

	installedPath := filepath.Join(homeDir, ".local", "bin", "shortcut")
	assertExecutableExists(t, installedPath)
	assertVersionOutput(t, installedPath)
	if !strings.Contains(output, "Installed to "+installedPath) {
		t.Fatalf("expected installer output to mention fallback install path %q, got:\n%s", installedPath, output)
	}

	if _, err := os.Stat(filepath.Join(disallowedDir, "shortcut")); err == nil {
		t.Fatalf("expected installer to skip language-managed PATH dir %q", disallowedDir)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat disallowed install path: %v", err)
	}
}

func TestInstallScriptHonorsExplicitInstallDir(t *testing.T) {
	homeDir := t.TempDir()
	explicitDir := filepath.Join(homeDir, "custom-bin")
	disallowedDir := filepath.Join(homeDir, ".nvm", "versions", "node", "v20.19.4", "bin")
	if err := os.MkdirAll(disallowedDir, 0o755); err != nil {
		t.Fatalf("mkdir disallowed dir: %v", err)
	}

	baseURL := releaseFixtureServer(t)
	output := runInstallScript(t, []string{"--version", fixtureVersion, "--install-dir", explicitDir}, map[string]string{
		"SHORTCUT_INSTALL_BASE_URL": baseURL,
		"HOME":                      homeDir,
		"NVM_DIR":                   filepath.Join(homeDir, ".nvm"),
		"PATH":                      joinPath(disallowedDir, requiredToolPath(t)),
	})

	installedPath := filepath.Join(explicitDir, "shortcut")
	assertExecutableExists(t, installedPath)
	assertVersionOutput(t, installedPath)
	if !strings.Contains(output, "Installed to "+installedPath) {
		t.Fatalf("expected installer output to mention explicit install path %q, got:\n%s", installedPath, output)
	}

	if _, err := os.Stat(filepath.Join(homeDir, ".local", "bin", "shortcut")); err == nil {
		t.Fatalf("expected explicit install dir to win over fallback locations")
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat fallback install path: %v", err)
	}
	if _, err := os.Stat(filepath.Join(disallowedDir, "shortcut")); err == nil {
		t.Fatalf("expected installer to ignore disallowed PATH dir when explicit dir is set")
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat disallowed install path: %v", err)
	}
}

func TestInstallScriptSkipsCustomGoBin(t *testing.T) {
	homeDir := t.TempDir()
	goBinRealDir := filepath.Join(t.TempDir(), "custom-gobin-real")
	if err := os.MkdirAll(goBinRealDir, 0o755); err != nil {
		t.Fatalf("mkdir gobin dir: %v", err)
	}
	goBinLinkDir := filepath.Join(t.TempDir(), "custom-gobin-link")
	if err := os.Symlink(goBinRealDir, goBinLinkDir); err != nil {
		t.Fatalf("symlink gobin dir: %v", err)
	}

	baseURL := releaseFixtureServer(t)
	output := runInstallScript(t, []string{"--version", fixtureVersion}, map[string]string{
		"SHORTCUT_INSTALL_BASE_URL": baseURL,
		"HOME":                      homeDir,
		"GOBIN":                     goBinLinkDir,
		"PATH":                      joinPath(goBinLinkDir, requiredToolPath(t)),
	})

	installedPath := filepath.Join(homeDir, ".local", "bin", "shortcut")
	assertExecutableExists(t, installedPath)
	assertVersionOutput(t, installedPath)
	if !strings.Contains(output, "Installed to "+installedPath) {
		t.Fatalf("expected installer output to mention fallback install path %q, got:\n%s", installedPath, output)
	}
	if _, err := os.Stat(filepath.Join(goBinRealDir, "shortcut")); err == nil {
		t.Fatalf("expected installer to skip canonicalized GOBIN dir %q", goBinRealDir)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat gobin install path: %v", err)
	}
}

func TestInstallScriptSkipsCustomPnpmHome(t *testing.T) {
	homeDir := t.TempDir()
	pnpmRealDir := filepath.Join(t.TempDir(), "custom-pnpm-real")
	if err := os.MkdirAll(pnpmRealDir, 0o755); err != nil {
		t.Fatalf("mkdir pnpm dir: %v", err)
	}
	pnpmLinkDir := filepath.Join(t.TempDir(), "custom-pnpm-link")
	if err := os.Symlink(pnpmRealDir, pnpmLinkDir); err != nil {
		t.Fatalf("symlink pnpm dir: %v", err)
	}

	baseURL := releaseFixtureServer(t)
	output := runInstallScript(t, []string{"--version", fixtureVersion}, map[string]string{
		"SHORTCUT_INSTALL_BASE_URL": baseURL,
		"HOME":                      homeDir,
		"PNPM_HOME":                 pnpmLinkDir,
		"PATH":                      joinPath(pnpmLinkDir, requiredToolPath(t)),
	})

	installedPath := filepath.Join(homeDir, ".local", "bin", "shortcut")
	assertExecutableExists(t, installedPath)
	assertVersionOutput(t, installedPath)
	if !strings.Contains(output, "Installed to "+installedPath) {
		t.Fatalf("expected installer output to mention fallback install path %q, got:\n%s", installedPath, output)
	}
	if _, err := os.Stat(filepath.Join(pnpmRealDir, "shortcut")); err == nil {
		t.Fatalf("expected installer to skip canonicalized PNPM_HOME dir %q", pnpmRealDir)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat pnpm install path: %v", err)
	}
}

func releaseFixtureServer(t *testing.T) string {
	t.Helper()

	fixtureDir := t.TempDir()
	archiveName := fmt.Sprintf("shortcut-cli_%s_%s_%s.tar.gz", fixtureVersion, runtimeGOOS(), runtimeGOARCH())
	checksumName := fmt.Sprintf("shortcut-cli_%s_checksums.txt", fixtureVersion)
	archivePath := filepath.Join(fixtureDir, archiveName)
	checksumPath := filepath.Join(fixtureDir, checksumName)

	checksum := writeFixtureArchive(t, archivePath)
	checksumFile := fmt.Sprintf("%s  %s\n", checksum, archiveName)
	if err := os.WriteFile(checksumPath, []byte(checksumFile), 0o644); err != nil {
		t.Fatalf("write checksum file: %v", err)
	}

	server := httptest.NewServer(http.FileServer(http.Dir(fixtureDir)))
	t.Cleanup(server.Close)
	return server.URL
}

func runInstallScript(t *testing.T, args []string, extraEnv map[string]string) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	scriptPath := filepath.Clean(filepath.Join(wd, "..", "..", "install.sh"))

	cmdArgs := append([]string{scriptPath}, args...)
	cmd := exec.Command("sh", cmdArgs...)
	cmd.Env = mergeEnv(extraEnv)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("install script failed: %v\n%s", err, string(output))
	}

	return string(output)
}

func requiredToolPath(t *testing.T) string {
	t.Helper()

	tools := []string{"sh", "curl", "tar", "mktemp", "install"}
	dirs := make([]string, 0, len(tools))
	seen := map[string]bool{}
	for _, tool := range tools {
		path, err := exec.LookPath(tool)
		if err != nil {
			t.Fatalf("lookpath %s: %v", tool, err)
		}
		dir := filepath.Dir(path)
		if seen[dir] {
			continue
		}
		seen[dir] = true
		dirs = append(dirs, dir)
	}

	return strings.Join(dirs, string(os.PathListSeparator))
}

func mergeEnv(extra map[string]string) []string {
	envMap := map[string]string{}
	for _, item := range os.Environ() {
		parts := strings.SplitN(item, "=", 2)
		key := parts[0]
		value := ""
		if len(parts) == 2 {
			value = parts[1]
		}
		envMap[key] = value
	}
	for key, value := range extra {
		envMap[key] = value
	}

	env := make([]string, 0, len(envMap))
	for key, value := range envMap {
		env = append(env, key+"="+value)
	}
	return env
}

func writeFixtureArchive(t *testing.T, archivePath string) string {
	t.Helper()

	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}
	defer file.Close()

	hash := sha256.New()
	multiWriter := io.MultiWriter(file, hash)
	gzipWriter := gzip.NewWriter(multiWriter)
	tarWriter := tar.NewWriter(gzipWriter)

	content := []byte("#!/bin/sh\nif [ \"$#\" -gt 0 ] && [ \"$1\" = \"version\" ]; then\n  printf 'Version: fixture\\n'\n  exit 0\nfi\nprintf 'fixture shortcut\\n'\n")
	header := &tar.Header{
		Name: "shortcut",
		Mode: 0o755,
		Size: int64(len(content)),
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		t.Fatalf("write tar header: %v", err)
	}
	if _, err := tarWriter.Write(content); err != nil {
		t.Fatalf("write tar content: %v", err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func assertExecutableExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat installed binary %q: %v", path, err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("expected %q to be executable, mode=%v", path, info.Mode())
	}
}

func assertVersionOutput(t *testing.T, path string) {
	t.Helper()
	output, err := exec.Command(path, "version").CombinedOutput()
	if err != nil {
		t.Fatalf("run installed binary: %v\n%s", err, string(output))
	}
	if !strings.Contains(string(output), "Version: fixture") {
		t.Fatalf("unexpected version output: %s", string(output))
	}
}

func joinPath(parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		filtered = append(filtered, part)
	}
	return strings.Join(filtered, string(os.PathListSeparator))
}

func runtimeGOOS() string {
	return runtime.GOOS
}

func runtimeGOARCH() string {
	return runtime.GOARCH
}
