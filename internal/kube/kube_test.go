package kube

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const kubeconfigTemplate = `apiVersion: v1
kind: Config
current-context: %[1]s
clusters:
- name: %[1]s
  cluster:
    server: %[2]s
contexts:
- name: %[1]s
  context:
    cluster: %[1]s
    user: %[1]s
users:
- name: %[1]s
  user:
    token: fake
`

func writeKubeconfig(t *testing.T, dir, name, server string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(fmt.Sprintf(kubeconfigTemplate, name, server)), 0o600)
	require.NoError(t, err)

	return path
}

func TestNewClientMultipleKubeconfigFiles(t *testing.T) {
	dir := t.TempDir()
	first := writeKubeconfig(t, dir, "cluster-a", "https://cluster-a.example.com")
	second := writeKubeconfig(t, dir, "cluster-b", "https://cluster-b.example.com")

	t.Setenv("KUBERNETES_SERVICE_HOST", "")
	t.Setenv("KUBECONFIG", first+string(os.PathListSeparator)+second)

	tests := []struct {
		name            string
		context         string
		expectedContext string
		expectedHost    string
	}{
		{
			name:            "current context from first file",
			context:         "",
			expectedContext: "cluster-a",
			expectedHost:    "https://cluster-a.example.com",
		},
		{
			name:            "explicit context from second file",
			context:         "cluster-b",
			expectedContext: "cluster-b",
			expectedHost:    "https://cluster-b.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.context)
			require.NoError(t, err)

			db, ok := client.(*k8sDB)
			require.True(t, ok)
			assert.Equal(t, tt.expectedContext, db.k8sContext)
			assert.Equal(t, tt.expectedHost, db.cfg.Host)
		})
	}

	t.Run("unknown context", func(t *testing.T) {
		_, err := NewClient("cluster-c")
		assert.Error(t, err)
	})
}
