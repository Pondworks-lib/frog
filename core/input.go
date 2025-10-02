package core

import (
	"bufio"
	"context"
	"os"
	"unicode"

	"golang.org/x/term"
)

type input struct {
	oldState *term.State
}

func newInput() *input { return &input{} }

func (i *input) raw() error {
	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	i.oldState = state
	enableVirtualTerminal()
	return nil
}

func (i *input) restore() {
	if i.oldState != nil {
		_ = term.Restore(int(os.Stdin.Fd()), i.oldState)
	}
}

func (i *input) readKeys(ctx context.Context, ch chan<- Msg) {
	r := bufio.NewReader(os.Stdin)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			b, err := r.ReadByte()
			if err != nil {
				return
			}
			switch b {
			case 3: // Ctrl+C
				ch <- KeyMsg{String: "\x03", Ctrl: true}
				return
			case 'q', 'Q':
				ch <- KeyMsg{Rune: rune(b), String: string(b)}
			case '\r', '\n':
				ch <- KeyMsg{String: "\r"}
			default:
				rm := KeyMsg{Rune: rune(b), String: string(b)}
				if !unicode.IsControl(rm.Rune) {
					ch <- rm
				}
			}
		}
	}
}
