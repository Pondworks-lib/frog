package validate

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/term"
)

type Code string

const (
	// Errors
	CodeNilModel           Code = "FROG001"
	CodeMissingView        Code = "FROG002"
	CodeEmptyView          Code = "FROG003"
	CodeViewNotString      Code = "FROG004"
	CodeViewHasBadRunes    Code = "FROG005"
	CodeViewPanic          Code = "FROG006"
	CodeMissingUpdate      Code = "FROG007"
	CodeBadUpdateSignature Code = "FROG008"
	CodeMissingInit        Code = "FROG009"
	CodeBadInitSignature   Code = "FROG010"

	// Warnings
	CodeViewVeryLarge   Code = "FROG101"
	CodeViewSuspicious  Code = "FROG102"
	CodeNonExportedType Code = "FROG103"
	CodeSlowView        Code = "FROG104"
	CodeSlowInit        Code = "FROG105"
	CodeHasNoMethods    Code = "FROG106"
	CodeUpdateNotMethod Code = "FROG107"
	CodeViewNotMethod   Code = "FROG108"
)

type Severity int

const (
	SeverityError Severity = iota
	SeverityWarning
)

func (s Severity) String() string {
	if s == SeverityWarning {
		return "warning"
	}
	return "error"
}

type Issue struct {
	Code       Code
	Severity   Severity
	Summary    string
	Detail     string
	Suggestion string
}

func (i Issue) String() string {
	sb := strings.Builder{}
	sb.WriteString(string(i.Code))
	sb.WriteByte(' ')
	sb.WriteString(i.Severity.String())
	sb.WriteString(": ")
	sb.WriteString(i.Summary)
	if i.Detail != "" {
		sb.WriteString("\n  - ")
		sb.WriteString(i.Detail)
	}
	if i.Suggestion != "" {
		sb.WriteString("\n  \u2192 ")
		sb.WriteString(i.Suggestion)
	}
	return sb.String()
}

type Report struct {
	issues []Issue
}

func (r *Report) Add(it Issue) { r.issues = append(r.issues, it) }
func (r *Report) OrNil() error {
	if len(r.issues) == 0 {
		return nil
	}
	return r
}
func (r *Report) HasErrors() bool {
	for _, it := range r.issues {
		if it.Severity == SeverityError {
			return true
		}
	}
	return false
}

func exampleMethodsHelp(modelName string) string {
	if modelName == "" {
		modelName = "MyModel"
	}
	return fmt.Sprintf(`Minimum:

    func (m %s) View() string {
        return "hello"
    }

    func (m %s) Update(msg frog.Msg) (frog.Model, frog.Cmd) {
        return m, nil
    }`, modelName, modelName)
}

func exampleUpdateHelp(modelName string) string {
	if modelName == "" {
		modelName = "MyModel"
	}
	return fmt.Sprintf(`Add:

    func (m %s) Update(msg frog.Msg) (frog.Model, frog.Cmd) {
        // handle messages here
        return m, nil
    }`, modelName)
}

// ----------------------------------------------------
// Color handling
// ----------------------------------------------------

type colorModeT int

const (
	ColorAuto colorModeT = iota
	ColorOn
	ColorOff
)

var colorMode = ColorAuto

func colorsEnabled() bool {
	// Explicit override beats everything
	switch strings.ToLower(os.Getenv("FROG_COLOR")) {
	case "1", "true", "on", "yes":
		return true
	case "0", "false", "off", "no":
		return false
	}

	switch colorMode {
	case ColorOn:
		return true
	case ColorOff:
		return false
	default:
		if os.Getenv("NO_COLOR") != "" {
			return false
		}
		if os.Getenv("TERM") == "dumb" {
			return false
		}
		// Prefer stdout if TTY; otherwise try stderr
		if f := os.Stdout; f != nil && term.IsTerminal(int(f.Fd())) {
			return true
		}
		if ef := os.Stderr; ef != nil && term.IsTerminal(int(ef.Fd())) {
			return true
		}
		return false
	}
}

const (
	cReset  = "\x1b[0m"
	cBold   = "\x1b[1m"
	cRed    = "\x1b[31m"
	cYellow = "\x1b[33m"
	cDim    = "\x1b[2m"
)

func paint(s, color string) string {
	if !colorsEnabled() {
		return s
	}
	return color + s + cReset
}

func (r *Report) Error() string {
	if len(r.issues) == 0 {
		return ""
	}

	// Errors first, then code order
	sort.SliceStable(r.issues, func(i, j int) bool {
		if r.issues[i].Severity != r.issues[j].Severity {
			return r.issues[i].Severity < r.issues[j].Severity
		}
		return string(r.issues[i].Code) < string(r.issues[j].Code)
	})

	full := os.Getenv("FROG_VALIDATE_FULL") == "1"

	if !full {
		it := r.issues[0]
		header := "frog validation failed:"
		if it.Severity == SeverityError {
			header = paint(header, cRed+cBold)
		} else {
			header = paint(header, cYellow+cBold)
		}
		return header + "\n - " + formatIssue(it)
	}

	sb := strings.Builder{}
	sb.WriteString(paint("frog validation failed:", cBold))
	sb.WriteByte('\n')

	// Errors section
	if r.HasErrors() {
		sb.WriteString(paint("Errors:", cRed+cBold))
		sb.WriteByte('\n')
		for _, it := range r.issues {
			if it.Severity == SeverityError {
				sb.WriteString(" - ")
				sb.WriteString(formatIssue(it))
				sb.WriteByte('\n')
			}
		}
	}

	// Warnings section
	hasWarn := false
	for _, it := range r.issues {
		if it.Severity == SeverityWarning {
			hasWarn = true
			break
		}
	}
	if hasWarn {
		sb.WriteString(paint("Warnings:", cYellow+cBold))
		sb.WriteByte('\n')
		for _, it := range r.issues {
			if it.Severity == SeverityWarning {
				sb.WriteString(" - ")
				sb.WriteString(formatIssue(it))
				sb.WriteByte('\n')
			}
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

func formatIssue(it Issue) string {
	s := it.String()
	switch it.Severity {
	case SeverityError:
		return paint(s, cRed)
	case SeverityWarning:
		return paint(s, cYellow)
	default:
		return s
	}
}

// ----------------------------------------------------
// Validation
// ----------------------------------------------------

// timeoutErr distinguishes timeouts from other runtime errors and can carry a location hint.
type timeoutErr struct {
	what string
	loc  string // "path/file.go:123" if we could locate the blocked frame
}

func (e timeoutErr) Error() string {
	if e.loc != "" {
		return fmt.Sprintf("%s at %s", e.what, e.loc)
	}
	return e.what
}

// ValidateModel checks the model shape and safely runs Init/View with timeout & recovery.
func ValidateModel(m any) error {
	rep := &Report{}

	// 1) nil
	if m == nil {
		rep.Add(Issue{
			Code:       CodeNilModel,
			Severity:   SeverityError,
			Summary:    "model is nil",
			Suggestion: "Pass a concrete value to frog.Run(), e.g. Run(MyModel{}).",
		})
		return rep
	}

	mv := reflect.ValueOf(m)
	mt := mv.Type()
	if !mv.IsValid() {
		rep.Add(Issue{Code: CodeNilModel, Severity: SeverityError, Summary: "model is invalid (zero reflect value)"})
		return rep
	}

	// unexported type → warning
	if !isExportedType(mt) {
		rep.Add(Issue{
			Code:       CodeNonExportedType,
			Severity:   SeverityWarning,
			Summary:    fmt.Sprintf("model type %q is not exported", mt.String()),
			Suggestion: "Consider exporting the model type for better reusability.",
		})
	}

	methods := discoverMethods(mv)
	if len(methods) == 0 {
		rep.Add(Issue{
			Code:       CodeHasNoMethods,
			Severity:   SeverityWarning,
			Summary:    "model has no methods; Frog requires at least View() and Update()",
			Suggestion: exampleMethodsHelp(mt.Name()),
		})
	}

	// ---- Init() first ----
	if vInit, ok := methods["Init"]; ok {
		// receiver only → NumIn()==1; 0 or 1 return values (permissive)
		if vInit.Type.NumIn() != 1 || vInit.Type.NumOut() > 1 {
			rep.Add(Issue{
				Code:       CodeBadInitSignature,
				Severity:   SeverityWarning,
				Summary:    "Init() has an unusual signature",
				Detail:     fmt.Sprintf("expected: func() or func() <one-value>, got: %s", prettyMethodType("Init", vInit.Type)),
				Suggestion: "Prefer: func() frog.Cmd or func() (frog.Cmd).",
			})
		} else {
			elapsed, err := safeCallInit(mv, vInit.Func, mt)
			switch e := err.(type) {
			case nil:
				if elapsed > 200*time.Millisecond {
					rep.Add(Issue{
						Code:       CodeSlowInit,
						Severity:   SeverityWarning,
						Summary:    fmt.Sprintf("Init() is slow (took %v)", elapsed),
						Suggestion: "Keep Init() snappy; heavy work should be done asynchronously via Cmd.",
					})
				}
			case timeoutErr:
				rep.Add(Issue{
					Code:       CodeSlowInit,
					Severity:   SeverityWarning,
					Summary:    "Init() exceeded 500ms",
					Detail:     e.Error(),
					Suggestion: "Ensure Init() just schedules background work and returns immediately.",
				})
			default:
				rep.Add(Issue{
					Code:       CodeSlowInit, // keep code stable; message is neutral
					Severity:   SeverityWarning,
					Summary:    "Init() encountered an unexpected error",
					Detail:     e.Error(),
					Suggestion: "Ensure Init() just schedules background work and returns immediately.",
				})
			}
		}
	} else {
		rep.Add(Issue{
			Code:       CodeMissingInit,
			Severity:   SeverityWarning,
			Summary:    "no Init() method",
			Suggestion: "Optional: define Init() to schedule timers or I/O via frog.Tick / commands.",
		})
	}

	// ---- View() ----
	if vView, ok := methods["View"]; !ok {
		rep.Add(Issue{
			Code:       CodeMissingView,
			Severity:   SeverityError,
			Summary:    "missing View() method",
			Suggestion: fmt.Sprintf("Add:\n\n    func (m %s) View() string { return \"...\" }\n", receiverName(mt)),
		})
	} else {
		// should be: func() string (reflect: NumIn==1 for receiver)
		if vView.Type.NumIn() != 1 || vView.Type.NumOut() != 1 || vView.Type.Out(0).Kind() != reflect.String {
			rep.Add(Issue{
				Code:       CodeViewNotString,
				Severity:   SeverityError,
				Summary:    "View() must have signature: func() string",
				Detail:     fmt.Sprintf("got: %s", prettyMethodType("View", vView.Type)),
				Suggestion: "Make sure View has no parameters and returns a string.",
			})
		} else {
			viewRes, elapsed, viewErr := safeCallView(mv, vView.Func, mt)
			switch e := viewErr.(type) {
			case nil:
				out := viewRes
				if out == "" {
					rep.Add(Issue{
						Code:       CodeEmptyView,
						Severity:   SeverityError,
						Summary:    "View() returned an empty string",
						Suggestion: "Always return at least a short string. A blank screen is not allowed.",
					})
				} else if strings.TrimSpace(out) == "" {
					rep.Add(Issue{
						Code:       CodeViewSuspicious,
						Severity:   SeverityWarning,
						Summary:    "View() returns only whitespace",
						Suggestion: "Return meaningful content; avoid only spaces/newlines.",
					})
				}
				if !utf8.ValidString(out) {
					rep.Add(Issue{
						Code:       CodeViewHasBadRunes,
						Severity:   SeverityError,
						Summary:    "View() returned invalid UTF-8",
						Suggestion: "Ensure the returned string is valid UTF-8.",
					})
				}
				if len(out) > 2_000_000 {
					rep.Add(Issue{
						Code:       CodeViewVeryLarge,
						Severity:   SeverityWarning,
						Summary:    "View() returned an extremely large string",
						Detail:     fmt.Sprintf("size=%d bytes", len(out)),
						Suggestion: "Consider incremental rendering or smaller views.",
					})
				}
				if elapsed > 200*time.Millisecond {
					rep.Add(Issue{
						Code:       CodeSlowView,
						Severity:   SeverityWarning,
						Summary:    fmt.Sprintf("View() is slow (took %v)", elapsed),
						Suggestion: "Keep View() fast; precompute data in Update() or background commands.",
					})
				}
			case timeoutErr:
				rep.Add(Issue{
					Code:       CodeSlowView,
					Severity:   SeverityWarning,
					Summary:    "View() exceeded 500ms",
					Detail:     e.Error(),
					Suggestion: "Keep View() fast; precompute data in Update() or background commands.",
				})
			default:
				rep.Add(Issue{
					Code:       CodeViewPanic, // codice storico, testo neutro
					Severity:   SeverityError,
					Summary:    "View() encountered an unexpected error",
					Detail:     e.Error(),
					Suggestion: "Ensure View() is side-effect free and handles zero state safely.",
				})
			}
		}
	}

	// ---- Update(Msg) (Model, Cmd) ----
	if vUpdate, ok := methods["Update"]; !ok {
		rep.Add(Issue{
			Code:       CodeMissingUpdate,
			Severity:   SeverityError,
			Summary:    "missing Update(...) method",
			Suggestion: exampleUpdateHelp(mt.Name()),
		})
	} else {
		// reflect includes receiver: expect 2 inputs (receiver + arg) and 2 outputs
		inN := vUpdate.Type.NumIn()
		outN := vUpdate.Type.NumOut()
		if inN != 2 || outN != 2 {
			rep.Add(Issue{
				Code:     CodeBadUpdateSignature,
				Severity: SeverityError,
				Summary:  "Update has an invalid signature",
				Detail: fmt.Sprintf("expected: func(msg frog.Msg) (frog.Model, frog.Cmd)\n        got: %s",
					prettyMethodType("Update", vUpdate.Type)),
				Suggestion: "Ensure Update takes exactly one parameter and returns (Model, Cmd).",
			})
		}
		if !vUpdate.Func.IsValid() {
			rep.Add(Issue{
				Code:       CodeUpdateNotMethod,
				Severity:   SeverityWarning,
				Summary:    "Update is not a valid method (maybe a field with the same name?)",
				Suggestion: "Define Update as a method on your model type.",
			})
		}
	}

	return rep.OrNil()
}

// ----------------------------------------------------
// helpers (reflection / formatting)
// ----------------------------------------------------

func discoverMethods(v reflect.Value) map[string]reflect.Method {
	out := map[string]reflect.Method{}
	vt := v.Type()

	for i := 0; i < vt.NumMethod(); i++ {
		m := vt.Method(i)
		if isExportedName(m.Name) {
			out[m.Name] = m
		}
	}
	if v.CanAddr() {
		pt := v.Addr().Type()
		for i := 0; i < pt.NumMethod(); i++ {
			m := pt.Method(i)
			if isExportedName(m.Name) {
				out[m.Name] = m // prefer pointer receiver if duplicates
			}
		}
	}
	return out
}

func isExportedName(name string) bool {
	if name == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(name)
	return r >= 'A' && r <= 'Z'
}

func isExportedType(t reflect.Type) bool {
	if t.Name() == "" {
		return true
	}
	return isExportedName(t.Name())
}

func receiverName(t reflect.Type) string {
	if t.Name() != "" {
		return t.Name()
	}
	return t.String()
}

func prettyMethodType(_ string, ft reflect.Type) string {
	sb := strings.Builder{}
	sb.WriteString("func(")
	for i := 0; i < ft.NumIn(); i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(typeName(ft.In(i)))
	}
	sb.WriteString(")")
	switch ft.NumOut() {
	case 0:
	case 1:
		sb.WriteByte(' ')
		sb.WriteString(typeName(ft.Out(0)))
	default:
		sb.WriteString(" (")
		for i := 0; i < ft.NumOut(); i++ {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(typeName(ft.Out(i)))
		}
		sb.WriteString(")")
	}
	return sb.String()
}

func typeName(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Ptr:
		return "*" + typeName(t.Elem())
	case reflect.Slice:
		return "[]" + typeName(t.Elem())
	case reflect.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), typeName(t.Elem()))
	case reflect.Map:
		return "map[" + typeName(t.Key()) + "]" + typeName(t.Elem())
	case reflect.Func:
		return prettyMethodType("", t)
	case reflect.Interface:
		if t.NumMethod() == 0 {
			return "interface{}"
		}
	}
	if t.PkgPath() != "" && t.Name() != "" {
		return t.PkgPath() + "." + t.Name()
	}
	if t.Name() != "" {
		return t.Name()
	}
	return t.String()
}

// ----------------------------------------------------
// Locate blocked frame on timeout (best effort)
// ----------------------------------------------------

func methodSymbols(t reflect.Type, method string) []string {
	pkg := t.PkgPath()
	name := t.Name()
	if pkg == "" || name == "" {
		// anonymous/unnamed type: fallback to bare method
		return []string{method + "("}
	}
	return []string{
		pkg + ".(*" + name + ")." + method + "(",
		pkg + "." + name + "." + method + "(",
	}
}

func findMethodLocInAllGoroutines(symbols []string) (string, bool) {
	// grow buffer until complete
	size := 64 * 1024
	for size <= 8*1024*1024 {
		buf := make([]byte, size)
		n := runtime.Stack(buf, true)
		if n < size {
			lines := strings.Split(string(buf[:n]), "\n")
			// scan pairs: func-line followed by file:line
			for i := 0; i+1 < len(lines); i++ {
				fn := lines[i]
				for _, sym := range symbols {
					if strings.Contains(fn, sym) {
						next := strings.TrimSpace(lines[i+1])
						// typically: /path/file.go:123 +0x..
						if idx := strings.Index(next, " +"); idx > 0 {
							return next[:idx], true
						}
						return next, true
					}
				}
			}
			return "", false
		}
		size *= 2
	}
	return "", false
}

// ----------------------------------------------------
// Safe calls with timeout & recovery
// ----------------------------------------------------

func safeCallView(mv reflect.Value, fn reflect.Value, mt reflect.Type) (out string, elapsed time.Duration, err error) {
	start := time.Now()
	done := make(chan struct{})
	var res string
	var callErr error

	go func() {
		defer func() {
			if r := recover(); r != nil {
				callErr = enrichError(r)
			}
			close(done)
		}()
		values := fn.Call([]reflect.Value{mv}) // only receiver
		if len(values) != 1 || values[0].Kind() != reflect.String {
			callErr = errors.New("View() did not return a string at runtime")
			return
		}
		res = values[0].String()
	}()

	select {
	case <-done:
		return res, time.Since(start), callErr
	case <-time.After(500 * time.Millisecond):
		loc, _ := findMethodLocInAllGoroutines(methodSymbols(mt, "View"))
		return "", 500 * time.Millisecond, timeoutErr{what: "View() timed out (>500ms)", loc: loc}
	}
}

func safeCallInit(mv reflect.Value, fn reflect.Value, mt reflect.Type) (elapsed time.Duration, err error) {
	start := time.Now()
	done := make(chan struct{})
	var callErr error

	go func() {
		defer func() {
			if r := recover(); r != nil {
				callErr = enrichError(r)
			}
			close(done)
		}()
		_ = fn.Call([]reflect.Value{mv}) // ignore returns
	}()

	select {
	case <-done:
		return time.Since(start), callErr
	case <-time.After(500 * time.Millisecond):
		loc, _ := findMethodLocInAllGoroutines(methodSymbols(mt, "Init"))
		return 500 * time.Millisecond, timeoutErr{what: "Init() timed out (>500ms)", loc: loc}
	}
}

func enrichError(r any) error {
	if e, ok := r.(error); ok {
		// add location
		return fmt.Errorf("unexpected error: %v (%s)", e, findCaller(4))
	}
	return fmt.Errorf("unexpected error: %v (%s)", r, findCaller(4))
}

func findCaller(skip int) string {
	pc := make([]uintptr, 8)
	n := runtime.Callers(skip, pc)
	frames := runtime.CallersFrames(pc[:n])
	f, _ := frames.Next()
	return fmt.Sprintf("%s:%d", f.File, f.Line)
}
