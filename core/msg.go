package core

import "time"

// Msg is any message delivered to Update.
type Msg interface{}

// KeyType classifies the kind of key pressed.
type KeyType int

const (
	KeyUnknown KeyType = iota
	KeyRune
	KeyEnter
	KeyBackspace
	KeyEsc
	KeyCtrlC
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyTab
	KeySpace
	KeyDelete
	KeyHome
	KeyEnd
	KeyPgUp
	KeyPgDn
)

// KeyMsg represents keyboard input in a normalized way.
type KeyMsg struct {
	Type   KeyType
	Rune   rune   // valid if Type == KeyRune
	String string // raw escape/control sequence (e.g., "\x1b[A", "\r")
	Alt    bool   // Alt (Meta) modifier
	Ctrl   bool   // Ctrl modifier (best-effort for some keys)
}

// TickMsg is emitted by Tick() after a duration.
type TickMsg struct{ At time.Time }

// QuitMsg requests graceful termination.
type QuitMsg struct{}

// ResizeMsg notifies terminal size changes (cols x rows).
type ResizeMsg struct {
	Width, Height int
}
