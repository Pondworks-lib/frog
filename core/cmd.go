package core

import "time"

type Cmd func() Msg

func Nil() Cmd { return nil }

func Batch(cmds ...Cmd) Cmd {
	if len(cmds) == 0 {
		return Nil()
	}
	return func() Msg {
		return cmds[0]()
	}
}

func Tick(d time.Duration) Cmd {
	if d <= 0 {
		d = time.Millisecond
	}
	return func() Msg {
		time.Sleep(d)
		return TickMsg{At: time.Now()}
	}
}

func Quit() Cmd {
	return func() Msg { return QuitMsg{} }
}
