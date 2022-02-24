package anchor

import (
	"context"
	"strings"
	"sync"

	api "github.com/wetware/ww/internal/api/cluster"
)

type AnchorServer struct {
	store map[string][]byte
	mu    sync.Mutex
}

func newAnchorServer() AnchorServer {
	return AnchorServer{store: make(map[string][]byte)}
}

func (as AnchorServer) Ls(context.Context, api.AnchorProvider_ls) error {
	return nil // TODO
}

func (as AnchorServer) Walk(ctx context.Context, call api.AnchorProvider_walk) error {
	capPath, err := call.Args().Path()
	if err != nil {
		return err
	}

	path := make([]string, capPath.Len())
	for i := 0; i < capPath.Len(); i++ {
		path[i], err = capPath.At(i)
		if err != nil {
			return err
		}
	}

	if !isValid(path) || len(path) < 3 {
		return ErrInvalidPath
	}

	results, err := call.AllocResults()
	if err != nil {
		return err
	}
	anchor, err := results.NewAnchor()
	if err != nil {
		return err
	}
	// TODO: level-3 RPC
	server := ContainerSever{
		path:  strings.Join(path, "/"),
		store: as.store,
		mu:    &as.mu,
	}
	anchor.SetContainer(api.Anchor_Container_ServerToClient(server, &defaultPolicy)) // TODO
	return nil
}

type ContainerSever struct {
	path  string
	store map[string][]byte

	mu *sync.Mutex
}

func (cs ContainerSever) Set(ctx context.Context, call api.Anchor_Container_set) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	data, err := call.Args().Data()
	if err != nil {
		return err
	}

	cs.store[cs.path] = data
	return nil
}

func (cs ContainerSever) Get(ctx context.Context, call api.Anchor_Container_get) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	results, err := call.AllocResults()
	if err != nil {
		return err
	}

	return results.SetData(cs.store[cs.path])
}
