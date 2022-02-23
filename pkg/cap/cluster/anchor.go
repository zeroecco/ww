package cluster

import (
	"context"
	"errors"

	"capnproto.org/go/capnp/v3"
	"github.com/hashicorp/go-multierror"
	"github.com/wetware/casm/pkg/cluster/routing"
)

// HACK:  attempt at writing a client...

var (
	ErrInvalidPath      = errors.New("invalid path")
	ErrInvalidOperation = errors.New("invalid operation")
)

type AnchorIterator interface {
	Next(ctx context.Context) bool
	Anchor() Anchor
}

type Anchor interface {
	Path() []string
	Set(context.Context, interface{}) error
}

type HostAnchor interface {
	Anchor
	Host() string
}

type AnchorClient struct {
	router RoutingTable
}

func NewAnchorClient(router RoutingTable) AnchorClient {
	return AnchorClient{router: router}
}

type HostAnchorImpl struct {
	path []string
	rec  routing.Record
}

func (hai *HostAnchorImpl) Path() []string {
	return hai.path
}

func (hai *HostAnchorImpl) Set(context.Context, interface{}) error {
	return multierror.Append(errors.New("host anchor"), ErrInvalidOperation)
}

func (hai *HostAnchorImpl) Host() string {
	return hai.rec.Peer().String()
}

type HostAnchorIterator struct {
	path    []string
	it      *Iterator
	release capnp.ReleaseFunc
}

func (hai HostAnchorIterator) Next(ctx context.Context) bool {
	return hai.it.Next(ctx)
}

func (hai HostAnchorIterator) Anchor() Anchor {
	rec := hai.it.Record()
	return &HostAnchorImpl{path: append(hai.path, rec.Peer().String()), rec: rec}
}

func (ac AnchorClient) Ls(ctx context.Context, path []string) (AnchorIterator, error) {
	if !isValid(path) {
		return nil, ErrInvalidPath
	}

	if len(path) == 1 {
		vf := ViewFactory{ac.router}
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
