package client

import (
	"context"
	"time"

	log "github.com/lthibault/log/pkg"
	"go.uber.org/fx"

	"github.com/ipfs/go-datastore"
	p2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/pnet"
	"github.com/libp2p/go-libp2p-core/routing"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/config"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"

	hostutil "github.com/lthibault/wetware/internal/util/host"
	ww "github.com/lthibault/wetware/pkg"
	"github.com/lthibault/wetware/pkg/discover"
	"github.com/lthibault/wetware/pkg/internal/eventloop"
)

func module(ctx context.Context, c *Client, opt []Option) fx.Option {
	return fx.Options(
		fx.NopLogger,
		fx.Supply(opt),
		fx.Provide(
			newCtx,
			userConfig,
			newRoutedHost,
			newTerminal,
			newPubsub,
			newClient,
		),
		fx.Populate(c),
		fx.Invoke(
			eventloop.DispatchNetwork,
			dialer(ctx),
		),
	)
}

type clientConfig struct {
	fx.In

	Log  log.Logger
	Host host.Host
	Term *terminal

	Namespace string `name:"ns"`
	PubSub    *pubsub.PubSub
}

func newClient(lx fx.Lifecycle, cfg clientConfig) Client {
	return Client{
		log:  cfg.Log.WithField("id", cfg.Host.ID()),
		term: cfg.Term,
		ps:   newTopicSet(cfg.Namespace, cfg.PubSub),
	}
}

type pubsubConfig struct {
	fx.In

	Ctx  context.Context
	Host host.Host
	DHT  routing.Routing
}

func newPubsub(lx fx.Lifecycle, cfg pubsubConfig) (*pubsub.PubSub, error) {
	return pubsub.NewGossipSub(
		cfg.Ctx,
		cfg.Host,
		pubsub.WithDiscovery(discovery.NewRoutingDiscovery(cfg.DHT)),
	)

}

func newTerminal(host host.Host) *terminal {
	return &terminal{
		local: host,
	}
}

type hostConfig struct {
	fx.In

	Ctx       context.Context
	Datastore datastore.Batching
	Secret    pnet.PSK
}

func (cfg hostConfig) options() []config.Option {
	return []config.Option{
		hostutil.MaybePrivate(cfg.Secret),
		p2p.Ping(false),
		p2p.NoListenAddrs, // also disables relay
		p2p.UserAgent(ww.ClientUAgent),
	}
}

type hostOut struct {
	fx.Out

	Host host.Host
	DHT  routing.Routing
}

func newRoutedHost(lx fx.Lifecycle, cfg hostConfig) (out hostOut, err error) {
	if out.Host, err = p2p.New(cfg.Ctx, cfg.options()...); err != nil {
		return
	}

	lx.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return out.Host.Close()
		},
	})

	out.DHT = dht.NewDHTClient(cfg.Ctx, out.Host, cfg.Datastore)
	out.Host = routedhost.Wrap(out.Host, out.DHT)
	return
}

func newCtx(lx fx.Lifecycle) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	lx.Append(fx.Hook{
		OnStop: func(context.Context) error {
			cancel()
			return nil
		},
	})

	return ctx
}

type userConfigOut struct {
	fx.Out

	Log       log.Logger
	Namespace string `name:"ns"`
	Secret    pnet.PSK

	Datastore datastore.Batching
	Discover  discover.Strategy
	Limit     int           `name:"discover_limit"`
	Timeout   time.Duration `name:"discover_timeout"`
}

func userConfig(opt []Option) (out userConfigOut, err error) {
	cfg := new(Config)
	for _, f := range withDefault(opt) {
		if err = f(cfg); err != nil {
			return
		}
	}

	out.Log = cfg.log.WithFields(log.F{
		"ns":   cfg.ns,
		"type": "client",
	})

	out.Namespace = cfg.ns
	out.Secret = cfg.psk
	out.Datastore = cfg.ds
	out.Discover = cfg.d
	out.Limit = cfg.queryLimit
	return
}
