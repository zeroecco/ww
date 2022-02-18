package cluster

import (
	"context"

	"capnproto.org/go/capnp/v3"
	"github.com/wetware/ww/internal/api/cluster"
)

// HACK:  attempt at writing a client...
type Anchor cluster.Anchor

func (a Anchor) Ls(ctx context.Context, ps func(cluster.Anchor_ls_Params) error) (FutureAnchor, capnp.ReleaseFunc) {
	f, release := cluster.Anchor(a).Ls(ctx, ps)
	return FutureAnchor(f), release
}

func (a Anchor) Walk(ctx context.Context, ps func(cluster.Anchor_walk_Params) error) (cluster.Anchor_walk_Results_Future, capnp.ReleaseFunc) {
	panic("NOT IMPLEMENTED")
}

type FutureAnchor cluster.Anchor_ls_Results_Future

type RootAnchor struct {
	Cluster RoutingTable
}

func (root RootAnchor) Ls(ctx context.Context, call cluster.Anchor_ls) error {
	res, err := call.AllocResults()
	if err != nil {
		return err
	}

	view := ViewFactory{root.Cluster}.NewClient(nil)

	return res.SetView(cluster.View(view))
}

func (root RootAnchor) Walk(ctx context.Context, call cluster.Anchor_walk) error {
	panic("NOT IMPLEMENTED")
}
