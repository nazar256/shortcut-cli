package openapi

import (
	"testing"
)

func TestDeriveGroup(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		tags     []string
		expected string
	}{
		{
			name:     "from tags",
			path:     "/api/v3/epics",
			tags:     []string{"Epics"},
			expected: "epics",
		},
		{
			name:     "from tags with spaces",
			path:     "/api/v3/story-links",
			tags:     []string{"Story Links"},
			expected: "story-links",
		},
		{
			name:     "fallback to path",
			path:     "/api/v3/epics",
			tags:     nil,
			expected: "epics",
		},
		{
			name:     "fallback to path with instance",
			path:     "/api/v3/epics/{epic-public-id}",
			tags:     []string{},
			expected: "epics",
		},
		{
			name:     "fallback to path with subresource",
			path:     "/api/v3/epics/{epic-public-id}/comments",
			tags:     nil,
			expected: "epics",
		},
		{
			name:     "no valid segments",
			path:     "/api/v3",
			tags:     nil,
			expected: "general",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := deriveGroup(tt.path, tt.tags)
			if actual != tt.expected {
				t.Errorf("deriveGroup(%q, %v) = %q, expected %q", tt.path, tt.tags, actual, tt.expected)
			}
		})
	}
}

func TestDeriveCommandName(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		operationID string
		expected    string
	}{
		{
			name:        "list epics",
			method:      "GET",
			path:        "/api/v3/epics",
			operationID: "listEpics",
			expected:    "list-epics",
		},
		{
			name:        "get epic",
			method:      "GET",
			path:        "/api/v3/epics/{epic-public-id}",
			operationID: "getEpic",
			expected:    "get-epic",
		},
		{
			name:        "create epic",
			method:      "POST",
			path:        "/api/v3/epics",
			operationID: "createEpic",
			expected:    "create-epic",
		},
		{
			name:        "update epic",
			method:      "PUT",
			path:        "/api/v3/epics/{epic-public-id}",
			operationID: "updateEpic",
			expected:    "update-epic",
		},
		{
			name:        "delete epic",
			method:      "DELETE",
			path:        "/api/v3/epics/{epic-public-id}",
			operationID: "deleteEpic",
			expected:    "delete-epic",
		},
		{
			name:        "list epic comments",
			method:      "GET",
			path:        "/api/v3/epics/{epic-public-id}/comments",
			operationID: "listEpicComments",
			expected:    "list-epic-comments",
		},
		{
			name:        "create epic comment",
			method:      "POST",
			path:        "/api/v3/epics/{epic-public-id}/comments",
			operationID: "createEpicComment",
			expected:    "create-epic-comment",
		},
		{
			name:        "get epic comment",
			method:      "GET",
			path:        "/api/v3/epics/{epic-public-id}/comments/{comment-public-id}",
			operationID: "getEpicComment",
			expected:    "get-epic-comment",
		},
		{
			name:        "update epic comment",
			method:      "PUT",
			path:        "/api/v3/epics/{epic-public-id}/comments/{comment-public-id}",
			operationID: "updateEpicComment",
			expected:    "update-epic-comment",
		},
		{
			name:        "delete epic comment",
			method:      "DELETE",
			path:        "/api/v3/epics/{epic-public-id}/comments/{comment-public-id}",
			operationID: "deleteEpicComment",
			expected:    "delete-epic-comment",
		},
		{
			name:        "fallback to operation ID",
			method:      "OPTIONS",
			path:        "/api/v3/epics",
			operationID: "optionsEpics",
			expected:    "options-epics",
		},
		{
			name:        "fallback to path list",
			method:      "GET",
			path:        "/api/v3/epics",
			operationID: "",
			expected:    "list",
		},
		{
			name:        "fallback to path get",
			method:      "GET",
			path:        "/api/v3/epics/{epic-public-id}",
			operationID: "",
			expected:    "get",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := deriveCommandName(tt.method, tt.path, tt.operationID)
			if actual != tt.expected {
				t.Errorf("deriveCommandName(%q, %q, %q) = %q, expected %q", tt.method, tt.path, tt.operationID, actual, tt.expected)
			}
		})
	}
}

func TestDeriveCommands(t *testing.T) {
	spec := &Spec{
		Paths: map[string]map[string]Operation{
			"/api/v3/epics": {
				"get": Operation{
					OperationID: "listEpics",
					Summary:     "List Epics",
					Tags:        []string{"Epics"},
				},
				"post": Operation{
					OperationID: "createEpic",
					Summary:     "Create Epic",
					Tags:        []string{"Epics"},
				},
			},
			"/api/v3/epics/{epic-public-id}": {
				"get": Operation{
					OperationID: "getEpic",
					Summary:     "Get Epic",
					Tags:        []string{"Epics"},
					Parameters: []Parameter{
						{
							Name: "epic-public-id",
							In:   "path",
						},
					},
				},
			},
			"/api/v3/stories": {
				"get": Operation{
					OperationID: "listStories",
					Summary:     "List Stories",
					Tags:        []string{"Stories"},
				},
				"post": Operation{
					OperationID: "createStory",
					Summary:     "Create Story",
					Tags:        []string{"Stories"},
					RequestBody: &RequestBody{Ref: "#/components/requestBodies/CreateStoryBody"},
				},
			},
		},
		Components: Components{
			RequestBodies: map[string]RequestBody{
				"CreateStoryBody": {
					Required: true,
					Content: map[string]MediaType{
						"application/json": {Schema: Schema{Type: "object"}},
					},
				},
			},
		},
	}

	commands, err := DeriveCommands(spec)
	if err != nil {
		t.Fatalf("DeriveCommands failed: %v", err)
	}

	if len(commands) != 5 {
		t.Fatalf("expected 5 commands, got %d", len(commands))
	}

	// Check sorting (epics first, then stories; within epics: create, get, list)
	expectedNames := []string{"create-epic", "get-epic", "list-epics", "create-story", "list-stories"}
	expectedGroups := []string{"epics", "epics", "epics", "stories", "stories"}

	for i, cmd := range commands {
		if cmd.Name != expectedNames[i] {
			t.Errorf("commands[%d].Name = %q, expected %q", i, cmd.Name, expectedNames[i])
		}
		if cmd.Group != expectedGroups[i] {
			t.Errorf("commands[%d].Group = %q, expected %q", i, cmd.Group, expectedGroups[i])
		}
	}

	// Check parameter mapping
	getEpicCmd := commands[1]
	if len(getEpicCmd.Parameters) != 1 {
		t.Fatalf("expected 1 parameter for get epic, got %d", len(getEpicCmd.Parameters))
	}
	if getEpicCmd.Parameters[0].Name != "epic-public-id" {
		t.Errorf("expected parameter name 'epic-public-id', got %q", getEpicCmd.Parameters[0].Name)
	}

	createStoryCmd := commands[3]
	if !createStoryCmd.HasBody || !createStoryCmd.BodyRequired {
		t.Fatalf("expected create-story to require request body, got %#v", createStoryCmd)
	}
}
