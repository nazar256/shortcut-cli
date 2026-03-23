package cli

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/nazar256/shortcut-cli/internal/config"
	"github.com/nazar256/shortcut-cli/internal/openapi"
	shortcutruntime "github.com/nazar256/shortcut-cli/internal/shortcut"
	"github.com/spf13/cobra"
)

func outputFormat(cmd *cobra.Command) string {
	value, _ := cmd.Root().PersistentFlags().GetString("output")
	return value
}

func newRuntime(cmd *cobra.Command) (*shortcutruntime.Runtime, error) {
	envFilePath, _ := cmd.Root().PersistentFlags().GetString("env-file")
	noEnvFile, _ := cmd.Root().PersistentFlags().GetBool("no-env-file")

	return shortcutruntime.NewRuntime(cmd.Context(), outputFormat(cmd), config.LoadOptions{
		EnvFilePath: envFilePath,
		NoEnvFile:   noEnvFile,
	})
}

func requireNoArgs(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(args, " "))
	}
	return nil
}

func pathParameterValues(meta openapi.CommandMetadata, args []string) (map[string]string, error) {
	pathParams := meta.PathParameters()
	if len(args) != len(pathParams) {
		return nil, fmt.Errorf("expected %d path arguments, got %d", len(pathParams), len(args))
	}

	values := make(map[string]string, len(pathParams))
	for index, param := range pathParams {
		values[param.Name] = args[index]
	}

	return values, nil
}

func parseValue(raw string, param openapi.CommandParameter) any {
	if raw == "" {
		return raw
	}

	switch param.Type {
	case "boolean":
		parsed, err := strconv.ParseBool(raw)
		if err == nil {
			return parsed
		}
	case "integer":
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err == nil {
			return parsed
		}
	case "number":
		parsed, err := strconv.ParseFloat(raw, 64)
		if err == nil {
			return parsed
		}
	}

	return raw
}

func commandContext(cmd *cobra.Command) context.Context {
	if ctx := cmd.Context(); ctx != nil {
		return ctx
	}
	return context.Background()
}

func withQuery(rawURL string, queryValues map[string]any) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	query := parsed.Query()
	for key, value := range queryValues {
		switch typed := value.(type) {
		case []any:
			for _, item := range typed {
				query.Add(key, fmt.Sprintf("%v", item))
			}
		default:
			query.Set(key, fmt.Sprintf("%v", typed))
		}
	}

	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}
