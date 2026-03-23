package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	helpTemplate := strings.TrimSpace(`{{with or .Long .Short}}{{. | trimTrailingWhitespaces}}{{end}}

Usage:
  {{.UseLine}}{{if .HasAvailableSubCommands}} [command]{{end}}

{{if .HasAvailableSubCommands}}Available Commands:
{{range .Commands}}{{if (and .IsAvailableCommand (not .Hidden))}}  {{rpad .Name .NamePadding }} {{.Short}}
{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}
Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}
Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .Example}}

Examples:
{{.Example}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:
{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}  {{rpad .CommandPath .CommandPathPadding }} {{.Short}}
{{end}}{{end}}{{end}}
{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}`)

	cmd := &cobra.Command{
		Use:           "shortcut",
		Short:         "Shortcut CLI",
		Long:          "Shortcut CLI for the official Shortcut REST API, with built-in help, concise defaults, and stable JSON output.",
		Example:       "  shortcut me\n  shortcut stories get 123\n  shortcut search stories --query 'id:sc-12345'\n  shortcut search syntax",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.SetFlagErrorFunc(func(command *cobra.Command, err error) error {
		return fmt.Errorf("%w\n\nRun `%s --help` for usage", err, command.CommandPath())
	})

	cmd.PersistentFlags().StringP("output", "o", "text", "Output format: text or json")
	cmd.PersistentFlags().String("env-file", "", "Path to dotenv file (disables automatic ./.env and ~/.env search)")
	cmd.PersistentFlags().Bool("no-env-file", false, "Disable dotenv loading entirely")

	cmd.AddCommand(NewMeCmd())
	cmd.AddCommand(NewDocsCmd())
	cmd.AddCommand(NewVersionCmd())
	cmd.AddCommand(NewAPICmd())
	cmd.AddCommand(NewResourceCmd("stories", "Read and update stories", "  shortcut stories get 123\n  shortcut stories query --body '{\"workflow_state_id\":500131237}'"))
	cmd.AddCommand(NewResourceCmd("epics", "Read and update epics", "  shortcut epics list\n  shortcut epics get 123"))
	cmd.AddCommand(NewResourceCmd("iterations", "Read iterations and their stories", "  shortcut iterations list\n  shortcut iterations get 123"))
	cmd.AddCommand(NewResourceCmd("workflows", "Read workflow definitions and states", "  shortcut workflows list\n  shortcut workflows get 500131231"))
	cmd.AddCommand(NewSearchCmd())

	applyHelpTemplate(cmd, helpTemplate)

	return cmd
}

func applyHelpTemplate(cmd *cobra.Command, template string) {
	cmd.SetHelpTemplate(template)
	for _, child := range cmd.Commands() {
		applyHelpTemplate(child, template)
	}
}
