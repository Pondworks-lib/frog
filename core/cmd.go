package core

import "time"

// Cmd represents an async action that eventually returns a Msg.
type Cmd func() Msg

// Nil returns no command.
func Nil() Cmd { return nil }

// Batch executes commands in order and returns the first produced message.
// (Subsequent scheduling is up to the Update loop.)
func Batch(cmds ...Cmd) Cmd {
	if len(cmds) == 0 {
		return Nil()
	}
	return func() Msg {
		for _, c := range cmds {
			if c == nil {
				continue
			}
			if m := c(); m != nil {
				return m
			}
		}
		return nil
	}
}

// Tick emits a TickMsg after d (min 1ms).
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
