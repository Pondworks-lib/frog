# Frog

Frog is the **core TUI framework** of the Pondworks ecosystem.  
It provides the runtime, message loop, input system, and renderer necessary to build modern terminal user interfaces in Go.

---

## Installation

```bash
go get github.com/pondworks-lib/frog@latest
```

---

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/pondworks-lib/frog"
)

type helloModel struct {
    count int
}

func (m helloModel) Init() frog.Cmd { return nil }

func (m helloModel) Update(msg frog.Msg) (frog.Model, frog.Cmd) {
    switch msg := msg.(type) {
    case frog.KeyMsg:
        if msg.String == "q" {
            return m, frog.Quit()
        }
        m.count++
    }
    return m, nil
}

func (m helloModel) View() string {
    return fmt.Sprintf("Hello Frog!\n\nKeys pressed: %d\n\nPress 'q' to quit.", m.count)
}

func main() {
    frog.Run(helloModel{})
}
```

---

## Features

- **MUV architecture** (Model, Update, View).
- **Message types**:
  - Keyboard (`KeyMsg`), Timer (`TickMsg`), Quit (`QuitMsg`), Resize (`ResizeMsg`).
  - **New in v0.0.5**: Mouse events (`MouseMsg`) and Bracketed Paste (`PasteMsg`).
- **Command system**: `Cmd`, `Tick`, `Quit`, `Nil`.
- **Renderer**: minimal diff-based ANSI renderer, with safe stripping when ANSI is disabled.
- **Input handling**: terminal raw mode, UTF-8 decoding, ESC sequences.
- **Color & Style system**: 16/256/TrueColor, chained style builder (`Fg`, `Bg`, `Bolded`, …).
- **Layout helpers**: `Center`, `PlaceBlock`, `Align*`.
- **Non-interactive mode** (`WithNonInteractive`) → render once, no loops.
- **Context-aware sessions**: `RunContext`, `NewAppWithContext`.
- **Logger support**: pluggable logging (`WithLogger`).
- **Feature toggles**: `WithMouse`, `WithBracketedPaste`, `WithAltScreen`.

---

## Examples (SOON)

Examples will be available in the `example/` directory once the API stabilizes:  

- `hello`: minimal hello world.  
- `dashboard`: mock CPU/RAM dashboard with live updates.  
- `login`: fake SSH login with fields and paste.  
- `progress`: animated progress bar.  

---

## Status

- Frog is in **pre-release** stage (`v0.0.5`).  
- APIs may evolve quickly until `v0.1.0`, when the core interface will be stabilized.

---

## Roadmap & Vision

- **v0.0.6 – 0.0.9**: robustness, ergonomics, API consistency, examples.  
- **v0.1.0**: first stable milestone, frozen core API.  
- **Future**:
  - `lily` : widgets (lists, tables, forms, progress bars).  
  - `ripple` : reactive event bus / pub-sub for TUIs.  
  - Additional repositories for layout, theming, and widget kits.  

---

## Notes

Projects `pad` (layout) and `splash` (theming/colors) were merged into Frog starting in v0.0.3.  
All layout and styling features are now part of this repository.
