package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestDocsSummaryWithoutToken(t *testing.T) {
	t.Setenv("SHORTCUT_API_TOKEN", "")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := NewRootCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"docs", "summary"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("docs summary should work without token: %v", err)
	}
}

func TestVersionWithoutToken(t *testing.T) {
	t.Setenv("SHORTCUT_API_TOKEN", "")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := NewRootCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("version should work without token: %v", err)
	}
}

func TestVersionTextIncludesBuildMetadataWhenPresent(t *testing.T) {
	originalVersion := Version
	originalCommit := Commit
	originalBuildDate := BuildDate
	Version = "v1.0.0"
	Commit = "abc1234"
	BuildDate = "2026-03-20T12:34:56Z"
	defer func() {
		Version = originalVersion
		Commit = originalCommit
		BuildDate = originalBuildDate
	}()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := NewRootCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("version should succeed: %v", err)
	}

	output := stdout.String()
	for _, want := range []string{"Version: v1.0.0", "Commit: abc1234", "Built: 2026-03-20T12:34:56Z"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in version output, got:\n%s", want, output)
		}
	}
}

func TestVersionJSONIncludesBuildMetadataFields(t *testing.T) {
	originalVersion := Version
	originalCommit := Commit
	originalBuildDate := BuildDate
	Version = "v1.0.0"
	Commit = "abc1234"
	BuildDate = "2026-03-20T12:34:56Z"
	defer func() {
		Version = originalVersion
		Commit = originalCommit
		BuildDate = originalBuildDate
	}()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := NewRootCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"version", "--output", "json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("version json should succeed: %v", err)
	}

	output := stdout.String()
	for _, want := range []string{"\"version\": \"v1.0.0\"", "\"commit\": \"abc1234\"", "\"build_date\": \"2026-03-20T12:34:56Z\""} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in version json output, got:\n%s", want, output)
		}
	}
}
