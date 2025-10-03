package core

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"strconv"
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
		// cannot enter raw mode (non-tty reader)
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
			case 27: // ESC: CSI, Alt+key, SGR mouse, bracketed paste
				if m := i.readEscape(r); m != nil {
					ch <- m
				}
				continue
			}

			// Other control bytes: ignore
			if b < 0x20 || b == 0x7f {
				continue
			}

			// UTF-8 rune
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

// readEscape decodes sequences after ESC. It can return KeyMsg, MouseMsg, PasteMsg.
func (i *input) readEscape(r *bufio.Reader) Msg {
	if r.Buffered() == 0 {
		return KeyMsg{Type: KeyEsc, String: "\x1b"}
	}

	nb, _ := r.ReadByte()
	switch nb {
	case '[':
		// Check for bracketed paste start: ESC [ 200 ~
		if i.peekSeq(r, "200~") {
			_, _ = r.Discard(len("200~")) // FIX: discard returns (int, error)
			return i.readBracketedPaste(r)
		}
		// SGR mouse starts with '<'
		if i.peekByte(r, '<') {
			_, _ = r.ReadByte() // consume '<'
			return i.readMouseSGR(r)
		}
		// Otherwise parse normal CSI keys
		return i.readCSI(r)
	default:
		// Likely Alt+key (Meta). Decode a rune from nb + more bytes if needed.
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

// readCSI parses a limited set of CSI codes (arrows, home/end, pgup/pgdn, delete).
func (i *input) readCSI(r *bufio.Reader) Msg {
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

// readMouseSGR parses SGR mouse events after "<" in the sequence ESC[<b;x;y(M|m)
func (i *input) readMouseSGR(r *bufio.Reader) Msg {
	readNum := func() (int, bool) {
		var buf bytes.Buffer
		for {
			if r.Buffered() == 0 {
				break
			}
			b, _ := r.ReadByte()
			if b >= '0' && b <= '9' {
				buf.WriteByte(b)
				continue
			}
			_ = r.UnreadByte()
			break
		}
		if buf.Len() == 0 {
			return 0, false
		}
		v, err := strconv.Atoi(buf.String())
		return v, err == nil
	}

	// <b ; x ; y (M|m)
	if b, ok := readNum(); ok {
		if c, _ := r.ReadByte(); c != ';' {
			return KeyMsg{Type: KeyEsc, String: "\x1b"}
		}
		x, okx := readNum()
		if !okx {
			return KeyMsg{Type: KeyEsc, String: "\x1b"}
		}
		if c, _ := r.ReadByte(); c != ';' {
			return KeyMsg{Type: KeyEsc, String: "\x1b"}
		}
		y, oky := readNum()
		if !oky {
			return KeyMsg{Type: KeyEsc, String: "\x1b"}
		}
		final, _ := r.ReadByte() // 'M' press/drag, 'm' release

		shift := (b & 4) != 0
		alt := (b & 8) != 0
		ctrl := (b & 16) != 0

		btn := MouseUnknown
		act := MousePress

		// Wheel
		if (b & 64) != 0 {
			act = MouseWheel
			if (b & 1) == 1 {
				btn = MouseWheelDown
			} else {
				btn = MouseWheelUp
			}
		} else {
			switch b & 3 {
			case 0:
				btn = MouseLeft
			case 1:
				btn = MouseMiddle
			case 2:
				btn = MouseRight
			default:
				btn = MouseUnknown
			}
			if final == 'm' {
				act = MouseRelease
			} else if (b & 32) != 0 {
				act = MouseDrag
			} else {
				act = MousePress
			}
		}

		return MouseMsg{
			Button: btn,
			Action: act,
			X:      x,
			Y:      y,
			Alt:    alt,
			Ctrl:   ctrl,
			Shift:  shift,
		}
	}

	return KeyMsg{Type: KeyEsc, String: "\x1b"}
}

// readBracketedPaste reads until ESC[201~ and returns the pasted payload.

const maxPaste = 1 << 20 // 1 MiB
func (i *input) readBracketedPaste(r *bufio.Reader) Msg {
	var buf bytes.Buffer
	for {
		b, err := r.ReadByte()
		if err != nil { break }
		if buf.Len() >= maxPaste {
			if b == 27 && i.peekSeq(r, "[201~") {
				_, _ = r.Discard(len("[201~"))
				break
			}
			continue
		}
		if b == 27 { // ESC
			if i.peekSeq(r, "[201~") {
				_, _ = r.Discard(len("[201~")) // FIX: discard returns (int, error)
				break
			}
			// Not the end sequence: include ESC and continue
			buf.WriteByte(b)
			continue
		}
		buf.WriteByte(b)
	}
	return PasteMsg{Text: buf.String()}
}

// helpers

func (i *input) peekSeq(r *bufio.Reader, s string) bool {
	if r.Buffered() < len(s) {
		return false
	}
	bs, err := r.Peek(len(s))
	return err == nil && string(bs) == s
}

func (i *input) peekByte(r *bufio.Reader, b byte) bool {
	if r.Buffered() < 1 {
		return false
	}
	bs, err := r.Peek(1)
	return err == nil && bs[0] == b
}
