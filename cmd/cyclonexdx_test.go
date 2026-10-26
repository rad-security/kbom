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
