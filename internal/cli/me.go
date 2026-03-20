package cli

import (
	"github.com/nazar256/shortcut-cli/internal/openapi"
	"github.com/spf13/cobra"
)

func NewMeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "me",
		Short:   "Show the authenticated Shortcut member",
		Long:    "Show the authenticated Shortcut member in a readable summary or JSON form.",
		Example: "  shortcut me\n  shortcut me -o json",
		Args:    requireNoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, err := newRuntime(cmd)
			if err != nil {
				return err
			}

			response, err := runtime.Client.GetCurrentMemberInfoWithResponse(commandContext(cmd))
			if err != nil {
				return err
			}

			if response.JSON200 != nil {
				return printCuratedOperation(runtime.Formatter, openapi.CommandMetadata{Name: "me", Group: "member"}, response.JSON200, curatedRenderOptions{})
			}

			decoded := decodeResponseBody(response.Body)
			if err := EnsureHTTPSuccess(response.HTTPResponse, decoded); err != nil {
				return err
			}

			return printCuratedOperation(runtime.Formatter, openapi.CommandMetadata{Name: "me", Group: "member"}, decoded, curatedRenderOptions{})
		},
	}

	return cmd
}
