package cmd

import (
	"fmt"

	"github.com/ksoclabs/kbom/internal/config"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the KBOM generator version",
	Long:  `All software has versions. This is KBOM's`,
	RunE:  runPrintVersion,
}

func runPrintVersion(cmd *cobra.Command, _ []string) error {
	fmt.Fprintf(out, "%s version %s\n", config.AppName, config.AppVersion)
	fmt.Fprintf(out, "build date: %s\n", config.BuildTime)
	fmt.Fprintf(out, "commit: %s\n\n", config.LastCommitHash)
	fmt.Fprintln(out, "https://github.com/ksoclabs/kbom")

	return nil
}
