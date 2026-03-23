package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nazar256/shortcut-cli/internal/openapi"
)

func TestCuratedResourceCommandsUseConciseNames(t *testing.T) {
	cases := []struct {
		args    []string
		present []string
		absent  []string
	}{
		{
			args:    []string{"stories", "--help"},
			present: []string{"get", "create", "query", "history", "create-comment", "update-task"},
			absent:  []string{"get-story", "create-story", "query-stories", "story-history"},
		},
		{
			args:    []string{"epics", "--help"},
			present: []string{"get", "list", "list-comments", "list-stories", "create-health"},
			absent:  []string{"get-epic", "list-epics", "list-epic-comments", "list-epic-stories"},
		},
		{
			args:    []string{"iterations", "--help"},
			present: []string{"get", "list", "stories"},
			absent:  []string{"get-iteration", "list-iterations", "list-iteration-stories"},
		},
		{
			args:    []string{"workflows", "--help"},
			present: []string{"get", "list"},
			absent:  []string{"get-workflow", "list-workflows"},
		},
	}

	for _, tc := range cases {
		output := executeResourceHelp(t, tc.args...)
		for _, want := range tc.present {
			if !strings.Contains(output, want) {
				t.Fatalf("expected %q in help output for %v:\n%s", want, tc.args, output)
			}
		}
		for _, unwanted := range tc.absent {
			if strings.Contains(output, unwanted) {
				t.Fatalf("did not expect %q in help output for %v:\n%s", unwanted, tc.args, output)
			}
		}
	}
}

func TestCuratedResourceLegacyAliasesStillWork(t *testing.T) {
	output := executeResourceHelp(t, "stories", "get-story", "--help")
	if !strings.Contains(output, "shortcut stories get 123") {
		t.Fatalf("expected canonical example in legacy alias help:\n%s", output)
	}
}

func TestCuratedResourceHelpOmitsAPIMethodAndPath(t *testing.T) {
	for _, args := range [][]string{{"stories", "get", "--help"}, {"epics", "get", "--help"}, {"iterations", "list", "--help"}, {"workflows", "get", "--help"}} {
		output := executeResourceHelp(t, args...)
		if strings.Contains(output, "Method:") || strings.Contains(output, "Path:") {
			t.Fatalf("did not expect API details in curated help for %v:\n%s", args, output)
		}
	}
}

func TestCuratedResourceOutputIsNotTransportEnvelope(t *testing.T) {
	jsonPayload, textPayload := shapeCuratedOperation(openapi.CommandMetadata{Name: "get", Group: "stories"}, map[string]any{
		"id":          123,
		"entity_type": "story",
		"name":        "Example story",
		"description": "A useful description.",
		"app_url":     "https://app.shortcut.com/story/123",
		"started":     true,
	}, curatedRenderOptions{})

	if payload, ok := jsonPayload.(map[string]any); !ok || payload["name"] != "Example story" {
		t.Fatalf("expected direct domain payload, got %#v", jsonPayload)
	}
	if strings.Contains(textPayload, "operation_id") || strings.Contains(textPayload, "resolved_path") {
		t.Fatalf("did not expect transport fields in text payload: %s", textPayload)
	}
}

func TestCuratedStoryTextIncludesFullDescription(t *testing.T) {
	description := strings.Repeat("Long story content. ", 30)
	_, textPayload := shapeCuratedOperation(openapi.CommandMetadata{Name: "get", Group: "stories"}, map[string]any{
		"id":          123,
		"entity_type": "story",
		"name":        "Example story",
		"description": description,
	}, curatedRenderOptions{})
	if !strings.Contains(textPayload, strings.TrimSpace(description)) {
		t.Fatalf("expected full description in text payload, got:\n%s", textPayload)
	}
}

func TestCuratedStoryCommentsAreHiddenByDefault(t *testing.T) {
	_, textPayload := shapeCuratedOperation(openapi.CommandMetadata{Name: "get", Group: "stories"}, map[string]any{
		"id":          123,
		"entity_type": "story",
		"name":        "Example story",
		"comments": []any{
			map[string]any{"text": "First comment"},
		},
	}, curatedRenderOptions{})

	if strings.Contains(textPayload, "Comments:") || strings.Contains(textPayload, "First comment") {
		t.Fatalf("expected comments to stay hidden by default, got:\n%s", textPayload)
	}
}

func TestCuratedStoryCommentsRenderWhenRequested(t *testing.T) {
	_, textPayload := shapeCuratedOperation(openapi.CommandMetadata{Name: "get", Group: "stories"}, map[string]any{
		"id":          123,
		"entity_type": "story",
		"name":        "Example story",
		"comments": []any{
			map[string]any{"text": "First comment"},
			map[string]any{"text": "Second comment"},
		},
	}, curatedRenderOptions{withComments: true})

	if !strings.Contains(textPayload, "Comments:") || !strings.Contains(textPayload, "1. First comment") || !strings.Contains(textPayload, "2. Second comment") {
		t.Fatalf("expected rendered comments when requested, got:\n%s", textPayload)
	}
}

func TestCuratedStoryWithCommentsDoesNotShowZeroCount(t *testing.T) {
	_, textPayload := shapeCuratedOperation(openapi.CommandMetadata{Name: "get", Group: "stories"}, map[string]any{
		"id":          123,
		"entity_type": "story",
		"name":        "Example story",
		"comments":    []any{},
	}, curatedRenderOptions{withComments: true})

	if strings.Contains(textPayload, "Comments: 0") {
		t.Fatalf("did not expect zero-count comments line, got:\n%s", textPayload)
	}
}

func TestCuratedMeTextSummaryIsReadable(t *testing.T) {
	_, textPayload := shapeCuratedOperation(openapi.CommandMetadata{Name: "me", Group: "member"}, map[string]any{
		"id":            float64(123456789012345),
		"entity_type":   "member",
		"name":          "Example Person",
		"mention_name":  "example-user",
		"email_address": "example@example.invalid",
		"role":          "member",
	}, curatedRenderOptions{})

	for _, want := range []string{
		"Member #123456789012345 Example Person",
		"Role: member",
		"Mention: @example-user",
		"Email: example@example.invalid",
	} {
		if !strings.Contains(textPayload, want) {
			t.Fatalf("expected %q in readable member summary, got:\n%s", want, textPayload)
		}
	}
}

func TestCuratedStoryStatePrefersCompletedOverStarted(t *testing.T) {
	_, textPayload := shapeCuratedOperation(openapi.CommandMetadata{Name: "get", Group: "stories"}, map[string]any{
		"id":          123,
		"entity_type": "story",
		"name":        "Done story",
		"started":     true,
		"completed":   true,
	}, curatedRenderOptions{})

	if !strings.Contains(textPayload, "State: completed") {
		t.Fatalf("expected completed state to win, got:\n%s", textPayload)
	}
	if strings.Contains(textPayload, "State: started") {
		t.Fatalf("did not expect started state when completed is true, got:\n%s", textPayload)
	}
}

func TestCuratedStoryHistorySummaryIsReadable(t *testing.T) {
	_, textPayload := shapeCuratedOperation(openapi.CommandMetadata{Name: "history", Group: "stories"}, map[string]any{
		"data": []any{
			map[string]any{
				"actor_name": "Example Person",
				"changed_at": "2026-03-20T10:11:12Z",
				"actions": []any{
					map[string]any{"action": "update", "entity_type": "story", "name": "Example story"},
				},
			},
		},
		"total": float64(1),
	}, curatedRenderOptions{})

	for _, want := range []string{"Example Person", "2026-03-20T10:11:12Z", "update story Example story"} {
		if !strings.Contains(textPayload, want) {
			t.Fatalf("expected %q in history summary, got:\n%s", want, textPayload)
		}
	}
}

func TestCuratedCollectionUsesPlainIntegerIDs(t *testing.T) {
	_, textPayload := shapeCuratedOperation(openapi.CommandMetadata{Name: "list", Group: "stories"}, map[string]any{
		"data": []any{
			map[string]any{
				"id":          float64(123456789012345),
				"entity_type": "story",
				"name":        "Example story",
			},
		},
		"total": float64(1),
	}, curatedRenderOptions{})

	if !strings.Contains(textPayload, "#123456789012345") {
		t.Fatalf("expected plain integer ID in collection summary, got:\n%s", textPayload)
	}
	if strings.Contains(textPayload, "e+") {
		t.Fatalf("did not expect scientific notation in collection summary, got:\n%s", textPayload)
	}
}

func executeResourceHelp(t *testing.T, args ...string) string {
	t.Helper()
	cmd := NewRootCmd()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("help command failed for %v: %v", args, err)
	}
	return stdout.String()
}
