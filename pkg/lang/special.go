package lang

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/wetware/ww/internal/mem"
	ww "github.com/wetware/ww/pkg"
	"github.com/wetware/ww/pkg/lang/core"
	capnp "zombiezen.com/go/capnproto2"
)

// ErrParseSpecial is returned when parsing a special form invocation
// fails due to malformed syntax.
var ErrParseSpecial = errors.New("invalid special form")

func parseDo(a core.Analyzer, env core.Env, args core.Seq) (core.Expr, error) {
	var de DoExpr
	err := core.ForEach(args, func(item ww.Any) (bool, error) {
		expr, err := a.Analyze(env, item)
		if err != nil {
			return true, err
		}
		de.Exprs = append(de.Exprs, expr)
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return de, nil
}

func parseIf(a core.Analyzer, env core.Env, args core.Seq) (core.Expr, error) {
	count, err := args.Count()
	if err != nil {
		return nil, err
	} else if count != 2 && count != 3 {
		return nil, core.Error{
			Cause:   fmt.Errorf("%w: if", ErrParseSpecial),
			Message: fmt.Sprintf("requires 2 or 3 arguments, got %d", count),
		}
	}

	exprs := [3]core.Expr{}
	for i := 0; i < count; i++ {
		f, err := args.First()
		if err != nil {
			return nil, err
		}

		expr, err := a.Analyze(env, f)
		if err != nil {
			return nil, err
		}
		exprs[i] = expr

		args, err = args.Next()
		if err != nil {
			return nil, err
		}
	}

	return IfExpr{
		Test: exprs[0],
		Then: exprs[1],
		Else: exprs[2],
	}, nil
}
func parseQuote(a core.Analyzer, _ core.Env, args core.Seq) (core.Expr, error) {
	if count, err := args.Count(); err != nil {
		return nil, err
	} else if count != 1 {
		return nil, core.Error{
			Cause:   fmt.Errorf("%w: quote", ErrParseSpecial),
			Message: fmt.Sprintf("requires exactly 1 argument, got %d", count),
		}
	}

	first, err := args.First()
	if err != nil {
		return nil, err
	}

	return ConstExpr{Form: first}, nil
}

func parseDef(a core.Analyzer, env core.Env, args core.Seq) (core.Expr, error) {
	e := core.Error{Cause: fmt.Errorf("%w: def", ErrParseSpecial)}

	if args == nil {
		return nil, e.With("requires exactly 2 args, got 0")
	}

	if count, err := args.Count(); err != nil {
		return nil, err
	} else if count != 2 {
		return nil, e.With(fmt.Sprintf(
			"requires exactly 2 arguments, got %d", count))
	}

	first, err := args.First()
	if err != nil {
		return nil, err
	}

	sym, ok := first.(core.Symbol)
	if !ok {
		return nil, e.With(fmt.Sprintf(
			"first arg must be symbol, not '%s'", reflect.TypeOf(first)))
	}

	symStr, err := sym.Symbol()
	if err != nil {
		return nil, err
	}

	rest, err := args.Next()
	if err != nil {
		return nil, err
	}

	second, err := rest.First()
	if err != nil {
		return nil, err
	}

	res, err := a.Analyze(env, second)
	if err != nil {
		return nil, err
	}

	return DefExpr{
		Name:  symStr,
		Value: res,
	}, nil
}

// parseFn parses the (fn name? [<params>*] <body>*) or
// (fn name? ([<params>*] <body>*)+) special forms and returns a function value.
func parseFn(a core.Analyzer, env core.Env, args core.Seq) (core.Expr, error) {
	fn, err := parseFnDef(env, args, false)
	return ConstExpr{fn}, err
}

// parseFn parses the (macro name? [<params>*] <body>*) or
// (macro name? ([<params>*] <body>*)+) special forms and returns a macro value.
func parseMacro(_ core.Analyzer, env core.Env, args core.Seq) (core.Expr, error) {
	fn, err := parseFnDef(env, args, true)
	return ConstExpr{fn}, err
}

func parseFnDef(env core.Env, seq core.Seq, macro bool) (ww.Any, error) {
	if seq == nil {
		return nil, errors.New("nil argument sequence")
	}

	args, err := core.ToSlice(seq)
	if err != nil {
		return nil, err
	}

	if len(args) < 1 {
		return nil, fmt.Errorf("%w: got %d, want at-least 1", core.ErrArity, len(args))
	}

	var b core.FuncBuilder
	b.Start(capnp.SingleSegment(nil))
	b.SetMacro(macro)

	// Set function name?
	if sym := args[0].Value(); sym.Which() == mem.Any_Which_symbol {
		name, err := sym.Symbol()
		if err != nil {
			return nil, err
		}

		b.SetName(name)
		args = args[1:]
	}

	// Set call signatures.
	switch mv := args[0].Value(); mv.Which() {
	case mem.Any_Which_vector:
		b.AddTarget(args[0], args[1:])

	case mem.Any_Which_list:
		for _, any := range args {
			if seq, ok := any.(core.Seq); ok {
				b.AddSeq(seq)
			}
		}

	default:
		return nil, errors.New("syntax error")

	}

	return b.Commit()
}

func parseGo(a core.Analyzer, env core.Env, seq core.Seq) (core.Expr, error) {
	return nil, errors.New("parseGo NOT IMPLEMENTED")
	// args, err := core.ToSlice(seq)
	// if err != nil {
	// 	return nil, err
	// }

	// if len(args) == 0 {
	// 	return nil, errors.Errorf("expected at least one argument, got %d", len(args))
	// }

	// if p, ok := procArgs(args).Remote(); ok {
	// 	return RemoteGoExpr{
	// 		Root: a.root,
	// 		Path: p,
	// 		Args: procArgs(args).Args(),
	// 	}, nil
	// }

	// return LocalGoExpr{
	// 	Args: procArgs(args).Args(),
	// }, nil
}

func parseEval(a core.Analyzer, env core.Env, seq core.Seq) (core.Expr, error) {
	var dex DoExpr
	return dex, core.ForEach(seq, func(item ww.Any) (bool, error) {
		expr, err := a.Analyze(env, item)
		if err == nil {
			dex.Exprs = append(dex.Exprs, expr)
		}
		return false, err
	})
}

func importParser(src PathProvider) SpecialParser {
	return func(a core.Analyzer, _ core.Env, seq core.Seq) (core.Expr, error) {
		if cnt, err := seq.Count(); err != nil {
			return nil, err
		} else if cnt != 1 {
			return nil, fmt.Errorf("expected 1 argument, got %d", cnt)
		}

		arg, err := seq.First()
		if err != nil {
			return nil, err
		}

		var s string
		var paths []string
		switch mv := arg.Value(); mv.Which() {
		case mem.Any_Which_keyword:
			if s, err = mv.Keyword(); err != nil {
				return nil, err
			}

			if s != "prelude" {
				return nil, fmt.Errorf("unrecognize kwarg '%s'", s)
			}

			paths, err = src.Paths().prelude()

		case mem.Any_Which_symbol:
			if s, err = mv.Symbol(); err != nil {
				return nil, err
			}

			if s, err = src.Paths().resolve(s); err == nil {
				paths = append(paths, s)
			}

		default:
			err = fmt.Errorf("invalid argument type %s", mv.Which())

		}

		return ImportExpr{Analyzer: a, Paths: paths}, err
	}
}

func (ps PathSet) prelude() (paths []string, err error) {
	var p string
	paths = make([]string, 0, len(ps))
	for _, root := range ps {
		if p, err = ps.resolvePath(root, ""); err == nil {
			paths = append(paths, p)
		}
	}

	return
}

func (ps PathSet) resolve(symbol string) (path string, err error) {
	if symbol[0] == '.' {
		return ps.resolvePath(".", strings.TrimLeft(symbol, "."))
	}

	for _, root := range ps {
		if path, err = ps.resolvePath(root, symbol); err == nil {
			break
		}
	}

	return path, err
}

func (PathSet) resolvePath(root, rel string) (string, error) {
	// TODO:  this does not handle /foo/bar/*.ww

	path := filepath.Clean(filepath.Join(root, rel))
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		path = filepath.Join(path, "__init__.ww")
		_, err = os.Stat(path)
	}

	return path, err
}
