package core

import (
	"fmt"
	"regexp"
	"strings"
)

// ---- Color profiles / capabilities ----

type ColorProfile int

const (
	ColorAuto ColorProfile = iota
	ColorNone
	ColorANSI16
	ColorANSI256
	ColorTrueColor
)

// ---- Color specification ----

type colorKind int

const (
	colorUnset colorKind = iota
	colorNamed16
	colorIndex256
	colorRGB
)

// Named base colors (0..7)
type NamedColor uint8

const (
	NamedBlack NamedColor = iota
	NamedRed
	NamedGreen
	NamedYellow
	NamedBlue
	NamedMagenta
	NamedCyan
	NamedWhite
)

type Color struct {
	kind   colorKind
	index  uint8 // for 256-colors
	r, g, b uint8
	named   NamedColor
	bright  bool // for 16-color bright variants
}

// Constructors
func Ansi16(name NamedColor, bright bool) Color {
	return Color{kind: colorNamed16, named: name, bright: bright}
}
func ANSI256(n uint8) Color { return Color{kind: colorIndex256, index: n} }
func RGB(r, g, b uint8) Color {
	return Color{kind: colorRGB, r: r, g: g, b: b}
}

// ---- Style with basic attributes ----

type Style struct {
	fg, bg   *Color
	Bold     bool
	Faint    bool
	Italic   bool
	Underline bool
	Blink    bool
	Reverse  bool
	Strike   bool
}

// Builder / chaining

func NewStyle() Style                 { return Style{} }
func (s Style) Fg(c Color) Style      { s.fg = &c; return s }
func (s Style) Bg(c Color) Style      { s.bg = &c; return s }
func (s Style) Bolded() Style         { s.Bold = true; return s }
func (s Style) Fainted() Style        { s.Faint = true; return s }
func (s Style) Italicized() Style     { s.Italic = true; return s }
func (s Style) Underlined() Style     { s.Underline = true; return s }
func (s Style) Blinking() Style       { s.Blink = true; return s }
func (s Style) Reversed() Style       { s.Reverse = true; return s }
func (s Style) Struck() Style         { s.Strike = true; return s }

// Render wraps text in ANSI SGR codes. It always emits ANSI; the renderer
func (s Style) Render(text string) string {
	codes := make([]string, 0, 6)

	// attributes
	if s.Bold {
		codes = append(codes, "1")
	}
	if s.Faint {
		codes = append(codes, "2")
	}
	if s.Italic {
		codes = append(codes, "3")
	}
	if s.Underline {
		codes = append(codes, "4")
	}
	if s.Blink {
		codes = append(codes, "5")
	}
	if s.Reverse {
		codes = append(codes, "7")
	}
	if s.Strike {
		codes = append(codes, "9")
	}

	// colors
	if s.fg != nil {
		codes = append(codes, s.fg.fgSGR()...)
	}
	if s.bg != nil {
		codes = append(codes, s.bg.bgSGR()...)
	}

	if len(codes) == 0 {
		return text
	}
	return fmt.Sprintf("\x1b[%sm%s\x1b[0m", strings.Join(codes, ";"), text)
}

func (c Color) fgSGR() []string {
	switch c.kind {
	case colorNamed16:
		// 30..37 (normal), 90..97 (bright)
		base := 30 + int(c.named)
		if c.bright {
			base = 90 + int(c.named)
		}
		return []string{fmt.Sprintf("%d", base)}
	case colorIndex256:
		return []string{"38", "5", fmt.Sprintf("%d", c.index)}
	case colorRGB:
		return []string{"38", "2", fmt.Sprintf("%d", c.r), fmt.Sprintf("%d", c.g), fmt.Sprintf("%d", c.b)}
	default:
		return nil
	}
}

func (c Color) bgSGR() []string {
	switch c.kind {
	case colorNamed16:
		// 40..47 (normal), 100..107 (bright)
		base := 40 + int(c.named)
		if c.bright {
			base = 100 + int(c.named)
		}
		return []string{fmt.Sprintf("%d", base)}
	case colorIndex256:
		return []string{"48", "5", fmt.Sprintf("%d", c.index)}
	case colorRGB:
		return []string{"48", "2", fmt.Sprintf("%d", c.r), fmt.Sprintf("%d", c.g), fmt.Sprintf("%d", c.b)}
	default:
		return nil
	}
}

// Convenience helpers
func Colorize(text string, fg *Color, bg *Color, bold bool) string {
	st := NewStyle()
	if fg != nil {
		st = st.Fg(*fg)
	}
	if bg != nil {
		st = st.Bg(*bg)
	}
	if bold {
		st = st.Bolded()
	}
	return st.Render(text)
}

// Named convenience (16-color)
var (
	ColorBlack       = Ansi16(NamedBlack, false)
	ColorRed         = Ansi16(NamedRed, false)
	ColorGreen       = Ansi16(NamedGreen, false)
	ColorYellow      = Ansi16(NamedYellow, false)
	ColorBlue        = Ansi16(NamedBlue, false)
	ColorMagenta     = Ansi16(NamedMagenta, false)
	ColorCyan        = Ansi16(NamedCyan, false)
	ColorWhite       = Ansi16(NamedWhite, false)
	ColorBrightBlack = Ansi16(NamedBlack, true)
	ColorBrightRed   = Ansi16(NamedRed, true)
	ColorBrightGreen = Ansi16(NamedGreen, true)
	ColorBrightYellow= Ansi16(NamedYellow, true)
	ColorBrightBlue  = Ansi16(NamedBlue, true)
	ColorBrightMagenta=Ansi16(NamedMagenta, true)
	ColorBrightCyan  = Ansi16(NamedCyan, true)
	ColorBrightWhite = Ansi16(NamedWhite, true)
)


var reANSISGR = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// StripANSI removes SGR sequences from a string.
func StripANSI(s string) string {
	return reANSISGR.ReplaceAllString(s, "")
}
