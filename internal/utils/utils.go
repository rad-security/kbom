package utils

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func BindFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Determine the naming convention of the flags when represented in the config file
		configName := f.Name

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && viper.IsSet(configName) {
			val := viper.Get(configName)
			err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
			if err != nil {
				log.Error().Err(err).Send()
				os.Exit(1)
			}
		}
	})
}
