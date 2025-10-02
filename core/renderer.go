package core

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

// Renderer is the rendering abstraction used by Session.
// Implementations should be safe to call from a single goroutine.
// (Session guarantees serialized calls to Render.)
type Renderer interface {
	// Clear prepares the terminal (hide cursor, clear, move home).
	Clear()
	// Render paints the given view to the terminal.
	Render(s string)
	// Close restores terminal state (e.g., show cursor).
	Close()
}

// ----------------------------------------------------------------------------
// Options

// RendererOption configures the ansi renderer.
type RendererOption func(*ansiRenderer)

// WithDiff enables or disables line-diff rendering.
// Default: true (diff enabled). If disabled, Render repaints the whole view.
func WithDiff(enabled bool) RendererOption {
	return func(r *ansiRenderer) { r.useDiff = enabled }
}

// NewRenderer creates a new ANSI renderer with options.
// You can keep using the internal newANSIRenderer for defaults;
// this exported constructor is for advanced/custom usage (tests, custom outs).
func NewRenderer(out io.Writer, opts ...RendererOption) Renderer {
	r := &ansiRenderer{
		out:     out,
		useDiff: true,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// ----------------------------------------------------------------------------
// ANSI implementation

type ansiRenderer struct {
	out     io.Writer
	mu      sync.Mutex
	last    string   // last view as a whole
	lines   []string // last view split by '\n'
	cleared bool
	useDiff bool
}

func newANSIRenderer(out io.Writer) *ansiRenderer {
	return &ansiRenderer{
		out:     out,
		useDiff: true,
	}
}

func (r *ansiRenderer) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Hide cursor + clear screen + cursor home
	fmt.Fprint(r.out, "\x1b[?25l\x1b[2J\x1b[H")
	r.cleared = true
	r.last = ""
	r.lines = nil
}

func (r *ansiRenderer) Render(s string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.cleared {
		// first call should always prepare the terminal
		r.clearLocked()
	}

	view := normalizeNewlines(s)

	// Short-circuit if identical
	if view == r.last {
		return
	}

	if !r.useDiff || len(r.lines) == 0 {
		// Full repaint: go home, print, erase tail
		fmt.Fprint(r.out, "\x1b[H")
		fmt.Fprint(r.out, view)
		fmt.Fprint(r.out, "\x1b[0J")
		r.last = view
		r.lines = splitKeep(view)
		return
	}

	// Diff by lines: update only changed rows, clear removed rows.
	newLines := splitKeep(view)

	max := len(newLines)
	if len(r.lines) > max {
		max = len(r.lines)
	}

	for i := 0; i < max; i++ {
		var oldLine, newLine string
		if i < len(r.lines) {
			oldLine = r.lines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}

		if i >= len(newLines) {
			// We had more lines previously; clear this line.
			moveCursor(r.out, i+1, 1)
			// Clear entire line
			fmt.Fprint(r.out, "\x1b[2K")
			continue
		}

		// Line changed?
		if oldLine != newLine {
			moveCursor(r.out, i+1, 1)
			// Print new content and clear to end of line
			fmt.Fprint(r.out, newLine)
			fmt.Fprint(r.out, "\x1b[0K")
		}
	}

	r.last = view
	r.lines = newLines
}

func (r *ansiRenderer) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Show cursor
	fmt.Fprint(r.out, "\x1b[?25h")
}

// ----------------------------------------------------------------------------
// Internals

func (r *ansiRenderer) clearLocked() {
	// called under lock
	fmt.Fprint(r.out, "\x1b[?25l\x1b[2J\x1b[H")
	r.cleared = true
	r.last = ""
	r.lines = nil
}

// normalizeNewlines converts CRLF/CR to LF so we can diff consistently.
func normalizeNewlines(s string) string {
	// Fast path: if no '\r', return as is.
	if !strings.ContainsRune(s, '\r') {
		return s
	}
	// Convert CRLF and bare CR to LF.
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

// splitKeep splits on '\n' but preserves the exact line contents
// (no trailing '\n' on the last line necessary). This allows us
// to address terminal rows precisely with 1-based coordinates.
func splitKeep(s string) []string {
	if s == "" {
		return nil
	}
	// strings.Split retains empty parts between separators.
	parts := strings.Split(s, "\n")
	return parts
}

// moveCursor positions the cursor to row, col (1-based).
func moveCursor(w io.Writer, row, col int) {
	fmt.Fprintf(w, "\x1b[%d;%dH", row, col)
}
