package core

import (
	"errors"
	"fmt"
	"strings"

	"github.com/wetware/ww/internal/mem"
	ww "github.com/wetware/ww/pkg"
	memutil "github.com/wetware/ww/pkg/util/mem"
	capnp "zombiezen.com/go/capnproto2"
)

var (
	errNoMatch = errors.New("does not match")

	_ Invokable = (*BoundFn)(nil)
)

// A CallTarget is an implementation of a specific call signature for Fn.
type CallTarget struct {
	Param    capnp.TextList
	Body     mem.Any_List
	Variadic bool
}

// Call evaluates the body.
func (t CallTarget) Call(a Analyzer, env Env) (any ww.Any, err error) {
	for i := 0; i < t.Body.Len(); i++ {
		if any, err = AsAny(t.Body.At(i)); err != nil {
			return
		}

		if any, err = Eval(env, a, any); err != nil {
			return
		}
	}

	if any == nil {
		return Nil{}, nil
	}

	return any, nil
}

// BoundFn is a function bound to a local env.
type BoundFn struct {
	Fn
	Env      Env
	Analyzer Analyzer
}

// Invoke the underlying function.
func (f BoundFn) Invoke(args []ww.Any) (ww.Any, error) {
	target, err := f.Fn.Match(len(args))
	if err != nil {
		return nil, err
	}

	var param string
	for i := 0; i < target.Param.Len(); i++ {
		if param, err = target.Param.At(i); err != nil {
			return nil, err
		}

		if err = f.Env.Bind(param, args[i]); err != nil {
			return nil, err
		}
	}

	return target.Call(f.Analyzer, f.Env)
}

// Fn is a multi-arity function or macro.
type Fn struct{ mem.Any }

// Value returns the memory value
func (fn Fn) Value() mem.Any { return fn.Any }

// String returns a human-readable representation of the function that
// is suitable for printing.
func (fn Fn) String() (string, error) {
	name, err := fn.Name()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Fn{%s}", name), nil
}

// Macro returns true if the function is a macro.
func (fn Fn) Macro() bool {
	raw, err := fn.Fn()
	if err != nil {
		panic(err)
	}

	return raw.Macro()
}

// Name of the function.
func (fn Fn) Name() (string, error) {
	raw, err := fn.Fn()
	if err != nil {
		return "", err
	}

	if raw.Which() == mem.Fn_Which_lambda {
		return "λ", nil
	}

	return raw.Name()
}

// Match the arguments to a call target.
func (fn Fn) Match(nargs int) (CallTarget, error) {
	raw, err := fn.Fn()
	if err != nil {
		return CallTarget{}, err
	}

	fs, err := raw.Funcs()
	if err != nil {
		return CallTarget{}, err
	}

	var f mem.Fn_Func
	for i := 0; i < fs.Len(); i++ {
		switch f = fs.At(i); f.Which() {
		case mem.Fn_Func_Which_nilary:
			if nargs != 0 {
				continue
			}

			body, err := f.Body()
			return CallTarget{Body: body}, err

		case mem.Fn_Func_Which_params:
			ps, err := f.Params()
			if err != nil {
				return CallTarget{}, err
			}

			nparam := ps.Len()
			body, err := f.Body()

			if f.Variadic() && nargs >= nparam-1 {
				return CallTarget{Body: body, Param: ps, Variadic: true}, err
			}

			if nargs == nparam {
				body, err := f.Body()
				return CallTarget{Body: body, Param: ps}, err
			}
		}
	}

	// Did not find suitable call target.  Return informative error message.
	name, err := fn.Name()
	if err != nil {
		return CallTarget{}, err
	}

	return CallTarget{}, fmt.Errorf("%w (%d) to '%s'", ErrArity, nargs, name)
}

// FuncBuilder is a factory type for Fn.
type FuncBuilder struct {
	any    mem.Any
	fn     mem.Fn
	sigs   []callSignature
	stages []func() error
}

// Start building a function.
func (b *FuncBuilder) Start(a capnp.Arena) {
	b.stages = make([]func() error, 0, 8)
	b.sigs = b.sigs[:0]

	b.addStage(func() (err error) {
		if b.any, err = memutil.Alloc(a); err != nil {
			return fmt.Errorf("alloc value: %w", err)
		}

		if b.fn, err = b.any.NewFn(); err != nil {
			return fmt.Errorf("alloc fn: %w", err)
		}

		return nil
	})
}

// SetMacro sets the macro flag.
func (b *FuncBuilder) SetMacro(macro bool) {
	b.addStage(func() error {
		b.fn.SetMacro(macro)
		return nil
	})
}

// SetName assigns a name to the function.
func (b *FuncBuilder) SetName(name string) {
	b.addStage(func() error {
		if name == "" {
			b.fn.SetLambda()
			return nil
		}

		if err := b.fn.SetName(name); err != nil {
			return fmt.Errorf("set name: %w", err)
		}

		return nil
	})
}

// Commit flushes any buffers and returns the constructed function.
// After a call to Commit(), users must call Start() before reusing b.
func (b *FuncBuilder) Commit() (Fn, error) {
	for _, stage := range append(b.stages, b.setFuncs) {
		if err := stage(); err != nil {
			return Fn{}, err
		}
	}

	return Fn{Any: b.any}, nil
}

// AddSeq parses the sequence `([<params>*] <body>*)` into a call target.
func (b *FuncBuilder) AddSeq(seq Seq) {
	sig, err := ToSlice(seq)
	if err != nil {
		b.addStage(func() error { return err })
	}

	b.AddTarget(sig[0], sig[1:])
}

// AddTarget parses the call signature `[<params>*] <body>*` into a call target.
func (b *FuncBuilder) AddTarget(args ww.Any, body []ww.Any) {
	b.addStage(func() error {
		if any := args.Value(); args.Value().Which() != mem.Any_Which_vector {
			return Error{
				Cause:   errors.New("invalid call signature"),
				Message: fmt.Sprintf("args must be Vector, not '%s'", any.Which()),
			}
		}

		if body == nil {
			return Error{
				Cause:   errors.New("invalid call signature"),
				Message: "empty body",
			}
		}

		ps, variadic, err := b.readParams(args.(Vector))
		if err != nil {
			return err
		}

		b.sigs = append(b.sigs, callSignature{
			Params:   ps,
			Variadic: variadic,
			Body:     body,
		})
		return nil
	})
}

func (b *FuncBuilder) addStage(fn func() error) { b.stages = append(b.stages, fn) }

func (b *FuncBuilder) setFuncs() error {
	if len(b.sigs) == 0 {
		return errors.New("no call signatures")
	}

	fs, err := b.fn.NewFuncs(int32(len(b.sigs)))
	if err != nil {
		return err
	}

	for i, sig := range b.sigs {
		f := fs.At(i)
		if err = sig.Populate(f); err != nil {
			break
		}
	}

	return err
}

func (b *FuncBuilder) readParams(v Vector) ([]string, bool, error) {
	cnt, err := v.Count()
	if err != nil || cnt == 0 {
		return nil, false, err
	}

	ps := make([]string, cnt)

	for i := range ps {
		entry, err := v.EntryAt(i)
		if err != nil {
			return nil, false, err
		}

		if entry.Value().Which() != mem.Any_Which_symbol {
			return nil, false, fmt.Errorf("expected symbol, got %s", entry.Value().Which())
		}

		if ps[i], err = entry.Value().Symbol(); err != nil {
			return nil, false, err
		}
	}

	var variadic bool
	if last := ps[cnt-1]; strings.HasSuffix(last, "...") {
		ps[cnt-1] = last[:len(last)-3]
		variadic = true
	}

	return ps, variadic, nil
}

type callSignature struct {
	Params   []string
	Variadic bool
	Body     []ww.Any
}

func (sig callSignature) Populate(f mem.Fn_Func) (err error) {
	if err = sig.populateBody(f); err == nil {
		err = sig.populateParams(f)
	}

	return
}

func (sig callSignature) populateParams(f mem.Fn_Func) error {
	if sig.Params == nil {
		f.SetNilary()
		return nil
	}

	as, err := f.NewParams(int32(len(sig.Params)))
	if err != nil {
		return err
	}

	for i, s := range sig.Params {
		if err = as.Set(i, s); err != nil {
			break
		}
	}

	if sig.Variadic {
		f.SetVariadic(true)
	}

	return err
}

func (sig callSignature) populateBody(f mem.Fn_Func) error {
	bs, err := f.NewBody(int32(len(sig.Body)))
	if err != nil {
		return err
	}

	for i, any := range sig.Body {
		if err = bs.Set(i, any.Value()); err != nil {
			break
		}
	}

	return err
}
