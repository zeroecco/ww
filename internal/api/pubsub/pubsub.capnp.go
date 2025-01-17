// Code generated by capnpc-go. DO NOT EDIT.

package pubsub

import (
	capnp "capnproto.org/go/capnp/v3"
	text "capnproto.org/go/capnp/v3/encoding/text"
	schemas "capnproto.org/go/capnp/v3/schemas"
	server "capnproto.org/go/capnp/v3/server"
	context "context"
)

type Topic struct{ Client *capnp.Client }

// Topic_TypeID is the unique identifier for the type Topic.
const Topic_TypeID = 0x986ea9282f106bb0

func (c Topic) Publish(ctx context.Context, params func(Topic_publish_Params) error) (Topic_publish_Results_Future, capnp.ReleaseFunc) {
	s := capnp.Send{
		Method: capnp.Method{
			InterfaceID:   0x986ea9282f106bb0,
			MethodID:      0,
			InterfaceName: "pubsub.capnp:Topic",
			MethodName:    "publish",
		},
	}
	if params != nil {
		s.ArgsSize = capnp.ObjectSize{DataSize: 0, PointerCount: 1}
		s.PlaceArgs = func(s capnp.Struct) error { return params(Topic_publish_Params{Struct: s}) }
	}
	ans, release := c.Client.SendCall(ctx, s)
	return Topic_publish_Results_Future{Future: ans.Future()}, release
}
func (c Topic) Subscribe(ctx context.Context, params func(Topic_subscribe_Params) error) (Topic_subscribe_Results_Future, capnp.ReleaseFunc) {
	s := capnp.Send{
		Method: capnp.Method{
			InterfaceID:   0x986ea9282f106bb0,
			MethodID:      1,
			InterfaceName: "pubsub.capnp:Topic",
			MethodName:    "subscribe",
		},
	}
	if params != nil {
		s.ArgsSize = capnp.ObjectSize{DataSize: 0, PointerCount: 1}
		s.PlaceArgs = func(s capnp.Struct) error { return params(Topic_subscribe_Params{Struct: s}) }
	}
	ans, release := c.Client.SendCall(ctx, s)
	return Topic_subscribe_Results_Future{Future: ans.Future()}, release
}

func (c Topic) AddRef() Topic {
	return Topic{
		Client: c.Client.AddRef(),
	}
}

func (c Topic) Release() {
	c.Client.Release()
}

// A Topic_Server is a Topic with a local implementation.
type Topic_Server interface {
	Publish(context.Context, Topic_publish) error

	Subscribe(context.Context, Topic_subscribe) error
}

// Topic_NewServer creates a new Server from an implementation of Topic_Server.
func Topic_NewServer(s Topic_Server, policy *server.Policy) *server.Server {
	c, _ := s.(server.Shutdowner)
	return server.New(Topic_Methods(nil, s), s, c, policy)
}

// Topic_ServerToClient creates a new Client from an implementation of Topic_Server.
// The caller is responsible for calling Release on the returned Client.
func Topic_ServerToClient(s Topic_Server, policy *server.Policy) Topic {
	return Topic{Client: capnp.NewClient(Topic_NewServer(s, policy))}
}

// Topic_Methods appends Methods to a slice that invoke the methods on s.
// This can be used to create a more complicated Server.
func Topic_Methods(methods []server.Method, s Topic_Server) []server.Method {
	if cap(methods) == 0 {
		methods = make([]server.Method, 0, 2)
	}

	methods = append(methods, server.Method{
		Method: capnp.Method{
			InterfaceID:   0x986ea9282f106bb0,
			MethodID:      0,
			InterfaceName: "pubsub.capnp:Topic",
			MethodName:    "publish",
		},
		Impl: func(ctx context.Context, call *server.Call) error {
			return s.Publish(ctx, Topic_publish{call})
		},
	})

	methods = append(methods, server.Method{
		Method: capnp.Method{
			InterfaceID:   0x986ea9282f106bb0,
			MethodID:      1,
			InterfaceName: "pubsub.capnp:Topic",
			MethodName:    "subscribe",
		},
		Impl: func(ctx context.Context, call *server.Call) error {
			return s.Subscribe(ctx, Topic_subscribe{call})
		},
	})

	return methods
}

// Topic_publish holds the state for a server call to Topic.publish.
// See server.Call for documentation.
type Topic_publish struct {
	*server.Call
}

// Args returns the call's arguments.
func (c Topic_publish) Args() Topic_publish_Params {
	return Topic_publish_Params{Struct: c.Call.Args()}
}

// AllocResults allocates the results struct.
func (c Topic_publish) AllocResults() (Topic_publish_Results, error) {
	r, err := c.Call.AllocResults(capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Topic_publish_Results{Struct: r}, err
}

// Topic_subscribe holds the state for a server call to Topic.subscribe.
// See server.Call for documentation.
type Topic_subscribe struct {
	*server.Call
}

// Args returns the call's arguments.
func (c Topic_subscribe) Args() Topic_subscribe_Params {
	return Topic_subscribe_Params{Struct: c.Call.Args()}
}

// AllocResults allocates the results struct.
func (c Topic_subscribe) AllocResults() (Topic_subscribe_Results, error) {
	r, err := c.Call.AllocResults(capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Topic_subscribe_Results{Struct: r}, err
}

type Topic_Handler struct{ Client *capnp.Client }

// Topic_Handler_TypeID is the unique identifier for the type Topic_Handler.
const Topic_Handler_TypeID = 0xd19c472616f2c6fb

func (c Topic_Handler) Handle(ctx context.Context, params func(Topic_Handler_handle_Params) error) (Topic_Handler_handle_Results_Future, capnp.ReleaseFunc) {
	s := capnp.Send{
		Method: capnp.Method{
			InterfaceID:   0xd19c472616f2c6fb,
			MethodID:      0,
			InterfaceName: "pubsub.capnp:Topic.Handler",
			MethodName:    "handle",
		},
	}
	if params != nil {
		s.ArgsSize = capnp.ObjectSize{DataSize: 0, PointerCount: 1}
		s.PlaceArgs = func(s capnp.Struct) error { return params(Topic_Handler_handle_Params{Struct: s}) }
	}
	ans, release := c.Client.SendCall(ctx, s)
	return Topic_Handler_handle_Results_Future{Future: ans.Future()}, release
}

func (c Topic_Handler) AddRef() Topic_Handler {
	return Topic_Handler{
		Client: c.Client.AddRef(),
	}
}

func (c Topic_Handler) Release() {
	c.Client.Release()
}

// A Topic_Handler_Server is a Topic_Handler with a local implementation.
type Topic_Handler_Server interface {
	Handle(context.Context, Topic_Handler_handle) error
}

// Topic_Handler_NewServer creates a new Server from an implementation of Topic_Handler_Server.
func Topic_Handler_NewServer(s Topic_Handler_Server, policy *server.Policy) *server.Server {
	c, _ := s.(server.Shutdowner)
	return server.New(Topic_Handler_Methods(nil, s), s, c, policy)
}

// Topic_Handler_ServerToClient creates a new Client from an implementation of Topic_Handler_Server.
// The caller is responsible for calling Release on the returned Client.
func Topic_Handler_ServerToClient(s Topic_Handler_Server, policy *server.Policy) Topic_Handler {
	return Topic_Handler{Client: capnp.NewClient(Topic_Handler_NewServer(s, policy))}
}

// Topic_Handler_Methods appends Methods to a slice that invoke the methods on s.
// This can be used to create a more complicated Server.
func Topic_Handler_Methods(methods []server.Method, s Topic_Handler_Server) []server.Method {
	if cap(methods) == 0 {
		methods = make([]server.Method, 0, 1)
	}

	methods = append(methods, server.Method{
		Method: capnp.Method{
			InterfaceID:   0xd19c472616f2c6fb,
			MethodID:      0,
			InterfaceName: "pubsub.capnp:Topic.Handler",
			MethodName:    "handle",
		},
		Impl: func(ctx context.Context, call *server.Call) error {
			return s.Handle(ctx, Topic_Handler_handle{call})
		},
	})

	return methods
}

// Topic_Handler_handle holds the state for a server call to Topic_Handler.handle.
// See server.Call for documentation.
type Topic_Handler_handle struct {
	*server.Call
}

// Args returns the call's arguments.
func (c Topic_Handler_handle) Args() Topic_Handler_handle_Params {
	return Topic_Handler_handle_Params{Struct: c.Call.Args()}
}

// AllocResults allocates the results struct.
func (c Topic_Handler_handle) AllocResults() (Topic_Handler_handle_Results, error) {
	r, err := c.Call.AllocResults(capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Topic_Handler_handle_Results{Struct: r}, err
}

type Topic_Handler_handle_Params struct{ capnp.Struct }

// Topic_Handler_handle_Params_TypeID is the unique identifier for the type Topic_Handler_handle_Params.
const Topic_Handler_handle_Params_TypeID = 0x89d849f1a30adbf2

func NewTopic_Handler_handle_Params(s *capnp.Segment) (Topic_Handler_handle_Params, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Topic_Handler_handle_Params{st}, err
}

func NewRootTopic_Handler_handle_Params(s *capnp.Segment) (Topic_Handler_handle_Params, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Topic_Handler_handle_Params{st}, err
}

func ReadRootTopic_Handler_handle_Params(msg *capnp.Message) (Topic_Handler_handle_Params, error) {
	root, err := msg.Root()
	return Topic_Handler_handle_Params{root.Struct()}, err
}

func (s Topic_Handler_handle_Params) String() string {
	str, _ := text.Marshal(0x89d849f1a30adbf2, s.Struct)
	return str
}

func (s Topic_Handler_handle_Params) Msg() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return []byte(p.Data()), err
}

func (s Topic_Handler_handle_Params) HasMsg() bool {
	return s.Struct.HasPtr(0)
}

func (s Topic_Handler_handle_Params) SetMsg(v []byte) error {
	return s.Struct.SetData(0, v)
}

// Topic_Handler_handle_Params_List is a list of Topic_Handler_handle_Params.
type Topic_Handler_handle_Params_List struct{ capnp.List }

// NewTopic_Handler_handle_Params creates a new list of Topic_Handler_handle_Params.
func NewTopic_Handler_handle_Params_List(s *capnp.Segment, sz int32) (Topic_Handler_handle_Params_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return Topic_Handler_handle_Params_List{l}, err
}

func (s Topic_Handler_handle_Params_List) At(i int) Topic_Handler_handle_Params {
	return Topic_Handler_handle_Params{s.List.Struct(i)}
}

func (s Topic_Handler_handle_Params_List) Set(i int, v Topic_Handler_handle_Params) error {
	return s.List.SetStruct(i, v.Struct)
}

func (s Topic_Handler_handle_Params_List) String() string {
	str, _ := text.MarshalList(0x89d849f1a30adbf2, s.List)
	return str
}

// Topic_Handler_handle_Params_Future is a wrapper for a Topic_Handler_handle_Params promised by a client call.
type Topic_Handler_handle_Params_Future struct{ *capnp.Future }

func (p Topic_Handler_handle_Params_Future) Struct() (Topic_Handler_handle_Params, error) {
	s, err := p.Future.Struct()
	return Topic_Handler_handle_Params{s}, err
}

type Topic_Handler_handle_Results struct{ capnp.Struct }

// Topic_Handler_handle_Results_TypeID is the unique identifier for the type Topic_Handler_handle_Results.
const Topic_Handler_handle_Results_TypeID = 0xf8d41329eb57bd62

func NewTopic_Handler_handle_Results(s *capnp.Segment) (Topic_Handler_handle_Results, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Topic_Handler_handle_Results{st}, err
}

func NewRootTopic_Handler_handle_Results(s *capnp.Segment) (Topic_Handler_handle_Results, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Topic_Handler_handle_Results{st}, err
}

func ReadRootTopic_Handler_handle_Results(msg *capnp.Message) (Topic_Handler_handle_Results, error) {
	root, err := msg.Root()
	return Topic_Handler_handle_Results{root.Struct()}, err
}

func (s Topic_Handler_handle_Results) String() string {
	str, _ := text.Marshal(0xf8d41329eb57bd62, s.Struct)
	return str
}

// Topic_Handler_handle_Results_List is a list of Topic_Handler_handle_Results.
type Topic_Handler_handle_Results_List struct{ capnp.List }

// NewTopic_Handler_handle_Results creates a new list of Topic_Handler_handle_Results.
func NewTopic_Handler_handle_Results_List(s *capnp.Segment, sz int32) (Topic_Handler_handle_Results_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0}, sz)
	return Topic_Handler_handle_Results_List{l}, err
}

func (s Topic_Handler_handle_Results_List) At(i int) Topic_Handler_handle_Results {
	return Topic_Handler_handle_Results{s.List.Struct(i)}
}

func (s Topic_Handler_handle_Results_List) Set(i int, v Topic_Handler_handle_Results) error {
	return s.List.SetStruct(i, v.Struct)
}

func (s Topic_Handler_handle_Results_List) String() string {
	str, _ := text.MarshalList(0xf8d41329eb57bd62, s.List)
	return str
}

// Topic_Handler_handle_Results_Future is a wrapper for a Topic_Handler_handle_Results promised by a client call.
type Topic_Handler_handle_Results_Future struct{ *capnp.Future }

func (p Topic_Handler_handle_Results_Future) Struct() (Topic_Handler_handle_Results, error) {
	s, err := p.Future.Struct()
	return Topic_Handler_handle_Results{s}, err
}

type Topic_publish_Params struct{ capnp.Struct }

// Topic_publish_Params_TypeID is the unique identifier for the type Topic_publish_Params.
const Topic_publish_Params_TypeID = 0x8810938879cb8443

func NewTopic_publish_Params(s *capnp.Segment) (Topic_publish_Params, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Topic_publish_Params{st}, err
}

func NewRootTopic_publish_Params(s *capnp.Segment) (Topic_publish_Params, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Topic_publish_Params{st}, err
}

func ReadRootTopic_publish_Params(msg *capnp.Message) (Topic_publish_Params, error) {
	root, err := msg.Root()
	return Topic_publish_Params{root.Struct()}, err
}

func (s Topic_publish_Params) String() string {
	str, _ := text.Marshal(0x8810938879cb8443, s.Struct)
	return str
}

func (s Topic_publish_Params) Msg() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return []byte(p.Data()), err
}

func (s Topic_publish_Params) HasMsg() bool {
	return s.Struct.HasPtr(0)
}

func (s Topic_publish_Params) SetMsg(v []byte) error {
	return s.Struct.SetData(0, v)
}

// Topic_publish_Params_List is a list of Topic_publish_Params.
type Topic_publish_Params_List struct{ capnp.List }

// NewTopic_publish_Params creates a new list of Topic_publish_Params.
func NewTopic_publish_Params_List(s *capnp.Segment, sz int32) (Topic_publish_Params_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return Topic_publish_Params_List{l}, err
}

func (s Topic_publish_Params_List) At(i int) Topic_publish_Params {
	return Topic_publish_Params{s.List.Struct(i)}
}

func (s Topic_publish_Params_List) Set(i int, v Topic_publish_Params) error {
	return s.List.SetStruct(i, v.Struct)
}

func (s Topic_publish_Params_List) String() string {
	str, _ := text.MarshalList(0x8810938879cb8443, s.List)
	return str
}

// Topic_publish_Params_Future is a wrapper for a Topic_publish_Params promised by a client call.
type Topic_publish_Params_Future struct{ *capnp.Future }

func (p Topic_publish_Params_Future) Struct() (Topic_publish_Params, error) {
	s, err := p.Future.Struct()
	return Topic_publish_Params{s}, err
}

type Topic_publish_Results struct{ capnp.Struct }

// Topic_publish_Results_TypeID is the unique identifier for the type Topic_publish_Results.
const Topic_publish_Results_TypeID = 0x9d3775c65b79b54c

func NewTopic_publish_Results(s *capnp.Segment) (Topic_publish_Results, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Topic_publish_Results{st}, err
}

func NewRootTopic_publish_Results(s *capnp.Segment) (Topic_publish_Results, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Topic_publish_Results{st}, err
}

func ReadRootTopic_publish_Results(msg *capnp.Message) (Topic_publish_Results, error) {
	root, err := msg.Root()
	return Topic_publish_Results{root.Struct()}, err
}

func (s Topic_publish_Results) String() string {
	str, _ := text.Marshal(0x9d3775c65b79b54c, s.Struct)
	return str
}

// Topic_publish_Results_List is a list of Topic_publish_Results.
type Topic_publish_Results_List struct{ capnp.List }

// NewTopic_publish_Results creates a new list of Topic_publish_Results.
func NewTopic_publish_Results_List(s *capnp.Segment, sz int32) (Topic_publish_Results_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0}, sz)
	return Topic_publish_Results_List{l}, err
}

func (s Topic_publish_Results_List) At(i int) Topic_publish_Results {
	return Topic_publish_Results{s.List.Struct(i)}
}

func (s Topic_publish_Results_List) Set(i int, v Topic_publish_Results) error {
	return s.List.SetStruct(i, v.Struct)
}

func (s Topic_publish_Results_List) String() string {
	str, _ := text.MarshalList(0x9d3775c65b79b54c, s.List)
	return str
}

// Topic_publish_Results_Future is a wrapper for a Topic_publish_Results promised by a client call.
type Topic_publish_Results_Future struct{ *capnp.Future }

func (p Topic_publish_Results_Future) Struct() (Topic_publish_Results, error) {
	s, err := p.Future.Struct()
	return Topic_publish_Results{s}, err
}

type Topic_subscribe_Params struct{ capnp.Struct }

// Topic_subscribe_Params_TypeID is the unique identifier for the type Topic_subscribe_Params.
const Topic_subscribe_Params_TypeID = 0xc772c6756fef5ba8

func NewTopic_subscribe_Params(s *capnp.Segment) (Topic_subscribe_Params, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Topic_subscribe_Params{st}, err
}

func NewRootTopic_subscribe_Params(s *capnp.Segment) (Topic_subscribe_Params, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Topic_subscribe_Params{st}, err
}

func ReadRootTopic_subscribe_Params(msg *capnp.Message) (Topic_subscribe_Params, error) {
	root, err := msg.Root()
	return Topic_subscribe_Params{root.Struct()}, err
}

func (s Topic_subscribe_Params) String() string {
	str, _ := text.Marshal(0xc772c6756fef5ba8, s.Struct)
	return str
}

func (s Topic_subscribe_Params) Handler() Topic_Handler {
	p, _ := s.Struct.Ptr(0)
	return Topic_Handler{Client: p.Interface().Client()}
}

func (s Topic_subscribe_Params) HasHandler() bool {
	return s.Struct.HasPtr(0)
}

func (s Topic_subscribe_Params) SetHandler(v Topic_Handler) error {
	if !v.Client.IsValid() {
		return s.Struct.SetPtr(0, capnp.Ptr{})
	}
	seg := s.Segment()
	in := capnp.NewInterface(seg, seg.Message().AddCap(v.Client))
	return s.Struct.SetPtr(0, in.ToPtr())
}

// Topic_subscribe_Params_List is a list of Topic_subscribe_Params.
type Topic_subscribe_Params_List struct{ capnp.List }

// NewTopic_subscribe_Params creates a new list of Topic_subscribe_Params.
func NewTopic_subscribe_Params_List(s *capnp.Segment, sz int32) (Topic_subscribe_Params_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return Topic_subscribe_Params_List{l}, err
}

func (s Topic_subscribe_Params_List) At(i int) Topic_subscribe_Params {
	return Topic_subscribe_Params{s.List.Struct(i)}
}

func (s Topic_subscribe_Params_List) Set(i int, v Topic_subscribe_Params) error {
	return s.List.SetStruct(i, v.Struct)
}

func (s Topic_subscribe_Params_List) String() string {
	str, _ := text.MarshalList(0xc772c6756fef5ba8, s.List)
	return str
}

// Topic_subscribe_Params_Future is a wrapper for a Topic_subscribe_Params promised by a client call.
type Topic_subscribe_Params_Future struct{ *capnp.Future }

func (p Topic_subscribe_Params_Future) Struct() (Topic_subscribe_Params, error) {
	s, err := p.Future.Struct()
	return Topic_subscribe_Params{s}, err
}

func (p Topic_subscribe_Params_Future) Handler() Topic_Handler {
	return Topic_Handler{Client: p.Future.Field(0, nil).Client()}
}

type Topic_subscribe_Results struct{ capnp.Struct }

// Topic_subscribe_Results_TypeID is the unique identifier for the type Topic_subscribe_Results.
const Topic_subscribe_Results_TypeID = 0x8470369ac91fcc32

func NewTopic_subscribe_Results(s *capnp.Segment) (Topic_subscribe_Results, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Topic_subscribe_Results{st}, err
}

func NewRootTopic_subscribe_Results(s *capnp.Segment) (Topic_subscribe_Results, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return Topic_subscribe_Results{st}, err
}

func ReadRootTopic_subscribe_Results(msg *capnp.Message) (Topic_subscribe_Results, error) {
	root, err := msg.Root()
	return Topic_subscribe_Results{root.Struct()}, err
}

func (s Topic_subscribe_Results) String() string {
	str, _ := text.Marshal(0x8470369ac91fcc32, s.Struct)
	return str
}

// Topic_subscribe_Results_List is a list of Topic_subscribe_Results.
type Topic_subscribe_Results_List struct{ capnp.List }

// NewTopic_subscribe_Results creates a new list of Topic_subscribe_Results.
func NewTopic_subscribe_Results_List(s *capnp.Segment, sz int32) (Topic_subscribe_Results_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0}, sz)
	return Topic_subscribe_Results_List{l}, err
}

func (s Topic_subscribe_Results_List) At(i int) Topic_subscribe_Results {
	return Topic_subscribe_Results{s.List.Struct(i)}
}

func (s Topic_subscribe_Results_List) Set(i int, v Topic_subscribe_Results) error {
	return s.List.SetStruct(i, v.Struct)
}

func (s Topic_subscribe_Results_List) String() string {
	str, _ := text.MarshalList(0x8470369ac91fcc32, s.List)
	return str
}

// Topic_subscribe_Results_Future is a wrapper for a Topic_subscribe_Results promised by a client call.
type Topic_subscribe_Results_Future struct{ *capnp.Future }

func (p Topic_subscribe_Results_Future) Struct() (Topic_subscribe_Results, error) {
	s, err := p.Future.Struct()
	return Topic_subscribe_Results{s}, err
}

type PubSub struct{ Client *capnp.Client }

// PubSub_TypeID is the unique identifier for the type PubSub.
const PubSub_TypeID = 0xf1cc149f1c06e50e

func (c PubSub) Join(ctx context.Context, params func(PubSub_join_Params) error) (PubSub_join_Results_Future, capnp.ReleaseFunc) {
	s := capnp.Send{
		Method: capnp.Method{
			InterfaceID:   0xf1cc149f1c06e50e,
			MethodID:      0,
			InterfaceName: "pubsub.capnp:PubSub",
			MethodName:    "join",
		},
	}
	if params != nil {
		s.ArgsSize = capnp.ObjectSize{DataSize: 0, PointerCount: 1}
		s.PlaceArgs = func(s capnp.Struct) error { return params(PubSub_join_Params{Struct: s}) }
	}
	ans, release := c.Client.SendCall(ctx, s)
	return PubSub_join_Results_Future{Future: ans.Future()}, release
}

func (c PubSub) AddRef() PubSub {
	return PubSub{
		Client: c.Client.AddRef(),
	}
}

func (c PubSub) Release() {
	c.Client.Release()
}

// A PubSub_Server is a PubSub with a local implementation.
type PubSub_Server interface {
	Join(context.Context, PubSub_join) error
}

// PubSub_NewServer creates a new Server from an implementation of PubSub_Server.
func PubSub_NewServer(s PubSub_Server, policy *server.Policy) *server.Server {
	c, _ := s.(server.Shutdowner)
	return server.New(PubSub_Methods(nil, s), s, c, policy)
}

// PubSub_ServerToClient creates a new Client from an implementation of PubSub_Server.
// The caller is responsible for calling Release on the returned Client.
func PubSub_ServerToClient(s PubSub_Server, policy *server.Policy) PubSub {
	return PubSub{Client: capnp.NewClient(PubSub_NewServer(s, policy))}
}

// PubSub_Methods appends Methods to a slice that invoke the methods on s.
// This can be used to create a more complicated Server.
func PubSub_Methods(methods []server.Method, s PubSub_Server) []server.Method {
	if cap(methods) == 0 {
		methods = make([]server.Method, 0, 1)
	}

	methods = append(methods, server.Method{
		Method: capnp.Method{
			InterfaceID:   0xf1cc149f1c06e50e,
			MethodID:      0,
			InterfaceName: "pubsub.capnp:PubSub",
			MethodName:    "join",
		},
		Impl: func(ctx context.Context, call *server.Call) error {
			return s.Join(ctx, PubSub_join{call})
		},
	})

	return methods
}

// PubSub_join holds the state for a server call to PubSub.join.
// See server.Call for documentation.
type PubSub_join struct {
	*server.Call
}

// Args returns the call's arguments.
func (c PubSub_join) Args() PubSub_join_Params {
	return PubSub_join_Params{Struct: c.Call.Args()}
}

// AllocResults allocates the results struct.
func (c PubSub_join) AllocResults() (PubSub_join_Results, error) {
	r, err := c.Call.AllocResults(capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return PubSub_join_Results{Struct: r}, err
}

type PubSub_join_Params struct{ capnp.Struct }

// PubSub_join_Params_TypeID is the unique identifier for the type PubSub_join_Params.
const PubSub_join_Params_TypeID = 0xfb4016d002794da7

func NewPubSub_join_Params(s *capnp.Segment) (PubSub_join_Params, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return PubSub_join_Params{st}, err
}

func NewRootPubSub_join_Params(s *capnp.Segment) (PubSub_join_Params, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return PubSub_join_Params{st}, err
}

func ReadRootPubSub_join_Params(msg *capnp.Message) (PubSub_join_Params, error) {
	root, err := msg.Root()
	return PubSub_join_Params{root.Struct()}, err
}

func (s PubSub_join_Params) String() string {
	str, _ := text.Marshal(0xfb4016d002794da7, s.Struct)
	return str
}

func (s PubSub_join_Params) Name() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.Text(), err
}

func (s PubSub_join_Params) HasName() bool {
	return s.Struct.HasPtr(0)
}

func (s PubSub_join_Params) NameBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytes(), err
}

func (s PubSub_join_Params) SetName(v string) error {
	return s.Struct.SetText(0, v)
}

// PubSub_join_Params_List is a list of PubSub_join_Params.
type PubSub_join_Params_List struct{ capnp.List }

// NewPubSub_join_Params creates a new list of PubSub_join_Params.
func NewPubSub_join_Params_List(s *capnp.Segment, sz int32) (PubSub_join_Params_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return PubSub_join_Params_List{l}, err
}

func (s PubSub_join_Params_List) At(i int) PubSub_join_Params {
	return PubSub_join_Params{s.List.Struct(i)}
}

func (s PubSub_join_Params_List) Set(i int, v PubSub_join_Params) error {
	return s.List.SetStruct(i, v.Struct)
}

func (s PubSub_join_Params_List) String() string {
	str, _ := text.MarshalList(0xfb4016d002794da7, s.List)
	return str
}

// PubSub_join_Params_Future is a wrapper for a PubSub_join_Params promised by a client call.
type PubSub_join_Params_Future struct{ *capnp.Future }

func (p PubSub_join_Params_Future) Struct() (PubSub_join_Params, error) {
	s, err := p.Future.Struct()
	return PubSub_join_Params{s}, err
}

type PubSub_join_Results struct{ capnp.Struct }

// PubSub_join_Results_TypeID is the unique identifier for the type PubSub_join_Results.
const PubSub_join_Results_TypeID = 0x9f6c50fbc67b1d88

func NewPubSub_join_Results(s *capnp.Segment) (PubSub_join_Results, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return PubSub_join_Results{st}, err
}

func NewRootPubSub_join_Results(s *capnp.Segment) (PubSub_join_Results, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return PubSub_join_Results{st}, err
}

func ReadRootPubSub_join_Results(msg *capnp.Message) (PubSub_join_Results, error) {
	root, err := msg.Root()
	return PubSub_join_Results{root.Struct()}, err
}

func (s PubSub_join_Results) String() string {
	str, _ := text.Marshal(0x9f6c50fbc67b1d88, s.Struct)
	return str
}

func (s PubSub_join_Results) Topic() Topic {
	p, _ := s.Struct.Ptr(0)
	return Topic{Client: p.Interface().Client()}
}

func (s PubSub_join_Results) HasTopic() bool {
	return s.Struct.HasPtr(0)
}

func (s PubSub_join_Results) SetTopic(v Topic) error {
	if !v.Client.IsValid() {
		return s.Struct.SetPtr(0, capnp.Ptr{})
	}
	seg := s.Segment()
	in := capnp.NewInterface(seg, seg.Message().AddCap(v.Client))
	return s.Struct.SetPtr(0, in.ToPtr())
}

// PubSub_join_Results_List is a list of PubSub_join_Results.
type PubSub_join_Results_List struct{ capnp.List }

// NewPubSub_join_Results creates a new list of PubSub_join_Results.
func NewPubSub_join_Results_List(s *capnp.Segment, sz int32) (PubSub_join_Results_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return PubSub_join_Results_List{l}, err
}

func (s PubSub_join_Results_List) At(i int) PubSub_join_Results {
	return PubSub_join_Results{s.List.Struct(i)}
}

func (s PubSub_join_Results_List) Set(i int, v PubSub_join_Results) error {
	return s.List.SetStruct(i, v.Struct)
}

func (s PubSub_join_Results_List) String() string {
	str, _ := text.MarshalList(0x9f6c50fbc67b1d88, s.List)
	return str
}

// PubSub_join_Results_Future is a wrapper for a PubSub_join_Results promised by a client call.
type PubSub_join_Results_Future struct{ *capnp.Future }

func (p PubSub_join_Results_Future) Struct() (PubSub_join_Results, error) {
	s, err := p.Future.Struct()
	return PubSub_join_Results{s}, err
}

func (p PubSub_join_Results_Future) Topic() Topic {
	return Topic{Client: p.Future.Field(0, nil).Client()}
}

const schema_f9d8a0180405d9ed = "x\xda\x94\x93Oh\xd4\\\x14\xc5\xef}/iJ\xe9" +
	"|\xed\xfb^\xad\x7fK\xa1T\xd1\x82S[A\xa1(" +
	"\x1d*R+\x0a\xc9\xd4\"\xa5\xb8H\xc6`\xa3\x93i" +
	"H\x9a\xc5 2\x08\x05\xc7\x85+7Z\x94\x8a.\xd4" +
	"\x9d\x1b\x17.\\\xb8\x19)v\xa1\xe0\xa2\xea\xda\x95\x14" +
	"iAPf\x11I2\x99d\xa62\xe2*\x8b\x1c\xce" +
	"=\xf7w\xcf;\xa2b\x86\x8c\x88\xc3\"\x80rBl" +
	"\xf3F\xd7\xfaW\x97\x8fYK\xc08\x02\x08\x12\xc0\xd1" +
	"\x1bd\x08A\xf0N-\xbd+\x96\xefv\x97\xc3?\"" +
	"\xfa\xbft\xf2?\x02r\x93\x8c\x03z[_:\x9el" +
	"N\xad\xdf\x06\xb6\xb7.\xb8C\xce\xfa\x82\xfb\x81\xe0\xc5" +
	"\xb5\xee\xe1\x83\xcf\x0b\xf7\x80\xa5\xa8\xb7\xf1I\x14v=" +
	"Z\xff\x05\x80\xfc\x15Y\xe6o\xc8N\x00\xbeJn\xf1" +
	">*\x01x\xe7^\x16\xe7*\xee\xf1\x87\x89 \"\xdd" +
	"\xe3\x07)\xf7]\xafT\xe5\xfc\x0a\xb0\x9e\xfa\x9c\x0d\xd2" +
	"\xe1\xcf\xf9\x11\xccy6\xf7}\xc1\xad\xd8o\x93Iw" +
	"\xd0\x01_\xd0G}A\xb5\xb2\xd5{`\xf2\xc1\x07`" +
	"\x9c\xc6\xa9\x00\xf9I\xfa\x99O\xf9\xe3\xf9i:\xc9\xcd" +
	" \xc8\x7f_\xdb\xf6\xad\xf4\xacmnK=C\x1f\xf3" +
	"K\x81x\x96N\xf2\x9b\x81X{}\xf1\xdb!\xfe\xf1" +
	"g\xc8 Hm\xd0\xac\x9f\xfa\xe9\xf9\"y\xdf\x9b\xa9" +
	"&S\xcfP\xe2\x87\x9a\xa5\xe3p\xd8\xb3\\\xcdq\xb5" +
	"t\x8e\xaaV\xc1\x1a\xbb\xb0`\x19\xb9\xb4\xe3jN\xce" +
	"64}0\xab;]n~\xd1\xf9\xa3\xccr\xb5\xbc" +
	"\xe1\xcc\x0f\xca\xaa\xad\x9a\xe8(\x02\x15\x00\x04\x04`\xa9" +
	"\x01\x00\xa5\x9d\xa2\xd2CP2\x9d+\x98\x02\x82)\xc0" +
	"\xba\x8d\x90\xb09\xa3\x16.\xe7u;=\x1f|C7" +
	"\x07\xe0_\xec0\xb2\xa3FN\x110\x89\x1a'J\xb5" +
	"\x01J;\x15\x01\xea\x85\xc2\xe8\xd4ld\x02\x08\xdb/" +
	"a|B\x8c\xfa\xc8vg\x810&\x95j\xbbf\xd0" +
	"\x8b\xe0\x00\xea\x19\x94\x11[\xa2\xc9\xea\x8e\x9b\xa7\xdb\xf8" +
	"\xc9\xae6\xedj\xe9\xab\x0bF!\x94,6-<\x1a" +
	"/\xdc\xbf\xe8;\"K6\x06\x19\xe0_.'\xab\xb6" +
	"\xa4\x9a\x0dG\x99\x88MK!l\x1bY\x8c\xaa\xc9\x96" +
	"4\x9f\x88\xea\xb6\x8c\xa8\x08\x01\xc5\xe8\xd5aT=\xc6" +
	"\xc6\x800Q\x1a\x0f\x9d\x1b\xd1`\xb4\xb54\xedj\xb1" +
	"ITN\x8c\xde\x16cC\x81I\x97O\xa6\xd1\xa2E" +
	"cB\x82\xd8\x0ar\xadTI\x1aC1\x8d\xae\x82j" +
	"\xea\xd8\x09\x04;\x01\x7f\x07\x00\x00\xff\xff\xfa\x01K\xa1"

func init() {
	schemas.Register(schema_f9d8a0180405d9ed,
		0x8470369ac91fcc32,
		0x8810938879cb8443,
		0x89d849f1a30adbf2,
		0x986ea9282f106bb0,
		0x9d3775c65b79b54c,
		0x9f6c50fbc67b1d88,
		0xc772c6756fef5ba8,
		0xd19c472616f2c6fb,
		0xf1cc149f1c06e50e,
		0xf8d41329eb57bd62,
		0xfb4016d002794da7)
}
