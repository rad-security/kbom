package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ksoclabs/kbom/internal/kube"
	"github.com/ksoclabs/kbom/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestGenerateKBOM(t *testing.T) {
	type testCase struct {
		name string

		// mocks
		clientMock kube.K8sClient
		idMock     string
		timeMock   string

		// flags
		output string
		format string

		expectedOut string
		expectedErr error
	}

	testCases := []testCase{
		{
			name: "metadata error",
			clientMock: &mockedK8sClient{
				metadata: func(context.Context) (string, string, error) {
					return "", "", fmt.Errorf("metadata error")
				},
			},
			expectedErr: fmt.Errorf("metadata error"),
		},
		{
			name: "location error",
			clientMock: &mockedK8sClient{
				location: func(context.Context) (*model.Location, error) {
					return nil, fmt.Errorf("location error")
				},
			},
			expectedErr: fmt.Errorf("location error"),
		},
		{
			name: "all nodes error",
			clientMock: &mockedK8sClient{
				allNodes: func(context.Context, bool) ([]model.Node, error) {
					return nil, fmt.Errorf("all nodes error")
				},
			},
			expectedErr: fmt.Errorf("all nodes error"),
		},
		{
			name: "all resources error",
			clientMock: &mockedK8sClient{
				allResources: func(context.Context, bool) (map[string]model.ResourceList, error) {
					return nil, fmt.Errorf("all resources error")
				},
			},
			expectedErr: fmt.Errorf("all resources error"),
		},
		{
			name: "all images error",
			clientMock: &mockedK8sClient{
				allImages: func(context.Context) ([]model.Image, error) {
					return nil, fmt.Errorf("all images error")
				},
			},
			expectedErr: fmt.Errorf("all images error"),
		},
		{
			name:        "print KBOM - stdout - wrong format",
			clientMock:  &mockedK8sClient{},
			timeMock:    "2023-04-26T10:00:00.000000+00:00",
			idMock:      "00000001",
			format:      "wrong",
			expectedErr: fmt.Errorf("format \"wrong\" is not supported"),
		},
		{
			name:        "print KBOM - wrong output - JSON",
			clientMock:  &mockedK8sClient{},
			timeMock:    "2023-04-26T10:00:00.000000+00:00",
			idMock:      "00000001",
			output:      "wrong",
			expectedErr: fmt.Errorf("output \"wrong\" is not supported"),
		},
		{
			name: "print full KBOM - stdout - json",
			clientMock: &mockedK8sClient{
				clusterName: func(context.Context) (string, error) {
					return "test-cluster", nil
				},
				metadata: func(context.Context) (string, string, error) {
					return "012345678", "1.25.1", nil
				},
				location: func(context.Context) (*model.Location, error) {
					return &model.Location{
						Name:   "aws",
						Region: "us-east-1",
						Zone:   "us-east-1a",
					}, nil
				},
				allNodes: func(context.Context, bool) ([]model.Node, error) {
					return []model.Node{
						{
							Name:     "ip-10-0-65-00.us-east-1.compute.internal",
							Type:     "t3.small",
							Hostname: "ip-10-0-65-00.us-east-1.compute.internal",
							Capacity: &model.Capacity{
								CPU:              "2",
								Memory:           "1970512Ki",
								Pods:             "11",
								EphemeralStorage: "524275692Ki",
							},
							Allocatable: &model.Capacity{
								CPU:              "1930m",
								Memory:           "1483088Ki",
								Pods:             "11",
								EphemeralStorage: "482098735124",
							},
							Labels: map[string]string{
								"beta.kubernetes.io/arch":          "amd64",
								"beta.kubernetes.io/instance-type": "t3.small",
								"beta.kubernetes.io/os":            "linux",
								"topology.kubernetes.io/region":    "us-west-2",
								"topology.kubernetes.io/zone":      "us-west-2a",
							},
							Annotations: map[string]string{
								"node.alpha.kubernetes.io/ttl": "0",
							},
							MachineID:               "00001",
							Architecture:            "amd64",
							ContainerRuntimeVersion: "containerd://1.6.8+bottlerocket",
							BootID:                  "00001",
							KernelVersion:           "5.15.59",
							KubeProxyVersion:        "v1.24.6",
							KubeletVersion:          "v1.24.6",
							OperatingSystem:         "linux",
							OsImage:                 "Bottlerocket OS 1.11.1 (aws-k8s-1.24)",
						},
						{
							Name:     "ip-10-0-65-01.us-east-1.compute.internal",
							Type:     "t3.small",
							Hostname: "ip-10-0-65-01.us-east-1.compute.internal",
							Capacity: &model.Capacity{
								CPU:              "2",
								Memory:           "1970512Ki",
								Pods:             "11",
								EphemeralStorage: "524275692Ki",
							},
							Allocatable: &model.Capacity{
								CPU:              "1930m",
								Memory:           "1483088Ki",
								Pods:             "11",
								EphemeralStorage: "482098735124",
							},
							Labels: map[string]string{
								"beta.kubernetes.io/arch":          "amd64",
								"beta.kubernetes.io/instance-type": "t3.small",
								"beta.kubernetes.io/os":            "linux",
								"topology.kubernetes.io/region":    "us-west-2",
								"topology.kubernetes.io/zone":      "us-west-2a",
							},
							Annotations: map[string]string{
								"node.alpha.kubernetes.io/ttl": "0",
							},
							MachineID:               "00002",
							Architecture:            "amd64",
							ContainerRuntimeVersion: "containerd://1.6.8+bottlerocket",
							BootID:                  "00002",
							KernelVersion:           "5.15.59",
							KubeProxyVersion:        "v1.24.6",
							KubeletVersion:          "v1.24.6",
							OperatingSystem:         "linux",
							OsImage:                 "Bottlerocket OS 1.11.1 (aws-k8s-1.24)",
						},
					}, nil
				},
				allImages: func(context.Context) ([]model.Image, error) {
					return []model.Image{
						{
							Name:      "nginx",
							Version:   "1.17.1",
							FullName:  "nginx:1.17.1",
							Digest:    "sha256:0000000000000000000000000000000000000000000000000000000000000001",
							Namespace: "default",
						},
						{
							Name:      "redis",
							Version:   "7.0.1",
							FullName:  "redis:7.0.1",
							Digest:    "sha256:0000000000000000000000000000000000000000000000000000000000000002",
							Namespace: "default",
						},
					}, nil
				},
				allResources: func(context.Context, bool) (map[string]model.ResourceList, error) {
					return map[string]model.ResourceList{
						"/v1, Resource=namespaces": {
							Kind:           "Namespace",
							APIVersion:     "v1",
							Namespaced:     false,
							ResourcesCount: 2,
							Resources: []model.Resource{
								{
									Name: "backend",
								},
								{
									Name: "frontend",
								},
							},
						},
					}, nil
				},
			},
			timeMock:    "2023-04-26T10:00:00.000000+00:00",
			idMock:      "00000001",
			expectedErr: nil,
			expectedOut: expectedOutJSON,
		},
		{
			name:        "print KBOM - stdout - yaml",
			clientMock:  &mockedK8sClient{},
			timeMock:    "2023-04-26T10:00:00.000000+00:00",
			idMock:      "00000001",
			format:      YAMLFormat.Name,
			expectedOut: expectedOutYAML,
		},
		{
			name:        "print KBOM - file - yaml",
			clientMock:  &mockedK8sClient{},
			timeMock:    "2023-04-26T10:00:00.000000+00:00",
			idMock:      "00000001",
			format:      YAMLFormat.Name,
			output:      FileOutput,
			expectedOut: expectedOutYAML,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &stdoutMock{buf: bytes.Buffer{}}
			out = mock
			kbomID = tc.idMock
			if tc.timeMock != "" {
				mockedTime, err := time.Parse(time.RFC3339, tc.timeMock)
				assert.NoError(t, err)
				generatedAt = mockedTime
			}

			if tc.format != "" {
				format = tc.format
			} else {
				format = JSONFormat.Name
			}

			if tc.output != "" {
				output = tc.output
			} else {
				output = StdOutput
			}

			err := generateKBOM(tc.clientMock)
			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			if output == FileOutput {
				filename := fmt.Sprintf("kbom-%s-2023-04-26-10-00-00.%s", mockCACert[:8], format)
				assert.FileExists(t, filename)
				file, err := os.Open(filename)
				assert.NoError(t, err)

				buf := new(bytes.Buffer)
				_, err = buf.ReadFrom(file)
				assert.NoError(t, err)
				file.Close()

				assert.Equal(t, tc.expectedOut, buf.String())
				assert.NoError(t, os.Remove(filename))
			} else {
				assert.Equal(t, tc.expectedOut, mock.buf.String())
			}
		})
	}
}

type mockedK8sClient struct {
	clusterName  func(context.Context) (string, error)
	metadata     func(context.Context) (string, string, error)
	location     func(context.Context) (*model.Location, error)
	allImages    func(context.Context) ([]model.Image, error)
	allNodes     func(context.Context, bool) ([]model.Node, error)
	allResources func(context.Context, bool) (map[string]model.ResourceList, error)
}

func (m *mockedK8sClient) ClusterName(ctx context.Context) (clusterName string, err error) {
	if m.clusterName == nil {
		return "test-cluster", nil
	}
	return m.clusterName(ctx)
}

func (m *mockedK8sClient) Metadata(ctx context.Context) (ver, ca string, err error) {
	if m.metadata == nil {
		return "1.25.1", mockCACert, nil
	}
	return m.metadata(ctx)
}

func (m *mockedK8sClient) Location(ctx context.Context) (*model.Location, error) {
	if m.location == nil {
		return nil, nil
	}
	return m.location(ctx)
}

func (m *mockedK8sClient) AllImages(ctx context.Context) ([]model.Image, error) {
	if m.allImages == nil {
		return nil, nil
	}
	return m.allImages(ctx)
}

func (m *mockedK8sClient) AllNodes(ctx context.Context, full bool) ([]model.Node, error) {
	if m.allNodes == nil {
		return nil, nil
	}
	return m.allNodes(ctx, full)
}

func (m *mockedK8sClient) AllResources(ctx context.Context, full bool) (map[string]model.ResourceList, error) {
	if m.allResources == nil {
		return nil, nil
	}
	return m.allResources(ctx, full)
}

var mockCACert = "1234567890"

var expectedOutJSON = `{
  "id": "00000001",
  "bom_format": "ksoc",
  "spec_version": "0.2",
  "generated_at": "2023-04-26T10:00:00Z",
  "generated_by": {
    "vendor": "KSOC Labs",
    "name": "unknown",
    "build_time": "unknown",
    "version": "unknown",
    "commit": "unknown",
    "commit_time": "unknown"
  },
  "cluster": {
    "name": "test-cluster",
    "ca_cert_digest": "1.25.1",
    "k8s_version": "012345678",
    "location": {
      "name": "aws",
      "region": "us-east-1",
      "zone": "us-east-1a"
    },
    "nodes_count": 2,
    "nodes": [
      {
        "name": "ip-10-0-65-00.us-east-1.compute.internal",
        "type": "t3.small",
        "hostname": "ip-10-0-65-00.us-east-1.compute.internal",
        "capacity": {
          "cpu": "2",
          "memory": "1970512Ki",
          "pods": "11",
          "ephemeral_storage": "524275692Ki"
        },
        "allocatable": {
          "cpu": "1930m",
          "memory": "1483088Ki",
          "pods": "11",
          "ephemeral_storage": "482098735124"
        },
        "labels": {
          "beta.kubernetes.io/arch": "amd64",
          "beta.kubernetes.io/instance-type": "t3.small",
          "beta.kubernetes.io/os": "linux",
          "topology.kubernetes.io/region": "us-west-2",
          "topology.kubernetes.io/zone": "us-west-2a"
        },
        "annotations": {
          "node.alpha.kubernetes.io/ttl": "0"
        },
        "machine_id": "00001",
        "architecture": "amd64",
        "container_runtime_version": "containerd://1.6.8+bottlerocket",
        "boot_id": "00001",
        "kernel_version": "5.15.59",
        "kube_proxy_version": "v1.24.6",
        "kubelet_version": "v1.24.6",
        "operating_system": "linux",
        "os_image": "Bottlerocket OS 1.11.1 (aws-k8s-1.24)"
      },
      {
        "name": "ip-10-0-65-01.us-east-1.compute.internal",
        "type": "t3.small",
        "hostname": "ip-10-0-65-01.us-east-1.compute.internal",
        "capacity": {
          "cpu": "2",
          "memory": "1970512Ki",
          "pods": "11",
          "ephemeral_storage": "524275692Ki"
        },
        "allocatable": {
          "cpu": "1930m",
          "memory": "1483088Ki",
          "pods": "11",
          "ephemeral_storage": "482098735124"
        },
        "labels": {
          "beta.kubernetes.io/arch": "amd64",
          "beta.kubernetes.io/instance-type": "t3.small",
          "beta.kubernetes.io/os": "linux",
          "topology.kubernetes.io/region": "us-west-2",
          "topology.kubernetes.io/zone": "us-west-2a"
        },
        "annotations": {
          "node.alpha.kubernetes.io/ttl": "0"
        },
        "machine_id": "00002",
        "architecture": "amd64",
        "container_runtime_version": "containerd://1.6.8+bottlerocket",
        "boot_id": "00002",
        "kernel_version": "5.15.59",
        "kube_proxy_version": "v1.24.6",
        "kubelet_version": "v1.24.6",
        "operating_system": "linux",
        "os_image": "Bottlerocket OS 1.11.1 (aws-k8s-1.24)"
      }
    ],
    "components": {
      "images": [
        {
          "full_name": "nginx:1.17.1",
          "name": "nginx",
          "version": "1.17.1",
          "digest": "sha256:0000000000000000000000000000000000000000000000000000000000000001",
          "namespace": "default"
        },
        {
          "full_name": "redis:7.0.1",
          "name": "redis",
          "version": "7.0.1",
          "digest": "sha256:0000000000000000000000000000000000000000000000000000000000000002",
          "namespace": "default"
        }
      ],
      "resources": {
        "/v1, Resource=namespaces": {
          "kind": "Namespace",
          "api_version": "v1",
          "namespaced": false,
          "count": 2,
          "resources": [
            {
              "name": "backend"
            },
            {
              "name": "frontend"
            }
          ]
        }
      }
    }
  }
}
`
var expectedOutYAML = `id: "00000001"
bomformat: ksoc
specversion: "0.2"
generatedat: 2023-04-26T10:00:00Z
generatedby:
  vendor: KSOC Labs
  name: unknown
  buildtime: unknown
  version: unknown
  commit: unknown
  committime: unknown
cluster:
  name: test-cluster
  cacertdigest: "1234567890"
  k8sversion: 1.25.1
  cniversion: ""
  location: null
  nodescount: 0
  nodes: []
  components:
    images: []
    resources: {}
`
