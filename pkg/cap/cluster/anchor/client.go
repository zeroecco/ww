package anchor

import (
	"context"
	"errors"

	"capnproto.org/go/capnp/v3/server"
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
)

type AnchorClient struct {
	router cluster.RoutingTable
}

func NewAnchorClient(router cluster.RoutingTable) AnchorClient {
	return AnchorClient{router: router}
}

func (ac AnchorClient) Ls(ctx context.Context, path []string) (AnchorIterator, error) {
	if !isValid(path) {
		return nil, ErrInvalidPath
	}

	if len(path) == 1 {
		vf := cluster.ViewFactory{View: ac.router}
		it, release := vf.NewClient(&defaultPolicy).Iter(ctx)
		return HostAnchorIterator{
			path:    []string{"/"},
			it:      it,
			release: release,
		}, nil
	} else {
		// TODO
	}
	return nil, nil
}

func isValid(path []string) bool {
	return true // TODO
}
