package core

import (
	"bufio"
	"context"
	"os"
	"unicode"
	"unicode/utf8"

	"golang.org/x/term"
)

type input struct{ oldState *term.State }

func newInput() *input { return &input{} }

// raw enables terminal raw mode; caller must defer i.restore().
func (i *input) raw() error {
	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	i.oldState = state
	enableVirtualTerminal() // no-op on non-Windows; enables ANSI on Windows
	return nil
}

func (i *input) restore() {
	if i.oldState != nil {
		_ = term.Restore(int(os.Stdin.Fd()), i.oldState)
	}
}

// readKeys parses bytes from stdin and sends KeyMsg until ctx is done.
// It recognizes common ASCII keys, arrows, Home/End, PgUp/PgDn, Delete,
// Enter, Backspace, Tab, Space, ESC, Ctrl-C, and Alt+<key>.
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
			case 3: // ^C
				ch <- KeyMsg{Type: KeyCtrlC, String: "\x03", Ctrl: true}
				continue

			case '\r', '\n':
				ch <- KeyMsg{Type: KeyEnter, String: "\r"}
				continue

			case 8, 127: // Backspace (ASCII/DEL)
				ch <- KeyMsg{Type: KeyBackspace, String: string(b)}
				continue

			case 9: // Tab
				ch <- KeyMsg{Type: KeyTab, String: "\t"}
				continue

			case ' ':
				ch <- KeyMsg{Type: KeySpace, Rune: ' ', String: " "}
				continue

			case 27: // ESC / CSI / Alt+key
				km := i.readEscape(r)
				ch <- km
				continue
			}

			// UTF-8 rune or control
			if b < 0x20 || b == 0x7f {
				// other control chars: ignore for now
				continue
			}

			// handle UTF-8 multibyte
			buf := []byte{b}
			if !utf8.FullRune(buf) {
				// try to complete a rune
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

// readEscape reads after an initial ESC (0x1b) and tries to recognize
// CSI sequences and Alt-modified keys. It never blocks indefinitely:
// it checks reader buffered content to decide.
func (i *input) readEscape(r *bufio.Reader) KeyMsg {
	// If nothing buffered, treat as bare ESC.
	if r.Buffered() == 0 {
		return KeyMsg{Type: KeyEsc, String: "\x1b"}
	}

	// Peek next byte to distinguish CSI or Alt+<key>.
	nb, _ := r.ReadByte()
	switch nb {
	case '[': // CSI sequences
		return i.readCSI(r)
	default:
		// Likely Alt+key (Meta). Try to decode a rune from nb + possibly more bytes.
		buf := []byte{nb}
		// complete UTF-8 if needed
		for r.Buffered() > 0 && !utf8.FullRune(buf) {
			b, _ := r.ReadByte()
			buf = append(buf, b)
		}
		if ru, _ := utf8.DecodeRune(buf); ru != utf8.RuneError && !unicode.IsControl(ru) {
			return KeyMsg{Type: KeyRune, Rune: ru, String: string(ru), Alt: true}
		}
		// Fallback: treat as ESC if not a printable rune.
		return KeyMsg{Type: KeyEsc, String: "\x1b"}
	}
}

// readCSI parses a limited set of Control Sequence Introducer codes.
// We support arrow keys, Home/End, PgUp/PgDn, Delete, with or without
// extra modifier parameters (ignored for now).
func (i *input) readCSI(r *bufio.Reader) KeyMsg {
	// The CSI format is ESC [ <params> <final>
	// We'll read until we hit a final byte in '@A-Za-z~'.
	params := []byte{}
	for {
		if r.Buffered() == 0 {
			// Incomplete sequence: fallback to ESC.
			return KeyMsg{Type: KeyEsc, String: "\x1b"}
		}
		b, _ := r.ReadByte()
		// Final bytes:
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
			// Tilde forms like "3~" (delete), "5~" (PgUp), "6~" (PgDn), "2~" (Ins)
			switch string(params) {
			case "3":
				return KeyMsg{Type: KeyDelete, String: "\x1b[3~"}
			case "5":
				return KeyMsg{Type: KeyPgUp, String: "\x1b[5~"}
			case "6":
				return KeyMsg{Type: KeyPgDn, String: "\x1b[6~"}
			case "2":
				// Insert – not mapped for now, return Esc fallback:
				return KeyMsg{Type: KeyEsc, String: "\x1b[2~"}
			default:
				// Unrecognized param; fallback:
				return KeyMsg{Type: KeyEsc, String: "\x1b[" + string(params) + "~"}
			}
		default:
			// Accumulate digits and separators (e.g., "1;5")
			if (b >= '0' && b <= '9') || b == ';' {
				params = append(params, b)
				continue
			}
			// Unknown final – fallback:
			return KeyMsg{Type: KeyEsc, String: "\x1b[" + string(params) + string(b)}
		}
	}
}
