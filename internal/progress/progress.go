// Package progress reports the progress of long running steps to the user.
//
// In interactive mode a spinner (and a progress bar for steps with a known total) is
// redrawn in place. Otherwise a single line is printed when a step finishes, so logs
// of non-interactive runs (CI, pipes) stay readable.
package progress

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

const (
	refreshInterval = 100 * time.Millisecond
	barWidth        = 20

	clearLine = "\r\033[K"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Reporter struct {
	mu          sync.Mutex // serializes writes of steps and logs
	w           io.Writer
	interactive bool
}

// New creates a Reporter writing to w. A nil w disables all output.
func New(w io.Writer, interactive bool) *Reporter {
	if w == nil {
		w = io.Discard
	}

	return &Reporter{w: w, interactive: interactive}
}

// Write lets other output, such as logs, be printed while a step is running. In
// interactive mode it clears the spinner line first; the spinner is redrawn on the next tick.
func (r *Reporter) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.interactive {
		if _, err := io.WriteString(r.w, clearLine); err != nil {
			return 0, err
		}
	}

	return r.w.Write(p)
}

func (r *Reporter) print(text string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	fmt.Fprint(r.w, text)
}

// Start begins a new step. Every step must be finished with Done or Fail.
func (r *Reporter) Start(title string) *Step {
	s := &Step{
		r:       r,
		title:   title,
		started: time.Now(),
		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
	}

	if r.interactive {
		go s.render()
	} else {
		close(s.stopped)
	}

	return s
}

type Step struct {
	r       *Reporter
	title   string
	started time.Time

	mu     sync.Mutex
	done   int
	total  int
	detail string

	stop     chan struct{}
	stopped  chan struct{}
	finished sync.Once
}

// Update sets the step progress. It is safe to call from multiple goroutines.
func (s *Step) Update(done, total int, detail string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.done, s.total, s.detail = done, total, detail
}

// Done finishes the step successfully with a short summary.
func (s *Step) Done(summary string) {
	s.finish("✓", summary)
}

// Fail finishes the step unsuccessfully.
func (s *Step) Fail() {
	s.finish("✗", s.title)
}

func (s *Step) finish(mark, summary string) {
	s.finished.Do(func() {
		close(s.stop)
		<-s.stopped

		prefix := ""
		if s.r.interactive {
			prefix = clearLine
		}

		s.r.print(fmt.Sprintf("%s%s %s (%s)\n", prefix, mark, summary, formatDuration(time.Since(s.started))))
	})
}

func (s *Step) render() {
	defer close(s.stopped)

	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	for frame := 0; ; frame++ {
		s.r.print(clearLine + s.line(spinnerFrames[frame%len(spinnerFrames)]))

		select {
		case <-s.stop:
			return
		case <-ticker.C:
		}
	}
}

func (s *Step) line(spinner string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	var b strings.Builder
	b.WriteString(spinner)
	b.WriteString(" ")
	b.WriteString(s.title)

	if s.total > 0 {
		filled := barWidth * s.done / s.total
		fmt.Fprintf(&b, " [%s%s] %d/%d", strings.Repeat("█", filled), strings.Repeat("░", barWidth-filled), s.done, s.total)
	}

	if s.detail != "" {
		b.WriteString(" ")
		b.WriteString(s.detail)
	}

	return b.String()
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return d.Round(time.Millisecond).String()
	}

	return d.Round(100 * time.Millisecond).String()
}
