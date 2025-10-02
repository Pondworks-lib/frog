package core

import (
	"bufio"
	"context"
	"io"
	"os"
	"unicode"
	"unicode/utf8"

	"golang.org/x/term"
)

type input struct {
	oldState *term.State
	inFile   *os.File // raw mode only if non-nil
	reader   io.Reader
}

func newInput(r io.Reader) *input {
	var f *os.File
	if rf, ok := r.(*os.File); ok {
		f = rf
	}
	return &input{inFile: f, reader: r}
}

func (i *input) raw() error {
	if i.inFile == nil {
		// cannot enter raw mode (non tty reader), act as non-interactive
		return nil
	}
	fd := int(i.inFile.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	i.oldState = state
	enableVirtualTerminal()
	return nil
}

func (i *input) restore() {
	if i.oldState != nil && i.inFile != nil {
		_ = term.Restore(int(i.inFile.Fd()), i.oldState)
	}
}

func (i *input) readKeys(ctx context.Context, ch chan<- Msg) {
	r := bufio.NewReader(i.reader)
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
			case 3:
				ch <- KeyMsg{Type: KeyCtrlC, String: "\x03", Ctrl: true}
				continue
			case '\r', '\n':
				ch <- KeyMsg{Type: KeyEnter, String: "\r"}
				continue
			case 8, 127:
				ch <- KeyMsg{Type: KeyBackspace, String: string(b)}
				continue
			case 9:
				ch <- KeyMsg{Type: KeyTab, String: "\t"}
				continue
			case ' ':
				ch <- KeyMsg{Type: KeySpace, Rune: ' ', String: " "}
				continue
			case 'q', 'Q':
				ch <- KeyMsg{Type: KeyQ, Rune: rune(b), String: string(b)}
				continue
			case 27:
				km := i.readEscape(r)
				ch <- km
				continue
			}

			if b < 0x20 || b == 0x7f {
				continue
			}

			buf := []byte{b}
			if !utf8.FullRune(buf) {
				for r.Buffered() > 0 && !utf8.FullRune(buf) {
					nb, _ := r.ReadByte()
					buf = append(buf, nb)
				}
			}
			if ru, _ := utf8.DecodeRune(buf); ru != utf8.RuneError && !unicode.IsControl(ru) {
				ch <- KeyMsg{Type: KeyRune, Rune: ru, String: string(ru)}
			}
		}
	}
}

func (i *input) readEscape(r *bufio.Reader) KeyMsg {
	if r.Buffered() == 0 {
		return KeyMsg{Type: KeyEsc, String: "\x1b"}
	}

	nb, _ := r.ReadByte()
	switch nb {
	case '[':
		return i.readCSI(r)
	default:
		buf := []byte{nb}
		for r.Buffered() > 0 && !utf8.FullRune(buf) {
			b, _ := r.ReadByte()
			buf = append(buf, b)
		}
		if ru, _ := utf8.DecodeRune(buf); ru != utf8.RuneError && !unicode.IsControl(ru) {
			return KeyMsg{Type: KeyRune, Rune: ru, String: string(ru), Alt: true}
		}
		return KeyMsg{Type: KeyEsc, String: "\x1b"}
	}
}

func (i *input) readCSI(r *bufio.Reader) KeyMsg {
	params := []byte{}
	for {
		if r.Buffered() == 0 {
			return KeyMsg{Type: KeyEsc, String: "\x1b"}
		}
		b, _ := r.ReadByte()
		switch b {
		case 'A':
			return KeyMsg{Type: KeyUp, String: "\x1b[A"}
		case 'B':
			return KeyMsg{Type: KeyDown, String: "\x1b[B"}
		case 'C':
			return KeyMsg{Type: KeyRight, String: "\x1b[C"}
		case 'D':
			return KeyMsg{Type: KeyLeft, String: "\x1b[D"}
		case 'H':
			return KeyMsg{Type: KeyHome, String: "\x1b[H"}
		case 'F':
			return KeyMsg{Type: KeyEnd, String: "\x1b[F"}
		case '~':
			switch string(params) {
			case "3":
				return KeyMsg{Type: KeyDelete, String: "\x1b[3~"}
			case "5":
				return KeyMsg{Type: KeyPgUp, String: "\x1b[5~"}
			case "6":
				return KeyMsg{Type: KeyPgDn, String: "\x1b[6~"}
			case "2":
				return KeyMsg{Type: KeyEsc, String: "\x1b[2~"}
			default:
				return KeyMsg{Type: KeyEsc, String: "\x1b[" + string(params) + "~"}
			}
		default:
			if (b >= '0' && b <= '9') || b == ';' {
				params = append(params, b)
				continue
			}
			return KeyMsg{Type: KeyEsc, String: "\x1b[" + string(params) + string(b)}
		}
	}
}
