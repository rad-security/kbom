package addons

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rad-security/kbom/internal/model"
)

func image(fullName, name, version string) model.Image {
	return model.Image{FullName: fullName, Name: name, Version: version}
}

func TestDetect(t *testing.T) {
	images := []model.Image{
		image("602401143452.dkr.ecr.us-west-2.amazonaws.com/amazon-k8s-cni:v1.19.2-eksbuild.1",
			"602401143452.dkr.ecr.us-west-2.amazonaws.com/amazon-k8s-cni", "v1.19.2-eksbuild.1"),
		image("602401143452.dkr.ecr.us-west-2.amazonaws.com/amazon-k8s-cni-init:v1.19.2-eksbuild.1",
			"602401143452.dkr.ecr.us-west-2.amazonaws.com/amazon-k8s-cni-init", "v1.19.2-eksbuild.1"),
		image("602401143452.dkr.ecr.us-west-2.amazonaws.com/eks/coredns:v1.11.4-eksbuild.2",
			"602401143452.dkr.ecr.us-west-2.amazonaws.com/eks/coredns", "v1.11.4-eksbuild.2"),
		image("registry.k8s.io/kube-proxy:v1.35.0", "registry.k8s.io/kube-proxy", "v1.35.0"),
		image("quay.io/jetstack/cert-manager-controller:v1.19.3", "quay.io/jetstack/cert-manager-controller", "v1.19.3"),
		image("quay.io/jetstack/cert-manager-webhook:v1.19.3", "quay.io/jetstack/cert-manager-webhook", "v1.19.3"),
		image("registry.k8s.io/kube-state-metrics/kube-state-metrics:v2.17.0",
			"registry.k8s.io/kube-state-metrics/kube-state-metrics", "v2.17.0"),
		image("registry.k8s.io/kube-state-metrics/kube-state-metrics:v1.9.8",
			"registry.k8s.io/kube-state-metrics/kube-state-metrics", "v1.9.8"),
		image("traefik:v3.3", "docker.io/library/traefik", "v3.3"),
		image("docker.io/library/nginx:1.27", "docker.io/library/nginx", "1.27"),
		image("example.com/team/my-coredns-fork:1.0", "example.com/team/my-coredns-fork", "1.0"),
	}

	expected := []model.Addon{
		{Name: "aws-vpc-cni", Category: CategoryCNI, Version: "v1.19.2-eksbuild.1",
			Images: []string{"602401143452.dkr.ecr.us-west-2.amazonaws.com/amazon-k8s-cni:v1.19.2-eksbuild.1"}},
		{Name: "coredns", Category: CategoryDNS, Version: "v1.11.4-eksbuild.2",
			CPE:    "cpe:2.3:a:coredns.io:coredns:1.11.4:*:*:*:*:*:*:*",
			Images: []string{"602401143452.dkr.ecr.us-west-2.amazonaws.com/eks/coredns:v1.11.4-eksbuild.2"}},
		{Name: "traefik", Category: CategoryIngress, Version: "v3.3", CPE: "cpe:2.3:a:traefik:traefik:3.3:*:*:*:*:*:*:*",
			Images: []string{"traefik:v3.3"}},
		{Name: "kube-state-metrics", Category: CategoryMetrics, Version: "v1.9.8",
			CPE:    "cpe:2.3:a:kubernetes:kube-state-metrics:1.9.8:*:*:*:*:*:*:*",
			Images: []string{"registry.k8s.io/kube-state-metrics/kube-state-metrics:v1.9.8"}},
		{Name: "kube-state-metrics", Category: CategoryMetrics, Version: "v2.17.0",
			CPE:    "cpe:2.3:a:kubernetes:kube-state-metrics:2.17.0:*:*:*:*:*:*:*",
			Images: []string{"registry.k8s.io/kube-state-metrics/kube-state-metrics:v2.17.0"}},
		{Name: "kube-proxy", Category: CategoryProxy, Version: "v1.35.0", Images: []string{"registry.k8s.io/kube-proxy:v1.35.0"}},
		{Name: "cert-manager", Category: CategorySecurity, Version: "v1.19.3",
			CPE:    "cpe:2.3:a:linuxfoundation:cert-manager:1.19.3:*:*:*:*:*:*:*",
			Images: []string{"quay.io/jetstack/cert-manager-controller:v1.19.3"}},
	}

	assert.Equal(t, expected, Detect(images))
}

func TestDetectGroupsImagesOfSameVersion(t *testing.T) {
	images := []model.Image{
		image("registry.k8s.io/coredns/coredns:v1.12.0", "registry.k8s.io/coredns/coredns", "v1.12.0"),
		image("registry.k8s.io/coredns/coredns:v1.12.0@sha256:abc", "registry.k8s.io/coredns/coredns", "v1.12.0"),
	}

	assert.Equal(t, []model.Addon{{
		Name:     "coredns",
		Category: CategoryDNS,
		Version:  "v1.12.0",
		CPE:      "cpe:2.3:a:coredns.io:coredns:1.12.0:*:*:*:*:*:*:*",
		Images:   []string{"registry.k8s.io/coredns/coredns:v1.12.0", "registry.k8s.io/coredns/coredns:v1.12.0@sha256:abc"},
	}}, Detect(images))
}

func TestDetectNoImages(t *testing.T) {
	assert.Empty(t, Detect(nil))
}

func TestCNIVersion(t *testing.T) {
	assert.Empty(t, CNIVersion(nil))
	assert.Empty(t, CNIVersion([]model.Addon{{Name: "coredns", Category: CategoryDNS, Version: "v1.12.0"}}))
	assert.Equal(t, "v1.17.0", CNIVersion([]model.Addon{
		{Name: "coredns", Category: CategoryDNS, Version: "v1.12.0"},
		{Name: "cilium", Category: CategoryCNI, Version: "v1.17.0"},
	}))
}
