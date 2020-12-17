package lang_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	ww "github.com/wetware/ww/pkg"
	"github.com/wetware/ww/pkg/lang"
	"github.com/wetware/ww/pkg/lang/core"
)

func TestRoot(t *testing.T) {
	t.Parallel()

	env := lang.NewEnv()
	assert.Equal(t, "[global frame]", env.Name())
	assert.Nil(t, env.Parent())
	assert.Same(t, env, env.Root())

	child := env.Child("foo", nil)
	root := child.Root()

	assert.NotSame(t, child, root)
	assert.Same(t, env, root)
}

func Test_rootEnv_Bind_Resolve(t *testing.T) {
	t.Parallel()

	var v ww.Any
	var err error

	env := lang.NewEnv()

	t.Run("Bind", func(t *testing.T) {
		assert.NoError(t, env.Bind("foo", core.True))
		assert.EqualError(t, env.Bind("", core.True), core.ErrInvalidName.Error()+": ")
	})

	t.Run("Resolve", func(t *testing.T) {
		v, err = env.Resolve("foo")
		assert.NoError(t, err)
		assert.Equal(t, core.True, v)

		v, err = env.Resolve("non-existent")
		assert.EqualError(t, err, core.ErrNotFound.Error())
		assert.Nil(t, v)
	})
}

func Test_mapEnv_Bind_Resolve(t *testing.T) {
	t.Parallel()

	var v ww.Any
	var err error

	root := lang.NewEnv()
	env := root.Child("test", nil)

	assert.Equal(t, "test", env.Name())
	assert.Same(t, env.Root(), env.Parent())

	t.Run("Bind", func(t *testing.T) {
		assert.NoError(t, env.Bind("foo", core.True))
		assert.EqualError(t, env.Bind("", core.True), core.ErrInvalidName.Error()+": ")
	})

	t.Run("Resolve", func(t *testing.T) {
		v, err = env.Resolve("foo")
		assert.NoError(t, err)
		assert.Equal(t, core.True, v)

		v, err = env.Resolve("non-existent")
		assert.EqualError(t, err, core.ErrNotFound.Error())
		assert.Nil(t, v)
	})

	t.Run("Child", func(t *testing.T) {
		child := env.Child("test-2", nil)
		assert.Equal(t, "test-2", child.Name())
		assert.Equal(t, env, child.Parent())
		assert.Same(t, root, child.Root())
	})
}
