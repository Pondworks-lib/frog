package core

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"golang.org/x/term"
)

type Renderer interface {
	Clear()
	Render(s string)
	Close()
}

// ---- Options

type RendererOption func(*ansiRenderer)

// WithDiff toggles line-diff rendering (default: enabled).
func WithDiff(enabled bool) RendererOption { return func(r *ansiRenderer) { r.useDiff = enabled } }

// WithColorProfile forces a specific color profile (overrides auto-detection).
func WithColorProfile(p ColorProfile) RendererOption { return func(r *ansiRenderer) { r.profile = p } }

// NewRenderer builds an ANSI renderer with options.
func NewRenderer(out io.Writer, opts ...RendererOption) Renderer {
	r := &ansiRenderer{
		out:     out,
		useDiff: true,
		profile: ColorAuto,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// ---- Implementation

type ansiRenderer struct {
	out     io.Writer
	mu      sync.Mutex
	last    string
	lines   []string
	cleared bool
	useDiff bool

	profile ColorProfile // ColorAuto by default; lazily resolved on first Clear/Render
}

func newANSIRenderer(out io.Writer) *ansiRenderer {
	return &ansiRenderer{
		out:     out,
		useDiff: true,
		profile: ColorAuto,
	}
}

func (r *ansiRenderer) ensureColorProfile() {
	if r.profile != ColorAuto {
		return
	}
	r.profile = detectColorProfile(r.out)
}

func (r *ansiRenderer) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.ensureColorProfile()

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
		r.clearLocked()
	}

	// Decide colors: strip if profile says None
	r.ensureColorProfile()
	view := normalizeNewlines(s)
	if r.profile == ColorNone {
		view = StripANSI(view)
	}

	// Short-circuit if identical
	if view == r.last {
		return
	}

	if !r.useDiff || len(r.lines) == 0 {
		// Full repaint
		fmt.Fprint(r.out, "\x1b[H")
		fmt.Fprint(r.out, view)
		fmt.Fprint(r.out, "\x1b[0J")
		r.last = view
		r.lines = splitKeep(view)
		return
	}

	// Diff by lines
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
			moveCursor(r.out, i+1, 1)
			fmt.Fprint(r.out, "\x1b[2K")
			continue
		}

		if oldLine != newLine {
			moveCursor(r.out, i+1, 1)
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
	fmt.Fprint(r.out, "\x1b[?25h")
}

// ---- Internals

func (r *ansiRenderer) clearLocked() {
	r.ensureColorProfile()
	fmt.Fprint(r.out, "\x1b[?25l\x1b[2J\x1b[H")
	r.cleared = true
	r.last = ""
	r.lines = nil
}

// Turn \r\n and \r into \n for stable diffs.
func normalizeNewlines(s string) string {
	if !strings.ContainsRune(s, '\r') {
		return s
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

func splitKeep(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func moveCursor(w io.Writer, row, col int) {
	fmt.Fprintf(w, "\x1b[%d;%dH", row, col)
}

// Honors NO_COLOR, checks TTY, then COLORTERM/TERM to choose 24-bit/256/16.
func detectColorProfile(out io.Writer) ColorProfile {
	// NO_COLOR -> no colors
	if v := strings.TrimSpace(os.Getenv("NO_COLOR")); v != "" {
		return ColorNone
	}

	// If not a terminal -> no colors
	if f, ok := out.(*os.File); ok {
		if !term.IsTerminal(int(f.Fd())) {
			return ColorNone
		}
	}

	// Truecolor?
	if strings.Contains(strings.ToLower(os.Getenv("COLORTERM")), "truecolor") {
		return ColorTrueColor
	}
	// 256 colors?
	if strings.Contains(strings.ToLower(os.Getenv("TERM")), "256color") {
		return ColorANSI256
	}
	// Conservative default
	return ColorANSI16
}
