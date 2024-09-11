package utils

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// BindFlags binds the viper config values to the flags
func BindFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		configName := f.Name

		if !f.Changed && viper.IsSet(configName) {
			val := viper.Get(configName)
			err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	})
}
