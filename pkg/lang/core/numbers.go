package core

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/wetware/ww/internal/mem"
	ww "github.com/wetware/ww/pkg"
	memutil "github.com/wetware/ww/pkg/util/mem"
	capnp "zombiezen.com/go/capnproto2"
)

var unit big.Int

func init() { unit.SetInt64(1) }

// Numerical values are comparable.
type Numerical interface {
	ww.Any
	Zero() bool
	Comp(ww.Any) (int, error)
}

// Int64 represents a 64-bit signed integer.
type Int64 struct{ mem.Any }

// NewInt64 .
func NewInt64(a capnp.Arena, i int64) (Int64, error) {
	mv, err := memutil.Alloc(a)
	if err == nil {
		mv.SetI64(i)
	}

	return Int64{mv}, err
}

// Value returns the memory value
func (i64 Int64) Value() mem.Any { return i64.Any }

// Int64 satsifies Int64
func (i64 Int64) Int64() int64 { return i64.I64() }

func (i64 Int64) String() string { return fmt.Sprintf("%d", i64.Int64()) }

// Comp returns 0 if the v == other, -1 if v < other, and 1 if v > other.
func (i64 Int64) Comp(other ww.Any) (int, error) {
	switch val := other.Value(); val.Which() {
	case mem.Any_Which_i64:
		return compI64(i64.Int64(), val.I64()), nil

	case mem.Any_Which_f64:
		var f big.Float
		return f.SetInt64(i64.Int64()).Cmp(big.NewFloat(val.F64())), nil

	case mem.Any_Which_bigInt:
		return big.NewInt(i64.Int64()).Cmp(other.(BigInt).BigInt()), nil

	case mem.Any_Which_bigFloat:
		var f big.Float
		return f.SetInt64(i64.Int64()).Cmp(other.(BigFloat).f), nil

	case mem.Any_Which_frac:
		var r big.Rat
		return r.SetInt64(i64.Int64()).Cmp(other.(Fraction).Rat()), nil

	default:
		return 0, ErrIncomparableTypes

	}
}

// Zero returns true if the value is 0.
func (i64 Int64) Zero() bool { return i64.Int64() == 0 }

// BigInt represents an arbitrary-length signed integer
type BigInt struct {
	i *big.Int
	mem.Any
}

// NewBigInt .
func NewBigInt(a capnp.Arena, i *big.Int) (BigInt, error) {
	mv, err := memutil.Alloc(a)
	if err == nil {
		err = mv.SetBigInt(i.Bytes())
	}

	return BigInt{i: i, Any: mv}, err
}

func asBigInt(any mem.Any) (BigInt, error) {
	var i big.Int
	if buf, err := any.BigInt(); err == nil {
		i.SetBytes(buf)
	}

	return BigInt{i: &i, Any: any}, nil
}

// Value returns the memory value
func (bi BigInt) Value() mem.Any { return bi.Any }

// BigInt satisfies BigInt
func (bi BigInt) BigInt() *big.Int { return bi.i }

func (bi BigInt) String() string { return bi.i.String() }

// Comp returns 0 if the v == other, -1 if v < other, and 1 if v > other.
func (bi BigInt) Comp(other ww.Any) (int, error) {
	switch val := other.Value(); val.Which() {
	case mem.Any_Which_i64:
		return bi.i.Cmp(big.NewInt(val.I64())), nil

	case mem.Any_Which_f64:
		var f big.Float
		return f.SetInt(bi.i).Cmp(big.NewFloat(val.F64())), nil

	case mem.Any_Which_bigInt:
		return bi.i.Cmp(other.(BigInt).BigInt()), nil

	case mem.Any_Which_bigFloat:
		var f big.Float
		return f.SetInt(bi.i).Cmp(other.(BigFloat).f), nil

	case mem.Any_Which_frac:
		var r big.Rat
		return r.SetFrac(bi.i, &unit).Cmp(other.(Fraction).Rat()), nil

	default:
		return 0, ErrIncomparableTypes
	}
}

// Zero returns true if the value is 0.
func (bi BigInt) Zero() bool { return bi.i.Sign() == 0 }

// Float64 represents a 64-bit floating-point number
type Float64 struct{ mem.Any }

// NewFloat64 .
func NewFloat64(a capnp.Arena, f float64) (Float64, error) {
	mv, err := memutil.Alloc(a)
	if err == nil {
		mv.SetF64(f)
	}

	return Float64{mv}, err
}

// Value returns the memory value
func (f64 Float64) Value() mem.Any { return f64.Any }

// Float64 satisfies Float64
func (f64 Float64) Float64() float64 { return f64.F64() }

func (f64 Float64) String() string { return strconv.FormatFloat(f64.Float64(), 'g', -1, 64) }

// Comp returns 0 if the v == other, -1 if v < other, and 1 if v > other.
func (f64 Float64) Comp(other ww.Any) (int, error) {
	switch val := other.Value(); val.Which() {
	case mem.Any_Which_i64:
		var bf big.Float
		return big.NewFloat(f64.Float64()).Cmp(bf.SetInt64(val.I64())), nil

	case mem.Any_Which_f64:
		return compF64(f64.Float64(), val.F64()), nil

	case mem.Any_Which_bigInt:
		var bf big.Float
		return big.NewFloat(f64.Float64()).Cmp(bf.SetInt(other.(BigInt).BigInt())), nil

	case mem.Any_Which_bigFloat:
		var bi big.Float
		bi.SetFloat64(f64.Float64())
		return bi.Cmp(other.(BigFloat).f), nil

	case mem.Any_Which_frac:
		var r big.Rat
		r.SetFloat64(f64.Float64())
		return r.Cmp(other.(Fraction).Rat()), nil

	default:
		return 0, ErrIncomparableTypes

	}
}

// Zero returns true if the value is 0.
func (f64 Float64) Zero() bool { return f64.Float64() == 0 }

// BigFloat represents an arbitrary-precision floating-point number.
type BigFloat struct {
	mem.Any
	f *big.Float
}

// NewBigFloat .
func NewBigFloat(a capnp.Arena, f *big.Float) (BigFloat, error) {
	mv, err := memutil.Alloc(a)
	if err == nil {
		err = mv.SetBigFloat(f.Text('g', -1))
	}

	return BigFloat{f: f, Any: mv}, err
}

func asBigFloat(v mem.Any) (bf BigFloat, err error) {
	bf.f = &big.Float{}
	bf.Any = v

	var s string
	if s, err = v.BigFloat(); err == nil {
		if _, ok := bf.f.SetString(s); !ok {
			err = fmt.Errorf("invalid bigfloat format '%s'", s)
		}
	}

	return
}

// Value returns the memory value
func (bf BigFloat) Value() mem.Any { return bf.Any }

// BigFloat satisfies BigFloat
func (bf BigFloat) BigFloat() *big.Float { return bf.f }

func (bf BigFloat) String() string { return bf.f.Text('g', -1) }

// Comp returns 0 if the v == other, -1 if v < other, and 1 if v > other.
func (bf BigFloat) Comp(other ww.Any) (int, error) {
	switch val := other.Value(); val.Which() {
	case mem.Any_Which_i64:
		var f big.Float
		return bf.f.Cmp(f.SetInt64(val.I64())), nil

	case mem.Any_Which_f64:
		return bf.f.Cmp(big.NewFloat(val.F64())), nil

	case mem.Any_Which_bigInt:
		var f big.Float
		return bf.f.Cmp(f.SetInt(other.(BigInt).BigInt())), nil

	case mem.Any_Which_bigFloat:
		return bf.f.Cmp(other.(BigFloat).f), nil

	case mem.Any_Which_frac:
		var r big.Rat
		bf.f.Rat(&r)
		return r.Cmp(other.(Fraction).Rat()), nil

	default:
		return 0, ErrIncomparableTypes

	}
}

// Zero returns true if the value is 0.
func (bf BigFloat) Zero() bool { return bf.f.Sign() == 0 }

// Fraction represents a rational number a/b of arbitrary precision.
type Fraction struct {
	r *big.Rat
	mem.Any
}

// NewFraction with built-in implementation.
func NewFraction(a capnp.Arena, r *big.Rat) (Fraction, error) {
	any, err := memutil.Alloc(a)
	if err != nil {
		return Fraction{}, err
	}

	f, err := any.NewFrac()
	if err != nil {
		return Fraction{}, err
	}

	if err = f.SetNumer(r.Num().Bytes()); err != nil {
		return Fraction{}, err
	}

	if err = f.SetDenom(r.Denom().Bytes()); err != nil {
		return Fraction{}, err
	}

	return Fraction{r: r, Any: any}, nil
}

func asFrac(any mem.Any) (Fraction, error) {
	fv, err := any.Frac()
	if err != nil {
		return Fraction{}, err
	}

	var nbuf, dbuf []byte
	if nbuf, err = fv.Numer(); err != nil {
		return Fraction{}, err
	}

	if dbuf, err = fv.Denom(); err != nil {
		return Fraction{}, err
	}

	var numer, denom big.Int
	numer.SetBytes(nbuf)
	denom.SetBytes(dbuf)

	var r big.Rat
	return Fraction{
		r:   r.SetFrac(&numer, &denom),
		Any: any,
	}, nil
}

// Value returns the memory value
func (f Fraction) Value() mem.Any { return f.Any }

// Rat satisfies Fraction
func (f Fraction) Rat() *big.Rat { return f.r }

func (f Fraction) String() string { return f.r.String() }

// Comp returns true if the other value is numerical and has the same value.
func (f Fraction) Comp(other ww.Any) (int, error) {
	switch val := other.Value(); val.Which() {
	case mem.Any_Which_i64:
		var r big.Rat
		return f.r.Cmp(r.SetFrac(big.NewInt(val.I64()), &unit)), nil

	case mem.Any_Which_f64:
		var r big.Rat
		return f.r.Cmp(r.SetFloat64(val.F64())), nil

	case mem.Any_Which_bigInt:
		var r big.Rat
		return f.r.Cmp(r.SetFrac(other.(BigInt).BigInt(), &unit)), nil

	case mem.Any_Which_bigFloat:
		var r big.Rat
		other.(BigFloat).f.Rat(&r)
		return f.r.Cmp(&r), nil

	case mem.Any_Which_frac:
		return f.r.Cmp(other.(Fraction).Rat()), nil

	default:
		return 0, ErrIncomparableTypes

	}
}

// Zero returns true if the value is 0.
func (f Fraction) Zero() bool { return f.r.Sign() == 0 }

func compI64(a, b int64) int {
	switch {
	case a == b:
		return 0
	case a > b:
		return 1
	default:
		return -1
	}
}

func compF64(a, b float64) int {
	switch {
	case a == b:
		return 0
	case a > b:
		return 1
	default:
		return -1
	}
}
