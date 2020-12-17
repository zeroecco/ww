package reader

import (
	ww "github.com/wetware/ww/pkg"
	"github.com/wetware/ww/pkg/lang/core"
)

var defaultSymTable = map[string]ww.Any{
	"nil":   core.Nil{},
	"false": core.False,
	"true":  core.True,
}

// Option values can be used with New() to configure the reader during init.
type Option func(*Reader)

// WithNumReader sets the number reader macro to be used by the Reader. Uses
// the default number reader if nil.
func WithNumReader(m Macro) Option {
	if m == nil {
		m = readNumber
	}
	return func(rd *Reader) {
		rd.numReader = m
	}
}

// WithSymbolReader sets the symbol reader macro to be used by the Reader.
func WithSymbolReader(m Macro) Option {
	if m == nil {
		return WithBuiltinSymbolReader()
	}
	return func(rd *Reader) {
		rd.symReader = m
	}
}

// WithBuiltinSymbolReader configures the default symbol reader with given
// symbol table.
func WithBuiltinSymbolReader() Option {
	m := symbolReader(defaultSymTable)
	return func(rd *Reader) {
		rd.symReader = m
	}
}

func withDefaults(opt []Option) []Option {
	return append([]Option{
		WithNumReader(nil),
		WithBuiltinSymbolReader(),
	}, opt...)
}
