package utils

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestBindFlags(t *testing.T) {
	// Set up a new cobra command
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	// Add some flags to the command
	cmd.Flags().String("foo", "", "foo flag")
	cmd.Flags().Int("bar", 0, "bar flag")

	// Initialize viper with some values
	viper.Set("foo", "foo-value")
	viper.Set("bar", 123)

	// Bind the viper config values to the command flags
	BindFlags(cmd)

	// Check that the flag values were set correctly
	fooFlag := cmd.Flags().Lookup("foo")
	if fooFlag.Value.String() != "foo-value" {
		t.Errorf("expected foo flag value to be 'foo-value', but got '%s'", fooFlag.Value.String())
	}

	barFlag := cmd.Flags().Lookup("bar")
	if barFlag.Value.String() != "123" {
		t.Errorf("expected bar flag value to be '123', but got '%s'", barFlag.Value.String())
	}
}
