package anchor

import (
	"context"

	"capnproto.org/go/capnp/v3"
	"github.com/wetware/ww/pkg/cap/cluster"
)

type HostAnchorImpl struct {
	path []string
	host string
}

func (hai *HostAnchorImpl) Path() []string {
	return hai.path
}

func (hai *HostAnchorImpl) Host() string {
	return hai.host
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
	return &HostAnchorImpl{path: append(hai.path, rec.Peer().String()), host: rec.Peer().String()}
}

func (hai *HostAnchorIterator) Finish() {
	hai.release()
}
