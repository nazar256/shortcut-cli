package cli

import (
	"testing"

	"github.com/nazar256/shortcut-cli/internal/openapi"
)

func TestBuildOperationCmdAddsQueryAndBodyFlags(t *testing.T) {
	meta := openapi.CommandMetadata{
		OperationID: "createStory",
		Name:        "create-story",
		Method:      "POST",
		Path:        "/api/v3/stories",
		Summary:     "Create Story",
		Description: "Creates a story.",
		Group:       "stories",
		Parameters: []openapi.CommandParameter{
			{Name: "project-id", In: "query", Required: true, Type: "integer"},
		},
		HasBody: true,
	}

	cmd := buildOperationCmd(meta)
	if cmd.Flags().Lookup("project-id") == nil {
		t.Fatal("expected project-id flag")
	}
	if cmd.Flags().Lookup("body") == nil {
		t.Fatal("expected body flag")
	}
	if cmd.Flags().Lookup("body-file") == nil {
		t.Fatal("expected body-file flag")
	}
}

func TestBuildOperationCmdAddsMultipartFlags(t *testing.T) {
	meta := openapi.CommandMetadata{
		OperationID: "uploadFiles",
		Name:        "upload-files",
		Method:      "POST",
		Path:        "/api/v3/files",
		Summary:     "Upload Files",
		Group:       "files",
		HasBody:     true,
		IsMultipart: true,
	}

	cmd := buildOperationCmd(meta)
	if cmd.Flags().Lookup("form") == nil {
		t.Fatal("expected form flag")
	}
	if cmd.Flags().Lookup("file") == nil {
		t.Fatal("expected file flag")
	}
}

func TestBuildOperationCmdUseIncludesPathArgs(t *testing.T) {
	meta := openapi.CommandMetadata{
		OperationID: "getEpic",
		Name:        "get-epic",
		Method:      "GET",
		Path:        "/api/v3/epics/{epic-public-id}",
		Summary:     "Get Epic",
		Group:       "epics",
		Parameters: []openapi.CommandParameter{
			{Name: "epic-public-id", In: "path", Required: true, Type: "integer"},
		},
	}

	cmd := buildOperationCmd(meta)
	if got, want := cmd.Use, "get-epic <epic-public-id>"; got != want {
		t.Fatalf("unexpected use string: got %q want %q", got, want)
	}
}
