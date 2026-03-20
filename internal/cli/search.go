package cli

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	shortcutopenapi "github.com/nazar256/shortcut-cli/internal/openapi"
	"github.com/nazar256/shortcut-cli/internal/output"
	shortcutspec "github.com/nazar256/shortcut-cli/openapi"
	"github.com/spf13/cobra"
)

func NewSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "search",
		Short:   "Search stories, epics, documents, and more",
		Long:    "Search Shortcut records. Pick a scope such as `stories`, `epics`, or `documents`, then use that subcommand's flags. Run `shortcut search syntax` to learn common Shortcut search operators for query-based searches.",
		Example: "  shortcut search stories --query 'id:sc-12345'\n  shortcut search stories --query 'owner:example-user is:started'\n  shortcut search all --query '\"checkout\" type:epic'\n  shortcut search syntax",
	}

	cmd.AddCommand(newSearchSyntaxCmd())

	spec, err := shortcutopenapi.ParseSpecBytes(shortcutspec.SpecBytes)
	if err != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("load OpenAPI spec: %w", err)
		}
		return cmd
	}

	commands, err := shortcutopenapi.DeriveCommands(spec)
	if err != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("derive command metadata: %w", err)
		}
		return cmd
	}

	searchCommands := make([]*cobra.Command, 0)
	for _, meta := range commands {
		if meta.Group != "search" {
			continue
		}
		searchCommands = append(searchCommands, buildSearchOperationCmd(meta))
	}

	sort.Slice(searchCommands, func(i, j int) bool {
		return searchCommands[i].Name() < searchCommands[j].Name()
	})
	cmd.AddCommand(searchCommands...)

	return cmd
}

func buildSearchOperationCmd(meta shortcutopenapi.CommandMetadata) *cobra.Command {
	displayName, aliases := searchCommandName(meta.Name)
	meta.Name = displayName
	cmd := buildOperationCmdForPrefix(meta, "shortcut search", false)
	cmd.Aliases = aliases
	cmd.Short = searchShortDescription(meta)
	cmd.Long = searchLongDescription(meta)
	cmd.Example = searchExamples(meta)
	cmd.Flags().Int("limit", 10, "Maximum number of results to print; use 0 for all")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		normalizeSearchFlags(cmd)
		runtime, err := newRuntime(cmd)
		if err != nil {
			return err
		}

		result, err := invokeOperation(runtime, cmd, meta, args)
		if err != nil {
			return err
		}

		limit, _ := cmd.Flags().GetInt("limit")
		detail, _ := cmd.Flags().GetString("detail")
		payload, text := shapeSearchOutput(result.Response, limit, detail == "full")
		if runtime.Formatter.Format() == output.FormatJSON {
			return runtime.Formatter.Print(payload)
		}
		return runtime.Formatter.Print(text)
	}

	if flag := cmd.Flags().Lookup("query"); flag != nil {
		flag.Usage = "Shortcut search expression, passed to the API as the `query` parameter. Run `shortcut search syntax` for examples."
	}

	return cmd
}

func searchCommandName(operationName string) (string, []string) {
	switch operationName {
	case "search":
		return "all", []string{"search"}
	case "search-stories":
		return "stories", []string{"search-stories"}
	case "search-epics":
		return "epics", []string{"search-epics"}
	case "search-documents":
		return "documents", []string{"search-documents"}
	case "search-iterations":
		return "iterations", []string{"search-iterations"}
	case "search-milestones":
		return "milestones", []string{"search-milestones"}
	case "search-objectives":
		return "objectives", []string{"search-objectives"}
	default:
		return operationName, nil
	}
}

func searchShortDescription(meta shortcutopenapi.CommandMetadata) string {
	shorts := map[string]string{
		"all":        "Search across Shortcut records",
		"stories":    "Search stories",
		"epics":      "Search epics",
		"documents":  "Search documents",
		"iterations": "Search iterations",
		"milestones": "Search milestones",
		"objectives": "Search objectives",
	}
	if value, ok := shorts[meta.Name]; ok {
		return value
	}
	return firstNonEmpty(meta.Summary, meta.OperationID, meta.Name)
}

func searchLongDescription(meta shortcutopenapi.CommandMetadata) string {
	if meta.Name == "documents" {
		parts := []string{searchShortDescription(meta) + "."}
		parts = append(parts, "This command uses structured flags like `--title`, `--archived`, `--created_by_me`, and `--followed_by_me` instead of a generic `--query` expression.")
		return strings.Join(parts, "\n\n")
	}

	parts := []string{searchShortDescription(meta) + "."}
	parts = append(parts, "`--query` is the Shortcut search expression sent to the API as the `query` parameter.")
	parts = append(parts, "Use a member's mention name for `owner:` and `requester:` (for example `owner:example-user`), not `@name` and not `owner:me`.")
	parts = append(parts, "Run `shortcut search syntax` to see common operators and example queries.")
	return strings.Join(parts, "\n\n")
}

func searchExamples(meta shortcutopenapi.CommandMetadata) string {
	base := "shortcut search " + meta.Name
	switch meta.Name {
	case "all":
		return base + " --query '\"checkout\" type:epic'\n" + base + " --query 'owner:example-user is:started'"
	case "stories":
		return base + " --query 'id:sc-12345'\n" + base + " --query 'owner:example-user is:started' --detail slim\n" + base + " --query 'label:\"ios\" updated:2026-03-01..2026-03-20'"
	case "epics":
		return base + " --query 'owner:example-user'\n" + base + " --query '\"migration\"' --detail slim"
	case "documents":
		return base + " --title 'search ranking'\n" + base + " --title 'checkout' --created_by_me true"
	case "iterations":
		return base + " --query 'name:Q1'\n" + base + " --query 'team:Engineering' --detail slim"
	case "milestones":
		return base + " --query 'status:started'"
	case "objectives":
		return base + " --query 'owner:example-user'"
	default:
		return buildOperationExamples(meta, "shortcut search")
	}
}

func newSearchSyntaxCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "syntax",
		Aliases: []string{"operators"},
		Short:   "Show search query syntax and examples",
		Long:    "Shortcut search syntax guide. The value passed to --query is sent unchanged to Shortcut's search API.",
		Args:    requireNoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := map[string]any{
				"summary": "Shortcut search query syntax",
				"behavior": []string{
					"The value passed to --query is sent unchanged to Shortcut's search API.",
					"Search terms are combined with AND logic by default.",
				},
				"operators": []string{
					"id:sc-12345",
					"owner:example-user",
					"requester:example-user",
					"team:Engineering",
					"type:bug",
					"state:\"In Review\"",
					"is:started",
					"label:\"ios\"",
					"project:\"Platform\"",
					"created:2026-03-01..2026-03-20",
					"updated:today",
					"!has:comment",
					"has:comment",
				},
				"notes": []string{
					"owner: and requester: require a full mention name without @.",
					"team: uses the team name, not the mention name.",
					"Dates use YYYY-MM-DD and ranges use .., for example 2026-03-01..2026-03-20.",
					"state: matches an exact workflow-state name; is: matches broad state types like unstarted/started/done.",
				},
				"examples": []string{
					"shortcut search stories --query 'id:sc-12345'",
					"shortcut search stories --query 'owner:example-user is:started'",
					"shortcut search stories --query 'label:\"ios\" updated:2026-03-01..2026-03-20'",
					"shortcut search all --query '\"checkout\" type:epic'",
				},
				"more": []string{
					"https://help.shortcut.com/hc/en-us/articles/360000046646-Search-Operators",
				},
			}

			formatter := output.NewFormatter(outputFormat(cmd), cmd.OutOrStdout(), cmd.ErrOrStderr())
			if formatter.Format() == output.FormatJSON {
				return formatter.Print(payload)
			}

			text := strings.Join([]string{
				"Shortcut search query syntax",
				"",
				"The value passed to --query is sent unchanged to Shortcut's search API.",
				"Search terms are combined with AND logic by default.",
				"",
				"Common operators:",
				"- id:sc-12345",
				"- owner:example-user",
				"- requester:example-user",
				"- team:Engineering",
				"- type:bug",
				"- state:\"In Review\"",
				"- is:started",
				"- label:\"ios\"",
				"- project:\"Platform\"",
				"- created:2026-03-01..2026-03-20",
				"- updated:today",
				"- !has:comment",
				"- has:comment",
				"",
				"Notes:",
				"- `owner:` and `requester:` require a full mention name without `@`.",
				"- `team:` uses the team name, not the mention name.",
				"- Dates use YYYY-MM-DD and ranges use `..`, for example `2026-03-01..2026-03-20`.",
				"- `state:` matches an exact workflow-state name; `is:` matches broad state types like unstarted/started/done.",
				"",
				"Examples:",
				"- shortcut search stories --query 'id:sc-12345'",
				"- shortcut search stories --query 'owner:example-user is:started'",
				"- shortcut search stories --query 'label:\"ios\" updated:2026-03-01..2026-03-20'",
				"- shortcut search all --query '\"checkout\" type:epic'",
				"",
				"More:",
				"- https://help.shortcut.com/hc/en-us/articles/360000046646-Search-Operators",
			}, "\n")
			return formatter.Print(text)
		},
	}
}

func normalizeSearchFlags(cmd *cobra.Command) {
	for _, name := range []string{"query", "title"} {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			continue
		}
		value, err := cmd.Flags().GetString(name)
		if err != nil || value == "" {
			continue
		}
		_ = cmd.Flags().Set(name, strings.TrimSpace(value))
	}
}

func shapeSearchOutput(response any, limit int, includeDescription bool) (map[string]any, string) {
	if payload, ok := response.(map[string]any); ok && hasSearchBuckets(payload) {
		return shapeMultiSearchOutput(payload, limit, includeDescription)
	}

	if payload, ok := response.(map[string]any); ok {
		items, _ := payload["data"].([]any)
		total := toInt(payload["total"])
		next, _ := payload["next"].(string)
		shownItems := items
		truncated := false
		if limit > 0 && len(items) > limit {
			shownItems = items[:limit]
			truncated = true
		}

		jsonPayload := map[string]any{
			"results": shownItems,
			"shown":   len(shownItems),
			"total":   total,
		}
		if next != "" {
			jsonPayload["next"] = next
		}

		return jsonPayload, formatSearchText(shownItems, total, next, limit, truncated, includeDescription)
	}

	return map[string]any{"result": response}, output.ToText(response)
}

func shapeMultiSearchOutput(payload map[string]any, limit int, includeDescription bool) (map[string]any, string) {
	buckets := []string{"stories", "epics", "iterations", "milestones"}
	jsonPayload := map[string]any{}
	lines := []string{}
	totalShown := 0
	totalMatches := 0
	remaining := limit

	for _, bucket := range buckets {
		bucketValue, ok := payload[bucket].(map[string]any)
		if !ok {
			continue
		}
		items, _ := bucketValue["data"].([]any)
		total := toInt(bucketValue["total"])
		next, _ := bucketValue["next"].(string)
		shownItems := items
		truncated := false
		if remaining == 0 && limit > 0 {
			shownItems = []any{}
			truncated = len(items) > 0
		} else if limit > 0 && len(items) > remaining {
			shownItems = items[:remaining]
			truncated = true
		}
		totalShown += len(shownItems)
		totalMatches += total
		if limit > 0 {
			remaining -= len(shownItems)
			if remaining < 0 {
				remaining = 0
			}
		}

		if total == 0 {
			continue
		}
		jsonPayload[bucket] = map[string]any{
			"results": shownItems,
			"shown":   len(shownItems),
			"total":   total,
		}
		if next != "" {
			jsonPayload[bucket].(map[string]any)["next"] = next
		}
		lines = append(lines, strings.Title(bucket)+":")
		lines = append(lines, formatSearchText(shownItems, total, next, limit, truncated, includeDescription))
		lines = append(lines, "")
	}

	jsonPayload["shown"] = totalShown
	jsonPayload["total"] = totalMatches
	if len(lines) == 0 {
		return jsonPayload, "No results."
	}
	return jsonPayload, strings.TrimSpace(strings.Join(lines, "\n"))
}

func formatSearchText(items []any, total int, next string, limit int, truncated bool, includeDescription bool) string {
	if len(items) == 0 {
		if total == 0 {
			return "No results."
		}
		return fmt.Sprintf("0 shown of %d results.", total)
	}

	lines := []string{fmt.Sprintf("Showing %d of %d results.", len(items), total)}
	for index, item := range items {
		if entry, ok := item.(map[string]any); ok {
			lines = append(lines, fmt.Sprintf("%d. %s", index+1, summarizeSearchItem(entry, includeDescription)))
			continue
		}
		lines = append(lines, fmt.Sprintf("%d. %s", index+1, output.ToText(item)))
	}

	if next != "" {
		lines = append(lines, "")
		lines = append(lines, "More results are available with the next page token.")
	}
	if truncated && limit > 0 {
		lines = append(lines, fmt.Sprintf("Tip: raise --limit above %d to show more from the current page.", limit))
	}

	return strings.Join(lines, "\n")
}

func hasSearchBuckets(payload map[string]any) bool {
	for _, key := range []string{"stories", "epics", "iterations", "milestones"} {
		if _, ok := payload[key]; ok {
			return true
		}
	}
	return false
}

func summarizeSearchItem(entry map[string]any, includeDescription bool) string {
	parts := []string{}
	if id, ok := entry["id"]; ok {
		parts = append(parts, "#"+formatScalar(id))
	}
	if entityType, ok := entry["entity_type"].(string); ok && entityType != "" {
		parts = append(parts, entityType)
	}
	if name, ok := entry["name"].(string); ok && name != "" {
		parts = append(parts, strconv.Quote(strings.TrimSpace(name)))
	} else if title, ok := entry["title"].(string); ok && title != "" {
		parts = append(parts, strconv.Quote(strings.TrimSpace(title)))
	}
	if appURL, ok := entry["app_url"].(string); ok && appURL != "" {
		parts = append(parts, appURL)
	}
	line := strings.Join(parts, " - ")
	if line == "" {
		line = output.ToText(entry)
	}
	if !includeDescription {
		return line
	}

	description := firstNonEmptyString(entry, "description", "summary")
	description = compactDescription(description)
	if description == "" {
		return line
	}

	return line + "\n   " + description
}

func firstNonEmptyString(entry map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := entry[key].(string); ok && strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func compactDescription(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	replacer := strings.NewReplacer("\r", " ", "\n", " ", "\t", " ", "<br>", " ", "<br/>", " ", "<br />", " ")
	trimmed = replacer.Replace(trimmed)
	trimmed = strings.Join(strings.Fields(trimmed), " ")
	if len(trimmed) <= 180 {
		return trimmed
	}
	return trimmed[:177] + "..."
}

func toInt(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return 0
	}
}
