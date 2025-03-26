package utils

import (
	"fmt"
	"github.com/golang-collections/collections/set"
	"github.com/rad-security/kbom/internal/model"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"
	"strings"

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

func ConvertListToSet(list []string) *set.Set {
	newSet := set.New()
	for _, v := range list {
		newSet.Insert(v)
	}
	return newSet
}

func GetFilterValue(namespaceFilter, resourceFilter string) (kbomFilter model.Filters) {
	if namespaceFilter != "" {
		kbomFilter.Namespace = strings.Split(strings.ReplaceAll(strings.ToLower(namespaceFilter), " ", ""), ",")
	}
	if resourceFilter != "" {
		kbomFilter.Resources = strings.Split(strings.ReplaceAll(strings.ToLower(resourceFilter), " ", ""), ",")
	}
	return
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
