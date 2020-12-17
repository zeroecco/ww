package shell

import (
	"context"
	"errors"

	"github.com/urfave/cli/v2"
	clientutil "github.com/wetware/ww/internal/util/client"
	ww "github.com/wetware/ww/pkg"
	"github.com/wetware/ww/pkg/client"
	anchorpath "github.com/wetware/ww/pkg/util/anchor/path"
	"go.uber.org/fx"
)

// AnchorProvider wraps the root anchor, allowing fx to
// provide it as a dependency before dialing into the cluster.
type AnchorProvider interface {
	Anchor() ww.Anchor
}

func newAnchorProvider(c *cli.Context, lx fx.Lifecycle) AnchorProvider {
	if !c.Bool("dial") {
		return nopAnchorProvider{}
	}

	var p clientProvider
	lx.Append(fx.Hook{
		OnStart: func(ctx context.Context) (err error) {
			p.c, err = clientutil.Dial(ctx, c)
			return
		},
		OnStop: func(c context.Context) error {
			return p.c.Close()
		},
	})

	return &p
}

type clientProvider struct{ c client.Client }

func (p clientProvider) Anchor() ww.Anchor { return p.c }

type nopAnchorProvider struct{}

func (nopAnchorProvider) Anchor() ww.Anchor { return nopAnchor{} }

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
