package shell

import (
	"errors"
	"io"
	"strings"

	"github.com/chzyer/readline"
	ww "github.com/wetware/ww/pkg"
	"github.com/wetware/ww/pkg/lang"
	"github.com/wetware/ww/pkg/lang/core"
	"github.com/wetware/ww/pkg/lang/reader"
	"go.uber.org/fx"
)

// Printer consumes output values from the REPL.
type Printer interface {
	Print(ww.Any) error
	PrintErr(error) error
}

// REPLConfig contain fx-injectable parameters for REPL.
type REPLConfig struct {
	fx.In

	VM       lang.VM
	Readline *readline.Instance
	Printer  Printer
}

func newREPL(cfg REPLConfig) *REPL {
	return &REPL{
		Input:     cfg.Readline,
		Evaluator: cfg.VM,
		Printer:   cfg.Printer,
	}
}

// REPL implements a read-eval-print loop for a generic Runtime.
type REPL struct {
	Input     *readline.Instance
	Evaluator lang.VM
	Printer   Printer

	err    error
	reader *reader.Reader

	LineNo    int
	CurrentNS string
}

// Init should be run once before iterating through the REPL.
func (repl *REPL) Init() {
	dummy := strings.NewReader("(error \"repl.Init()::unreachable\")")
	repl.reader = reader.New(dummy)
	repl.reader.File = "<uninitialized>"

	repl.err = repl.Evaluator.Init()
}

// Err returns the error that halted the REPL.
func (repl *REPL) Err() error {
	if errors.Is(repl.err, io.EOF) {
		return nil
	}

	return repl.err
}

// More returns true if there is more data in the
// byte stream.
func (repl *REPL) More() bool { return repl.err == nil }

// SetPrompt for the next read iteration.
func (repl *REPL) SetPrompt(prompt string) { repl.Input.SetPrompt(prompt) }

// Next iteration of the Read-Evaluate-Print loop.
func (repl *REPL) Next() {
	if repl.err != nil {
		return
	}

	if err := repl.next(); err != nil {
		switch err.(type) {
		case reader.Error, core.Error:
			repl.err = repl.Printer.PrintErr(err)

		default:
			repl.err = err
		}
	}
}

func (repl *REPL) next() error {
	forms, err := repl.read()
	if err != nil {
		return err
	}

	if forms, err = repl.eval(forms); err != nil {
		return err
	}

	if err = repl.print(forms); err != nil {
		return err
	}

	return nil
}

func (repl *REPL) read() ([]ww.Any, error) {
	var b strings.Builder
	b.Grow(32) // best-effort pre-allocation

	for {
		line, err := repl.Input.Readline()
		if err != nil {
			if repl.Input.Clean(); err == readline.ErrInterrupt {
				continue
			}

			return nil, err
		}

		if strings.TrimSpace(line) == "" {
			continue
		}

		b.WriteString(line)
		b.WriteString("\n")

		repl.reader.Reset(strings.NewReader(b.String()))
		repl.reader.File = "REPL"

		forms, err := repl.reader.All()
		if err != nil {
			if errors.Is(err, reader.ErrEOF) {
				repl.LineNo++
				continue
			}
		}

		return forms, err
	}
}

func (repl *REPL) eval(forms []ww.Any) (res []ww.Any, err error) {
	res = make([]ww.Any, len(forms))
	for i, form := range forms {
		if res[i], err = repl.Evaluator.Eval(form); err != nil {
			break
		}
	}

	return
}

func (repl *REPL) print(forms []ww.Any) (err error) {
	if len(forms) > 0 {
		err = repl.Printer.Print(forms[len(forms)-1])
	}

	return
}
