package client

import "context"

type Anchor interface {
	Ls(context.Context) (AnchorSet, error)
	Walk(context.Context, []string) Anchor
}

type AnchorSet interface {
	Anchor() Anchor
}
