package progress

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// syncBuffer guards bytes.Buffer, which the render goroutine writes to concurrently.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func TestNonInteractive(t *testing.T) {
	buf := &syncBuffer{}
	r := New(buf, false)

	s := r.Start("Listing resources")
	s.Update(1, 2, "pods")
	s.Done("Found 2 resource types")

	s = r.Start("Listing nodes")
	s.Fail()

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Len(t, lines, 2)
	assert.Regexp(t, `^✓ Found 2 resource types \(.+\)$`, lines[0])
	assert.Regexp(t, `^✗ Listing nodes \(.+\)$`, lines[1])
	assert.NotContains(t, buf.String(), "\033[")
}

func TestInteractive(t *testing.T) {
	buf := &syncBuffer{}
	r := New(buf, true)

	s := r.Start("Listing resources")
	s.Update(5, 10, "pods")
	time.Sleep(3 * refreshInterval)
	s.Done("Found 10 resource types")
	s.Done("ignored second call")

	got := buf.String()
	assert.Contains(t, got, "Listing resources [██████████░░░░░░░░░░] 5/10 pods")
	assert.Regexp(t, `\r\x1b\[K✓ Found 10 resource types \(.+\)\n$`, got)
	assert.Equal(t, 1, strings.Count(got, "✓"))
}

func TestNilWriter(t *testing.T) {
	s := New(nil, true).Start("step")
	s.Update(1, 1, "")
	s.Done("done")
}
