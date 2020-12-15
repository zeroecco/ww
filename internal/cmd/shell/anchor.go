package shell

import (
	"context"
	"errors"

	ww "github.com/wetware/ww/pkg"
	anchorpath "github.com/wetware/ww/pkg/util/anchor/path"
)

type nopAnchor []string

func (a nopAnchor) Name() string {
	if anchorpath.Root(a) {
		return ""
	}

	return a[len(a)-1]
}

func (a nopAnchor) Path() []string { return a }

func (nopAnchor) Ls(context.Context) ([]ww.Anchor, error) { return nil, nil }

func (a nopAnchor) Walk(_ context.Context, path []string) ww.Anchor {
	return append(a, path...)
}

func (a nopAnchor) Load(context.Context) (ww.Any, error) {
	return nil, errors.New("anchor not found")
}

func (a nopAnchor) Store(context.Context, ww.Any) error {
	if anchorpath.Root(a) {
		return errors.New("not implemented")
	}

	return errors.New("anchor not found")
}

func (a nopAnchor) Go(context.Context, ...ww.Any) (ww.Any, error) {
	if anchorpath.Root(a) {
		return nil, errors.New("not implemented")
	}

	return nil, errors.New("anchor not found")
}
