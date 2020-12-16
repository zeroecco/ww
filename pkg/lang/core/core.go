package core

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"

	"github.com/spy16/slurp/core"
	"github.com/wetware/ww/internal/mem"
	ww "github.com/wetware/ww/pkg"
	capnp "zombiezen.com/go/capnproto2"
)

var (
	// ErrIncomparableTypes is returned if two types cannot be meaningfully
	// compared to each other.
	ErrIncomparableTypes = errors.New("incomparable types")

	// ErrIndexOutOfBounds is returned when a sequence's index is out of range.
	ErrIndexOutOfBounds = errors.New("index out of bounds")

	// ErrNotFound is returned by Env when a the corresponding entity for a name,
	// binding or module path is not found.
	ErrNotFound = core.ErrNotFound

	// ErrArity is returned when an Invokable is invoked with wrong number
	// of arguments.
	ErrArity = core.ErrArity

	// ErrNotInvokable is returned by InvokeExpr when the target is not invokable.
	ErrNotInvokable = errors.New("non-invokable type")

	// ErrIllegalState is returned when an operation is attempted against a datatype
	// that has the right type but an inappropriate value.
	ErrIllegalState = errors.New("illegal state")

	// ErrMemory is returned when an operation is attempted against an illegally
	// formatted datatype.
	ErrMemory = errors.New("memory error")

	errType = reflect.TypeOf((*error)(nil)).Elem()
	anyType = reflect.TypeOf((*ww.Any)(nil)).Elem()
)

type (
	// Env represents the environment in which forms are evaluated.
	Env = core.Env

	// Expr represents an expression that can be evaluated against an env.
	Expr = core.Expr

	// Error is returned by all slurp operations. Cause indicates the underlying
	// error type. Use errors.Is() with Cause to check for specific errors.
	Error = core.Error
)

// New returns a root Env that can be used to execute forms.
// It binds the prelude to the environment before returning.
func New() Env { return core.New(nil) }

// Analyzer implementation is responsible for performing syntax analysis
// on given form.
type Analyzer interface {
	core.Analyzer
}

// Eval a form.
func Eval(env Env, a Analyzer, form ww.Any) (ww.Any, error) {
	expr, err := a.Analyze(env, form)
	if err != nil || expr == nil {
		return nil, err
	}

	res, err := expr.Eval(env)
	if err != nil || res == nil {
		return Nil{}, err
	}

	return res.(ww.Any), nil
}

// Invokable represents a value that can be invoked as a function.
type Invokable interface {
	// Invoke is called if this value appears as the first argument of
	// invocation form (i.e., list).
	Invoke(args []ww.Any) (ww.Any, error)
}

// Countable types can report the number of elements they contain.
type Countable interface {
	// Count provides the number of elements contained.
	Count() (int, error)
}

// Container is an aggregate of values.
type Container interface {
	ww.Any
	Countable
	Conj(...ww.Any) (Container, error)
}

// Comparable type.
type Comparable interface {
	// Comp compares the magnitude of the comparable c with that of other.
	// It returns 0 if the magnitudes are equal, -1 if c < other, and 1 if c > other.
	Comp(other ww.Any) (int, error)
}

// EqualityProvider can test for equality.
type EqualityProvider interface {
	Eq(ww.Any) (bool, error)
}

// ErrStringer is equivalent to fmt.Stringer, except that it may return a
// non-nil error.
type ErrStringer interface {
	String() (string, error)
}

// Render a value into a human-readable representation suitable for printing.
// Ouptut from Render IS NOT guaranteed to be parseable by reader.Reader.
func Render(item ww.Any) (string, error) {
	switch any := item.Value(); any.Which() {
	case mem.Any_Which_nil:
		return "nil", nil

	case mem.Any_Which_bool:
		return Bool{any}.String(), nil

	case mem.Any_Which_str:
		return any.Str()

	case mem.Any_Which_symbol:
		return any.Symbol()

	case mem.Any_Which_keyword:
		return Keyword{any}.String()

	case mem.Any_Which_path:
		return UnboundPath{any}.String()

	case mem.Any_Which_char:
		return Char{any}.String(), nil

	case mem.Any_Which_i64:
		return Int64{any}.String(), nil

	case mem.Any_Which_f64:
		return Float64{any}.String(), nil

	}

	switch v := item.(type) {
	case ErrStringer:
		return v.String()

	case Seq:
		return seqToString(v)

	case fmt.Stringer:
		return v.String(), nil
	}

	return fmt.Sprintf("%#v", item), nil
}

// IsNil returns true if value is native go `nil` or `Nil{}`.
func IsNil(v ww.Any) bool {
	if v == nil {
		return true
	}

	return v.Value().Which() == mem.Any_Which_nil
}

// IsTruthy returns true if the value has a logical vale of `true`.
func IsTruthy(v ww.Any) (bool, error) {
	if IsNil(v) {
		return false, nil
	}

	switch val := v.(type) {
	case Bool:
		return val.Bool(), nil

	case Numerical:
		return !val.Zero(), nil

	case Countable:
		i, err := val.Count()
		return i == 0, err

	default:
		return true, nil

	}
}

// Eq returns true is the two values are equal
func Eq(a, b ww.Any) (bool, error) {
	// Nil is only equal to itself
	if IsNil(a) && IsNil(b) {
		return true, nil
	}

	// Check for usable interfaces on object A
	switch val := a.(type) {
	case Comparable:
		i, err := val.Comp(b)
		return i == 0, err

	case EqualityProvider:
		return val.Eq(b)

	}

	// Check for usable interfaces on object B
	switch val := b.(type) {
	case Comparable:
		i, err := val.Comp(b)
		return i == 0, err

	case EqualityProvider:
		return val.Eq(b)

	}

	// Identical types with the same canonical representation are equal.
	if a.Value().Which() == b.Value().Which() {
		ca, err := Canonical(a)
		if err != nil {
			return false, err
		}

		cb, err := Canonical(b)
		if err != nil {
			return false, err
		}

		return bytes.Equal(ca, cb), nil
	}

	// Disparate types are unequal by default.
	return false, nil
}

// Pop an item from an ordered collection.
// For a list, returns a new list without the first item.
// For a vector, returns a new vector without the last item.
// If the collection is empty, returns a wrapped ErrIllegalState.
func Pop(cont Container) (ww.Any, error) {
	switch v := cont.(type) {
	case Vector:
		return v.Pop()

	case Seq:
		cnt, err := v.Count()
		if err != nil {
			return nil, err
		}

		if cnt == 0 {
			return nil, fmt.Errorf("%w: cannot pop from empty seq", ErrIllegalState)
		}

		return v.Next()

	}

	return nil, fmt.Errorf("cannot pop from %s", cont.Value().Which())
}

// Conj returns a new collection with the supplied
// values "conjoined".
//
// For lists, the value is added at the head.
// For vectors, the value is added at the tail.
// `(conj nil item)` returns `(item)``.
func Conj(any ww.Any, xs ...ww.Any) (Container, error) {
	if IsNil(any) {
		return NewList(capnp.SingleSegment(nil), xs...)
	}

	if c, ok := any.(Container); ok {
		return c.Conj(xs...)
	}

	return nil, fmt.Errorf("cannot conj with %T", any)
}

// Canonical representation of an arbitrary value.
func Canonical(any ww.Any) ([]byte, error) {
	return capnp.Canonicalize(any.Value().Struct)
}

// AsAny lifts a mem.Any to a ww.Any.
func AsAny(any mem.Any) (item ww.Any, err error) {
	switch any.Which() {
	case mem.Any_Which_nil:
		item = Nil{}
	case mem.Any_Which_bool:
		item = Bool{any}
	case mem.Any_Which_i64:
		item = Int64{any}
	case mem.Any_Which_f64:
		item = Float64{any}
	case mem.Any_Which_bigInt:
		item, err = asBigInt(any)
	case mem.Any_Which_bigFloat:
		item, err = asBigFloat(any)
	case mem.Any_Which_frac:
		item, err = asFrac(any)
	case mem.Any_Which_char:
		item = Char{any}
	case mem.Any_Which_str:
		item = String{any}
	case mem.Any_Which_keyword:
		item = Keyword{any}
	case mem.Any_Which_symbol:
		item = Symbol{any}
	case mem.Any_Which_path:
		item = UnboundPath{any}
	case mem.Any_Which_list:
		item, err = asList(any)
	case mem.Any_Which_vector:
		item, err = asVector(any)

	// case mem.Any_Which_proc:
	// 	item = RemoteProcess{v}
	default:
		err = fmt.Errorf("unknown value type '%s'", any.Which())
	}

	return
}
