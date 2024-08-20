package utils

import (
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

func GetVersion(item unstructured.Unstructured) (version string, ok bool) {

	obj := item.Object
	if obj == nil {
		return "", false
	}

	spec, ok := obj["spec"].(map[string]interface{})
	if !ok {
		return "", false
	}

	version, ok = spec["version"].(string)
	return
}
