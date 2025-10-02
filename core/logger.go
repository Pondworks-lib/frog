package core

import (
	"fmt"
	"io"
	"log"
	"os"
)

type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

type noopLogger struct{}

func (noopLogger) Debugf(string, ...any) {}
func (noopLogger) Infof(string, ...any)  {}
func (noopLogger) Warnf(string, ...any)  {}
func (noopLogger) Errorf(string, ...any) {}

type stdLogger struct {
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	err   *log.Logger
}

func newStdLogger(out io.Writer) Logger {
	if out == nil {
		out = os.Stderr
	}
	flags := log.Lmsgprefix
	return &stdLogger{
		debug: log.New(out, "DEBUG ", flags),
		info:  log.New(out, "INFO  ", flags),
		warn:  log.New(out, "WARN  ", flags),
		err:   log.New(out, "ERROR ", flags),
	}
}

func (l *stdLogger) Debugf(f string, a ...any) { l.debug.Printf(f, a...) }
func (l *stdLogger) Infof(f string, a ...any)  { l.info.Printf(f, a...) }
func (l *stdLogger) Warnf(f string, a ...any)  { l.warn.Printf(f, a...) }
func (l *stdLogger) Errorf(f string, a ...any) { l.err.Printf(f, a...) }

// fmtLogger is a tiny adapter around an io.Writer (for tests).
type fmtLogger struct{ w io.Writer }

func (l fmtLogger) Debugf(f string, a ...any) { fmt.Fprintf(l.w, f+"\n", a...) }
func (l fmtLogger) Infof(f string, a ...any)  { fmt.Fprintf(l.w, f+"\n", a...) }
func (l fmtLogger) Warnf(f string, a ...any)  { fmt.Fprintf(l.w, f+"\n", a...) }
func (l fmtLogger) Errorf(f string, a ...any) { fmt.Fprintf(l.w, f+"\n", a...) }
