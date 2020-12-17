// Package lang contains the wetware language iplementation
package lang

import (
	"fmt"

	capnp "zombiezen.com/go/capnproto2"

	ww "github.com/wetware/ww/pkg"
	"github.com/wetware/ww/pkg/lang/core"
	// _ "github.com/wetware/ww/pkg/lang/core/proc" // register default process types
)

// VM is a virtual machine that can evaluate forms.
type VM struct {
	Analyzer core.Analyzer
	Env      core.Env
}

// Eval a form.
func (vm VM) Eval(form ww.Any) (ww.Any, error) {
	return core.Eval(vm.Analyzer, vm.Env, form)
}

// Init populates the default environment and loads the prelude.
// This is done seperately from 'New' in order to support lazy initialization.
func (vm VM) Init() (err error) {
	if err = vm.loadBuiltins(); err == nil {
		err = vm.loadPrelude()
	}
	return
}

func (vm VM) loadBuiltins() error {
	return bindAll(vm.Env,
		text("__version__", ww.Version),
		text("__author__", "Louis Thibault"),
		text("__copyright__", "2020, Louis Thibault\nAll rights reserved."),

		anchor(),
		comparison(),
		container(),
		function("error", "__error__", core.Raise),
		function("nil?", "__isnil__", core.IsNil),
		function("type", "__type__", fnTypeOf),
		function("not", "__not__", fnNot),
		function("read", "__read__", fnRead),
		function("render", "__render__", core.Render),
		function("print", "__print__", fnPrint))
}

func (vm VM) loadPrelude() error {
	// We're effectively running `(import :prelude)`

	sym, err := core.NewSymbol(capnp.SingleSegment(nil), "import")
	if err != nil {
		return err
	}

	kw, err := core.NewKeyword(capnp.SingleSegment(nil), "prelude")
	if err != nil {
		return err
	}

	// (import :prelude)
	form, err := core.NewList(capnp.SingleSegment(nil), sym, kw)
	if err != nil {
		return err
	}

	if _, err = vm.Eval(form); err != nil {
		err = fmt.Errorf("load prelude: %w", err)
	}
	return err
}
