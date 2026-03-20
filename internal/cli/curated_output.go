package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	shortcutopenapi "github.com/nazar256/shortcut-cli/internal/openapi"
	"github.com/nazar256/shortcut-cli/internal/output"
)

func printCuratedOperation(formatter *output.Formatter, meta shortcutopenapi.CommandMetadata, response any, options curatedRenderOptions) error {
	jsonPayload, textPayload := shapeCuratedOperation(meta, response, options)
	if formatter.Format() == output.FormatJSON {
		return formatter.Print(jsonPayload)
	}
	return formatter.Print(textPayload)
}

func shapeCuratedOperation(meta shortcutopenapi.CommandMetadata, response any, options curatedRenderOptions) (any, string) {
	if payload := structToMap(response); payload != nil {
		if items, ok := payload["data"].([]any); ok {
			total := toInt(payload["total"])
			next, _ := payload["next"].(string)
			jsonPayload := map[string]any{
				"results": items,
				"shown":   len(items),
			}
			if total > 0 {
				jsonPayload["total"] = total
			}
			if next != "" {
				jsonPayload["next"] = next
			}
			return jsonPayload, formatCuratedCollection(meta, items, total, next)
		}
		return payload, formatCuratedSingle(meta, payload, options)
	}

	if payload, ok := response.(map[string]any); ok {
		if items, ok := payload["data"].([]any); ok {
			total := toInt(payload["total"])
			next, _ := payload["next"].(string)
			jsonPayload := map[string]any{
				"results": items,
				"shown":   len(items),
			}
			if total > 0 {
				jsonPayload["total"] = total
			}
			if next != "" {
				jsonPayload["next"] = next
			}
			return jsonPayload, formatCuratedCollection(meta, items, total, next)
		}
		return payload, formatCuratedSingle(meta, payload, options)
	}

	if items, ok := response.([]any); ok {
		jsonPayload := map[string]any{"results": items, "shown": len(items)}
		return jsonPayload, formatCuratedCollection(meta, items, len(items), "")
	}

	return response, output.ToText(response)
}

func formatCuratedSingle(meta shortcutopenapi.CommandMetadata, payload map[string]any, options curatedRenderOptions) string {
	name := firstNonEmptyString(payload, "name", "title")
	if name == "" {
		return output.ToText(payload)
	}

	parts := []string{}
	entityType := firstNonEmptyString(payload, "entity_type")
	entityLabel := strings.Title(entityType)
	if entityLabel == "" {
		entityLabel = strings.Title(resourceSingular(meta.Group))
	}
	if id, ok := payload["id"]; ok {
		parts = append(parts, fmt.Sprintf("%s #%s", entityLabel, formatScalar(id)))
	} else {
		parts = append(parts, entityLabel)
	}
	parts = append(parts, name)

	lines := []string{strings.Join(parts, " ")}
	if description := strings.TrimSpace(firstNonEmptyString(payload, "description", "summary")); description != "" {
		lines = append(lines, "Description: "+description)
	}
	if storyType := firstNonEmptyString(payload, "story_type"); storyType != "" {
		lines = append(lines, "Type: "+storyType)
	}
	if role := firstNonEmptyString(payload, "role"); role != "" {
		lines = append(lines, "Role: "+role)
	}
	if estimate, ok := payload["estimate"]; ok {
		lines = append(lines, fmt.Sprintf("Estimate: %s", formatScalar(estimate)))
	}
	if labels := labelNames(payload); len(labels) > 0 {
		lines = append(lines, "Labels: "+strings.Join(labels, ", "))
	}
	if owners := ownerList(payload); len(owners) > 0 {
		lines = append(lines, "Owners: "+strings.Join(owners, ", "))
	}
	if mention := firstNonEmptyString(payload, "mention_name"); mention != "" {
		lines = append(lines, "Mention: @"+mention)
	}
	if email := firstNonEmptyString(payload, "email_address"); email != "" {
		lines = append(lines, "Email: "+email)
	}
	if appURL, ok := payload["app_url"].(string); ok && appURL != "" {
		lines = append(lines, "URL: "+appURL)
	}
	if interval := iterationRange(payload); interval != "" {
		lines = append(lines, "Dates: "+interval)
	}
	if workflowStates := workflowStateSummary(payload); workflowStates != "" {
		lines = append(lines, "States: "+workflowStates)
	}
	if state := inferCuratedState(meta, payload); state != "" {
		lines = append(lines, "State: "+state)
	}
	if options.withComments {
		if commentText := renderComments(payload); commentText != "" {
			lines = append(lines, "")
			lines = append(lines, commentText)
		}
	}
	return strings.Join(lines, "\n")
}

func formatCuratedCollection(meta shortcutopenapi.CommandMetadata, items []any, total int, next string) string {
	if len(items) == 0 {
		if total == 0 {
			return "No results."
		}
		return fmt.Sprintf("0 shown of %d results.", total)
	}

	headingTotal := total
	if headingTotal == 0 {
		headingTotal = len(items)
	}
	lines := []string{fmt.Sprintf("Showing %d of %d results.", len(items), headingTotal)}
	for index, item := range items {
		if entry, ok := item.(map[string]any); ok {
			lines = append(lines, fmt.Sprintf("%d. %s", index+1, summarizeCollectionItem(meta, entry)))
			continue
		}
		lines = append(lines, fmt.Sprintf("%d. %s", index+1, output.ToText(item)))
	}
	if next != "" {
		lines = append(lines, "")
		lines = append(lines, "More results are available with the next page token.")
	}
	return strings.Join(lines, "\n")
}

func summarizeCollectionItem(meta shortcutopenapi.CommandMetadata, entry map[string]any) string {
	if meta.Group == "stories" && meta.Name == "history" {
		return summarizeStoryHistoryItem(entry)
	}
	if meta.Group == "stories" || meta.Group == "search" {
		return summarizeSearchItem(entry, false)
	}

	parts := []string{}
	if id, ok := entry["id"]; ok {
		parts = append(parts, "#"+formatScalar(id))
	}
	if entityType := firstNonEmptyString(entry, "entity_type"); entityType != "" {
		parts = append(parts, entityType)
	}
	if name := firstNonEmptyString(entry, "name", "title"); name != "" {
		parts = append(parts, strconv.Quote(strings.TrimSpace(name)))
	}
	if state := inferCuratedState(meta, entry); state != "" {
		parts = append(parts, "state:"+state)
	}
	if url := firstNonEmptyString(entry, "app_url"); url != "" {
		parts = append(parts, url)
	}
	return strings.Join(parts, " - ")
}

func inferCuratedState(meta shortcutopenapi.CommandMetadata, payload map[string]any) string {
	if meta.Group == "workflows" {
		return "configured"
	}
	if state := firstNonEmptyString(payload, "state"); state != "" {
		return state
	}
	if status := firstNonEmptyString(payload, "status"); status != "" {
		return status
	}
	if completed, ok := payload["completed"].(bool); ok && completed {
		return "completed"
	}
	if archived, ok := payload["archived"].(bool); ok && archived {
		return "archived"
	}
	if started, ok := payload["started"].(bool); ok && started {
		return "started"
	}
	return ""
}

func summarizeStoryHistoryItem(entry map[string]any) string {
	parts := []string{}
	if actorName := firstNonEmptyString(entry, "actor_name"); actorName != "" {
		parts = append(parts, actorName)
	}
	if changedAt := firstNonEmptyString(entry, "changed_at"); changedAt != "" {
		parts = append(parts, changedAt)
	}
	if actions := summarizeHistoryActions(entry); actions != "" {
		parts = append(parts, actions)
	}
	if len(parts) > 0 {
		return strings.Join(parts, " - ")
	}
	return output.ToText(entry)
}

func summarizeHistoryActions(entry map[string]any) string {
	actions, ok := entry["actions"].([]any)
	if !ok || len(actions) == 0 {
		return ""
	}
	parts := make([]string, 0, len(actions))
	for _, raw := range actions {
		action, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		segment := strings.TrimSpace(strings.Join([]string{
			firstNonEmptyString(action, "action"),
			firstNonEmptyString(action, "entity_type"),
			firstNonEmptyString(action, "name"),
		}, " "))
		segment = strings.Join(strings.Fields(segment), " ")
		if segment != "" {
			parts = append(parts, segment)
		}
	}
	return strings.Join(parts, "; ")
}

func workflowStateSummary(payload map[string]any) string {
	states, ok := payload["states"].([]any)
	if !ok || len(states) == 0 {
		return ""
	}
	names := make([]string, 0, len(states))
	for _, value := range states {
		entry, ok := value.(map[string]any)
		if !ok {
			continue
		}
		name := firstNonEmptyString(entry, "name")
		if name == "" {
			continue
		}
		typeName := firstNonEmptyString(entry, "type")
		if typeName != "" {
			names = append(names, fmt.Sprintf("%s (%s)", name, strings.ToLower(typeName)))
			continue
		}
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}

func labelNames(payload map[string]any) []string {
	labels, ok := payload["labels"].([]any)
	if !ok {
		return nil
	}
	names := make([]string, 0, len(labels))
	for _, label := range labels {
		entry, ok := label.(map[string]any)
		if !ok {
			continue
		}
		name := firstNonEmptyString(entry, "name")
		if name != "" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func ownerList(payload map[string]any) []string {
	for _, key := range []string{"owner_ids", "follower_ids"} {
		if values, ok := payload[key].([]any); ok && len(values) > 0 {
			items := make([]string, 0, len(values))
			for _, value := range values {
				items = append(items, fmt.Sprintf("%v", value))
			}
			return items
		}
	}
	return nil
}

func idList(payload map[string]any, key string) []string {
	values, ok := payload[key].([]any)
	if !ok {
		return nil
	}
	items := make([]string, 0, len(values))
	for _, value := range values {
		items = append(items, formatScalar(value))
	}
	return items
}

func iterationRange(payload map[string]any) string {
	start := firstNonEmptyString(payload, "start_date")
	end := firstNonEmptyString(payload, "end_date")
	if start == "" && end == "" {
		return ""
	}
	if start == "" {
		return end
	}
	if end == "" {
		return start
	}
	return start + " → " + end
}

func renderComments(payload map[string]any) string {
	comments, ok := payload["comments"].([]any)
	if !ok || len(comments) == 0 {
		return ""
	}
	lines := []string{"Comments:"}
	for index, raw := range comments {
		comment, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		text := strings.TrimSpace(firstNonEmptyString(comment, "text"))
		if text == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%d. %s", index+1, text))
	}
	if len(lines) == 1 {
		return ""
	}
	return strings.Join(lines, "\n")
}

func structToMap(value any) map[string]any {
	if value == nil {
		return nil
	}
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	var payload map[string]any
	if err := json.Unmarshal(bytes, &payload); err != nil {
		return nil
	}
	return payload
}

func formatScalar(value any) string {
	switch typed := value.(type) {
	case float64:
		return fmt.Sprintf("%.0f", typed)
	case float32:
		return fmt.Sprintf("%.0f", typed)
	default:
		return fmt.Sprintf("%v", value)
	}
}
