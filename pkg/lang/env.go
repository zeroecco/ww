package lang

import (
	"fmt"
	"strings"
	"sync"

	ww "github.com/wetware/ww/pkg"
	"github.com/wetware/ww/pkg/lang/core"
)

const (
	width = 32
	mask  = width - 1
)

var (
	_ core.Env = (*rootEnv)(nil)
	_ core.Env = (*mapEnv)(nil)
)

// rootEnv implements Env using a sharded map.  This is
// useful since the global environment subject to heavy
// concurrent access.  rootEnv's children are standard
// mapEnvs.
type rootEnv struct {
	mu   sync.RWMutex
	vars map[string]ww.Any
}

// NewEnv returns a root environment.
func NewEnv() core.Env {
	return &rootEnv{vars: map[string]ww.Any{}}
}

func (*rootEnv) Name() string       { return "[global frame]" }
func (*rootEnv) Parent() core.Env   { return nil }
func (env *rootEnv) Root() core.Env { return env }
func (env *rootEnv) Child(name string, vars map[string]ww.Any) core.Env {
	if vars == nil {
		vars = map[string]ww.Any{}
	}

	return mapEnv{name: name, parent: env, vars: vars}
}

func (env *rootEnv) Bind(name string, item ww.Any) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("%w: %s", core.ErrInvalidName, name)
	}

	env.mu.Lock()
	defer env.mu.Unlock()

	env.vars[name] = item
	return nil
}

func (env *rootEnv) Resolve(name string) (ww.Any, error) {
	env.mu.RLock()
	defer env.mu.RUnlock()

	if v, found := env.vars[name]; found {
		return v, nil
	}

	return nil, core.ErrNotFound
}

type mapEnv struct {
	parent core.Env
	name   string
	vars   map[string]ww.Any
}

func (env mapEnv) Name() string     { return env.name }
func (env mapEnv) Parent() core.Env { return env.parent }

func (env mapEnv) Root() core.Env {
	res := env.parent
	for res.Parent() != nil {
		res = res.Parent()
	}

	return res
}

func (env mapEnv) Child(name string, vars map[string]ww.Any) core.Env {
	if vars == nil {
		vars = map[string]ww.Any{}
	}
	return mapEnv{
		name:   name,
		parent: env,
		vars:   vars,
	}
}

func (env mapEnv) Bind(name string, val ww.Any) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("%w: %s", core.ErrInvalidName, name)
	}

	env.vars[name] = val
	return nil
}

func (env mapEnv) Resolve(name string) (ww.Any, error) {
	if v, found := env.vars[name]; found {
		return v, nil
	}
	return nil, core.ErrNotFound
}
