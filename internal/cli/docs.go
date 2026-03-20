package cli

import (
	"fmt"
	"sort"
	"strings"

	shortcutopenapi "github.com/nazar256/shortcut-cli/internal/openapi"
	"github.com/nazar256/shortcut-cli/internal/output"
	shortcutspec "github.com/nazar256/shortcut-cli/openapi"
	"github.com/spf13/cobra"
)

func NewDocsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "docs",
		Short:   "Show built-in API documentation summaries",
		Long:    "Prints spec-derived documentation directly from the vendored Shortcut OpenAPI definition embedded in the CLI.",
		Example: "  shortcut docs summary\n  shortcut docs operation stories get-story",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "summary",
		Short: "Show a high-level overview of the embedded Shortcut API spec",
		Args:  requireNoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			formatter := output.NewFormatter(outputFormat(cmd), cmd.OutOrStdout(), cmd.ErrOrStderr())

			spec, err := shortcutopenapi.ParseSpecBytes(shortcutspec.SpecBytes)
			if err != nil {
				return err
			}

			commands, err := shortcutopenapi.DeriveCommands(spec)
			if err != nil {
				return err
			}

			groups := map[string]int{}
			for _, command := range commands {
				groups[command.Group]++
			}

			names := make([]string, 0, len(groups))
			for name := range groups {
				names = append(names, name)
			}
			sort.Strings(names)

			summary := map[string]any{
				"openapi":     spec.OpenAPI,
				"operations":  len(commands),
				"groups":      groups,
				"group_names": names,
			}

			if formatter.Format() == output.FormatText {
				lines := []string{fmt.Sprintf("OpenAPI version: %s", spec.OpenAPI), fmt.Sprintf("Operations: %d", len(commands)), "Groups:"}
				for _, name := range names {
					lines = append(lines, fmt.Sprintf("- %s (%d)", name, groups[name]))
				}
				return formatter.Print(strings.Join(lines, "\n"))
			}

			return formatter.Print(summary)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "operation <group> <command>",
		Short: "Show a single operation's help metadata",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			formatter := output.NewFormatter(outputFormat(cmd), cmd.OutOrStdout(), cmd.ErrOrStderr())

			spec, err := shortcutopenapi.ParseSpecBytes(shortcutspec.SpecBytes)
			if err != nil {
				return err
			}

			commands, err := shortcutopenapi.DeriveCommands(spec)
			if err != nil {
				return err
			}

			for _, meta := range commands {
				if meta.Group == args[0] && meta.Name == args[1] {
					return formatter.Print(meta)
				}
			}

			return fmt.Errorf("operation %s %s not found", args[0], args[1])
		},
	})

	return cmd
}
