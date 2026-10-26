package model

import (
	"reflect"
	"testing"
)

// TestYAMLTagsMatchJSON ensures YAML output uses the same field names as JSON output.
func TestYAMLTagsMatchJSON(t *testing.T) {
	seen := map[reflect.Type]bool{}

	var walk func(typ reflect.Type)
	walk = func(typ reflect.Type) {
		for typ.Kind() == reflect.Pointer || typ.Kind() == reflect.Slice || typ.Kind() == reflect.Map {
			typ = typ.Elem()
		}

		if typ.Kind() != reflect.Struct || typ.PkgPath() != reflect.TypeFor[KBOM]().PkgPath() || seen[typ] {
			return
		}
		seen[typ] = true

		for field := range typ.Fields() {
			if !field.IsExported() {
				continue
			}

			jsonTag, yamlTag := field.Tag.Get("json"), field.Tag.Get("yaml")
			if jsonTag != yamlTag {
				t.Errorf("%s.%s: json tag %q and yaml tag %q differ", typ.Name(), field.Name, jsonTag, yamlTag)
			}

			walk(field.Type)
		}
	}

	walk(reflect.TypeFor[KBOM]())

	if len(seen) < 5 {
		t.Fatalf("expected to check all model types, checked only %d", len(seen))
	}
}
