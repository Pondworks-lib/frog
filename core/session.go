package core

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/term"
)

// Option configures a Session at construction.
type Option func(*Session)

// Session runs a Model, coordinating input, rendering and lifecycle.
type Session struct {
	m         Model
	renderer  Renderer
	input     *input
	msgCh     chan Msg
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	startOnce sync.Once
	stopOnce  sync.Once

	altScreen bool
	msgBuf    int // channel capacity (default 64)
}

// WithRenderer uses a custom renderer (useful in tests).
func WithRenderer(r Renderer) Option { return func(p *Session) { p.renderer = r } }

// WithAltScreen switches to the terminal alternate screen while the session runs.
func WithAltScreen() Option { return func(p *Session) { p.altScreen = true } }

// WithMsgBuffer sets the size of the internal message buffer.
func WithMsgBuffer(n int) Option {
	return func(p *Session) {
		if n > 0 {
			p.msgBuf = n
		}
	}
}

// NewSession creates a session for a given Model.
func NewSession(m Model, opts ...Option) *Session {
	ctx, cancel := context.WithCancel(context.Background())
	p := &Session{
		m:        m,
		renderer: newANSIRenderer(os.Stdout),
		input:    newInput(),
		msgBuf:   64,
		ctx:      ctx,
		cancel:   cancel,
	}
	for _, o := range opts {
		o(p)
	}
	// late init for msgCh (after msgBuf is known)
	p.msgCh = make(chan Msg, p.msgBuf)
	return p
}

// Run starts the session and blocks until completion or error.
// It ensures raw mode, optional alt screen, input & resize watchers,
// initial render, and the main Update loop.
func (p *Session) Run() error {
	var runErr error
	p.startOnce.Do(func() {
		if err := p.input.raw(); err != nil {
			runErr = fmt.Errorf("raw mode: %w", err)
			return
		}
		defer p.input.restore()

		// Alt screen on/off
		if p.altScreen {
			fmt.Fprint(os.Stdout, "\x1b[?1049h")
			defer fmt.Fprint(os.Stdout, "\x1b[?1049l")
		}

		// Input reader
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			p.input.readKeys(p.ctx, p.msgCh)
		}()

		// Portable resize watcher (poll); we can optimize with SIGWINCH later
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			p.watchSize(p.ctx, p.msgCh)
		}()

		// OS signals (Ctrl+C et al.)
		sigCh := make(chan os.Signal, 2)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(sigCh)

		// Initial cycle
		cmd := p.m.Init()
		p.renderer.Clear()
		p.renderer.Render(p.m.View())

		if cmd != nil {
			go func(c Cmd) { p.msgCh <- c() }(cmd)
		}

		// Main loop
	loop:
		for {
			select {
			case <-p.ctx.Done():
				break loop
			case <-sigCh:
				p.cancel()
			case msg := <-p.msgCh:
				if _, ok := msg.(QuitMsg); ok {
					break loop
				}
				var next Cmd
				p.m, next = p.m.Update(msg)
				p.renderer.Render(p.m.View())
				if next != nil {
					go func(c Cmd) { p.msgCh <- c() }(next)
				}
			}
		}

		// Shutdown
		p.stopOnce.Do(func() {
			p.cancel()
			p.wg.Wait()
			p.renderer.Close()
		})
	})
	return runErr
}

// Send injects a message from outside (tests or background jobs).
func (p *Session) Send(msg Msg) {
	select {
	case p.msgCh <- msg:
	default:
		// drop if full; configurable via WithMsgBuffer
	}
}

// Quit requests a graceful shutdown (helper).
func (p *Session) Quit() { p.Send(QuitMsg{}) }

// watchSize polls terminal size and emits ResizeMsg on change.
// It sends an initial ResizeMsg if measuring succeeds.
func (p *Session) watchSize(ctx context.Context, out chan<- Msg) {
	fd := int(os.Stdout.Fd())
	lastW, lastH := 0, 0
	// initial measurement
	if w, h, err := term.GetSize(fd); err == nil {
		lastW, lastH = w, h
		out <- ResizeMsg{Width: w, Height: h}
	}
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if w, h, err := term.GetSize(fd); err == nil {
				if w != lastW || h != lastH {
					lastW, lastH = w, h
					out <- ResizeMsg{Width: w, Height: h}
				}
			}
		}
	}
}
