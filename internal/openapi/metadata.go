package openapi

import (
	"sort"
	"strings"
)

// CommandMetadata represents the derived metadata for a CLI command.
type CommandMetadata struct {
	OperationID         string
	Name                string
	Method              string
	Path                string
	Summary             string
	Description         string
	Group               string // e.g., "epics", "stories"
	Parameters          []CommandParameter
	HasBody             bool
	BodyRequired        bool
	IsMultipart         bool
	RequestContentTypes []string
}

// CommandParameter represents a parameter for a CLI command.
type CommandParameter struct {
	Name        string
	In          string // "query", "path", "header"
	Description string
	Required    bool
	Type        string
	Format      string
	IsArray     bool
	Enum        []string
}

// DeriveCommands extracts and normalizes command metadata from the OpenAPI spec.
func DeriveCommands(spec *Spec) ([]CommandMetadata, error) {
	var commands []CommandMetadata

	for _, path := range spec.SortedPaths() {
		methods := spec.Paths[path]
		for method, op := range methods {
			// Skip non-standard methods if any
			method = strings.ToUpper(method)
			if method != "GET" && method != "POST" && method != "PUT" && method != "DELETE" && method != "PATCH" {
				continue
			}

			group := deriveGroup(path, op.Tags)
			name := deriveCommandName(method, path, op.OperationID)

			var params []CommandParameter
			for _, p := range op.Parameters {
				isArray := p.Schema.Type == "array" && p.Schema.Items != nil
				paramType := p.Schema.Type
				if isArray {
					paramType = p.Schema.Items.Type
				}
				params = append(params, CommandParameter{
					Name:        p.Name,
					In:          p.In,
					Description: p.Description,
					Required:    p.Required,
					Type:        paramType,
					Format:      p.Schema.Format,
					IsArray:     isArray,
					Enum:        p.Schema.Enum,
				})
			}

			var hasBody, bodyRequired, isMultipart bool
			requestContentTypes := []string{}
			resolvedBody := spec.ResolveRequestBody(op.RequestBody)
			if resolvedBody != nil {
				hasBody = true
				bodyRequired = resolvedBody.Required
				for contentType := range resolvedBody.Content {
					requestContentTypes = append(requestContentTypes, contentType)
				}
				sort.Strings(requestContentTypes)
				if _, ok := resolvedBody.Content["multipart/form-data"]; ok {
					isMultipart = true
				}
			}

			cmd := CommandMetadata{
				OperationID:         op.OperationID,
				Name:                name,
				Method:              method,
				Path:                path,
				Summary:             op.Summary,
				Description:         op.Description,
				Group:               group,
				Parameters:          params,
				HasBody:             hasBody,
				BodyRequired:        bodyRequired,
				IsMultipart:         isMultipart,
				RequestContentTypes: requestContentTypes,
			}
			commands = append(commands, cmd)
		}
	}

	// Sort commands for stable output
	sort.Slice(commands, func(i, j int) bool {
		if commands[i].Group != commands[j].Group {
			return commands[i].Group < commands[j].Group
		}
		return commands[i].Name < commands[j].Name
	})

	return commands, nil
}

func (m CommandMetadata) PathParameters() []CommandParameter {
	params := make([]CommandParameter, 0)
	for _, param := range m.Parameters {
		if param.In == "path" {
			params = append(params, param)
		}
	}
	return params
}

func (m CommandMetadata) QueryParameters() []CommandParameter {
	params := make([]CommandParameter, 0)
	for _, param := range m.Parameters {
		if param.In == "query" {
			params = append(params, param)
		}
	}
	return params
}

// deriveGroup attempts to find a logical grouping for the command.
// It prefers tags, but falls back to the first significant path segment.
func deriveGroup(path string, tags []string) string {
	if len(tags) > 0 && tags[0] != "" {
		// Normalize tag to lowercase, replace spaces with dashes
		return strings.ToLower(strings.ReplaceAll(tags[0], " ", "-"))
	}

	// Fallback to path parsing
	// e.g., /api/v3/epics/{epic-public-id} -> epics
	segments := strings.Split(strings.TrimPrefix(path, "/"), "/")
	for _, seg := range segments {
		if seg == "api" || strings.HasPrefix(seg, "v") {
			continue // skip /api/v3
		}
		if !strings.HasPrefix(seg, "{") {
			return seg
		}
	}
	return "general"
}

// deriveCommandName generates a stable CLI command name.
func deriveCommandName(method, path, operationID string) string {
	if operationID != "" {
		// Convert camelCase to kebab-case
		var kebab strings.Builder
		for i, r := range operationID {
			if r >= 'A' && r <= 'Z' {
				if i > 0 {
					kebab.WriteRune('-')
				}
				kebab.WriteRune(r + 32)
			} else {
				kebab.WriteRune(r)
			}
		}
		return kebab.String()
	}

	// Fallback if no operationId
	segments := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(segments) >= 2 && segments[0] == "api" && strings.HasPrefix(segments[1], "v") {
		segments = segments[2:]
	}

	if len(segments) == 0 {
		return strings.ToLower(method)
	}

	lastSegment := segments[len(segments)-1]
	isInstance := strings.HasPrefix(lastSegment, "{") && strings.HasSuffix(lastSegment, "}")

	var suffix string
	if len(segments) > 2 {
		if isInstance {
			subResource := segments[len(segments)-2]
			if !strings.HasPrefix(subResource, "{") {
				suffix = "-" + strings.TrimSuffix(subResource, "s")
			}
		} else {
			if method == "POST" {
				suffix = "-" + strings.TrimSuffix(lastSegment, "s")
			} else {
				suffix = "-" + lastSegment
			}
		}
	}

	switch method {
	case "GET":
		if isInstance {
			return "get" + suffix
		}
		return "list" + suffix
	case "POST":
		return "create" + suffix
	case "PUT", "PATCH":
		return "update" + suffix
	case "DELETE":
		return "delete" + suffix
	}

	return strings.ToLower(method)
}
