// Package addons detects well-known cluster add-ons (CNI, DNS, CSI drivers, ingress controllers, ...)
// and their versions from the container images running in the cluster.
package addons

import (
	"cmp"
	"slices"
	"strings"

	"github.com/distribution/reference"

	"github.com/rad-security/kbom/internal/model"
)

const (
	CategoryCNI           = "cni"
	CategoryDNS           = "dns"
	CategoryProxy         = "proxy"
	CategoryMetrics       = "metrics"
	CategoryCSI           = "csi"
	CategoryIngress       = "ingress"
	CategoryServiceMesh   = "service-mesh"
	CategorySecurity      = "security"
	CategoryAutoscaling   = "autoscaling"
	CategoryGitOps        = "gitops"
	CategoryObservability = "observability"
)

type addon struct {
	name     string
	category string
	// images are repository path suffixes (without registry) of the image that identifies the add-on
	images []string
}

// catalog of known add-ons, every add-on is identified by its main image only, so sidecars and
// helper images (e.g. CSI livenessprobe) don't create duplicates
var catalog = []addon{
	{"aws-vpc-cni", CategoryCNI, []string{"amazon-k8s-cni"}},
	{"azure-cni", CategoryCNI, []string{"containernetworking/azure-cns"}},
	{"calico", CategoryCNI, []string{"calico/node"}},
	{"cilium", CategoryCNI, []string{"cilium/cilium"}},
	{"flannel", CategoryCNI, []string{"flannel/flannel", "flannel-io/flannel", "rancher/mirrored-flannelcni-flannel"}},
	{"kindnet", CategoryCNI, []string{"kindest/kindnetd"}},
	{"weave", CategoryCNI, []string{"weaveworks/weave-kube"}},

	{"coredns", CategoryDNS, []string{"coredns"}},
	{"kube-dns", CategoryDNS, []string{"k8s-dns-kube-dns"}},

	{"kube-proxy", CategoryProxy, []string{"kube-proxy"}},

	{"metrics-server", CategoryMetrics, []string{"metrics-server"}},
	{"kube-state-metrics", CategoryMetrics, []string{"kube-state-metrics"}},

	{"aws-ebs-csi-driver", CategoryCSI, []string{"aws-ebs-csi-driver"}},
	{"aws-efs-csi-driver", CategoryCSI, []string{"aws-efs-csi-driver"}},
	{"azuredisk-csi-driver", CategoryCSI, []string{"azuredisk-csi"}},
	{"azurefile-csi-driver", CategoryCSI, []string{"azurefile-csi"}},
	{"gcp-pd-csi-driver", CategoryCSI, []string{"gcp-compute-persistent-disk-csi-driver"}},
	{"secrets-store-csi-driver", CategoryCSI, []string{"csi-secrets-store/driver"}},
	{"local-path-provisioner", CategoryCSI, []string{"local-path-provisioner"}},

	{"aws-load-balancer-controller", CategoryIngress, []string{"aws-load-balancer-controller"}},
	{"ingress-nginx", CategoryIngress, []string{"ingress-nginx/controller"}},
	{"traefik", CategoryIngress, []string{"traefik"}},

	{"istio", CategoryServiceMesh, []string{"istio/pilot"}},
	{"linkerd", CategoryServiceMesh, []string{"linkerd/controller"}},

	{"cert-manager", CategorySecurity, []string{"jetstack/cert-manager-controller"}},

	{"karpenter", CategoryAutoscaling, []string{"karpenter/controller"}},
	{"cluster-autoscaler", CategoryAutoscaling, []string{"autoscaling/cluster-autoscaler", "cluster-autoscaler"}},
	{"keda", CategoryAutoscaling, []string{"kedacore/keda"}},

	{"argocd", CategoryGitOps, []string{"argoproj/argocd"}},
	{"flux", CategoryGitOps, []string{"fluxcd/source-controller"}},

	{"prometheus-operator", CategoryObservability, []string{"prometheus-operator/prometheus-operator"}},
}

// Detect returns the known add-ons found in images, one entry per add-on and version, sorted by
// category, name and version.
func Detect(images []model.Image) []model.Addon {
	found := map[[2]string]*model.Addon{}
	for _, img := range images {
		a, ok := match(img.Name)
		if !ok {
			continue
		}

		key := [2]string{a.name, img.Version}
		if found[key] == nil {
			found[key] = &model.Addon{
				Name:     a.name,
				Category: a.category,
				Version:  img.Version,
				CPE:      CPE(cpeProducts[a.name], img.Version),
			}
		}
		found[key].Images = append(found[key].Images, img.FullName)
	}

	result := make([]model.Addon, 0, len(found))
	for _, a := range found {
		slices.Sort(a.Images)
		a.Images = slices.Compact(a.Images)
		result = append(result, *a)
	}

	slices.SortFunc(result, func(a, b model.Addon) int {
		return cmp.Or(cmp.Compare(a.Category, b.Category), cmp.Compare(a.Name, b.Name), cmp.Compare(a.Version, b.Version))
	})

	return result
}

// CNIVersion returns the version of the first detected CNI plugin.
func CNIVersion(addons []model.Addon) string {
	for _, a := range addons {
		if a.Category == CategoryCNI {
			return a.Version
		}
	}

	return ""
}

func match(imageName string) (addon, bool) {
	path := imageName
	if named, err := reference.ParseNormalizedNamed(imageName); err == nil {
		path = reference.Path(named)
	}

	for _, a := range catalog {
		for _, suffix := range a.images {
			if path == suffix || strings.HasSuffix(path, "/"+suffix) {
				return a, true
			}
		}
	}

	return addon{}, false
}
