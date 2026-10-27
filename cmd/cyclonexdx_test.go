package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rad-security/kbom/internal/model"
)

func TestTransformToCycloneDXWithoutLocation(t *testing.T) {
	kbom := &model.KBOM{Cluster: model.Cluster{Name: "test-cluster", K8sVersion: "1.35.8"}}

	assert.NotPanics(t, func() {
		bom := transformToCycloneDXBOM(kbom)
		assert.NotNil(t, bom.Metadata.Component)
	})
}

func TestTransformToCycloneDXUniqueResourceRefs(t *testing.T) {
	kbom := &model.KBOM{Cluster: model.Cluster{
		Name:       "test-cluster",
		K8sVersion: "1.35.8",
		Components: model.Components{Resources: map[string]model.ResourceList{
			"clusterroles": {
				Kind:       "ClusterRole",
				APIVersion: "rbac.authorization.k8s.io/v1",
				Resources:  []model.Resource{{Name: "admin"}},
			},
			"clusterrolebindings": {
				Kind:       "ClusterRoleBinding",
				APIVersion: "rbac.authorization.k8s.io/v1",
				Resources:  []model.Resource{{Name: "admin"}},
			},
		}},
	}}

	bom := transformToCycloneDXBOM(kbom)

	refs := map[string]bool{}
	for _, c := range *bom.Components {
		assert.False(t, refs[c.BOMRef], "duplicate bom-ref %s for %s", c.BOMRef, c.Name)
		refs[c.BOMRef] = true
	}
	assert.Len(t, refs, 2)
}
