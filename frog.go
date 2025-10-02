// Package frog is the top-level API for the Pondworks TUI framework.
// It provides the runtime, message system, and rendering loop for building TUIs in Go.
package frog

import "github.com/pondworks-lib/frog/core"

type (
	App    = core.Session
	Option = core.Option

	// MUV types re-export
	Model    = core.Model
	Msg      = core.Msg
	KeyMsg   = core.KeyMsg
	KeyType  = core.KeyType
	TickMsg  = core.TickMsg
	QuitMsg  = core.QuitMsg
	Cmd      = core.Cmd
	ResizeMsg = core.ResizeMsg
)

// Re-export key constants so users can refer to frog.KeyEnter, frog.KeyCtrlC, etc.
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

// NewApp creates a new UI session.
func NewApp(m Model, opts ...Option) *App { return core.NewSession(m, opts...) }

// Run is a convenience entrypoint running the session to completion.
func Run(m Model, opts ...Option) error { return core.NewSession(m, opts...).Run() }

// Helpers re-export
var (
	Tick           = core.Tick
	Quit           = core.Quit
	Nil            = core.Nil
	WithRenderer   = core.WithRenderer
	WithAltScreen  = core.WithAltScreen
	WithMsgBuffer  = core.WithMsgBuffer
)
