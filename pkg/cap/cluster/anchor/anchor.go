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
	Finish()
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
	Set(context.Context, []byte) error
	Get(context.Context) ([]byte, error)
}

type ProcessAnchor interface {
	Anchor
	//TODO
}

type ChannelAnchor interface {
	Anchor
	//TODO
}
