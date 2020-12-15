package lang

import (
	"bytes"
	"context"
	"errors"
	"os"

	"github.com/spy16/slurp/builtin"
	score "github.com/spy16/slurp/core"
	ww "github.com/wetware/ww/pkg"
	"github.com/wetware/ww/pkg/lang/core"
	"github.com/wetware/ww/pkg/lang/reader"
	memutil "github.com/wetware/ww/pkg/util/mem"
	capnp "zombiezen.com/go/capnproto2"
)

var (
	_ core.Expr = (*ConstExpr)(nil)
	_ core.Expr = (*IfExpr)(nil)
	_ core.Expr = (*ResolveExpr)(nil)
	_ core.Expr = (*DefExpr)(nil)
	_ core.Expr = (*InvokeExpr)(nil)
	_ core.Expr = (*PathExpr)(nil)
	// _ core.Expr = (*LocalGoExpr)(nil)
	// _ core.Expr = (*RemoteGoExpr)(nil)
	_ core.Expr = (*InvokeExpr)(nil)
	// _ core.Expr = (*)(nil)
)

type (
	// DoExpr represents the (do expr*) form.
	DoExpr = builtin.DoExpr

	// QuoteExpr expression represents a quoted form
	QuoteExpr = builtin.QuoteExpr
)

// ConstExpr returns the Const value wrapped inside when evaluated. It has
// no side-effect on the VM.
type ConstExpr struct{ Form ww.Any }

// Eval returns the constant value unmodified.
func (ce ConstExpr) Eval(_ core.Env) (score.Any, error) { return ce.Form, nil }

// IfExpr represents the if-then-else form.
type IfExpr struct{ Test, Then, Else core.Expr }

// Eval evaluates the then or else expr based on truthiness of the test
// expr result.
func (ife IfExpr) Eval(env core.Env) (score.Any, error) {
	target := ife.Else
	if ife.Test != nil {
		test, err := ife.Test.Eval(env)
		if err != nil {
			return nil, err
		}

		ok, err := core.IsTruthy(test.(ww.Any))
		if err != nil {
			return nil, err
		}

		if ok {
			target = ife.Then
		}
	}

	if target == nil {
		return core.Nil{}, nil
	}
	return target.Eval(env)
}

// ResolveExpr resolves a symbol from the given environment.
type ResolveExpr struct{ Symbol core.Symbol }

// Eval resolves the symbol in the given environment or its parent env
// and returns the result. Returns ErrNotFound if the symbol was not
// found in the entire hierarchy.
func (re ResolveExpr) Eval(env core.Env) (v score.Any, err error) {
	var sym string
	if sym, err = re.Symbol.Symbol(); err != nil {
		return
	}

	for env != nil {
		if v, err = env.Resolve(sym); errors.Is(err, core.ErrNotFound) {
			// not found in the current frame. check parent.
			env = env.Parent()
			continue
		}

		// found the symbol or there was some unexpected error.
		break
	}

	return
}

// DefExpr represents the (def name value) binding form.
type DefExpr struct {
	Name  string
	Value core.Expr
}

// Eval creates the binding with the name and value in Root env.
func (de DefExpr) Eval(env core.Env) (score.Any, error) {
	var val score.Any
	var err error
	if de.Value != nil {
		val, err = de.Value.Eval(env)
		if err != nil {
			return nil, err
		}
	} else {
		val = core.Nil{}
	}

	if err := score.Root(env).Bind(de.Name, val); err != nil {
		return nil, err
	}

	return core.NewSymbol(capnp.SingleSegment(nil), de.Name)
}

// CallExpr invokes a function body when evaluated.
type CallExpr struct {
	Fn       core.Fn
	Analyzer core.Analyzer
	Args     []core.Expr
}

// Eval calls the function.
func (cex CallExpr) Eval(env core.Env) (score.Any, error) {
	// Get the call target that corresponds to the number of
	// arguments supplied.
	ct, err := cex.Fn.Match(len(cex.Args))
	if err != nil {
		return nil, err
	}

	// Bind evaluated arguments to parameter names to build
	// a map of local variables.
	scope := make(map[string]score.Any, len(ct.Param))
	var vargs []ww.Any
	for i, arg := range cex.Args {
		any, err := arg.Eval(env)
		if err != nil {
			return nil, err
		}

		if ct.Variadic && i >= len(ct.Param)-1 {
			vargs = append(vargs, any.(ww.Any))
			continue
		}

		scope[ct.Param[i]] = any
	}

	if vargs != nil {
		vs, err := core.NewList(capnp.SingleSegment(nil), vargs...)
		if err != nil {
			return nil, err
		}

		scope[ct.Param[len(ct.Param)-1]] = vs
	}

	// Analyze the call target's body to obtain evaluable expressions.
	body := make([]core.Expr, len(ct.Body))
	for i, form := range ct.Body {
		if body[i], err = cex.Analyzer.Analyze(env, form); err != nil {
			return nil, err
		}
	}

	// Derive a child environment and evaluate the function body as a
	// do expression.
	return DoExpr{Exprs: body}.Eval(env.Child(ct.Name, scope))
}

// InvokeExpr performs invocation of target when evaluated.
type InvokeExpr struct {
	Target core.Invokable
	Args   []core.Expr
}

// Eval evaluates the target expr and invokes the result if it is an
// Invokable  Returns error otherwise.
func (ie InvokeExpr) Eval(env core.Env) (any score.Any, err error) {
	args := make([]ww.Any, len(ie.Args))
	for i, ae := range ie.Args {
		if any, err = ae.Eval(env); err != nil {
			return
		}

		args[i] = any.(ww.Any)
	}

	return ie.Target.Invoke(args...)
}

// PathExpr binds a path to an Anchor
type PathExpr struct {
	Root ww.Anchor
	Path core.UnboundPath
}

// Eval walks the specified path, starting from the root anchor,
// and binds the result to the path.
func (pex PathExpr) Eval(core.Env) (score.Any, error) {
	parts, err := pex.Path.Parts()
	if err != nil {
		return core.BoundPath{}, err
	}

	return core.BoundPath{
		Any:    pex.Path.Any,
		Anchor: pex.Root.Walk(context.TODO(), parts),
	}, nil
}

// VectorExpr .
type VectorExpr struct {
	eval   func(core.Env, ww.Any) (ww.Any, error)
	Vector core.Vector
}

// Eval returns a new vector whose contents are the evaluated values
// of the objects contained by the evaluated vector. Elements are evaluated left to right
func (vex VectorExpr) Eval(env core.Env) (score.Any, error) {
	cnt, err := vex.Vector.Count()
	if err != nil || cnt == 0 {
		return vex.Vector, err
	}

	// TODO(performace):  this is just begging for a transient.

	for i := 0; i < cnt; i++ {
		any, err := vex.Vector.EntryAt(i)
		if err != nil {
			return nil, err
		}

		other, err := vex.eval(env, any)
		if err != nil {
			return nil, err
		}

		// no need to canonicalize here.  If different, it's because the value changed.
		if bytes.Equal(memutil.Bytes(any.Value()), memutil.Bytes(other.Value())) {
			continue
		}

		if vex.Vector, err = vex.Vector.Assoc(i, other); err != nil {
			return nil, err
		}
	}

	return vex.Vector, nil
}

// // LocalGoExpr starts a local process.  Local processes cannot be addressed by remote
// // hosts.
// type LocalGoExpr struct {
// 	Args []ww.Any
// }

// // Eval resolves starts the process.
// func (lx LocalGoExpr) Eval(env core.Env) (score.Any, error) {
// 	return nil, errors.New("LocalGoExpr NOT IMPLEMENTED")
// }

// // RemoteGoExpr starts a global process.  Global processes may be bound to an Anchor,
// // rendering them addressable by remote hosts.
// type RemoteGoExpr struct {
// 	Root ww.Anchor
// 	Path core.Path
// 	Args []ww.Any
// }

// // Eval resolves the anchor and starts the process.
// func (rx RemoteGoExpr) Eval(core.Env) (score.Any, error) {
// 	path, err := rx.Path.Parts()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return rx.Root.Walk(context.Background(), path).
// 		Go(context.Background(), rx.Args...)
// }

// ImportExpr .
type ImportExpr struct {
	Analyzer core.Analyzer
	Paths    []string
}

// Eval loads the module files from the supplied paths
func (lex ImportExpr) Eval(env core.Env) (any score.Any, err error) {
	var dex DoExpr
	for _, path := range lex.Paths {
		if dex.Exprs, err = lex.loadOne(env, path); err != nil {
			break
		}

		if any, err = dex.Eval(env); err != nil {
			break
		}
	}

	return
}

func (lex ImportExpr) loadOne(env core.Env, path string) ([]core.Expr, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	forms, err := reader.New(f).All()
	if err != nil {
		return nil, err
	}

	exprs := make([]core.Expr, len(forms))
	for i, form := range forms {
		if exprs[i], err = lex.Analyzer.Analyze(env, form); err != nil {
			break
		}
	}

	return exprs, err
}
