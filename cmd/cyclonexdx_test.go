package cmd

import (
	"testing"

	cyclonedx "github.com/CycloneDX/cyclonedx-go"
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

func TestTransformToCycloneDXAddons(t *testing.T) {
	img := model.Image{
		FullName: "registry.k8s.io/coredns/coredns:v1.12.0",
		Name:     "registry.k8s.io/coredns/coredns",
		Version:  "v1.12.0",
	}
	kbom := &model.KBOM{Cluster: model.Cluster{
		Name:       "test-cluster",
		K8sVersion: "1.35.8",
		Components: model.Components{
			Images: []model.Image{img},
			Addons: []model.Addon{{Name: "coredns", Category: "dns", Version: "v1.12.0", Images: []string{img.FullName}}},
		},
	}}

	bom := transformToCycloneDXBOM(kbom)

	var addonRef string
	for _, c := range *bom.Components {
		if c.Name == "coredns" {
			addonRef = c.BOMRef
			assert.Equal(t, "v1.12.0", c.Version)
			assert.Contains(t, *c.Properties, cyclonedx.Property{Name: "rad:kbom:k8s:addon:category", Value: "dns"})
		}
	}
	assert.NotEmpty(t, addonRef, "addon component is missing")

	deps := map[string][]string{}
	for _, d := range *bom.Dependencies {
		deps[d.Ref] = *d.Dependencies
	}
	assert.Equal(t, []string{img.PkgID()}, deps[addonRef])
	assert.Contains(t, deps[kbom.Cluster.BOMRef()], addonRef)
}
