package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunGenerateSchema(t *testing.T) {
	mock := &stdoutMock{buf: bytes.Buffer{}}
	out = mock

	err := runGenerateSchema(nil, []string{})
	assert.NoError(t, err)

	assert.Equal(t, expectedSchema, mock.buf.String())
}

type stdoutMock struct {
	buf bytes.Buffer
}

func (m *stdoutMock) Write(p []byte) (n int, err error) {
	return m.buf.Write(p)
}

func (m *stdoutMock) Close() error {
	return nil
}

var expectedSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://github.com/rad-security/kbom/internal/model/kbom",
  "$ref": "#/$defs/KBOM",
  "$defs": {
    "Capacity": {
      "properties": {
        "cpu": {
          "type": "string"
        },
        "memory": {
          "type": "string"
        },
        "pods": {
          "type": "string"
        },
        "ephemeral_storage": {
          "type": "string"
        }
      },
      "additionalProperties": false,
      "type": "object",
      "required": [
        "cpu",
        "memory",
        "pods",
        "ephemeral_storage"
      ]
    },
    "Cluster": {
      "properties": {
        "name": {
          "type": "string"
        },
        "ca_cert_digest": {
          "type": "string"
        },
        "k8s_version": {
          "type": "string"
        },
        "cni_version": {
          "type": "string"
        },
        "location": {
          "$ref": "#/$defs/Location"
        },
        "nodes_count": {
          "type": "integer"
        },
        "nodes": {
          "items": {
            "$ref": "#/$defs/Node"
          },
          "type": "array"
        },
        "components": {
          "$ref": "#/$defs/Components"
        }
      },
      "additionalProperties": false,
      "type": "object",
      "required": [
        "name",
        "ca_cert_digest",
        "k8s_version",
        "location",
        "nodes_count",
        "nodes",
        "components"
      ]
    },
    "Components": {
      "properties": {
        "images": {
          "items": {
            "$ref": "#/$defs/Image"
          },
          "type": "array"
        },
        "resources": {
          "additionalProperties": {
            "$ref": "#/$defs/ResourceList"
          },
          "type": "object"
        }
      },
      "additionalProperties": false,
      "type": "object",
      "required": [
        "resources"
      ]
    },
    "Image": {
      "properties": {
        "full_name": {
          "type": "string"
        },
        "name": {
          "type": "string"
        },
        "version": {
          "type": "string"
        },
        "digest": {
          "type": "string"
        }
      },
      "additionalProperties": false,
      "type": "object",
      "required": [
        "full_name",
        "name",
        "version",
        "digest"
      ]
    },
    "KBOM": {
      "properties": {
        "id": {
          "type": "string"
        },
        "bom_format": {
          "type": "string"
        },
        "spec_version": {
          "type": "string"
        },
        "generated_at": {
          "type": "string",
          "format": "date-time"
        },
        "generated_by": {
          "$ref": "#/$defs/Tool"
        },
        "cluster": {
          "$ref": "#/$defs/Cluster"
        }
      },
      "additionalProperties": false,
      "type": "object",
      "required": [
        "id",
        "bom_format",
        "spec_version",
        "generated_at",
        "generated_by",
        "cluster"
      ]
    },
    "Location": {
      "properties": {
        "name": {
          "type": "string"
        },
        "region": {
          "type": "string"
        },
        "zone": {
          "type": "string"
        }
      },
      "additionalProperties": false,
      "type": "object",
      "required": [
        "name",
        "region",
        "zone"
      ]
    },
    "Node": {
      "properties": {
        "name": {
          "type": "string"
        },
        "type": {
          "type": "string"
        },
        "hostname": {
          "type": "string"
        },
        "capacity": {
          "$ref": "#/$defs/Capacity"
        },
        "allocatable": {
          "$ref": "#/$defs/Capacity"
        },
        "labels": {
          "additionalProperties": {
            "type": "string"
          },
          "type": "object"
        },
        "annotations": {
          "additionalProperties": {
            "type": "string"
          },
          "type": "object"
        },
        "machine_id": {
          "type": "string"
        },
        "architecture": {
          "type": "string"
        },
        "container_runtime_version": {
          "type": "string"
        },
        "boot_id": {
          "type": "string"
        },
        "kernel_version": {
          "type": "string"
        },
        "kube_proxy_version": {
          "type": "string"
        },
        "kubelet_version": {
          "type": "string"
        },
        "operating_system": {
          "type": "string"
        },
        "os_image": {
          "type": "string"
        }
      },
      "additionalProperties": false,
      "type": "object",
      "required": [
        "name",
        "type",
        "hostname",
        "capacity",
        "allocatable",
        "labels",
        "annotations",
        "machine_id",
        "architecture",
        "container_runtime_version",
        "boot_id",
        "kernel_version",
        "kube_proxy_version",
        "kubelet_version",
        "operating_system",
        "os_image"
      ]
    },
    "Resource": {
      "properties": {
        "kind": {
          "type": "string"
        },
        "api_version": {
          "type": "string"
        },
        "name": {
          "type": "string"
        },
        "namespace": {
          "type": "string"
        },
        "additional_properties": {
          "additionalProperties": {
            "type": "string"
          },
          "type": "object"
        }
      },
      "additionalProperties": false,
      "type": "object",
      "required": [
        "name"
      ]
    },
    "ResourceList": {
      "properties": {
        "kind": {
          "type": "string"
        },
        "api_version": {
          "type": "string"
        },
        "namespaced": {
          "type": "boolean"
        },
        "count": {
          "type": "integer"
        },
        "resources": {
          "items": {
            "$ref": "#/$defs/Resource"
          },
          "type": "array"
        }
      },
      "additionalProperties": false,
      "type": "object",
      "required": [
        "kind",
        "api_version",
        "namespaced",
        "count"
      ]
    },
    "Tool": {
      "properties": {
        "vendor": {
          "type": "string"
        },
        "name": {
          "type": "string"
        },
        "build_time": {
          "type": "string"
        },
        "version": {
          "type": "string"
        },
        "commit": {
          "type": "string"
        },
        "commit_time": {
          "type": "string"
        }
      },
      "additionalProperties": false,
      "type": "object",
      "required": [
        "vendor",
        "name",
        "build_time",
        "version",
        "commit",
        "commit_time"
      ]
    }
  }
}
`
