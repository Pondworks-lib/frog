package core

import "time"

// Cmd is an async action that returns a Msg once completed.
type Cmd func() Msg

// Nil returns no command.
func Nil() Cmd { return nil }

// Tick returns a command that emits a TickMsg after d.
func Tick(d time.Duration) Cmd {
	if d <= 0 {
		d = time.Millisecond
	}
	return func() Msg {
		time.Sleep(d)
		return TickMsg{At: time.Now()}
	}
}

// Quit requests a graceful termination.
func Quit() Cmd { return func() Msg { return QuitMsg{} } }
