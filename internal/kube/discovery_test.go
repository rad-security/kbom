package kube

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	fakediscovery "k8s.io/client-go/discovery/fake"
)

type stubDiscovery struct {
	fakediscovery.FakeDiscovery

	resources []*metav1.APIResourceList
	err       error
}

func (s *stubDiscovery) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	return s.resources, s.err
}

func TestPreferredResources(t *testing.T) {
	pods := []*metav1.APIResourceList{{GroupVersion: "v1", APIResources: []metav1.APIResource{{Name: "pods"}}}}
	groupFailed := &discovery.ErrGroupDiscoveryFailed{
		Groups: map[schema.GroupVersion]error{
			{Group: "metrics.k8s.io", Version: "v1beta1"}: errors.New("the server is currently unable to handle the request"),
		},
	}

	tests := []struct {
		name        string
		resources   []*metav1.APIResourceList
		err         error
		expected    []*metav1.APIResourceList
		expectedErr bool
	}{
		{
			name:      "all groups available",
			resources: pods,
			expected:  pods,
		},
		{
			name:      "some groups unavailable",
			resources: pods,
			err:       groupFailed,
			expected:  pods,
		},
		{
			name:        "all groups unavailable",
			err:         groupFailed,
			expectedErr: true,
		},
		{
			name:        "other error",
			resources:   pods,
			err:         errors.New("connection refused"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := preferredResources(&stubDiscovery{resources: tt.resources, err: tt.err})
			if tt.expectedErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}
