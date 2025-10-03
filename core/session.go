package core

import (
	"context"
	"fmt"
	"io"
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
	m        Model
	renderer Renderer
	input    *input

	// IO
	out io.Writer
	in  io.Reader

	// control
	msgCh          chan Msg
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	startOnce      sync.Once
	stopOnce       sync.Once
	altScreen      bool
	msgBuf         int
	resizeInterval time.Duration
	nonInteractive bool

	// features
	enableMouse         bool
	enableBracketedPaste bool

	logger Logger
}

// WithRenderer sets a custom renderer (useful in tests).
func WithRenderer(r Renderer) Option { return func(p *Session) { p.renderer = r } }

// WithAltScreen switches to the terminal alternate screen while the session runs.
func WithAltScreen() Option { return func(p *Session) { p.altScreen = true } }

// WithMsgBuffer sets the size of the internal message buffer (default 64).
func WithMsgBuffer(n int) Option {
	return func(p *Session) {
		if n > 0 {
			p.msgBuf = n
		}
	}
}

// WithOut sets the output writer (default os.Stdout).
func WithOut(w io.Writer) Option { return func(p *Session) { p.out = w } }

// WithIn sets the input reader (default os.Stdin).
func WithIn(r io.Reader) Option { return func(p *Session) { p.in = r } }

// WithResizeInterval sets the polling interval for terminal size (default 150ms).
func WithResizeInterval(d time.Duration) Option {
	return func(p *Session) {
		if d > 0 {
			p.resizeInterval = d
		}
	}
}

// WithNonInteractive forces non-interactive mode (no raw mode, no input loop).
func WithNonInteractive() Option { return func(p *Session) { p.nonInteractive = true } }

// WithLogger sets a custom logger (defaults to std logger on stderr).
func WithLogger(l Logger) Option { return func(p *Session) { p.logger = l } }

// WithMouse enables SGR mouse reporting.
func WithMouse() Option { return func(p *Session) { p.enableMouse = true } }

// WithBracketedPaste enables bracketed paste (ESC[200~ .. ESC[201~]).
func WithBracketedPaste() Option { return func(p *Session) { p.enableBracketedPaste = true } }

// NewSession creates a session for a given Model.
func NewSession(m Model, opts ...Option) *Session {
	return NewSessionWithContext(context.Background(), m, opts...)
}

// NewSessionWithContext creates a session bound to the provided context.
func NewSessionWithContext(ctx context.Context, m Model, opts ...Option) *Session {
	if ctx == nil {
		ctx = context.Background()
	}
	cctx, cancel := context.WithCancel(ctx)

	p := &Session{
		m:              m,
		out:            os.Stdout,
		in:             os.Stdin,
		msgBuf:         64,
		ctx:            cctx,
		cancel:         cancel,
		resizeInterval: 150 * time.Millisecond,
		logger:         newStdLogger(os.Stderr),
	}
	for _, o := range opts {
		o(p)
	}

	// IO-derived components
	if p.renderer == nil {
		p.renderer = newANSIRenderer(p.out)
	}
	p.input = newInput(p.in)

	// channel
	p.msgCh = make(chan Msg, p.msgBuf)
	return p
}

// Run starts the session and blocks until completion or error.
func (p *Session) Run() (runErr error) {
	p.startOnce.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				p.logger.Errorf("panic: %v", r)
				p.stopOnce.Do(func() {
					p.cancel()
					p.wg.Wait()
					p.renderer.Close()
					p.input.restore()
				})
				runErr = fmt.Errorf("panic: %v", r)
			}
		}()

		// Determine interactive/tty
		isTTY := func(w io.Writer) bool {
			if f, ok := w.(*os.File); ok {
				return term.IsTerminal(int(f.Fd()))
			}
			return false
		}
		autoNonInteractive := !isTTY(p.out)
		effectiveNonInteractive := p.nonInteractive || autoNonInteractive

		if effectiveNonInteractive {
			// no raw, no loops; render once, strip ANSI
			cmd := p.m.Init()
			_ = cmd
			view := p.m.View()
			fmt.Fprintln(p.out, StripANSI(view))
			return
		}

		// Interactive path
		if err := p.input.raw(); err != nil {
			runErr = fmt.Errorf("raw mode: %w", err)
			return
		}
		defer p.input.restore()

		// Alt screen
		if p.altScreen {
			fmt.Fprint(p.out, "\x1b[?1049h")
			defer fmt.Fprint(p.out, "\x1b[?1049l")
		}

		// Feature toggles
		if p.enableMouse {
			// 1000: report clicks, 1002: button-motion, 1006: SGR mode
			fmt.Fprint(p.out, "\x1b[?1000h\x1b[?1002h\x1b[?1006h")
			defer fmt.Fprint(p.out, "\x1b[?1000l\x1b[?1002l\x1b[?1006l")
		}
		if p.enableBracketedPaste {
			fmt.Fprint(p.out, "\x1b[?2004h")
			defer fmt.Fprint(p.out, "\x1b[?2004l")
		}

		// Input reader
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			p.input.readKeys(p.ctx, p.msgCh)
		}()

		// Size watcher (poll)
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			p.watchSize(p.ctx, p.msgCh)
		}()

		// OS signals
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

			case s := <-sigCh:
				p.logger.Infof("signal: %v", s)
				p.msgCh <- QuitMsg{}

			case msg := <-p.msgCh:
				if msg == nil {
					continue
				}
				newModel, cmd := p.m.Update(msg)
				p.m = newModel
				p.renderer.Render(p.m.View())
				if cmd != nil {
					go func(c Cmd) { p.msgCh <- c() }(cmd)
				}
				if _, ok := msg.(QuitMsg); ok {
					break loop
				}
			}
		}


		// 
		// p.stopOnce.Do(func() {
		// 	p.cancel()
		// 	p.wg.Wait()
		// 	p.renderer.Close()
		// })
		p.stopOnce.Do(func() {
			p.cancel()
			p.renderer.Close()
			p.input.restore()

			done := make(chan struct {})
			go func() { p.wg.Wait(); close(done) }()
			select {
			case <- done:
			case <-time.After(200 * time.Millisecond):
			}
		})
	})
	return runErr
}

// Send injects a message from outside (tests or background jobs).
func (p *Session) Send(msg Msg) {
	select {
	case p.msgCh <- msg:
	default:
	}
}

// Quit requests a graceful shutdown (helper).
func (p *Session) Quit() { p.Send(QuitMsg{}) }

// watchSize polls terminal size and emits ResizeMsg on change.
func (p *Session) watchSize(ctx context.Context, out chan<- Msg) {
	fd := func(w io.Writer) int {
		if f, ok := w.(*os.File); ok {
			return int(f.Fd())
		}
		return int(os.Stdout.Fd())
	}(p.out)

	lastW, lastH := 0, 0
	if w, h, err := term.GetSize(fd); err == nil {
		lastW, lastH = w, h
		out <- ResizeMsg{Width: w, Height: h}
	}
	ticker := time.NewTicker(p.resizeInterval)
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
