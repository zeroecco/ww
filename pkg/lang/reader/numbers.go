package reader

import (
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"strconv"
	"strings"

	capnp "zombiezen.com/go/capnproto2"

	ww "github.com/wetware/ww/pkg"
	"github.com/wetware/ww/pkg/lang/core"
)

func readNumber(rd *Reader, init rune) (v ww.Any, err error) {
	beginPos := rd.Position()

	numStr, err := readNumToken(rd, init)
	if err != nil {
		return nil, err
	}

	decimalPoint := strings.ContainsRune(numStr, '.')
	isRadix := strings.ContainsRune(numStr, 'r')
	isScientific := strings.ContainsRune(numStr, 'e')
	isFrac := strings.ContainsRune(numStr, '/')

	switch {
	case isRadix && (decimalPoint || isScientific || isFrac):
		err = ErrNumberFormat

	case isScientific:
		v, err = parseScientific(numStr)

	case decimalPoint:
		v, err = parseFloat(numStr)

	case isRadix:
		v, err = parseRadix(numStr)

	case isFrac:
		v, err = parseFrac(numStr)

	default:
		v, err = parseInt(numStr)

	}

	if err != nil {
		err = rd.annotateErr(err, beginPos, numStr)
	}

	return
}

func parseInt(numStr string) (core.Numerical, error) {
	v, err := strconv.ParseInt(numStr, 0, 64)
	switch {
	case err == nil:
		return core.NewInt64(capnp.SingleSegment(nil), v)

	case errors.Is(err, strconv.ErrRange):
		if b, ok := (&big.Int{}).SetString(numStr, 0); ok {
			return core.NewBigInt(capnp.SingleSegment(nil), b)
		}

	}
	return nil, fmt.Errorf("%w (int64): '%s'", ErrNumberFormat, numStr)
}

func parseFloat(numStr string) (core.Numerical, error) {
	v, err := strconv.ParseFloat(numStr, 64)
	switch {
	case err == nil:
		// TODO(performance):  pre-allocate arena
		return core.NewFloat64(capnp.SingleSegment(nil), v)

	case errors.Is(err, strconv.ErrRange):
		var f big.Float
		if _, ok := f.SetString(numStr); !ok {
			return nil, fmt.Errorf("%w (bigfloat): '%s'", ErrNumberFormat, numStr)
		}

		// TODO(performance):  pre-allocate arena
		return core.NewBigFloat(capnp.SingleSegment(nil), &f)

	default:
		return nil, ErrNumberFormat

	}
}

func parseRadix(numStr string) (core.Numerical, error) {
	parts := strings.Split(numStr, "r")
	if len(parts) != 2 {
		return nil, fmt.Errorf("%w (radix notation): '%s'", ErrNumberFormat, numStr)
	}

	base, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w (radix notation): '%s'", ErrNumberFormat, numStr)
	}

	repr := parts[1]
	if base < 0 {
		base = -1 * base
		repr = "-" + repr
	}

	v, err := strconv.ParseInt(repr, int(base), 64)
	if errors.Is(err, strconv.ErrRange) {
		var bi big.Int
		if _, ok := bi.SetString(repr, int(base)); !ok {
			return nil, fmt.Errorf("%w (radix notation): '%s'", ErrNumberFormat, numStr)
		}

		return core.NewBigInt(capnp.SingleSegment(nil), &bi)
	}
	if err != nil {
		return nil, fmt.Errorf("%w (radix notation): '%s'", ErrNumberFormat, numStr)
	}

	return core.NewInt64(capnp.SingleSegment(nil), v)
}

func parseScientific(numStr string) (_ core.Numerical, err error) {
	parts := strings.Split(numStr, "e")
	if len(parts) != 2 {
		return nil, fmt.Errorf("%w (scientific notation): '%s'", ErrNumberFormat, numStr)
	}

	base, pow, err := parseScientificFloat64(parts)
	switch {
	case err == nil:
		if num := base * math.Pow(10, pow); !math.IsInf(num, 0) {
			return core.NewFloat64(capnp.SingleSegment(nil), num)
		}
		fallthrough

	case errors.Is(err, strconv.ErrRange):
		if f, ok := (&big.Float{}).SetString(numStr); ok {
			return core.NewBigFloat(capnp.SingleSegment(nil), f)
		}

	}

	return nil, fmt.Errorf("%w (scientific): '%s'", ErrNumberFormat, numStr)
}

func parseScientificFloat64(parts []string) (base float64, pow float64, err error) {
	mant, exp := parts[0], parts[1]
	if base, err = strconv.ParseFloat(mant, 64); err != nil {
		return
	}

	if pow, err = strconv.ParseFloat(exp, 64); err != nil {
		return
	}

	if pow != math.Trunc(pow) {
		return 0, 0, fmt.Errorf("invalid exponent: '%s'", exp)
	}

	// TODO:  remove this check once we're satisfied that fuzzing doesn't
	// trigger this condition.
	if math.IsInf(pow, 0) || math.IsNaN(pow) {
		panic("unreachable")
	}

	return
}

func parseFrac(numStr string) (core.Numerical, error) {
	parts := strings.Split(numStr, "/")
	if len(parts) != 2 || parts[1] == "" {
		return nil, fmt.Errorf("%w (fractional notation): '%s'", ErrNumberFormat, numStr)
	}

	rat, err := parseRatInt64(parts)
	switch {
	case err == nil:
		break

	case errors.Is(err, strconv.ErrRange):
		if rat, err = parseRatBigInt(parts); err == nil {
			break
		}
		fallthrough

	default:
		return nil, err
	}

	return core.NewFraction(capnp.SingleSegment(nil), rat)
}

func parseRatInt64(parts []string) (*big.Rat, error) {
	numer, err := strconv.ParseInt(parts[0], 0, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid numerator '%s'", ErrNumberFormat, parts[0])
	}

	denom, err := strconv.ParseInt(parts[1], 0, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid denominator '%s'", ErrNumberFormat, parts[1])
	}

	if denom == 0 {
		return nil, core.ErrDivideByZero
	}

	return big.NewRat(numer, denom), nil
}

func parseRatBigInt(parts []string) (*big.Rat, error) {
	var ok bool
	var numer, denom *big.Int
	if numer, ok = numer.SetString(parts[0], 0); !ok {
		return nil, fmt.Errorf("%w: invalid numerator '%s'", ErrNumberFormat, parts[0])
	}
	if denom, ok = denom.SetString(parts[1], 0); !ok {
		return nil, fmt.Errorf("%w: invalid denominator '%s'", ErrNumberFormat, parts[1])
	}

	if denom.Sign() == 0 {
		return nil, core.ErrDivideByZero
	}

	var r big.Rat
	return r.SetFrac(numer, denom), nil
}

// Token reads one token from the reader and returns. If init is not -1, it is included
// as first character in the token.
func readNumToken(rd *Reader, init rune) (string, error) {
	var b strings.Builder
	if init != -1 {
		b.WriteRune(init)
	}

	for {
		r, err := rd.NextRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return b.String(), err
		}

		if r != '/' && rd.IsTerminal(r) {
			rd.Unread(r)
			break
		}

		b.WriteRune(r)
	}

	return b.String(), nil
}
