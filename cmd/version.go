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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s version %s\n", config.AppName, config.AppVersion)
		fmt.Printf("build date: %s\n", config.BuildTime)
		fmt.Printf("commit: %s\n\n", config.LastCommitHash)
		fmt.Println("https://github.com/ksoclabs/kbom")
	},
}
