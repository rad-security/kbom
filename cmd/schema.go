package cmd

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
	"github.com/ksoclabs/kbom/internal/model"
	"github.com/spf13/cobra"
)

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Print the KBOM json file schema",
	RunE:  runGenerateSchema,
}

func runGenerateSchema(cmd *cobra.Command, _ []string) error {
	schema := jsonschema.Reflect(&model.KBOM{})
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")

	return enc.Encode(schema)
}
