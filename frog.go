package frog

import "github.com/pondworks-lib/frog/core"

type (
	App    = core.Session
	Option = core.Option

	// MUV types
	Model   = core.Model
	Msg     = core.Msg
	KeyMsg  = core.KeyMsg
	TickMsg = core.TickMsg
	QuitMsg = core.QuitMsg
	Cmd     = core.Cmd
)

func NewApp(m Model, opts ...Option) *App {
	return core.NewSession(m, opts...)
}

func Run(m Model, opts ...Option) error {
	return core.NewSession(m, opts...).Run()
}

// Helpers re-export
var (
	Tick         = core.Tick
	Quit         = core.Quit
	Nil          = core.Nil
	WithRenderer = core.WithRenderer
)
