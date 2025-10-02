package core

import (
	"fmt"
	"io"
	"strings"
)

type Renderer interface {
	Clear()
	Render(s string)
	Close()
}

type ansiRenderer struct {
	out     io.Writer
	last    string
	cleared bool
}

func newANSIRenderer(out io.Writer) *ansiRenderer {
	return &ansiRenderer{out: out}
}

func (r *ansiRenderer) Clear() {
	fmt.Fprint(r.out, "\x1b[?25l\x1b[2J\x1b[H")
	r.cleared = true
	r.last = ""
}

func (r *ansiRenderer) Render(s string) {
	if !r.cleared {
		r.Clear()
	}
	if s == r.last {
		return
	}
	var b strings.Builder
	b.WriteString("\x1b[H")
	b.WriteString(s)
	b.WriteString("\x1b[0J")
	fmt.Fprint(r.out, b.String())
	r.last = s
}

func (r *ansiRenderer) Close() {
	fmt.Fprint(r.out, "\x1b[?25h")
}
