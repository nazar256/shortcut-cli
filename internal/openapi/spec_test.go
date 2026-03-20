package openapi

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSpec(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "spec.json")

	specContent := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Shortcut API",
			"version": "3.0"
		},
		"components": {
			"requestBodies": {
				"ExampleBody": {
					"required": true,
					"content": {
						"application/json": {
							"schema": {
								"type": "object"
							}
						}
					}
				}
			}
		},
		"paths": {
			"/api/v3/epics": {
				"get": {
					"operationId": "listEpics",
					"summary": "List Epics",
					"tags": ["Epics"],
					"parameters": [
						{
							"name": "includes_description",
							"in": "query",
							"required": false,
							"schema": {
								"type": "boolean"
							}
						}
					]
				}
			}
		}
	}`

	if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
		t.Fatalf("failed to write temp spec file: %v", err)
	}

	spec, err := ParseSpec(specPath)
	if err != nil {
		t.Fatalf("ParseSpec failed: %v", err)
	}

	if spec.OpenAPI != "3.0.0" {
		t.Errorf("expected openapi version 3.0.0, got %q", spec.OpenAPI)
	}

	if spec.Info.Title != "Shortcut API" {
		t.Errorf("expected title 'Shortcut API', got %q", spec.Info.Title)
	}

	epicsPath, ok := spec.Paths["/api/v3/epics"]
	if !ok {
		t.Fatalf("expected path /api/v3/epics to exist")
	}

	getOp, ok := epicsPath["get"]
	if !ok {
		t.Fatalf("expected get operation to exist")
	}

	if getOp.OperationID != "listEpics" {
		t.Errorf("expected operationId 'listEpics', got %q", getOp.OperationID)
	}

	if len(getOp.Parameters) != 1 {
		t.Fatalf("expected 1 parameter, got %d", len(getOp.Parameters))
	}

	param := getOp.Parameters[0]
	if param.Name != "includes_description" {
		t.Errorf("expected parameter name 'includes_description', got %q", param.Name)
	}
	if param.Schema.Type != "boolean" {
		t.Errorf("expected parameter schema type 'boolean', got %q", param.Schema.Type)
	}
	if pathList := spec.SortedPaths(); len(pathList) != 1 || pathList[0] != "/api/v3/epics" {
		t.Fatalf("unexpected sorted paths: %#v", pathList)
	}

	resolved := spec.ResolveRequestBody(&RequestBody{Ref: "#/components/requestBodies/ExampleBody"})
	if resolved == nil || !resolved.Required {
		t.Fatalf("expected resolved request body with required=true, got %#v", resolved)
	}
}
