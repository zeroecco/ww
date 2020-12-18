package reader

import (
	"bytes"
	"io"
	"math/big"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ww "github.com/wetware/ww/pkg"
	"github.com/wetware/ww/pkg/lang/core"
	capnp "zombiezen.com/go/capnproto2"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		r        io.Reader
		fileName string
	}{
		{
			name:     "WithStringReader",
			r:        strings.NewReader(":test"),
			fileName: "<string>",
		},
		{
			name:     "WithBytesReader",
			r:        bytes.NewReader([]byte(":test")),
			fileName: "<bytes>",
		},
		{
			name:     "WihFile",
			r:        os.NewFile(0, "test.lisp"),
			fileName: "test.lisp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rd := New(tt.r)
			require.NotNil(t, rd)
			assert.Equal(t, tt.fileName, rd.File)
		})
	}
}

func TestReader_SetMacro(t *testing.T) {
	t.Run("UnsetDefaultMacro", func(t *testing.T) {
		rd := New(strings.NewReader("~hello"))
		rd.SetMacro('~', false, nil) // remove unquote operator

		want := mustSymbol("~hello")

		got, err := rd.One()
		if err != nil {
			t.Errorf("unexpected error: %#v", err)
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got = %+v, want = %+v", got, want)
		}
	})

	t.Run("DispatchMacro", func(t *testing.T) {
		rd := New(strings.NewReader("#$123"))
		// `#$` returns string "USD"
		rd.SetMacro('$', true, func(rd *Reader, init rune) (ww.Any, error) {
			return mustString("USD"), nil
		}) // remove unquote operator

		want := mustString("USD")

		got, err := rd.One()
		if err != nil {
			t.Errorf("unexpected error: %#v", err)
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got = %+v, want = %+v", got, want)
		}
	})

	t.Run("CustomMacro", func(t *testing.T) {
		rd := New(strings.NewReader("~hello"))
		rd.SetMacro('~', false, func(rd *Reader, _ rune) (ww.Any, error) {
			var ru []rune
			for {
				r, err := rd.NextRune()
				if err != nil {
					if err == io.EOF {
						break
					}
					return nil, err
				}

				if rd.IsTerminal(r) {
					break
				}
				ru = append(ru, r)
			}

			return mustString(string(ru)), nil
		}) // override unquote operator

		want := mustString("hello")

		got, err := rd.One()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got = %+v, want = %+v", got, want)
		}
	})
}

func TestReader_All(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		want    []ww.Any
		wantErr bool
	}{
		{
			name: "ValidLiteralSample",
			src:  `123 "Hello World" 12.34 -0xF +010 true nil 0b1010 \a :hello`,
			want: []ww.Any{
				mustInt64(123),
				mustString("Hello World"),
				mustFloat64(12.34),
				mustInt64(-15),
				mustInt64(8),
				core.True,
				core.Nil{},
				mustInt64(10),
				mustChar('a'),
				mustKeyword("hello"),
			},
		},
		{
			name: "WithComment",
			src:  `:valid-keyword ; comment should return errSkip`,
			want: []ww.Any{mustKeyword("valid-keyword")},
		},
		{
			name:    "UnterminatedString",
			src:     `:valid-keyword "unterminated string literal`,
			wantErr: true,
		},
		{
			name: "CommentFollowedByForm",
			src:  `; comment should return errSkip` + "\n" + `:valid-keyword`,
			want: []ww.Any{mustKeyword("valid-keyword")},
		},
		{
			name:    "UnterminatedList",
			src:     `:valid-keyword (add 1 2`,
			wantErr: true,
		},
		{
			name:    "EOFAfterQuote",
			src:     `:valid-keyword '`,
			wantErr: true,
		},
		{
			name:    "CommentAfterQuote",
			src:     `:valid-keyword ';hello world`,
			wantErr: true,
		},
		{
			name:    "UnbalancedParenthesis",
			src:     `())`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(strings.NewReader(tt.src)).All()
			if (err != nil) != tt.wantErr {
				t.Errorf("All() error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("All() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestReader_One(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name:    "Empty",
			src:     "",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "QuotedEOF",
			src:     "';comment is a no-op form\n",
			wantErr: true,
		},
		{
			name:    "ListEOF",
			src:     "( 1",
			wantErr: true,
		},
		{
			name: "UnQuote",
			src:  "~(x 3)",
			want: mustList(
				mustSymbol("unquote"),
				mustList(
					mustSymbol("x"),
					mustInt64(3),
				),
			),
		},
	})
}

func TestReader_One_Number(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name: "NumberWithLeadingSpaces",
			src:  "    +1234",
			want: mustInt64(1234),
		},
		{
			name: "PositiveInt",
			src:  "+1245",
			want: mustInt64(1245),
		},
		{
			name: "NegativeInt",
			src:  "-234",
			want: mustInt64(-234),
		},
		{
			name: "PositiveFloat",
			src:  "+1.334",
			want: mustFloat64(1.334),
		},
		{
			name: "NegativeFloat",
			src:  "-1.334",
			want: mustFloat64(-1.334),
		},
		{
			name: "PositiveHex",
			src:  "0x124",
			want: mustInt64(0x124),
		},
		{
			name: "NegativeHex",
			src:  "-0x124",
			want: mustInt64(-0x124),
		},
		{
			name: "PositiveOctal",
			src:  "0123",
			want: mustInt64(0123),
		},
		{
			name: "NegativeOctal",
			src:  "-0123",
			want: mustInt64(-0123),
		},
		{
			name: "PositiveBinary",
			src:  "0b10",
			want: mustInt64(2),
		},
		{
			name: "NegativeBinary",
			src:  "-0b10",
			want: mustInt64(-2),
		},
		{
			name: "PositiveBase2Radix",
			src:  "2r10",
			want: mustInt64(2),
		},
		{
			name: "NegativeBase2Radix",
			src:  "-2r10",
			want: mustInt64(-2),
		},
		{
			name: "PositiveBase4Radix",
			src:  "4r123",
			want: mustInt64(27),
		},
		{
			name: "NegativeBase4Radix",
			src:  "-4r123",
			want: mustInt64(-27),
		},
		{
			name: "ScientificSimple",
			src:  "1e10",
			want: mustFloat64(1e10),
		},
		{
			name: "ScientificNegativeExponent",
			src:  "1e-10",
			want: mustFloat64(1e-10),
		},
		{
			name: "ScientificWithDecimal",
			src:  "1.5e10",
			want: mustFloat64(1.5e+10),
		},
		{
			name:    "FloatStartingWith0",
			src:     "012.3",
			want:    mustFloat64(012.3),
			wantErr: false,
		},
		{
			name:    "InvalidValue",
			src:     "1ABe13",
			wantErr: true,
		},
		{
			name:    "InvalidScientificFormat",
			src:     "1e13e10",
			wantErr: true,
		},
		{
			name:    "InvalidExponent",
			src:     "1e1.3",
			wantErr: true,
		},
		{
			name:    "InvalidRadixFormat",
			src:     "1r2r3",
			wantErr: true,
		},
		{
			name:    "RadixBase3WithDigit4",
			src:     "-3r1234",
			wantErr: true,
		},
		{
			name:    "RadixMissingValue",
			src:     "2r",
			wantErr: true,
		},
		{
			name:    "RadixInvalidBase",
			src:     "2ar",
			wantErr: true,
		},
		{
			name:    "RadixWithFloat",
			src:     "2.3r4",
			wantErr: true,
		},
		{
			name:    "DecimalPointInBinary",
			src:     "0b1.0101",
			wantErr: true,
		},
		{
			name:    "InvalidDigitForOctal",
			src:     "08",
			wantErr: true,
		},
		{
			name:    "IllegalNumberFormat",
			src:     "9.3.2",
			wantErr: true,
		},
		{
			name: "BigScientific",
			src:  "1.84467440737095516159999e20",
			want: mustBigFloat("1.84467440737095516159999e20"),
		},
	})
}

func TestReader_One_String(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name: "SimpleString",
			src:  `"hello"`,
			want: mustString("hello"),
		},
		{
			name: "EscapeQuote",
			src:  `"double quote is \""`,
			want: mustString(`double quote is "`),
		},
		{
			name: "EscapeTab",
			src:  `"hello\tworld"`,
			want: mustString("hello\tworld"),
		},
		{
			name: "EscapeSlash",
			src:  `"hello\\world"`,
			want: mustString(`hello\world`),
		},
		{
			name:    "UnexpectedEOF",
			src:     `"double quote is`,
			wantErr: true,
		},
		{
			name:    "InvalidEscape",
			src:     `"hello \x world"`,
			wantErr: true,
		},
		{
			name:    "EscapeEOF",
			src:     `"hello\`,
			wantErr: true,
		},
	})
}

func TestReader_One_Keyword(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name: "SimpleASCII",
			src:  `:test`,
			want: mustKeyword("test"),
		},
		{
			name: "LeadingTrailingSpaces",
			src:  "          :test          ",
			want: mustKeyword("test"),
		},
		{
			name: "SimpleUnicode",
			src:  `:∂`,
			want: mustKeyword("∂"),
		},
		{
			name: "WithSpecialChars",
			src:  `:this-is-valid?`,
			want: mustKeyword("this-is-valid?"),
		},
		{
			name: "FollowedByMacroChar",
			src:  `:this-is-valid'hello`,
			want: mustKeyword("this-is-valid"),
		},
	})
}

func TestReader_One_Character(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name: "ASCIILetter",
			src:  `\a`,
			want: mustChar('a'),
		},
		{
			name: "ASCIIDigit",
			src:  `\1`,
			want: mustChar('1'),
		},
		{
			name: "Unicode",
			src:  `\∂`,
			want: mustChar('∂'),
		},
		{
			name: "Newline",
			src:  `\newline`,
			want: mustChar('\n'),
		},
		{
			name: "FormFeed",
			src:  `\formfeed`,
			want: mustChar('\f'),
		},
		{
			name: "Unicode",
			src:  `\u00AE`,
			want: mustChar('®'),
		},
		{
			name:    "InvalidUnicode",
			src:     `\uHELLO`,
			wantErr: true,
		},
		{
			name:    "OutOfRangeUnicode",
			src:     `\u-100`,
			wantErr: true,
		},
		{
			name:    "UnknownSpecial",
			src:     `\hello`,
			wantErr: true,
		},
		{
			name:    "EOF",
			src:     `\`,
			wantErr: true,
		},
	})
}

func TestReader_One_Symbol(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name: "SimpleASCII",
			src:  `hello`,
			want: mustSymbol("hello"),
		},
		{
			name: "Unicode",
			src:  `find-∂`,
			want: mustSymbol("find-∂"),
		},
		{
			name: "SingleChar",
			src:  `+`,
			want: mustSymbol("+"),
		},
	})
}

func TestReader_One_List(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name: "EmptyList",
			src:  `()`,
			want: mustList(),
		},
		{
			name: "ListWithOneEntry",
			src:  `(help)`,
			want: mustList(mustSymbol("help")),
		},
		{
			name: "ListWithMultipleEntry",
			src:  `(+ 0xF 3.1413)`,
			want: mustList(
				mustSymbol("+"),
				mustInt64(15),
				mustFloat64(3.1413),
			),
		},
		{
			name: "ListWithCommaSeparator",
			src:  `(+,0xF,3.1413)`,
			want: mustList(
				mustSymbol("+"),
				mustInt64(15),
				mustFloat64(3.1413),
			),
		},
		{
			name: "MultiLine",
			src: `(+
                      0xF
                      3.1413
					)`,
			want: mustList(
				mustSymbol("+"),
				mustInt64(15),
				mustFloat64(3.1413),
			),
		},
		{
			name: "MultiLineWithComments",
			src: `(+         ; plus operator adds numerical values
                      0xF    ; hex representation of 15
                      3.1413 ; value of math constant pi
                  )`,
			want: mustList(
				mustSymbol("+"),
				mustInt64(15),
				mustFloat64(3.1413),
			),
		},
		{
			name:    "UnexpectedEOF",
			src:     "(+ 1 2 ",
			wantErr: true,
		},
	})
}

func TestReader_One_Vector(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name: "EmptyVector",
			src:  `[]`,
			want: core.EmptyVector,
		},
		{
			name: "VectorWithOneEntry",
			src:  `[help]`,
			want: mustVec(mustSymbol("help")),
		},
		{
			name: "VectorWithMultipleEntry",
			src:  `[+ 0xF 3.1413]`,
			want: mustVec(
				mustSymbol("+"),
				mustInt64(15),
				mustFloat64(3.1413),
			),
		},
		{
			name: "VectorWithCommaSeparator",
			src:  `[+,0xF,3.1413]`,
			want: mustVec(
				mustSymbol("+"),
				mustInt64(15),
				mustFloat64(3.1413),
			),
		},
		{
			name: "MultiLine",
			src: `[+
                      0xF
                      3.1413
					]`,
			want: mustVec(
				mustSymbol("+"),
				mustInt64(15),
				mustFloat64(3.1413),
			),
		},
		{
			name: "MultiLineWithComments",
			src: `[+         ; plus operator adds numerical values
                      0xF    ; hex representation of 15
                      3.1413 ; value of math constant pi
                  ]`,
			want: mustVec(
				mustSymbol("+"),
				mustInt64(15),
				mustFloat64(3.1413),
			),
		},
		{
			name:    "UnexpectedEOF",
			src:     "[+ 1 2 ",
			wantErr: true,
		},
	})
}

type readerTestCase struct {
	name    string
	src     string
	want    ww.Any
	wantErr bool
}

func executeReaderTests(t *testing.T, tests []readerTestCase) {
	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(strings.NewReader(tt.src)).One()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				eq, err := core.Eq(tt.want, got)
				require.NoError(t, err)
				assert.True(t, eq,
					"expected '%s', got '%s'.",
					mustRender(tt.want), mustRender(got))
			}
		})
	}
}

func mustSymbol(s string) core.Symbol {
	v, _ := core.NewSymbol(capnp.SingleSegment(nil), s)
	return v
}

func mustString(s string) core.String {
	v, _ := core.NewString(capnp.SingleSegment(nil), s)
	return v
}

func mustKeyword(s string) core.Keyword {
	v, _ := core.NewKeyword(capnp.SingleSegment(nil), s)
	return v
}

func mustChar(r rune) core.Char {
	v, _ := core.NewChar(capnp.SingleSegment(nil), r)
	return v
}

func mustInt64(i int64) core.Int64 {
	v, _ := core.NewInt64(capnp.SingleSegment(nil), i)
	return v
}

func mustFloat64(f float64) core.Float64 {
	v, _ := core.NewFloat64(capnp.SingleSegment(nil), f)
	return v
}

func mustList(items ...ww.Any) core.List {
	l, _ := core.NewList(capnp.SingleSegment(nil), items...)
	return l
}

func mustVec(items ...ww.Any) core.Vector {
	v, _ := core.NewVector(capnp.SingleSegment(nil), items...)
	return v
}

func mustBigFloat(q string) core.BigFloat {
	var f big.Float
	if _, ok := f.SetString(q); !ok {
		panic("invalid float " + q)
	}

	bf, err := core.NewBigFloat(capnp.SingleSegment(nil), &f)
	if err != nil {
		panic(err)
	}

	return bf
}

func mustRender(item ww.Any) string {
	s, err := core.Render(item)
	if err != nil {
		panic(err)
	}

	return s
}
