package cli

import (
	"bytes"
	"strings"
	"testing"
)

func executeHelp(t *testing.T, args ...string) string {
	t.Helper()
	cmd := NewRootCmd()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("help command failed: %v", err)
	}

	return stdout.String()
}

func TestRootHelpDoesNotUseAIFriendlyLabel(t *testing.T) {
	output := executeHelp(t, "--help")
	if strings.Contains(output, "AI-friendly") {
		t.Fatalf("unexpected AI-friendly wording in help:\n%s", output)
	}
}

func TestRootHelpIncludesEnvFileFlags(t *testing.T) {
	output := executeHelp(t, "--help")
	if !strings.Contains(output, "--env-file") {
		t.Fatalf("expected root help to include --env-file flag:\n%s", output)
	}
	if !strings.Contains(output, "--no-env-file") {
		t.Fatalf("expected root help to include --no-env-file flag:\n%s", output)
	}
}

func TestSearchHelpExplainsSearchFlow(t *testing.T) {
	output := executeHelp(t, "search", "--help")
	if !strings.Contains(output, "shortcut search syntax") {
		t.Fatalf("expected search help to mention syntax command:\n%s", output)
	}
	if !strings.Contains(output, "Pick a scope") {
		t.Fatalf("expected search help to explain scope selection:\n%s", output)
	}
}

func TestSearchHelpCommandWorks(t *testing.T) {
	output := executeHelp(t, "search", "help")
	if !strings.Contains(output, "shortcut search syntax") {
		t.Fatalf("expected search help command to show curated help:\n%s", output)
	}
}

func TestSearchStoriesHelpIncludesExamples(t *testing.T) {
	output := executeHelp(t, "search", "stories", "--help")
	if !strings.Contains(output, "Examples:") {
		t.Fatalf("expected examples section:\n%s", output)
	}
	if !strings.Contains(output, "shortcut search stories") {
		t.Fatalf("expected top-level search stories examples:\n%s", output)
	}
	if !strings.Contains(output, "shortcut search syntax") {
		t.Fatalf("expected syntax guidance:\n%s", output)
	}
}

func TestLegacySearchAliasStillWorks(t *testing.T) {
	output := executeHelp(t, "search", "search-stories", "--help")
	if !strings.Contains(output, "shortcut search stories") {
		t.Fatalf("expected alias help to render canonical examples:\n%s", output)
	}
}

func TestSearchDocumentsHelpUsesStructuredFlags(t *testing.T) {
	output := executeHelp(t, "search", "documents", "--help")
	if strings.Contains(output, "sent to the API as the `query` parameter") {
		t.Fatalf("documents help should not describe query parameter flow:\n%s", output)
	}
	if !strings.Contains(output, "--title") {
		t.Fatalf("expected documents help to show structured flags:\n%s", output)
	}
}
