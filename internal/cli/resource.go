package cli

import (
	"fmt"
	"sort"
	"strings"

	shortcutopenapi "github.com/nazar256/shortcut-cli/internal/openapi"
	shortcutspec "github.com/nazar256/shortcut-cli/openapi"
	"github.com/spf13/cobra"
)

func NewResourceCmd(name string, short string, example string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     name,
		Short:   short,
		Example: example,
	}

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

	resourceCommands := make([]*cobra.Command, 0)
	for _, meta := range commands {
		if meta.Group != name {
			continue
		}
		resourceCommands = append(resourceCommands, buildResourceOperationCmd(name, meta))
	}

	sort.Slice(resourceCommands, func(i, j int) bool {
		return resourceCommands[i].Name() < resourceCommands[j].Name()
	})
	cmd.AddCommand(resourceCommands...)

	for _, child := range cmd.Commands() {
		child.Short = conciseShort(name, child.Short)
	}

	if len(cmd.Commands()) == 0 {
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		}
		return cmd
	}

	cmd.Long = fmt.Sprintf("%s. Use the subcommands below to work with %s.", short, name)
	return cmd
}

func buildResourceOperationCmd(resource string, meta shortcutopenapi.CommandMetadata) *cobra.Command {
	conciseName, aliases := conciseResourceCommandName(resource, meta.Name)
	meta.Name = conciseName
	cmd := buildOperationCmdForPrefix(meta, "shortcut "+resource, false)
	cmd.Aliases = aliases
	cmd.Short = conciseShort(resource, cmd.Short)
	return cmd
}

func conciseResourceCommandName(resource string, name string) (string, []string) {
	trimmed := strings.TrimPrefix(name, resourceSingular(resource)+"-")
	trimmed = strings.TrimPrefix(trimmed, resource+"-")
	trimmed = strings.TrimSuffix(trimmed, "-"+resource)
	trimmed = strings.TrimSuffix(trimmed, "-"+resourceSingular(resource))
	trimmed = strings.ReplaceAll(trimmed, "-"+resourceSingular(resource)+"-", "-")
	trimmed = strings.ReplaceAll(trimmed, "-"+resource+"-", "-")
	trimmed = strings.Trim(trimmed, "-")
	if trimmed == "" {
		return name, nil
	}
	return trimmed, []string{name}
}

func conciseShort(resource string, short string) string {
	resourceTitle := strings.Title(resource)
	singularTitle := strings.Title(resourceSingular(resource))
	replacements := []string{
		" " + resourceTitle, "",
		" " + singularTitle, "",
		resourceTitle + " ", "",
		singularTitle + " ", "",
	}
	updated := strings.NewReplacer(replacements...).Replace(short)
	updated = strings.Join(strings.Fields(updated), " ")
	if updated == "" {
		return short
	}
	return updated
}

func resourceSingular(resource string) string {
	if strings.HasSuffix(resource, "ies") {
		return strings.TrimSuffix(resource, "ies") + "y"
	}
	if strings.HasSuffix(resource, "s") {
		return strings.TrimSuffix(resource, "s")
	}
	return resource
}
