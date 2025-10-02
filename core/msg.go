package core

import "time"

type Msg interface{}

type KeyMsg struct {
	Rune   rune
	String string // es: "\r", "\x03"
	Alt    bool
	Ctrl   bool
}

type TickMsg struct {
	At time.Time
}

type QuitMsg struct{}

type ResizeMsg struct {
	Width  int
	Height int
}
