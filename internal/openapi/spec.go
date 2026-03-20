package openapi

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// Spec represents the root of the OpenAPI specification.
type Spec struct {
	OpenAPI    string                          `json:"openapi"`
	Info       Info                            `json:"info"`
	Paths      map[string]map[string]Operation `json:"paths"`
	Components Components                      `json:"components,omitempty"`
}

type Components struct {
	RequestBodies map[string]RequestBody `json:"requestBodies,omitempty"`
}

// Info provides metadata about the API.
type Info struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

// Operation describes a single API operation on a path.
type Operation struct {
	OperationID string                 `json:"operationId"`
	Summary     string                 `json:"summary"`
	Description string                 `json:"description"`
	Tags        []string               `json:"tags"`
	Parameters  []Parameter            `json:"parameters"`
	RequestBody *RequestBody           `json:"requestBody,omitempty"`
	Responses   map[string]APIResponse `json:"responses,omitempty"`
}

// RequestBody describes the body of a request.
type RequestBody struct {
	Ref         string               `json:"$ref,omitempty"`
	Description string               `json:"description"`
	Required    bool                 `json:"required"`
	Content     map[string]MediaType `json:"content"`
}

// MediaType describes the media type of a request body.
type MediaType struct {
	Schema Schema `json:"schema"`
}

// APIResponse describes an operation response.
type APIResponse struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// Parameter describes a single operation parameter.
type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Schema      Schema `json:"schema"`
}

// Schema describes the data type of a parameter.
type Schema struct {
	Type   string   `json:"type"`
	Format string   `json:"format,omitempty"`
	Enum   []string `json:"enum,omitempty"`
	Items  *Schema  `json:"items,omitempty"`
}

// ParseSpec parses an OpenAPI specification from a file.
func ParseSpec(filePath string) (*Spec, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open spec file: %w", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	return ParseSpecBytes(bytes)
}

// ParseSpecBytes parses an OpenAPI specification from bytes.
func ParseSpecBytes(data []byte) (*Spec, error) {
	var spec Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec: %w", err)
	}

	return &spec, nil
}

// SortedPaths returns a stable sorted list of path keys.
func (s *Spec) SortedPaths() []string {
	paths := make([]string, 0, len(s.Paths))
	for path := range s.Paths {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func (s *Spec) ResolveRequestBody(body *RequestBody) *RequestBody {
	if body == nil {
		return nil
	}
	if body.Ref == "" {
		return body
	}

	const prefix = "#/components/requestBodies/"
	if !strings.HasPrefix(body.Ref, prefix) {
		return body
	}

	name := strings.TrimPrefix(body.Ref, prefix)
	resolved, ok := s.Components.RequestBodies[name]
	if !ok {
		return body
	}

	copy := resolved
	if body.Description != "" {
		copy.Description = body.Description
	}
	if body.Required {
		copy.Required = true
	}
	return &copy
}
