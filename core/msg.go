package core

import "time"

// Msg is any message delivered to Update.
type Msg interface{}

// ---------- Keys ----------

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
	KeyQ
)

type KeyMsg struct {
	Type   KeyType
	Rune   rune
	String string
	Alt    bool
	Ctrl   bool
}

// ---------- Time / Quit / Resize ----------

type TickMsg struct{ At time.Time }

type QuitMsg struct{}

type ResizeMsg struct {
	Width, Height int
}

// ---------- Bracketed Paste ----------

type PasteMsg struct {
	Text string
}

// ---------- Mouse (SGR) ----------

type MouseButton int

const (
	MouseUnknown MouseButton = iota
	MouseLeft
	MouseMiddle
	MouseRight
	MouseWheelUp
	MouseWheelDown
)

type MouseAction int

const (
	MousePress MouseAction = iota
	MouseRelease
	MouseDrag
	MouseWheel
)

type MouseMsg struct {
	Button MouseButton
	Action MouseAction
	X, Y   int // 1-based terminal coords
	Alt    bool
	Ctrl   bool
	Shift  bool
}
