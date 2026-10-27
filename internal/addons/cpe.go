package addons

import (
	"regexp"
	"strings"
)

// cpeProducts maps add-ons to their NVD CPE vendor and product. Add-ons are only listed when the
// detected image tag is the product version, e.g. Flux is detected from the source-controller
// image, whose version is not the Flux version.
var cpeProducts = map[string]string{ //nolint:gosec // CPE names, not credentials
	"argocd":                   "argoproj:argo_cd",
	"calico":                   "tigera:calico",
	"cert-manager":             "linuxfoundation:cert-manager",
	"cilium":                   "cilium:cilium",
	"coredns":                  "coredns.io:coredns",
	"flannel":                  "flannel-io:flannel",
	"ingress-nginx":            "kubernetes:ingress-nginx",
	"istio":                    "istio:istio",
	"kube-state-metrics":       "kubernetes:kube-state-metrics",
	"linkerd":                  "linkerd:linkerd",
	"secrets-store-csi-driver": "kubernetes:secrets_store_csi_driver",
	"traefik":                  "traefik:traefik",
}

// KubernetesCPEProduct is the NVD CPE vendor and product of Kubernetes.
const KubernetesCPEProduct = "kubernetes:kubernetes"

var semverRe = regexp.MustCompile(`^v?(\d+\.\d+(?:\.\d+)?)(?:[-+].*)?$`)

// CPE returns the CPE 2.3 name of the application product ("vendor:product") in version, e.g.
// cpe:2.3:a:coredns.io:coredns:1.11.4:*:*:*:*:*:*:*. Build suffixes of the version
// (v1.11.4-eksbuild.2) are dropped. It returns "" if product is empty or version is not a
// release version (e.g. latest).
func CPE(product, version string) string {
	if product == "" {
		return ""
	}

	m := semverRe.FindStringSubmatch(version)
	if m == nil {
		return ""
	}

	return "cpe:2.3:a:" + product + ":" + m[1] + strings.Repeat(":*", 7)
}
