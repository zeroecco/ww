package client

import (
	"context"
	"fmt"
	"path"
	"runtime"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	"github.com/lthibault/log"
	"github.com/wetware/casm/pkg/boot"
	rpcutil "github.com/wetware/ww/internal/util/rpc"
	"github.com/wetware/ww/pkg/cap"
)

type Dialer struct {
	join discovery.Discoverer

	ns      string
	log     log.Logger
	host    HostFactory
	routing RoutingFactory
	pubsub  PubSubFactory
	cap     cap.Dialer
}

func NewDialer(join discovery.Discoverer, opt ...Option) Dialer {
	d := Dialer{join: join}
	for _, option := range withDefault(opt) {
		option(&d)
	}
	return d
}

// Dial joins a cluster via 'addr', using the default Dialer.
func Dial(ctx context.Context, addr string, opt ...Option) (Node, error) {
	return DialDiscover(ctx, addrString(addr), opt...)
}

// DialDiscover joins a cluster via the supplied discovery service,
// using the default dialer.
func DialDiscover(ctx context.Context, d discovery.Discoverer, opt ...Option) (Node, error) {
	return NewDialer(d, opt...).Dial(ctx)
}

// Dial creates a client and connects it to a cluster.  The context
// can be safely cancelled when 'Dial' returns.
func (d Dialer) Dial(ctx context.Context) (n Node, err error) {
	// Libp2p often binds the lifecycle of various types to that of
	// the context passed into their respective constructors (e.g.
	// host.Host).  Throughout the Wetware ecosystem, we enforce the
	// idiom that contexts passed to constructors are used to abort
	// construction, NOT shutdown the resulting type.
	//
	// This deferred call to 'cancel' is a guard against passing the
	// dial context to a constructor that expects long-lived context.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if rf, ok := d.host.(RoutingHook); ok {
		d.routing = &routingHook{RoutingFactory: d.routing}
		rf.SetRouting(d.routing)
	}

	var (
		h    host.Host
		ps   PubSub
		r    routing.Routing
		conn *rpc.Conn
	)

	h, err = d.host.New(context.Background())
	if err != nil {
		return n, fmt.Errorf("host: %w", err)
	}

	// Ensure we do not leak resources (e.g. bind addresses) if any
	// of our subsequent factories fail.
	//
	// Note that a 'HostFactory' implementation could potentially
	// return a 'Host' instsance that was previously created and is
	// still in use elsewhere. Under such conditions, closing would
	// be an error.
	//
	// Routing, PubSub and Conn are all tied to the lifetime of the
	// 'Host' instance.  They need not be explicitly closed.
	runtime.SetFinalizer(&h, func(h *host.Host) { _ = (*h).Close() })

	r, err = d.routing.New(h)
	if err != nil {
		return n, fmt.Errorf("dht: %w", err)
	}

	ps, err = d.pubsub.New(h, n.Routing)
	if err != nil {
		return n, fmt.Errorf("pubsub: %w", err)
	}

	c, err := d.cap.Dial(ctx, cap.NewBinder(func(s network.Stream) (*capnp.Client, error) {
		if conn, err = d.bind(ctx, s); err != nil {
			return nil, err
		}

		return conn.Bootstrap(ctx), nil
	}))

	return Node{
		PubSub:  ps,
		Routing: r,
		Conn:    conn,
		c:       c,
	}, nil
}

func (d Dialer) bind(ctx context.Context, s network.Stream) (*rpc.Conn, error) {
	return rpc.NewConn(d.transportFor(s), &rpc.Options{
		ErrorReporter: d.reporter(),
	}), nil
}

func (d Dialer) transportFor(s network.Stream) rpc.Transport {
	if path.Base(string(s.Protocol())) == "packed" {
		return rpc.NewPackedStreamTransport(s)
	}

	return rpc.NewStreamTransport(s)
}

func (d Dialer) reporter() rpcutil.ErrReporterFunc {
	log := d.log.WithField("ns", d.ns)
	return func(err error) {
		log.WithError(err).Debug("rpc error")
	}
}

type addrString string

func (addr addrString) FindPeers(ctx context.Context, ns string, opt ...discovery.Option) (<-chan peer.AddrInfo, error) {
	info, err := peer.AddrInfoFromString(string(addr))
	if err != nil {
		return nil, fmt.Errorf("addr: %w", err)
	}

	return boot.StaticAddrs{*info}.FindPeers(ctx, ns, opt...)
}