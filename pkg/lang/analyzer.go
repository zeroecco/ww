package lang

import (
	"errors"
	"strings"
	"sync"

	"github.com/wetware/ww/internal/mem"
	ww "github.com/wetware/ww/pkg"
	"github.com/wetware/ww/pkg/lang/core"
)

var _ core.Analyzer = (*Analyzer)(nil)

// ErrNoExpand is returned during macro-expansion to indicate the form
// was not expanded.
var ErrNoExpand = errors.New("no macro expansion")

// SpecialParser defines a special form.
type SpecialParser func(core.Analyzer, core.Env, core.Seq) (core.Expr, error)

// PathProvider returns a set of paths from which wetware source code
// can be loaded.
type PathProvider interface {
	Paths() PathSet
}

// PathSet is a collection of source directories.
type PathSet []string

// Analyzer for wetware syntax
type Analyzer struct {
	Root ww.Anchor
	Src  PathProvider

	init    sync.Once
	special map[string]SpecialParser
}

// Analyze performs syntactic analysis of given form and returns an Expr
// that can be evaluated for result against an Env.
func (a *Analyzer) Analyze(env core.Env, any ww.Any) (core.Expr, error) {
	a.init.Do(func() {
		if a.Src == nil {
			a.Src = noSrc{}
		}

		a.special = map[string]SpecialParser{
			"do":    parseDo,
			"if":    parseIf,
			"def":   parseDef,
			"fn":    parseFn,
			"macro": parseMacro,
			"quote": parseQuote,
			// "go": goParser(a.Root),
			"eval":   parseEval,
			"import": importParser(a.Src),
		}
	})

	if core.IsNil(any) {
		return ConstExpr{Form: core.Nil{}}, nil
	}

	form, err := a.macroExpand(env, any)
	if err != nil {
		if !errors.Is(err, ErrNoExpand) {
			return nil, err
		}

		form = any // no expansion; use unmodified form
	}

	switch f := form.(type) {
	case core.Symbol:
		return ResolveExpr{f}, nil

	case core.UnboundPath:
		return PathExpr{
			Root: a.Root,
			Path: f,
		}, nil

	case core.Vector:
		return VectorExpr{
			eval:   a.Eval,
			Vector: f,
		}, nil

	case core.Seq:
		return a.analyzeSeq(env, f)

	case core.Fn:
		return FnExpr{Analyzer: a, Fn: f}, nil

	}

	return ConstExpr{form}, nil
}

func (a *Analyzer) analyzeSeq(env core.Env, seq core.Seq) (core.Expr, error) {
	// Return an empty sequence unmodified.
	cnt, err := seq.Count()
	if err != nil || cnt == 0 {
		return ConstExpr{seq}, err
	}

	// Analyze the call target.  This is the first item in the sequence.
	// Call targets come in several flavors.
	target, err := seq.First()
	if err != nil {
		return nil, err
	}

	if seq, err = seq.Next(); err != nil {
		return nil, err
	}

	// The call target may be a special form.  In this case, we need to get the
	// corresponding parser function, which will take care of parsing/analyzing
	// the tail.
	if any := target.Value(); any.Which() == mem.Any_Which_symbol {
		s, err := any.Symbol()
		if err != nil {
			return nil, err
		}

		if parse, found := a.special[s]; found {
			return parse(a, env, seq)
		}
	}

	// Target is not a special form; resolve and evaluate to get an
	// invokable value or function.
	if target, err = a.Eval(env, target); err != nil {
		return nil, err
	}

	// The call target is not a special form.  It is some kind of invokation.
	// Unpack arguments.
	as, err := a.unpackArgs(env, seq)
	if err != nil {
		return nil, err
	}

	// Analyze arguments.
	args := make([]core.Expr, len(as))
	for i, arg := range as {
		if args[i], err = a.Analyze(env, arg); err != nil {
			return nil, err
		}
	}

	// Call target is not a special form and must be a Invokable. Analyze
	// the arguments and create an InvokeExpr.
	iex := InvokeExpr{Args: args}
	iex.Target, err = a.Analyze(env, target)
	return iex, err
}

func (a *Analyzer) unpackArgs(env core.Env, seq core.Seq) (args []ww.Any, err error) {
	if seq == nil {
		return
	}

	if args, err = core.ToSlice(seq); err != nil || len(args) == 0 {
		return
	}

	varg := args[len(args)-1]
	any := varg.Value()

	// not vargs?
	if any.Which() != mem.Any_Which_symbol {
		return
	}

	var sym string
	if sym, err = any.Symbol(); err != nil {
		return
	}

	// no unpacking?
	if !strings.HasSuffix(sym, "...") {
		return
	}

	// It's a varg.  Symbol or collection literal form?
	switch sym {
	case "...":
		if len(args) < 2 {
			err = errors.New("invalid syntax (no vargs to unpack)")
			return
		}

		varg = args[len(args)-2]
		args = args[:len(args)-1]

	default:
		// foo...
		if varg, err = resolve(env, sym[:len(sym)-3]); err != nil {
			return
		}
	}

	// Evaluate the varg.  This will notably unquote sequences.
	if varg, err = a.Eval(env, varg); err != nil {
		return
	}

	// Coerce the varg into a sequence.
	switch v := varg.(type) {
	case core.Seq:
		// '(:foo :bar)...'
		seq = v

	case core.Seqable:
		// '[:foo :bar]...'
		if seq, err = v.Seq(); err != nil {
			return
		}

	default:
		err = errors.New("invalid syntax (vargs)")
		return

	}

	args = args[:len(args)-1] // pop last
	err = core.ForEach(seq, func(item ww.Any) (bool, error) {
		args = append(args, item)
		return false, nil
	})

	return
}

// Eval a form against the supplied environment.
func (a *Analyzer) Eval(env core.Env, form ww.Any) (ww.Any, error) {
	expr, err := a.Analyze(env, form)
	if err != nil {
		return nil, err
	}

	return expr.Eval(env)
}

func (a *Analyzer) macroExpand(env core.Env, form ww.Any) (ww.Any, error) {
	seq, ok := form.(core.Seq)
	if !ok {
		return nil, ErrNoExpand
	}

	cnt, err := seq.Count()
	if err != nil {
		return nil, err
	}
	if cnt == 0 {
		return nil, ErrNoExpand
	}

	first, err := seq.First()
	if err != nil {
		return nil, err
	}

	var v interface{}
	if any := first.Value(); any.Which() == mem.Any_Which_symbol {
		var rex ResolveExpr
		rex.Symbol.Any = any
		if v, err = rex.Eval(env); err != nil {
			return nil, ErrNoExpand
		}
	}

	fn, ok := v.(core.Fn)
	if !ok || !fn.Macro() {
		return nil, ErrNoExpand
	}

	// pop head; seq is now args.
	if seq, err = seq.Next(); err != nil {
		return nil, err
	}

	// N.B.:  macro functions receive unevaluated values, so
	//		  no analyze/eval here.
	args, err := core.ToSlice(seq)
	if err != nil {
		return nil, err
	}

	name, err := fn.Name()
	if err != nil {
		return nil, err
	}

	return core.BoundFn{
		Analyzer: a,
		Fn:       fn,
		Env:      env.Child(name, nil),
	}.Invoke(args)
}

func resolve(env core.Env, symbol string) (any ww.Any, err error) {
	var v interface{}
	for env != nil {
		if v, err = env.Resolve(symbol); err != nil && !errors.Is(err, core.ErrNotFound) {
			// found symbol, or there was some unexpected error
			break
		}

		// not found in the current frame. check parent.
		env = env.Parent()
	}

	if err == nil {
		any = v.(ww.Any)
	}

	return
}

type noSrc struct{}

func (noSrc) Paths() PathSet { return PathSet{} }
