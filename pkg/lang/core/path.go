package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/wetware/ww/internal/mem"
	ww "github.com/wetware/ww/pkg"
	anchorpath "github.com/wetware/ww/pkg/util/anchor/path"
	memutil "github.com/wetware/ww/pkg/util/mem"
	capnp "zombiezen.com/go/capnproto2"
)

var (
	_ Path = (*BoundPath)(nil)
	_ Path = (*UnboundPath)(nil)

	_ Invokable = (*BoundPath)(nil)

	rootPath UnboundPath
)

func init() {
	var err error
	if rootPath.Any, err = memutil.Alloc(capnp.SingleSegment(nil)); err != nil {
		panic(err)
	}

	if err = rootPath.SetPath(""); err != nil {
		panic(err)
	}
}

// Path points to a unique Anchor.
type Path interface {
	ww.Any
	Parts() ([]string, error)
}

// NewPath .
func NewPath(a capnp.Arena, p string) (UnboundPath, error) {
	if p = strings.Trim(p, "/"); p == "" {
		return rootPath, nil
	}

	any, err := memutil.Alloc(a)
	if err == nil {
		err = any.SetPath(p)
	}

	return UnboundPath{any}, err
}

// UnboundPath is an 'unresolved' path that has not been associated with
// its corresponding Anchor.
type UnboundPath struct{ mem.Any }

// Value returns the memory value for p.
func (p UnboundPath) Value() mem.Any { return p.Any }

// Parts of the path.
func (p UnboundPath) Parts() ([]string, error) {
	path, err := p.Path()
	return anchorpath.Parts(path), err
}

// String returns a human-readable representation of the path that is suitable
// for printing.
func (p UnboundPath) String() (string, error) {
	path, err := p.Any.Path()
	return "/" + path, err
}

// BoundPath is a 'resolved' path that has been associated with its
// corresponding Anchor.
type BoundPath struct {
	mem.Any
	Anchor ww.Anchor
}

// Value returns the memory value for p.
func (p BoundPath) Value() mem.Any { return p.Any }

// Parts of the path.
func (p BoundPath) Parts() ([]string, error) {
	return UnboundPath{p.Any}.Parts()
}

// String returns a human-readable representation of the path that is suitable
// for printing.
func (p BoundPath) String() (string, error) {
	return UnboundPath{p.Any}.String()
}

// Invoke is used to get/set the anchor's value.
func (p BoundPath) Invoke(args ...ww.Any) (ww.Any, error) {
	switch len(args) {
	case 0:
		return p.Anchor.Load(context.TODO())

	case 1:
		return Nil{}, p.Anchor.Store(context.TODO(), args[0])

	}

	return nil, fmt.Errorf("expected 0 or 1 argument, got %d", len(args))
}
