package reader

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"strings"
	"unicode"

	"github.com/libp2p/go-libp2p-core/network"
	ww "github.com/wetware/ww/pkg"
	"github.com/wetware/ww/pkg/lang/core"
	capnp "zombiezen.com/go/capnproto2"
)

const dispatchTrigger = '#'

var (
	symTable = map[string]ww.Any{
		"nil":   core.Nil{},
		"false": core.False,
		"true":  core.True,
	}

	escapeMap = map[rune]rune{
		'"':  '"',
		'n':  '\n',
		'\\': '\\',
		't':  '\t',
		'a':  '\a',
		'f':  '\a',
		'r':  '\r',
		'b':  '\b',
		'v':  '\v',
	}

	charLiterals = make(map[string]core.Char, 6)
)

func init() {
	for k, v := range map[string]rune{
		"tab":       '\t',
		"space":     ' ',
		"newline":   '\n',
		"return":    '\r',
		"backspace": '\b',
		"formfeed":  '\f',
	} {
		c, err := core.NewChar(capnp.SingleSegment(nil), v)
		if err != nil {
			panic(err)
		}

		charLiterals[k] = c
	}
}

// Reader consumes characters from a stream and parses them into symbolic expressions
// or forms. Reader is customizable through Macro implementations which can be set as
// handler for specific trigger runes.
type Reader struct {
	File string

	rs          io.RuneReader
	buf         []rune
	line, col   int
	lastCol     int
	dispatching bool
	dispatch    map[rune]Macro
	macros      map[rune]Macro
}

// New returns a lisp reader instance which can read forms from r. Returned instance
// supports only primitive data  types from value package by default. Support for
// custom forms can be added using SetMacro(). File name is inferred from the value &
// type information of 'r' OR can be set manually on the Reader instance returned.
func New(r io.Reader) *Reader {
	return &Reader{
		File: inferFileName(r),
		rs:   bufio.NewReader(r),
		macros: map[rune]Macro{
			'"':  readString,
			';':  readComment,
			':':  readKeyword,
			'\\': readCharacter,
			'(':  readList,
			')':  UnmatchedDelimiter(),
			'[':  readVector,
			']':  UnmatchedDelimiter(),
			'\'': quoteFormReader("quote"),
			'~':  quoteFormReader("unquote"),
			'`':  quoteFormReader("syntax-quote"),
			'/':  readPath,
		},
		dispatch: map[rune]Macro{},
	}
}

// Reset clears all state from the reader and assigns the specified byte stream
// as input.
func (rd *Reader) Reset(r io.Reader) {
	rd.dispatching = false
	rd.buf = rd.buf[:0]
	rd.rs = bufio.NewReader(r)
	rd.File = inferFileName(r)
}

// All consumes characters from stream until EOF and returns a list of all the forms
// parsed. Any no-op forms (e.g., comment) will not be included in the result.
func (rd *Reader) All() ([]ww.Any, error) {
	forms := make([]ww.Any, 0, 8)

	for {
		form, err := rd.One()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		forms = append(forms, form)
	}

	return forms, nil
}

// One consumes characters from underlying stream until a complete form is parsed and
// returns the form while ignoring the no-op forms like comments. Except EOF, all other
// errors will be wrapped with reader Error type along with the positional information
// obtained using Position().
func (rd *Reader) One() (ww.Any, error) {
	for {
		form, err := rd.readOne()
		if err != nil {
			if errors.Is(err, ErrSkip) {
				continue
			}
			return nil, err
		}
		return form, nil
	}
}

// IsTerminal returns true if the rune should terminate a form. Macro trigger runes
// defined in the read table and all whitespace characters are considered terminal.
// "," is also considered a whitespace character and hence a terminal.
func (rd *Reader) IsTerminal(r rune) bool {
	if isSpace(r) {
		return true
	}

	if rd.dispatching {
		_, found := rd.dispatch[r]
		if found {
			return true
		}
	}

	_, found := rd.macros[r]
	return found
}

// SetMacro sets the given reader macro as the handler for init rune in the read table.
// Overwrites if a macro is already present. If the macro value given is nil, entry for
// the init rune will be removed from the read table. isDispatch decides if the macro is
// a dispatch macro and takes effect only after a '#' sign.
func (rd *Reader) SetMacro(init rune, isDispatch bool, macro Macro) {
	if isDispatch {
		if macro == nil {
			delete(rd.dispatch, init)
			return
		}
		rd.dispatch[init] = macro
	} else {
		if macro == nil {
			delete(rd.macros, init)
			return
		}
		rd.macros[init] = macro
	}
}

// NextRune returns next rune from the stream and advances the stream.
func (rd *Reader) NextRune() (rune, error) {
	var r rune
	if len(rd.buf) > 0 {
		r = rd.buf[0]
		rd.buf = rd.buf[1:]
	} else {
		temp, _, err := rd.rs.ReadRune()
		if err != nil {
			return -1, err
		}
		r = temp
	}

	if r == '\n' {
		rd.line++
		rd.lastCol = rd.col
		rd.col = 0
	} else {
		rd.col++
	}
	return r, nil
}

// Unread returns runes consumed from the stream back to the stream. Un-reading more
// runes than read is guaranteed to work but will cause inconsistency in  positional
// information of the Reader.
func (rd *Reader) Unread(runes ...rune) {
	newLine := false
	for _, r := range runes {
		if r == '\n' {
			newLine = true
			break
		}
	}

	if newLine {
		rd.line--
		rd.col = rd.lastCol
	} else {
		rd.col--
	}

	rd.buf = append(runes, rd.buf...)
}

// Position returns information about the stream including file name and the position
// of the reader.
func (rd Reader) Position() Position {
	file := strings.TrimSpace(rd.File)
	return Position{
		File: file,
		Ln:   rd.line + 1,
		Col:  rd.col,
	}
}

// SkipSpaces consumes and discards runes from stream repeatedly until a character that
// is not a whitespace is identified. Along with standard unicode whitespace characters,
// "," is also considered a whitespace and discarded.
func (rd *Reader) SkipSpaces() error {
	for {
		r, err := rd.NextRune()
		if err != nil {
			return err
		}

		if !isSpace(r) {
			rd.Unread(r)
			break
		}
	}

	return nil
}

// Token reads one token from the reader and returns. If init is not -1, it is included
// as first character in the token.
func (rd *Reader) Token(init rune) (string, error) {
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

		if rd.IsTerminal(r) {
			rd.Unread(r)
			break
		}

		b.WriteRune(r)
	}

	return b.String(), nil
}

// Container reads multiple forms until 'end' rune is reached. Should be used to read
// collection types like List etc. formType is only used to annotate errors.
func (rd Reader) Container(end rune, formType string, f func(ww.Any) error) error {
	for {
		if err := rd.SkipSpaces(); err != nil {
			if err == io.EOF {
				return Error{Cause: ErrEOF}
			}
			return err
		}

		r, err := rd.NextRune()
		if err != nil {
			if err == io.EOF {
				return Error{Cause: ErrEOF}
			}
			return err
		}

		if r == end {
			break
		}
		rd.Unread(r)

		expr, err := rd.readOne()
		if err != nil {
			if err == ErrSkip {
				continue
			}
			return err
		}

		// TODO(performance):  verify `f` is inlined by the compiler
		if err = f(expr); err != nil {
			return err
		}
	}

	return nil
}

// readOne is same as One() but always returns un-annotated errors.
func (rd *Reader) readOne() (ww.Any, error) {
	if err := rd.SkipSpaces(); err != nil {
		return nil, err
	}

	r, err := rd.NextRune()
	if err != nil {
		return nil, err
	}

	if unicode.IsNumber(r) {
		return readNumber(rd, r)
	} else if r == '+' || r == '-' {
		r2, err := rd.NextRune()
		if err != nil && err != io.EOF {
			return nil, err
		}

		if err != io.EOF {
			rd.Unread(r2)
			if unicode.IsNumber(r2) {
				return readNumber(rd, r)
			}
		}
	}

	macro, found := rd.macros[r]
	if found {
		return macro(rd, r)
	}

	if r == dispatchTrigger {
		f, err := rd.execDispatch()
		if f != nil || err != nil {
			return f, err
		}
	}

	return readSymbol(rd, r)
}

func (rd *Reader) execDispatch() (ww.Any, error) {
	r2, err := rd.NextRune()
	if err != nil {
		// ignore the error and let readOne handle it.
		return nil, nil
	}

	dispatchMacro, found := rd.dispatch[r2]
	if !found {
		rd.Unread(r2)
		return nil, nil
	}

	rd.dispatching = true
	defer func() {
		rd.dispatching = false
	}()

	form, err := dispatchMacro(rd, r2)
	if err != nil {
		return nil, err
	}

	return form, nil
}

func (rd *Reader) annotateErr(err error, beginPos Position, form string) error {
	if err == io.EOF || err == ErrSkip {
		return err
	}

	readErr := Error{}
	if e, ok := err.(Error); ok {
		readErr = e
	} else {
		readErr = Error{Cause: err}
	}

	readErr.Form = form
	readErr.Begin = beginPos
	readErr.End = rd.Position()
	return readErr
}

func inferFileName(rs io.Reader) string {
	switch r := rs.(type) {
	case *os.File:
		return r.Name()

	case *strings.Reader:
		return "<string>"

	case *bytes.Reader, *bytes.Buffer:
		return "<bytes>"

	case network.Stream:
		return fmt.Sprintf("<%s>", r.Conn().RemoteMultiaddr())

	case net.Conn:
		return fmt.Sprintf("<con:%s>", r.RemoteAddr())

	default:
		return fmt.Sprintf("<%s>", reflect.TypeOf(rs))
	}
}

// Position represents the positional information about a value read
// by reader.
type Position struct {
	File string
	Ln   int
	Col  int
}

func (p Position) String() string {
	if p.File == "" {
		p.File = "<unknown>"
	}
	return fmt.Sprintf("%s:%d:%d", p.File, p.Ln, p.Col)
}
