package kube

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestClusterNameFromKubeconfig(t *testing.T) {
	tests := map[string]string{
		"arn:aws:eks:us-west-2:652031173150:cluster/pe-sbx-us-west-2":       "pe-sbx-us-west-2",
		"arn:aws-us-gov:eks:us-gov-west-1:123456789012:cluster/gov-cluster": "gov-cluster",
		"gke_my-project_europe-central2_test":                               "test",
		"gke_my-project_us-central1-a_prod-cluster":                         "prod-cluster",
		"sandbox-cluster-1.us-west-2.eksctl.io":                             "sandbox-cluster-1",
		"kind-sbx-workflows":                                                "",
		"my-cluster":                                                        "",
		"":                                                                  "",
	}

	for cluster, expected := range tests {
		t.Run(cluster, func(t *testing.T) {
			assert.Equal(t, expected, clusterNameFromKubeconfig(cluster))
		})
	}
}

func TestClusterName(t *testing.T) {
	eksctlNode := &v1.Node{ObjectMeta: metav1.ObjectMeta{
		Name:   "node-1",
		Labels: map[string]string{eksctlClusterNameLabel: "label-cluster"},
	}}

	tests := []struct {
		name              string
		k8sContext        string
		kubeconfigCluster string
		nodes             []*v1.Node
		expected          string
	}{
		{
			name:              "known kubeconfig cluster format",
			k8sContext:        "my-alias",
			kubeconfigCluster: "arn:aws:eks:us-west-2:652031173150:cluster/pe-sbx-us-west-2",
			expected:          "pe-sbx-us-west-2",
		},
		{
			name:              "unknown kubeconfig cluster format uses context",
			k8sContext:        "kind-test",
			kubeconfigCluster: "kind-test",
			expected:          "kind-test",
		},
		{
			name:     "in-cluster uses eksctl node label",
			nodes:    []*v1.Node{eksctlNode},
			expected: "label-cluster",
		},
		{
			name:     "in-cluster without label is unknown",
			nodes:    []*v1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "node-1"}}},
			expected: "",
		},
		{
			name:     "in-cluster without nodes is unknown",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewClientset()
			for _, n := range tt.nodes {
				_, err := client.CoreV1().Nodes().Create(context.Background(), n, metav1.CreateOptions{})
				require.NoError(t, err)
			}

			k := &k8sDB{client: client, k8sContext: tt.k8sContext, kubeconfigCluster: tt.kubeconfigCluster}
			name, err := k.ClusterName(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.expected, name)
		})
	}
}

func TestClusterID(t *testing.T) {
	k := &k8sDB{client: fake.NewClientset(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kube-system", UID: "uid-1234"}})}
	id, err := k.ClusterID(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "uid-1234", id)

	k = &k8sDB{client: fake.NewClientset()}
	id, err = k.ClusterID(context.Background())
	require.NoError(t, err)
	assert.Empty(t, id)
}

func TestNewClientKubeconfigCluster(t *testing.T) {
	dir := t.TempDir()
	path := writeKubeconfig(t, dir, "my-alias", "https://example.com")

	t.Setenv("KUBERNETES_SERVICE_HOST", "")
	t.Setenv("KUBECONFIG", path)

	client, err := NewClient("")
	require.NoError(t, err)
	assert.Equal(t, "my-alias", client.(*k8sDB).kubeconfigCluster)
}
