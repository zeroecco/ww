package cluster

import (
	"sync"

	syncutil "github.com/lthibault/util/sync"
)

type releaseFunc func()

type node struct {
	ref ref

	Name   string
	parent *node

	mu sync.RWMutex
	cs map[string]*node
}

func (n *node) Path() []string {
	if n.parent == nil {
		return make([]string, 0, 16) // best-effort pre-alloc
	}

	return append(n.parent.Path(), n.Name)
}

func (n *node) Acquire() *node {
	n.mu.RLock()
	defer n.mu.RUnlock()

	n.ref.Incr()
	return n
}

func (n *node) Release() {
	if n.parent != nil {
		defer n.parent.Release()
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	n.ref.Decr()

	for k, v := range n.cs {
		if v.ref.Zero() {
			v.mu.Lock()

			// still zero?
			if v.ref.Zero() {
				delete(n.cs, k)
			}

			v.mu.Unlock()
		}
	}
}

func (n *node) Children() (map[string]*node, releaseFunc) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if len(n.cs) == 0 {
		return nil, func() {}
	}

	m := make(map[string]*node, len(n.cs)) // TODO:  pool
	for k, v := range n.cs {
		n.ref.Incr()
		m[k] = v.Acquire()
	}

	return m, func() {
		for _, v := range m { // TODO:  return n.cs to pool?
			v.Release()
		}
	}
}

func (n *node) Walk(path []string) (u *node) {
	if len(path) == 0 {
		return n
	}

	n.mu.Lock()
	if n.cs == nil {
		n.cs = make(map[string]*node, 1)
	}

	if u = n.cs[path[0]]; u == nil {
		u = &node{Name: path[0], parent: n}
		n.cs[path[0]] = u
	}
	u.Acquire()
	n.mu.Unlock()

	return u.Walk(path[1:])
}

type ref syncutil.Ctr

func (r *ref) Incr()      { (*syncutil.Ctr)(r).Incr() }
func (r *ref) Decr()      { (*syncutil.Ctr)(r).Decr() }
func (r *ref) Zero() bool { return (*syncutil.Ctr)(r).Int() == 0 }
