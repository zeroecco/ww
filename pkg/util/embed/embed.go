// Package embed provides utilities for embedding ww server nodes into applications.
package embed

import (
	"context"
	"time"

	ds "github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/pnet"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/lthibault/log"
	"go.uber.org/fx"

	"github.com/wetware/casm/pkg/cluster"
	"github.com/wetware/casm/pkg/cluster/pulse"
	"github.com/wetware/ww/pkg/cap"
	"github.com/wetware/ww/pkg/client"
	"github.com/wetware/ww/pkg/server"
)

// ServerConfig can be populated by Fx.
type ServerConfig struct {
	fx.In

	Logger      log.Logger           `optional:"true"`
	ListenAddrs []string             `optional:"true"`
	Topics      []string             `group:"topics"`
	NS          string               `optional:"true" name:"ns"`
	TTL         time.Duration        `optional:"true" name:"ttl"`
	Meta        pulse.Preparer       `optional:"true"`
	Secret      pnet.PSK             `optional:"true"`
	Cluster     discovery.Advertiser `optional:"true"`
	Boot        discovery.Discoverer `optional:"true"`
	Routing     cluster.RoutingTable `optional:"true"`
	Ready       pubsub.RouterReady   `optional:"true"`
	Datastore   ds.Batching          `optional:"true"`
}

// Server returns a fully configured 'server.Node', suitable for
// embedding in applications. The contents of 'cfg' is populated
// by Fx.
func Server(cfg ServerConfig) server.Node {
	return server.New(
		server.WithLogger(cfg.Logger),
		server.WithTopics(cfg.Topics...),
		server.WithHost(&server.RoutedHost{
			ListenAddrs: cfg.ListenAddrs,
			Secret:      cfg.Secret,
			// Auth:               ,  // TODO
			// PrivKey:            ,  // TODO
		}),
		server.WithDHT(server.DualDHTFactory{
			// ...
		}),
		server.WithPubSub(server.GossipsubFactory{
			Logger:     cfg.Logger,
			Advertiser: cfg.Cluster,
			Discoverer: cfg.Boot,
			Discovery: server.PexDiscovery{
				Logger:    cfg.Logger,
				NS:        cfg.NS,
				Cluster:   cfg.Cluster,
				Boot:      cfg.Boot,
				Datastore: cfg.Datastore,
			},
		}),
		server.WithCluster(server.ClusterConfig{
			NS:      cfg.NS,
			Log:     cfg.Logger,
			TTL:     cfg.TTL,
			Meta:    cfg.Meta,
			Routing: cfg.Routing,
			Ready:   cfg.Ready,
		}))
}

// DialConfig can be populated by Fx.
type DialConfig struct {
	fx.In

	Logger  log.Logger
	Join    discovery.Discoverer
	NS      string                `optional:"true" name:"ns"`
	Host    client.HostFactory    `optional:"true"`
	Routing client.RoutingFactory `optional:"true"`
	PubSub  client.PubSubFactory  `optional:"true"`
	Cap     cap.Dialer            `optional:"true"`
}

// Dialer returns a fully configured 'ClientDialer', suitable for
// embedding in applications.  The contents of 'cfg' is populated
// by Fx.
func Dialer(cfg DialConfig) ClientDialer { return ClientDialer(cfg) }

type ClientDialer DialConfig

func (d ClientDialer) Dial(ctx context.Context) (client.Node, error) {
	return client.DialDiscover(ctx, d.Join,
		client.WithLogger(d.Logger.WithField("ns", d.NS)),
		client.WithHost(d.Host),
		client.WithPubSub(d.PubSub),
		client.WithCapability(d.Cap),
		client.WithRouting(d.Routing),
		client.WithNamespace(d.NS))
}