package addons

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCPE(t *testing.T) {
	tests := []struct {
		product  string
		version  string
		expected string
	}{
		{KubernetesCPEProduct, "1.35.8", "cpe:2.3:a:kubernetes:kubernetes:1.35.8:*:*:*:*:*:*:*"},
		{"coredns.io:coredns", "v1.11.4-eksbuild.2", "cpe:2.3:a:coredns.io:coredns:1.11.4:*:*:*:*:*:*:*"},
		{"traefik:traefik", "v3.3", "cpe:2.3:a:traefik:traefik:3.3:*:*:*:*:*:*:*"},
		{"istio:istio", "1.19.0+build.1", "cpe:2.3:a:istio:istio:1.19.0:*:*:*:*:*:*:*"},
		{"istio:istio", "latest", ""},
		{"istio:istio", "", ""},
		{"", "1.0.0", ""},
	}

	for _, tt := range tests {
		t.Run(tt.product+"@"+tt.version, func(t *testing.T) {
			assert.Equal(t, tt.expected, CPE(tt.product, tt.version))
		})
	}
}

func TestCPEProductsAreDetectable(t *testing.T) {
	names := map[string]bool{}
	for _, a := range catalog {
		names[a.name] = true
	}

	for name := range cpeProducts {
		assert.True(t, names[name], "CPE product for unknown add-on %s", name)
	}
}
