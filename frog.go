// Package frog is the top-level API for the Pondworks TUI framework.
// It provides the runtime, message system, and rendering loop for building TUIs in Go.
package frog

import (
	"io"

	"github.com/pondworks-lib/frog/core"
)

type (
	App    = core.Session
	Option = core.Option

	// MUV types
	Model     = core.Model
	Msg       = core.Msg
	KeyMsg    = core.KeyMsg
	KeyType   = core.KeyType
	TickMsg   = core.TickMsg
	QuitMsg   = core.QuitMsg
	Cmd       = core.Cmd
	ResizeMsg = core.ResizeMsg

	// Styling
	Style        = core.Style
	Color        = core.Color
	ColorProfile = core.ColorProfile

	// Renderer options (advanced)
	RendererOption = core.RendererOption

	// Layout
	AlignH = core.AlignH
	AlignV = core.AlignV
)

// Key constants
const (
	KeyUnknown   = core.KeyUnknown
	KeyRune      = core.KeyRune
	KeyEnter     = core.KeyEnter
	KeyBackspace = core.KeyBackspace
	KeyEsc       = core.KeyEsc
	KeyCtrlC     = core.KeyCtrlC
	KeyUp        = core.KeyUp
	KeyDown      = core.KeyDown
	KeyLeft      = core.KeyLeft
	KeyRight     = core.KeyRight
	KeyTab       = core.KeyTab
	KeySpace     = core.KeySpace
	KeyDelete    = core.KeyDelete
	KeyHome      = core.KeyHome
	KeyEnd       = core.KeyEnd
	KeyPgUp      = core.KeyPgUp
	KeyPgDn      = core.KeyPgDn
)

// Color profile constants
const (
	ColorAuto      = core.ColorAuto
	ColorNone      = core.ColorNone
	ColorANSI16    = core.ColorANSI16
	ColorANSI256   = core.ColorANSI256
	ColorTrueColor = core.ColorTrueColor
)

// Named colors (16-color)
var (
	ColorBlack         = core.ColorBlack
	ColorRed           = core.ColorRed
	ColorGreen         = core.ColorGreen
	ColorYellow        = core.ColorYellow
	ColorBlue          = core.ColorBlue
	ColorMagenta       = core.ColorMagenta
	ColorCyan          = core.ColorCyan
	ColorWhite         = core.ColorWhite
	ColorBrightBlack   = core.ColorBrightBlack
	ColorBrightRed     = core.ColorBrightRed
	ColorBrightGreen   = core.ColorBrightGreen
	ColorBrightYellow  = core.ColorBrightYellow
	ColorBrightBlue    = core.ColorBrightBlue
	ColorBrightMagenta = core.ColorBrightMagenta
	ColorBrightCyan    = core.ColorBrightCyan
	ColorBrightWhite   = core.ColorBrightWhite
)

// Style helpers
var (
	NewStyle  = core.NewStyle
	ANSI256   = core.ANSI256
	RGB       = core.RGB
	Colorize  = core.Colorize
	StripANSI = core.StripANSI
)

// App helpers
func NewApp(m Model, opts ...Option) *App { return core.NewSession(m, opts...) }
func Run(m Model, opts ...Option) error   { return core.NewSession(m, opts...).Run() }

// Session options
var (
	Tick           = core.Tick
	Quit           = core.Quit
	Nil            = core.Nil
	WithRenderer   = core.WithRenderer
	WithAltScreen  = core.WithAltScreen
	WithMsgBuffer  = core.WithMsgBuffer
)

// Renderer power-user API
func NewRenderer(out io.Writer, opts ...RendererOption) core.Renderer {
	return core.NewRenderer(out, opts...)
}
var (
	WithDiff         = core.WithDiff
	WithColorProfile = core.WithColorProfile
)

// Layout helpers
const (
	AlignLeft   = core.AlignLeft
	AlignCenter = core.AlignCenter
	AlignRight  = core.AlignRight
	AlignTop    = core.AlignTop
	AlignMiddle = core.AlignMiddle
	AlignBottom = core.AlignBottom
)

var (
	Center     = core.Center
	PlaceBlock = core.PlaceBlock
)
