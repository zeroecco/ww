// Code generated by capnpc-go. DO NOT EDIT.

package api

import (
	capnp "zombiezen.com/go/capnproto2"
	text "zombiezen.com/go/capnproto2/encoding/text"
	schemas "zombiezen.com/go/capnproto2/schemas"
)

// Heartbeat is a liveliness message that is broadcast over pubsub.
type Heartbeat struct{ capnp.Struct }

// Heartbeat_TypeID is the unique identifier for the type Heartbeat.
const Heartbeat_TypeID = 0xbbeb920e5748c15b

func NewHeartbeat(s *capnp.Segment) (Heartbeat, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return Heartbeat{st}, err
}

func NewRootHeartbeat(s *capnp.Segment) (Heartbeat, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return Heartbeat{st}, err
}

func ReadRootHeartbeat(msg *capnp.Message) (Heartbeat, error) {
	root, err := msg.RootPtr()
	return Heartbeat{root.Struct()}, err
}

func (s Heartbeat) String() string {
	str, _ := text.Marshal(0xbbeb920e5748c15b, s.Struct)
	return str
}

func (s Heartbeat) Id() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.Text(), err
}

func (s Heartbeat) HasId() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Heartbeat) IdBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytes(), err
}

func (s Heartbeat) SetId(v string) error {
	return s.Struct.SetText(0, v)
}

func (s Heartbeat) Ttl() int64 {
	return int64(s.Struct.Uint64(0))
}

func (s Heartbeat) SetTtl(v int64) {
	s.Struct.SetUint64(0, uint64(v))
}

// Heartbeat_List is a list of Heartbeat.
type Heartbeat_List struct{ capnp.List }

// NewHeartbeat creates a new list of Heartbeat.
func NewHeartbeat_List(s *capnp.Segment, sz int32) (Heartbeat_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1}, sz)
	return Heartbeat_List{l}, err
}

func (s Heartbeat_List) At(i int) Heartbeat { return Heartbeat{s.List.Struct(i)} }

func (s Heartbeat_List) Set(i int, v Heartbeat) error { return s.List.SetStruct(i, v.Struct) }

func (s Heartbeat_List) String() string {
	str, _ := text.MarshalList(0xbbeb920e5748c15b, s.List)
	return str
}

// Heartbeat_Promise is a wrapper for a Heartbeat promised by a client call.
type Heartbeat_Promise struct{ *capnp.Pipeline }

func (p Heartbeat_Promise) Struct() (Heartbeat, error) {
	s, err := p.Pipeline.Struct()
	return Heartbeat{s}, err
}

const schema_b3f8acfcffafd8e8 = "x\xda\x1c\xca1J\x03A\x1c\x85\xf1\xf7fw]\x91" +
	" \x0e\xc4\xc6&\x13\xb1\xd0\xc2\x18\xc1\xcaJ\xc1\"\xa5" +
	"SY\x08\xca\x7f\x93\xc1,\xc4d\xd9\x99\xc4;X\xea" +
	"\x01\xbc\x80`a)\x16\x16^A<\x82`ge!" +
	"\x8c$\xed\xf7\xfd\xd6N\x8e\x94\xce.\x01\x9bfK\xf1" +
	"\xfc\xadw\xb6z\xf7\xfd\x02\xdb$\xe3\xd7\xe7S\xfc{" +
	"\xfc}F\xc6\x1c\xd0\xeb?\xba\x9d\xebvK\xcb\x0d\x18" +
	"\xa5*\xf7\x86Nj\x15\x0a'\xa1\xd3\x97j\\\x1d\xf6" +
	"\x9c\xd4\xadE8%mJ\x15/\xee\x1f\xec\xeb\xc7\xed" +
	";l\xaax\xdc%\x1b\xc0>WT\x9c\xd3\xb9\xcc\x83" +
	")\xbd\x113*gnT\x8e\x9d\xf7\xe6\xday/W" +
	"\xce\x84\xa1,fQOd\xd0\x17\x1f\xccd\xe6jS" +
	"M\x0b?-:\xa0]NR %\xa0w6\x00\xbb" +
	"\x95\xd0v\x15\xc9&\xe7mw\x13\xb0\xdb\x09\xed\x81b" +
	"R\x0e\xd8\x80b\x03\xccC\x181\x83b\x06\xfe\x07\x00" +
	"\x00\xff\xff\x13\xc7CL"

func init() {
	schemas.Register(schema_b3f8acfcffafd8e8,
		0xbbeb920e5748c15b)
}
