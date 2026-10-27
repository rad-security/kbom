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

func TestClusterPackageURL(t *testing.T) {
	for version, expected := range map[string]string{
		"1.35.8":           "pkg:golang/k8s.io/kubernetes@v1.35.8",
		"v1.35.8":          "pkg:golang/k8s.io/kubernetes@v1.35.8",
		"1.35.8-eks-3025e": "pkg:golang/k8s.io/kubernetes@v1.35.8",
		"1.35.8+k3s1":      "pkg:golang/k8s.io/kubernetes@v1.35.8",
		"":                 "",
	} {
		c := Cluster{K8sVersion: version}
		if got := c.PackageURL(); got != expected {
			t.Errorf("PackageURL(%q) = %q, want %q", version, got, expected)
		}
	}
}
