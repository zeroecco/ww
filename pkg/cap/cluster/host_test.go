package cluster_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	mx "github.com/wetware/matrix/pkg"
	"github.com/wetware/ww/pkg/cap/cluster"
	"github.com/wetware/ww/pkg/vat"
)

func TestHostWalk(t *testing.T) {
	t.Parallel()
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sim := mx.New(ctx)
	hs := sim.MustHostSet(ctx, 2)

	vat := vat.Network{NS: "test-host", Host: hs[0]}
	server := cluster.NewHostAnchorServer(vat)
	vat.Export(cluster.AnchorCapability, server)

	client := server.NewClient()

	a1, err := client.Walk(ctx, []string{"foo"})
	require.NoError(t, err)
	require.NotNil(t, a1)
	expectedPath := []string{hs[0].ID().String(), "foo"}
	require.Equal(t, expectedPath, a1.Path())
}

func TestHostLs(t *testing.T) {
	t.Parallel()
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sim := mx.New(ctx)
	hs := sim.MustHostSet(ctx, 2)

	vat := vat.Network{NS: "test-host", Host: hs[0]}
	server := cluster.NewHostAnchorServer(vat)
	vat.Export(cluster.AnchorCapability, server)

	client := server.NewClient()

	_, err := client.Walk(ctx, []string{"foo"})
	require.NoError(t, err)
	expectedPath := []string{hs[0].ID().String(), "foo"}

	it, err := client.Ls(ctx)
	require.NoError(t, err)
	require.NotNil(t, it)
	require.True(t, it.Next(ctx))
	require.Equal(t, expectedPath, it.Anchor().Path())
	require.False(t, it.Next(ctx))
	require.Nil(t, it.Err())
}