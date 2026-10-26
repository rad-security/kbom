package cmd

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"

	"github.com/rad-security/kbom/internal/utils"
)

const (
	confDir = ".config/rad"
)

var (
	verbose    bool
	k8sContext string

	out io.WriteCloser = os.Stdout
)

var rootCmd = &cobra.Command{
	Use:   "kbom",
	Short: "KBOM - Kubernetes Bill of Materials",

	PersistentPreRun: setup,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(GenerateCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(schemaCmd)

	rootCmd.PersistentFlags().StringVarP(&k8sContext, "context", "c", "", "Kubernetes context to use, defaults to current context")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging (DEBUG and below)")

	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
}

func setup(cmd *cobra.Command, _ []string) {
	initLogger()
	initConfig()
	utils.BindFlags(cmd)
}

func initConfig() {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	viper.AddConfigPath(home)
	viper.SetConfigType("json")
	viper.SetConfigName(path.Join(confDir, "kbom.json"))

	if err := viper.ReadInConfig(); err != nil {
		// It's okay if there isn't a config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Println("error reading config file")
			os.Exit(1)
		}
	}

	// Environment variables can't have dashes
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

func initLogger() {
	defaultLogger := zerolog.New(os.Stderr)

	logLevel := zerolog.InfoLevel
	if verbose {
		logLevel = zerolog.TraceLevel
	}

	zerolog.SetGlobalLevel(logLevel)

	// use color logger when run in terminal; logs never go to stdout, which is reserved for the KBOM
	if isTerminalWriter(os.Stderr) {
		defaultLogger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	log.Logger = defaultLogger.With().Timestamp().Stack().Logger()
}

func isTerminalWriter(w io.Writer) bool {
	f, ok := w.(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
}
