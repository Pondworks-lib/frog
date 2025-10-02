package core

import "fmt"

type App struct {
    name string
}

func NewApp(name string) *App {
    return &App{name: name}
}

func (a *App) Run() {
    fmt.Printf("🐸 Frog TUI App: %s is running...\n", a.name)
}
