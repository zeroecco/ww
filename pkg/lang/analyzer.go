package lang

import (
	"errors"
	"strings"

	"github.com/spy16/slurp/builtin"
	score "github.com/spy16/slurp/core"
	"github.com/wetware/ww/internal/mem"
	ww "github.com/wetware/ww/pkg"
	"github.com/wetware/ww/pkg/lang/core"
)

var _ core.Analyzer = (*analyzer)(nil)

// SpecialParser defines a special form.
type SpecialParser func(core.Analyzer, core.Env, core.Seq) (core.Expr, error)

type analyzer struct {
	root    ww.Anchor
	special map[string]SpecialParser
}

func newAnalyzer(root ww.Anchor, paths []string) (core.Analyzer, error) {
	if root == nil {
		return nil, errors.New("nil anchor")
	}

	return analyzer{
		root: root,
		special: map[string]SpecialParser{
			"do":    parseDo,
			"if":    parseIf,
			"def":   parseDef,
			"fn":    parseFn,
			"macro": parseMacro,
			"quote": parseQuote,
			// "go": c.Go,
			"eval":   parseEval,
			"import": importer(paths).Parse,
		},
	}, nil
}

// Analyze performs syntactic analysis of given form and returns an Expr
// that can be evaluated for result against an Env.
func (a analyzer) Analyze(env core.Env, rawForm score.Any) (core.Expr, error) {
	return a.analyze(env, rawForm.(ww.Any))
}

// analyze allows private methods of `analyzer` to by pass the initial
// type assertion for `ww.Any`.
func (a analyzer) analyze(env core.Env, any ww.Any) (core.Expr, error) {
	if core.IsNil(any) {
		return builtin.ConstExpr{Const: core.Nil{}}, nil
	}

	form, err := a.macroExpand(env, any)
	if err != nil {
		if !errors.Is(err, builtin.ErrNoExpand) {
			return nil, err
		}

		form = any // no expansion; use unmodified form
	}

	switch f := form.(type) {
	case core.Symbol:
		return ResolveExpr{f}, nil

	case core.UnboundPath:
		return PathExpr{
			Root: a.root,
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
		return FnExpr{Fn: f}, nil

	}

	return ConstExpr{form}, nil
}

func (a analyzer) analyzeSeq(env core.Env, seq core.Seq) (core.Expr, error) {
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
		if args[i], err = a.analyze(env, arg); err != nil {
			return nil, err
		}
	}

	// Call target is not a special form and must be a Invokable. Analyze
	// the arguments and create an InvokeExpr.
	iex := InvokeExpr{Args: args}
	iex.Target, err = a.analyze(env, target)
	return iex, err
}

func (a analyzer) unpackArgs(env core.Env, seq core.Seq) (args []ww.Any, err error) {
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

func (a analyzer) Eval(env core.Env, any ww.Any) (ww.Any, error) {
	expr, err := a.analyze(env, any)
	if err != nil {
		return nil, err
	}

	v, err := expr.Eval(env)
	if err == nil {
		any = v.(ww.Any)
	}

	return any, err
}

func (a analyzer) macroExpand(env core.Env, form ww.Any) (ww.Any, error) {
	seq, ok := form.(core.Seq)
	if !ok {
		return nil, builtin.ErrNoExpand
	}

	cnt, err := seq.Count()
	if err != nil {
		return nil, err
	}
	if cnt == 0 {
		return nil, builtin.ErrNoExpand
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
			return nil, builtin.ErrNoExpand
		}
	}

	fn, ok := v.(core.Fn)
	if !ok || !fn.Macro() {
		return nil, builtin.ErrNoExpand
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
