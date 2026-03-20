package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	shortcutopenapi "github.com/nazar256/shortcut-cli/internal/openapi"
	shortcutruntime "github.com/nazar256/shortcut-cli/internal/shortcut"
	shortcutspec "github.com/nazar256/shortcut-cli/openapi"
	"github.com/spf13/cobra"
)

func NewAPICmd() *cobra.Command {
	spec, err := shortcutopenapi.ParseSpecBytes(shortcutspec.SpecBytes)
	apiCmd := &cobra.Command{
		Use:     "api",
		Short:   "Direct access to all Shortcut API operations",
		Long:    "The api command exposes the full official Shortcut REST API. Commands are grouped by resource and generated from the vendored OpenAPI spec.",
		Example: "  shortcut api stories get-story 123\n  shortcut api epics list-epics",
	}

	if err != nil {
		apiCmd.RunE = func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("load OpenAPI spec: %w", err)
		}
		return apiCmd
	}

	commands, err := shortcutopenapi.DeriveCommands(spec)
	if err != nil {
		apiCmd.RunE = func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("derive command metadata: %w", err)
		}
		return apiCmd
	}

	groupCommands := map[string]*cobra.Command{}
	orderedGroups := make([]string, 0)

	for _, meta := range commands {
		groupCmd, exists := groupCommands[meta.Group]
		if !exists {
			groupCmd = &cobra.Command{
				Use:   meta.Group,
				Short: fmt.Sprintf("Operations for %s", meta.Group),
			}
			groupCommands[meta.Group] = groupCmd
			orderedGroups = append(orderedGroups, meta.Group)
		}

		groupCmd.AddCommand(buildOperationCmd(meta))
	}

	sort.Strings(orderedGroups)
	for _, group := range orderedGroups {
		apiCmd.AddCommand(groupCommands[group])
	}

	return apiCmd
}

func buildOperationCmd(meta shortcutopenapi.CommandMetadata) *cobra.Command {
	return buildOperationCmdForPrefix(meta, "shortcut api "+meta.Group, true)
}

func buildOperationCmdForPrefix(meta shortcutopenapi.CommandMetadata, prefix string, includeAPIDetails bool) *cobra.Command {
	use := meta.Name
	for _, param := range meta.PathParameters() {
		use += " <" + param.Name + ">"
	}

	cmd := &cobra.Command{
		Use:     use,
		Short:   firstNonEmpty(meta.Summary, meta.OperationID, meta.Name),
		Long:    buildOperationLongDescription(meta, includeAPIDetails),
		Args:    cobra.ExactArgs(len(meta.PathParameters())),
		Example: buildOperationExamples(meta, prefix),
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, err := newRuntime(cmd)
			if err != nil {
				return err
			}

			if includeAPIDetails {
				payload, err := executeOperation(runtime, cmd, meta, args)
				if err != nil {
					return err
				}
				return runtime.Formatter.Print(payload)
			}

			result, err := invokeOperation(runtime, cmd, meta, args)
			if err != nil {
				return err
			}

			return printCuratedOperation(runtime.Formatter, meta, applyCuratedLimit(cmd, result.Response), curatedRenderOptionsFromCmd(cmd, meta))
		},
	}

	for _, param := range meta.QueryParameters() {
		addParameterFlag(cmd, param)
	}

	if !includeAPIDetails && strings.HasPrefix(meta.Name, "list") {
		cmd.Flags().Int("limit", 20, "Maximum number of results to print; use 0 for all")
	}
	if !includeAPIDetails && (meta.Group == "stories" || meta.Group == "epics") && meta.Name == "get" {
		cmd.Flags().Bool("with-comments", false, "Include comments in text output")
	}

	if meta.HasBody {
		if meta.IsMultipart {
			cmd.Flags().StringSlice("form", nil, "Multipart form fields in key=value form")
			cmd.Flags().StringSlice("file", nil, "Multipart file fields in key=path form")
		} else {
			cmd.Flags().String("body", "", "Inline JSON request body")
			cmd.Flags().String("body-file", "", "Read JSON request body from file")
		}
	}

	return cmd
}

func addParameterFlag(cmd *cobra.Command, param shortcutopenapi.CommandParameter) {
	description := strings.TrimSpace(param.Description)
	if description == "" {
		description = "Query parameter"
	}
	if len(param.Enum) > 0 {
		description += " Allowed: " + strings.Join(param.Enum, ", ") + "."
	}

	if param.IsArray {
		cmd.Flags().StringSlice(param.Name, nil, description)
	} else {
		cmd.Flags().String(param.Name, "", description)
	}

	if param.Required {
		_ = cmd.MarkFlagRequired(param.Name)
	}
}

type requestDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type operationResult struct {
	Meta               shortcutopenapi.CommandMetadata
	Path               string
	PathParameters     map[string]string
	QueryParameters    map[string]any
	RequestBody        map[string]any
	RequestContentType string
	Status             int
	Response           any
}

func executeOperation(runtime *shortcutruntime.Runtime, cmd *cobra.Command, meta shortcutopenapi.CommandMetadata, args []string) (map[string]any, error) {
	result, err := invokeOperation(runtime, cmd, meta, args)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"operation_id":         result.Meta.OperationID,
		"method":               result.Meta.Method,
		"path":                 result.Meta.Path,
		"resolved_path":        result.Path,
		"path_parameters":      result.PathParameters,
		"query_parameters":     result.QueryParameters,
		"request_body":         result.RequestBody,
		"request_content_type": result.RequestContentType,
		"status":               result.Status,
		"response":             result.Response,
	}, nil
}

func invokeOperation(runtime *shortcutruntime.Runtime, cmd *cobra.Command, meta shortcutopenapi.CommandMetadata, args []string) (*operationResult, error) {
	pathValues, err := pathParameterValues(meta, args)
	if err != nil {
		return nil, err
	}

	queryValues := map[string]any{}
	for _, param := range meta.QueryParameters() {
		if param.IsArray {
			values, _ := cmd.Flags().GetStringSlice(param.Name)
			if len(values) > 0 {
				parsed := make([]any, 0, len(values))
				for _, value := range values {
					parsed = append(parsed, parseValue(value, param))
				}
				queryValues[param.Name] = parsed
			}
			continue
		}

		value, _ := cmd.Flags().GetString(param.Name)
		if value != "" {
			queryValues[param.Name] = parseValue(value, param)
		}
	}

	requestBody, contentType, bodySummary, err := collectRequestBody(cmd, meta)
	if err != nil {
		return nil, err
	}

	resolvedPath := meta.Path
	for key, value := range pathValues {
		resolvedPath = strings.ReplaceAll(resolvedPath, "{"+key+"}", value)
	}

	requestURL, err := withQuery(strings.TrimRight(runtime.Config.BaseURL, "/")+resolvedPath, queryValues)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(commandContext(cmd), meta.Method, requestURL, requestBody)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	response, err := runtime.GetHTTP().Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	decodedBody := decodeResponseBody(responseBytes)
	if err := EnsureHTTPSuccess(response, decodedBody); err != nil {
		return nil, err
	}

	return &operationResult{
		Meta:               meta,
		Path:               resolvedPath,
		PathParameters:     pathValues,
		QueryParameters:    queryValues,
		RequestBody:        bodySummary,
		RequestContentType: contentType,
		Status:             response.StatusCode,
		Response:           decodedBody,
	}, nil
}

func collectRequestBody(cmd *cobra.Command, meta shortcutopenapi.CommandMetadata) (io.Reader, string, map[string]any, error) {
	if !meta.HasBody {
		return nil, "", map[string]any{"type": "none"}, nil
	}

	if meta.IsMultipart {
		forms, _ := cmd.Flags().GetStringSlice("form")
		files, _ := cmd.Flags().GetStringSlice("file")
		reader, contentType, err := buildMultipartBody(forms, files)
		if err != nil {
			return nil, "", nil, err
		}
		return reader, contentType, map[string]any{"type": "multipart/form-data", "form": forms, "files": files}, nil
	}

	bodyInline, _ := cmd.Flags().GetString("body")
	bodyFile, _ := cmd.Flags().GetString("body-file")
	if bodyInline != "" && bodyFile != "" {
		return nil, "", nil, fmt.Errorf("use only one of --body or --body-file")
	}

	if bodyInline == "" && bodyFile == "" {
		if meta.BodyRequired {
			return nil, "", nil, fmt.Errorf("request body is required")
		}
		return nil, "", map[string]any{"type": "none"}, nil
	}

	raw := bodyInline
	if bodyFile != "" {
		bytes, err := os.ReadFile(bodyFile)
		if err != nil {
			return nil, "", nil, fmt.Errorf("read body file: %w", err)
		}
		raw = string(bytes)
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return bytes.NewBufferString(trimmed), "application/json", map[string]any{"type": "empty"}, nil
	}

	var decoded any
	if err := json.Unmarshal([]byte(trimmed), &decoded); err != nil {
		return nil, "", nil, fmt.Errorf("parse JSON body: %w", err)
	}

	return bytes.NewBufferString(trimmed), "application/json", map[string]any{"type": "application/json", "payload": decoded}, nil
}

func collectBodySummary(cmd *cobra.Command, meta shortcutopenapi.CommandMetadata) (map[string]any, error) {
	if meta.IsMultipart {
		forms, _ := cmd.Flags().GetStringSlice("form")
		files, _ := cmd.Flags().GetStringSlice("file")
		return map[string]any{
			"type":  "multipart/form-data",
			"form":  forms,
			"files": files,
		}, nil
	}

	bodyInline, _ := cmd.Flags().GetString("body")
	bodyFile, _ := cmd.Flags().GetString("body-file")
	if bodyInline != "" && bodyFile != "" {
		return nil, fmt.Errorf("use only one of --body or --body-file")
	}

	if bodyInline == "" && bodyFile == "" {
		if meta.BodyRequired {
			return nil, fmt.Errorf("request body is required")
		}
		return map[string]any{"type": "none"}, nil
	}

	raw := bodyInline
	if bodyFile != "" {
		bytes, err := os.ReadFile(bodyFile)
		if err != nil {
			return nil, fmt.Errorf("read body file: %w", err)
		}
		raw = string(bytes)
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return map[string]any{"type": "empty"}, nil
	}

	var decoded any
	if err := json.Unmarshal([]byte(trimmed), &decoded); err != nil {
		return nil, fmt.Errorf("parse JSON body: %w", err)
	}

	return map[string]any{
		"type":    "application/json",
		"payload": decoded,
	}, nil
}

func buildOperationLongDescription(meta shortcutopenapi.CommandMetadata, includeAPIDetails bool) string {
	parts := []string{}
	if meta.Description != "" {
		parts = append(parts, meta.Description)
	}
	if includeAPIDetails {
		parts = append(parts, fmt.Sprintf("Method: %s", meta.Method))
		parts = append(parts, fmt.Sprintf("Path: %s", meta.Path))
		if len(meta.RequestContentTypes) > 0 {
			parts = append(parts, "Accepts body content types: "+strings.Join(meta.RequestContentTypes, ", "))
		}
		return strings.Join(parts, "\n\n")
	}
	if meta.HasBody {
		if meta.IsMultipart {
			parts = append(parts, "Use --form for fields and --file for attachments.")
		} else {
			parts = append(parts, "Use --body or --body-file to provide request data.")
		}
	}
	return strings.Join(parts, "\n\n")
}

func applyCuratedLimit(cmd *cobra.Command, response any) any {
	flag := cmd.Flags().Lookup("limit")
	if flag == nil {
		return response
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil || limit <= 0 {
		return response
	}
	payload, ok := response.(map[string]any)
	if !ok {
		return response
	}
	items, ok := payload["data"].([]any)
	if !ok || len(items) <= limit {
		return response
	}
	cloned := map[string]any{}
	for key, value := range payload {
		cloned[key] = value
	}
	cloned["data"] = items[:limit]
	return cloned
}

type curatedRenderOptions struct {
	withComments bool
}

func curatedRenderOptionsFromCmd(cmd *cobra.Command, meta shortcutopenapi.CommandMetadata) curatedRenderOptions {
	options := curatedRenderOptions{}
	if (meta.Group == "stories" || meta.Group == "epics") && meta.Name == "get" {
		options.withComments, _ = cmd.Flags().GetBool("with-comments")
	}
	return options
}

func buildOperationExamples(meta shortcutopenapi.CommandMetadata, prefix string) string {
	base := prefix + " " + meta.Name
	pathArgs := make([]string, 0)
	for _, param := range meta.PathParameters() {
		pathArgs = append(pathArgs, sampleValue(param))
	}
	if len(pathArgs) > 0 {
		base += " " + strings.Join(pathArgs, " ")
	}

	queryArgs := make([]string, 0)
	for _, param := range meta.QueryParameters() {
		queryArgs = append(queryArgs, fmt.Sprintf("--%s %s", param.Name, sampleValue(param)))
		if len(queryArgs) == 2 {
			break
		}
	}
	if len(queryArgs) > 0 {
		base += " " + strings.Join(queryArgs, " ")
	}

	if meta.HasBody {
		if meta.IsMultipart {
			base += " --form story_id=123 --file file0=./attachment.txt"
		} else {
			base += " --body '{\"name\":\"example\"}'"
		}
	}

	return base
}

func sampleValue(param shortcutopenapi.CommandParameter) string {
	if len(param.Enum) > 0 {
		return param.Enum[0]
	}

	switch param.Type {
	case "boolean":
		return "true"
	case "integer", "number":
		return "123"
	default:
		return fmt.Sprintf("<%s>", param.Name)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return "Shortcut API operation"
}

func marshalMultipartSummary(forms []string, files []string) (string, error) {
	buffer := &bytes.Buffer{}
	writer := multipart.NewWriter(buffer)

	for _, entry := range forms {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid form field %q, expected key=value", entry)
		}
		if err := writer.WriteField(parts[0], parts[1]); err != nil {
			return "", err
		}
	}

	for _, entry := range files {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid file field %q, expected key=path", entry)
		}
		fileWriter, err := writer.CreateFormFile(parts[0], filepath.Base(parts[1]))
		if err != nil {
			return "", err
		}
		file, err := os.Open(parts[1])
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(fileWriter, file); err != nil {
			file.Close()
			return "", err
		}
		file.Close()
	}

	if err := writer.Close(); err != nil {
		return "", err
	}

	return writer.FormDataContentType(), nil
}

func buildMultipartBody(forms []string, files []string) (*bytes.Buffer, string, error) {
	buffer := &bytes.Buffer{}
	writer := multipart.NewWriter(buffer)

	for _, entry := range forms {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			return nil, "", fmt.Errorf("invalid form field %q, expected key=value", entry)
		}
		if err := writer.WriteField(parts[0], parts[1]); err != nil {
			return nil, "", err
		}
	}

	for _, entry := range files {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			return nil, "", fmt.Errorf("invalid file field %q, expected key=path", entry)
		}
		fileWriter, err := writer.CreateFormFile(parts[0], filepath.Base(parts[1]))
		if err != nil {
			return nil, "", err
		}
		file, err := os.Open(parts[1])
		if err != nil {
			return nil, "", err
		}
		if _, err := io.Copy(fileWriter, file); err != nil {
			file.Close()
			return nil, "", err
		}
		file.Close()
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return buffer, writer.FormDataContentType(), nil
}

func decodeResponseBody(body []byte) any {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return nil
	}

	var decoded any
	if err := json.Unmarshal(body, &decoded); err == nil {
		return decoded
	}

	return trimmed
}

var _ = marshalMultipartSummary
