// Package frog is the top-level API for the Pondworks TUI framework.
// It provides the runtime, message system, and rendering loop for building TUIs in Go.
package frog

import (
	"context"
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

	// Mouse & Paste
	MouseMsg    = core.MouseMsg
	MouseButton = core.MouseButton
	MouseAction = core.MouseAction
	PasteMsg    = core.PasteMsg

	// Styling
	Style        = core.Style
	Color        = core.Color
	ColorProfile = core.ColorProfile

	// Renderer options (advanced)
	RendererOption = core.RendererOption

	// Layout
	AlignH = core.AlignH
	AlignV = core.AlignV

	// Logger
	Logger = core.Logger
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
	KeyQ         = core.KeyQ
)

// Mouse constants
const (
	MouseUnknown   = core.MouseUnknown
	MouseLeft      = core.MouseLeft
	MouseMiddle    = core.MouseMiddle
	MouseRight     = core.MouseRight
	MouseWheelUp   = core.MouseWheelUp
	MouseWheelDown = core.MouseWheelDown
)

const (
	MousePress = core.MousePress
	MouseRelease = core.MouseRelease
	MouseDrag = core.MouseDrag
	MouseWheel = core.MouseWheel
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

// Context-aware entrypoints
func NewAppWithContext(ctx context.Context, m Model, opts ...Option) *App {
	return core.NewSessionWithContext(ctx, m, opts...)
}
func RunContext(ctx context.Context, m Model, opts ...Option) error {
	return core.NewSessionWithContext(ctx, m, opts...).Run()
}

// Session options
var (
	Tick               = core.Tick
	Quit               = core.Quit
	Nil                = core.Nil
	WithRenderer       = core.WithRenderer
	WithAltScreen      = core.WithAltScreen
	WithMsgBuffer      = core.WithMsgBuffer
	WithOut            = core.WithOut
	WithIn             = core.WithIn
	WithResizeInterval = core.WithResizeInterval
	WithNonInteractive = core.WithNonInteractive
	WithLogger         = core.WithLogger
	WithMouse          = core.WithMouse
	WithBracketedPaste = core.WithBracketedPaste
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
