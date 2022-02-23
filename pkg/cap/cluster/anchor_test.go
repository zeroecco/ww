package cluster_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/wetware/ww/pkg/cap/cluster/anchor"
)

func TestRootAnchor(t *testing.T) {
	t.Parallel()
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dl := time.Now().Add(time.Second * 10)
	var rt = make(routingTable, 65)
	for i := range rt {
		rt[i] = record{
			id:  newID(),
			ttl: time.Second * 10,
			seq: uint64(i),
			dl:  dl,
		}
	}

	client := anchor.NewAnchorClient(rt)

	it, err := client.Ls(ctx, []string{"/"})
	require.NoError(t, err)

	for i := 0; it.Next(ctx); i++ {
		a := it.Anchor()
		require.NotNil(t, a)
		path := a.Path()
		require.Equal(t, "/"+rt[i].Peer().String(), strings.Join(path, ""))

		ha, ok := a.(anchor.HostAnchor)
		require.True(t, ok)
		require.Equal(t, rt[i].Peer().String(), ha.Host())
	}

	for i := 0; i < len(rt); i++ {
		a, err := client.Walk(ctx, append([]string{"/"}, rt[i].Peer().String()))
		require.NotNil(t, a)
		require.NoError(t, err)

		path := a.Path()
		require.Equal(t, "/"+rt[i].Peer().String(), strings.Join(path, ""))

		ha, ok := a.(anchor.HostAnchor)
		require.True(t, ok)
		require.Equal(t, rt[i].Peer().String(), ha.Host())
	}
}

func TestInvalidHostAnchor(t *testing.T) {
	t.Parallel()
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dl := time.Now().Add(time.Second * 10)
	var rt = make(routingTable, 65)
	for i := range rt {
		rt[i] = record{
			id:  newID(),
			ttl: time.Second * 10,
			seq: uint64(i),
			dl:  dl,
		}
	}

	client := anchor.NewAnchorClient(rt)
	a, err := client.Walk(ctx, append([]string{"/"}, newID().String()))

	require.Nil(t, a)
	require.ErrorIs(t, err, anchor.ErrInvalidPath)
}
