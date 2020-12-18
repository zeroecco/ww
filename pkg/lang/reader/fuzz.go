// +build gofuzz

/*
 *	Fuzz test harness.
 *
 *  To build:  make fuzz-target
 *  To run:    make fuzz
 */

package reader

import (
	"bytes"

	"github.com/wetware/ww/pkg/lang/core"
)

var reader *Reader

func init() {
	reader = New(bytes.NewReader(nil))
}

// Fuzz testing hook.
func Fuzz(data []byte) int {
	reader.Reset(bytes.NewReader(data))

	form, err := reader.One()
	if err != nil {
		if form != nil {
			panic("non-nil form on read error")
		}
		return 0
	}

	s, err := core.Render(form)
	if err != nil {
		if s != "" {
			panic("non-empty string on render error")
		}
	}

	return 1
}
