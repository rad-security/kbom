package cmd

import (
	"bytes"
	"testing"
)

func TestRoot(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	Execute()

	// Check if output contains expected string
	expected := "KBOM - Kubernetes Bill of Materials"
	if !bytes.Contains(buf.Bytes(), []byte(expected)) {
		t.Errorf("Execute() output = %q, want %q", buf.String(), expected)
	}
}
