# Frog

Frog is the core TUI framework of the Pondworks ecosystem.  
It provides the runtime, message loop, and renderer necessary to build terminal user interfaces in Go.

---

## Installation

```bash
go get github.com/pondworks-lib/frog@latest
```

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

func (m helloModel) Init() frog.Cmd {
    return nil
}

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
    return fmt.Sprintf(
        "Hello Frog!\n\nKeys pressed: %d\n\nPress 'q' to quit.",
        m.count,
    )
}

func main() {
    frog.Run(helloModel{})
}

```
---

## Features
- MUV architecture (Model, Update, View)

- Message types: KeyMsg, TickMsg, QuitMsg, ResizeMsg

- Command system: Cmd, Tick, Quit, Nil

- Minimal ANSI renderer (with diffing)

- Input handling with terminal raw mode

- Cross-platform support for enabling ANSI (Windows / Unix)

---

## Status
- Frog is in ```pre-release``` stage.
- APIs may change until we reach v1.0.0.

--

## Roadmap & Vision

- lily : widgets (lists, tables, forms, progress bar)

- pad : layout / panels / flex / grid

- splash : theming, styles, colors

- ripple : event bus / pub-sub for UI

- More examples & community contributions **soon**!
