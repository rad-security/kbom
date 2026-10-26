package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClusterIdentity(t *testing.T) {
	detected := func(context.Context) (string, error) { return "detected", nil }
	unknown := func(context.Context) (string, error) { return "", nil }
	id := func(context.Context) (string, error) { return "uid-1234", nil }

	tests := []struct {
		name         string
		flag         string
		clusterName  func(context.Context) (string, error)
		expectedName string
		expectedErr  bool
	}{
		{name: "flag overrides detected name", flag: "from-flag", clusterName: detected, expectedName: "from-flag"},
		{name: "detected name", clusterName: detected, expectedName: "detected"},
		{name: "falls back to cluster ID", clusterName: unknown, expectedName: "uid-1234"},
		{
			name:        "detection error",
			clusterName: func(context.Context) (string, error) { return "", errors.New("forbidden") },
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusterName = tt.flag
			t.Cleanup(func() { clusterName = "" })

			name, clusterID, err := clusterIdentity(context.Background(), &mockedK8sClient{clusterName: tt.clusterName, clusterID: id})
			if tt.expectedErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedName, name)
			assert.Equal(t, "uid-1234", clusterID)
		})
	}
}
