package anchor

import (
	"context"

	"capnproto.org/go/capnp/v3"
	"github.com/wetware/casm/pkg/cluster/routing"
	"github.com/wetware/ww/pkg/cap/cluster"
)

type HostAnchorImpl struct {
	path []string
	rec  routing.Record
}

func (hai *HostAnchorImpl) Path() []string {
	return hai.path
}

func (hai *HostAnchorImpl) Host() string {
	return hai.rec.Peer().String()
}

type HostAnchorIterator struct {
	path    []string
	it      *cluster.Iterator
	release capnp.ReleaseFunc
}

func (hai *HostAnchorIterator) Next(ctx context.Context) bool {
	return hai.it.Next(ctx)
}

func (hai *HostAnchorIterator) Anchor() Anchor {
	rec := hai.it.Record()
	return &HostAnchorImpl{path: append(hai.path, rec.Peer().String()), rec: rec}
}

func (hai *HostAnchorIterator) Finish() {
	hai.release()
}
