package pubsub

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/server"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/lthibault/log"
	ctxutil "github.com/lthibault/util/ctx"
	api "github.com/wetware/ww/internal/api/pubsub"
)

var ErrClosed = errors.New("closed")

var defaultPolicy = server.Policy{
	MaxConcurrentCalls: 64,
}

type TopicJoiner interface {
	Join(string, ...pubsub.TopicOpt) (*pubsub.Topic, error)
}

// Provider wraps a PubSub and provides vat.ClientProvider.
//
// In order to export a given topic through multiple capabilities,
// Provider tracks existing topics internally.  See 'Join' for more details.
type Provider struct {
	cq  chan struct{}
	log log.Logger

	ps TopicJoiner

	mu sync.RWMutex
	wg sync.WaitGroup // blocks shutdown until all tasks are released
	ts map[string]*refCountedTopic
}

func New(ns string, ps TopicJoiner, opt ...Option) *Provider {
	var f = &Provider{
		cq: make(chan struct{}),
		ps: ps,
		ts: make(map[string]*refCountedTopic),
	}

	for _, option := range withDefault(opt) {
		option(f)
	}

	return f
}

func (p *Provider) Close() (err error) {
	if p != nil {
		select {
		case <-p.cq:
			err = fmt.Errorf("already %w", ErrClosed)
		default:
			close(p.cq)
			p.wg.Wait()
		}
	}

	return
}

func (p *Provider) Client() *capnp.Client {
	return api.PubSub_ServerToClient(p, &defaultPolicy).Client
}

func (p *Provider) Join(ctx context.Context, call api.PubSub_join) error {
	call.Ack()

	name, err := call.Args().Name()
	if err != nil {
		return err
	}

	t, err := p.getOrCreate(name)
	if err != nil {
		return err
	}

	res, err := call.AllocResults()
	if err != nil {
		return err
	}

	return res.SetTopic(api.Topic_ServerToClient(t, &defaultPolicy))
}

func (p *Provider) getOrCreate(topic string) (*refCountedTopic, error) {
	p.mu.RLock()

	// fast path - already exists?
	if t, ok := p.ts[topic]; ok {
		defer p.mu.RUnlock()
		return t.AddRef(), nil
	}

	// slow path
	p.mu.RUnlock()
	p.mu.Lock()
	defer p.mu.Unlock()

	// topic may have been added while acquiring write-lock
	if t, ok := p.ts[topic]; ok {
		return t.AddRef(), nil
	}

	// join topic
	return p.joinTopic(topic)
}

// joinTopic and assign a refcounted topic to tm.ts.  Callers MUST hold a
// write-lock on f.mu.
func (p *Provider) joinTopic(topic string) (*refCountedTopic, error) {
	t, err := p.ps.Join(topic)
	if err != nil {
		return nil, err
	}

	p.wg.Add(1)
	release := func() {
		defer p.wg.Done()

		p.mu.Lock()
		defer p.mu.Unlock()

		delete(p.ts, topic)

		if err := t.Close(); err != nil {
			p.log.
				WithError(err).
				Errorf("unable to close topic %s", topic)
		}
	}

	rt := &refCountedTopic{
		log:     p.log.WithField("topic", topic),
		ctx:     ctxutil.C(p.cq),
		topic:   t,
		ref:     1,
		release: release,
	}

	p.ts[topic] = rt
	return rt, nil
}

type refCountedTopic struct {
	ctx   context.Context // root context for subscriptions
	log   log.Logger
	topic *pubsub.Topic

	mu  sync.Mutex
	ref int // number of refs from capnp.Client instances

	release capnp.ReleaseFunc // caller MUST hold mu
}

// AddRef MUST be called each time a new capnp.Client is
// created for t.
func (t *refCountedTopic) AddRef() *refCountedTopic {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.ref++
	return t
}

func (t *refCountedTopic) Release() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.ref--; t.ref == 0 {
		t.release()
	}
}

// The refCountedTopic is unique for each *pubsub.Topic, and is
// therefore shared across multiple capnp.Client instances. For
// this reason, Shutdown MAY be called multiple times.
func (t *refCountedTopic) Shutdown() { t.Release() }

func (t *refCountedTopic) Publish(ctx context.Context, call api.Topic_publish) error {
	if t.ctx.Err() != nil {
		return ErrClosed
	}

	b, err := call.Args().Msg()
	if err == nil {
		err = t.topic.Publish(ctx, b)
	}

	return err
}

func (t *refCountedTopic) Subscribe(_ context.Context, call api.Topic_subscribe) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	sub, err := t.topic.Subscribe()
	if err != nil {
		return err
	}

	if t.ref == 0 {
		return ErrClosed
	}

	t.ref++
	go t.handle(sub, call.Args().Handler().AddRef())

	return nil
}

func (t *refCountedTopic) handle(sub *pubsub.Subscription, h api.Topic_Handler) {
	defer t.Release()
	defer sub.Cancel()
	defer h.Release()

	for {
		m, err := sub.Next(t.ctx)
		if err != nil {
			return
		}

		if t.send(h, m) != nil {
			return
		}
	}
}

func (t *refCountedTopic) send(h api.Topic_Handler, m *pubsub.Message) error {
	f, release := h.Handle(t.ctx, message(m))
	defer release()

	_, err := f.Struct()
	return err
}

func message(m *pubsub.Message) func(api.Topic_Handler_handle_Params) error {
	return func(ps api.Topic_Handler_handle_Params) error {
		return ps.SetMsg(m.Data)
	}
}
