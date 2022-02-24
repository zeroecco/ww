package anchor

import (
	"context"
	"errors"
	"strings"

	"capnproto.org/go/capnp/v3/server"
	"github.com/hashicorp/go-multierror"
	api "github.com/wetware/ww/internal/api/cluster"
	"github.com/wetware/ww/pkg/cap/cluster"
)

var defaultPolicy = server.Policy{
	// HACK:  raise MaxConcurrentCalls to mitigate known deadlock condition.
	//        https://github.com/capnproto/go-capnproto2/issues/189
	MaxConcurrentCalls: 64,
}

var (
	ErrInvalidPath      = errors.New("invalid path")
	ErrInvalidOperation = errors.New("invalid operation")
	ErrUnknownAnchor    = errors.New("unknown anchor")
)

type AnchorClient struct {
	router cluster.RoutingTable
	ap     api.AnchorProvider
}

func NewAnchorClient(router cluster.RoutingTable) AnchorClient {
	return AnchorClient{
		router: router,
		ap:     api.AnchorProvider_ServerToClient(newAnchorServer(), &defaultPolicy),
	}
}

func (ac AnchorClient) Ls(ctx context.Context, path []string) (AnchorIterator, error) {
	if !isValid(path) {
		return nil, ErrInvalidPath
	}

	if len(path) == 1 {
		vf := cluster.ViewFactory{View: ac.router}
		it, release := vf.NewClient(&defaultPolicy).Iter(ctx)
		return &HostAnchorIterator{
			path:    []string{"/"},
			it:      it,
			release: release,
		}, nil
	} else {
		// TODO
	}
	return nil, nil
}

func (ac AnchorClient) Walk(ctx context.Context, path []string) (Anchor, error) {
	if !isValid(path) {
		return nil, ErrInvalidPath
	}
	if len(path) == 1 { // root anchor
		return RootAnchor{}, nil
	} else if len(path) == 2 {
		vf := cluster.ViewFactory{View: ac.router}
		it, release := vf.NewClient(&defaultPolicy).Iter(ctx)
		defer release()

		for it.Next(ctx) {
			if strings.Compare(it.Record().Peer().String(), path[1]) == 0 {
				return &HostAnchorImpl{path: path, host: it.Record().Peer().String()}, nil
			}
		}
		return nil, multierror.Append(ErrInvalidPath, errors.New("host anchor does not exist"))
	} else {
		fut, _ := ac.ap.Walk(ctx, func(ap api.AnchorProvider_walk_Params) error {
			textList, err := ap.NewPath(int32(len(path)))
			if err != nil {
				return err
			}
			for i, p := range path {
				if err := textList.Set(i, p); err != nil {
					return err
				}
			}
			return nil
		})
		// TODO: defer release()

		select {
		case <-fut.Done():
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		results, err := fut.Struct()
		if err != nil {
			return nil, err
		}
		anchor, err := results.Anchor()
		if err != nil {
			return nil, err
		}

		return toAnchor(path, anchor)
	}
}

func isValid(path []string) bool {
	return true // TODO
}

type RootAnchor struct{}

func (ra RootAnchor) Path() []string {
	return []string{"/"}
}

func toAnchor(path []string, anchor api.Anchor) (Anchor, error) {
	return ContainerAnchorImpl{path: path, cap: anchor.Container()}, nil
}
