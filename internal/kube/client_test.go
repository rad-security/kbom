package kube

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func TestLocationWithoutNodes(t *testing.T) {
	k := &k8sDB{client: fake.NewClientset()}

	loc, err := k.Location(context.Background())
	require.NoError(t, err)
	assert.Nil(t, loc)
}

func TestLocationFromNodeLabels(t *testing.T) {
	k := &k8sDB{client: fake.NewClientset(&v1.Node{ObjectMeta: metav1.ObjectMeta{
		Name: "node-1",
		Labels: map[string]string{
			"topology.kubernetes.io/region": "us-west-2",
			"topology.kubernetes.io/zone":   "us-west-2a",
		},
	}})}

	loc, err := k.Location(context.Background())
	require.NoError(t, err)
	require.NotNil(t, loc)
	assert.Equal(t, "us-west-2", loc.Region)
	assert.Equal(t, "us-west-2a", loc.Zone)
}

func TestMetadata(t *testing.T) {
	caData := []byte("ca")
	expectedDigest := fmt.Sprintf("%x", sha256.Sum256(caData))

	tests := []struct {
		name            string
		gitVersion      string
		expectedVersion string
		expectedErr     bool
	}{
		{name: "valid version", gitVersion: "v1.35.8-eks-abc", expectedVersion: "1.35.8"},
		{name: "invalid version", gitVersion: "not-a-version", expectedErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewClientset()
			client.Discovery().(*fakediscovery.FakeDiscovery).FakedServerVersion = &version.Info{GitVersion: tt.gitVersion}
			k := &k8sDB{client: client, cfg: &rest.Config{TLSClientConfig: rest.TLSClientConfig{CAData: caData}}}

			k8sVersion, caDigest, err := k.Metadata(context.Background())
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedVersion, k8sVersion)
			assert.Equal(t, expectedDigest, caDigest)
		})
	}
}
