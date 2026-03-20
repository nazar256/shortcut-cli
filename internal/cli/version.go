package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/nazar256/shortcut-cli/internal/output"
	"github.com/spf13/cobra"
)

var Version = "dev"
var Commit = "unknown"
var BuildDate = "unknown"

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Show CLI build version",
		Example: "  shortcut version\n  shortcut version --output json",
		Args:    requireNoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			formatter := output.NewFormatter(outputFormat(cmd), cmd.OutOrStdout(), cmd.ErrOrStderr())
			payload := map[string]string{
				"version":    Version,
				"commit":     Commit,
				"build_date": BuildDate,
			}
			if formatter.Format() == output.FormatJSON {
				return formatter.Print(payload)
			}

			lines := []string{fmt.Sprintf("Version: %s", Version)}
			if strings.TrimSpace(Commit) != "" && Commit != "unknown" {
				lines = append(lines, fmt.Sprintf("Commit: %s", Commit))
			}
			if buildDateText := formatBuildDate(BuildDate); buildDateText != "" {
				lines = append(lines, fmt.Sprintf("Built: %s", buildDateText))
			}
			return formatter.Print(strings.Join(lines, "\n"))
		},
	}
}

func formatBuildDate(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || trimmed == "unknown" {
		return ""
	}
	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return trimmed
	}
	return parsed.UTC().Format(time.RFC3339)
}
