package core

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Option func(*Session)

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
}

func WithRenderer(r Renderer) Option {
	return func(p *Session) { p.renderer = r }
}

func NewSession(m Model, opts ...Option) *Session {
	ctx, cancel := context.WithCancel(context.Background())
	p := &Session{
		m:        m,
		renderer: newANSIRenderer(os.Stdout),
		input:    newInput(),
		msgCh:    make(chan Msg, 64),
		ctx:      ctx,
		cancel:   cancel,
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

func (p *Session) Run() error {
	var runErr error
	p.startOnce.Do(func() {
		if err := p.input.raw(); err != nil {
			runErr = fmt.Errorf("raw mode: %w", err)
			return
		}
		defer p.input.restore()

		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			p.input.readKeys(p.ctx, p.msgCh)
		}()

		sigCh := make(chan os.Signal, 2)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(sigCh)

		cmd := p.m.Init()
		p.renderer.Clear()
		p.renderer.Render(p.m.View())

		if cmd != nil {
			go func(c Cmd) { p.msgCh <- c() }(cmd)
		}

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

		p.stopOnce.Do(func() {
			p.cancel()
			p.wg.Wait()
			p.renderer.Close()
		})
	})
	return runErr
}

func (p *Session) Send(msg Msg) {
	select {
	case p.msgCh <- msg:
	default:
	}
}
