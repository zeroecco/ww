package anchor

import (
	"context"
)

type AnchorProvider interface {
	Ls(ctx context.Context, path []string) (AnchorIterator, error)
	Walk(ctx context.Context, path []string) (Anchor, error)
}

type AnchorIterator interface {
	Next(ctx context.Context) bool
	Anchor() Anchor
}

type Anchor interface {
	Path() []string
}

type HostAnchor interface {
	Anchor
	Host() string
}

type ContainerAnchor interface {
	Anchor
	Set(context.Context, interface{}) error
}

type DataAnchor interface {
	ContainerAnchor
	Get(context.Context) interface{}
}

type ProcessAnchor interface {
	ContainerAnchor
	//TODO
}

type ChannelAnchor interface {
	ContainerAnchor
	//TODO
}
