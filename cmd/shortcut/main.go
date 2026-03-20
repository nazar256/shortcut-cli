package main

import (
	"os"

	"github.com/nazar256/shortcut-cli/internal/cli"
)

func main() {
	rootCmd := cli.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		outputFormat, _ := rootCmd.PersistentFlags().GetString("output")
		cli.RenderError(outputFormat, err, os.Stderr)
		os.Exit(1)
	}
}
