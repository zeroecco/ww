package core_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mock_vendor "github.com/wetware/ww/internal/test/mock/vendor"
	ww "github.com/wetware/ww/pkg"
	"github.com/wetware/ww/pkg/lang/core"
	capnp "zombiezen.com/go/capnproto2"
)

const width = 32

func TestPersistentVector(t *testing.T) {
	t.Parallel()

	const count = 4096

	var err error
	var v core.Vector = core.EmptyVector
	for i := 0; i < count; i++ {
		v, err = v.Cons(mustInt(i))
		require.NoError(t, err, "cons error on iteration %d", i)

		assertVectorTypeOK(t, v)
	}

	for i := 0; i < count; i++ {
		v, err = v.Pop()
		require.NoError(t, err, "pop error on iteration %d", i)

		assertVectorTypeOK(t, v)
	}

	assert.IsType(t, core.EmptyPersistentVector{}, v)
}

func TestEmptyVector(t *testing.T) {
	t.Parallel()

	t.Run("New", func(t *testing.T) {
		t.Parallel()

		v, err := core.NewVector(capnp.SingleSegment(nil))
		require.NoError(t, err)
		require.NotNil(t, v)
		assert.IsType(t, core.EmptyPersistentVector{}, v)

		cnt, err := v.Count()
		require.NoError(t, err)
		assert.Zero(t, cnt)
	})

	t.Run("Count", func(t *testing.T) {
		t.Parallel()

		cnt, err := core.EmptyVector.Count()
		assert.NoError(t, err)
		assert.Zero(t, cnt)
	})

	t.Run("Render", func(t *testing.T) {
		t.Parallel()

		s, err := core.Render(core.EmptyVector)
		assert.NoError(t, err)
		assert.Equal(t, "[]", s)
	})

	t.Run("Assoc", func(t *testing.T) {
		t.Run("Append", func(t *testing.T) {
			v, err := core.EmptyVector.Assoc(0, mustInt(0))
			assert.NoError(t, err)
			assert.IsType(t, core.ShallowPersistentVector{}, v)

			cnt, err := v.Count()
			assert.NoError(t, err)
			assert.Equal(t, 1, cnt)
		})

		t.Run("Overflow", func(t *testing.T) {
			v, err := core.EmptyVector.Assoc(1, mustInt(0))
			assert.EqualError(t, err, core.ErrIndexOutOfBounds.Error())
			assert.Nil(t, v)
		})
	})

	t.Run("EntryAt", func(t *testing.T) {
		t.Parallel()

		v, err := core.EmptyVector.EntryAt(0)
		assert.EqualError(t, err, core.ErrIndexOutOfBounds.Error())
		assert.Nil(t, v)
	})

	t.Run("Pop", func(t *testing.T) {
		t.Parallel()

		tail, err := core.EmptyVector.Pop()
		assert.True(t, errors.Is(err, core.ErrIllegalState),
			"'%s' is not ErrIllegalState", err)

		assert.Nil(t, tail)
	})

	t.Run("Cons", func(t *testing.T) {
		t.Parallel()

		item := mustInt(0)

		v, err := core.EmptyVector.Cons(item)
		assert.NoError(t, err)
		assert.NotEqual(t, core.EmptyVector, v)
		assert.IsType(t, core.ShallowPersistentVector{}, v)

		any, err := v.EntryAt(0)
		assert.NoError(t, err)

		eq, err := core.Eq(item, any)
		assert.NoError(t, err)
		assert.True(t, eq)
	})

	t.Run("Conj", func(t *testing.T) {
		t.Parallel()

		t.Run("Nop", func(t *testing.T) {
			t.Parallel()

			v, err := core.EmptyVector.Conj()
			assert.NoError(t, err)
			assert.IsType(t, core.EmptyPersistentVector{}, v)
		})

		v, err := core.EmptyVector.Conj(mustInt(0))
		assert.NoError(t, err)

		v2, err := core.NewVector(nil)
		assert.NoError(t, err)

		v3, err := core.Conj(v2, mustInt(0))
		assert.NoError(t, err)

		eq, err := core.Eq(v, v3)
		assert.NoError(t, err)
		assert.True(t, eq, "vector v should be equal to v2.")
	})

	t.Run("Seq", func(t *testing.T) {
		t.Parallel()

		seq, err := core.EmptyVector.Seq()
		assert.NoError(t, err)

		cnt, err := seq.Count()
		assert.NoError(t, err)
		assert.Zero(t, cnt)
	})
}

func TestShallowPersistentVector(t *testing.T) {
	t.Parallel()

	const count = width

	t.Run("New", func(t *testing.T) {
		t.Parallel()

		v, err := core.NewVector(capnp.SingleSegment(nil), valueRange(count)...)
		require.NoError(t, err)
		assert.IsType(t, core.ShallowPersistentVector{}, v)

		cnt, err := v.Count()
		assert.NoError(t, err)
		assert.Equal(t, count, cnt)

		t.Run("AllocError", func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			arena := mock_vendor.NewMockArena(ctrl)
			arena.EXPECT().NumSegments().
				Return(int64(0)).
				Times(1)

			arena.EXPECT().
				Allocate(gomock.Any(), gomock.Any()).
				Return(capnp.SegmentID(0), nil, errors.New("mock arena allocation failure"))

			_, err := core.NewVector(arena, valueRange(count)...)
			require.Error(t, err)
		})
	})

	t.Run("Count", func(t *testing.T) {
		t.Parallel()

		for i := 1; i <= count; i++ {
			withParallelIndex(t, i, func(t *testing.T, i int) {
				v, err := core.NewVector(capnp.SingleSegment(nil), valueRange(i)...)
				require.NoError(t, err)

				cnt, err := v.Count()
				assert.NoError(t, err)
				assert.Equal(t, i, cnt)
			})
		}
	})

	t.Run("Render", func(t *testing.T) {
		t.Parallel()

		for i := 1; i <= count; i++ {
			withParallelIndex(t, i, func(t *testing.T, i int) {
				v, err := core.NewVector(capnp.SingleSegment(nil), valueRange(i)...)
				require.NoError(t, err)

				s, err := core.Render(v)
				assert.NoError(t, err)

				assert.Equal(t, renderRange(i), s)
			})
		}
	})

	t.Run("Assoc", func(t *testing.T) {
		t.Run("Update", func(t *testing.T) {
			v, err := core.NewVector(capnp.SingleSegment(nil), valueRange(16)...)
			require.NoError(t, err)

			v, err = v.Assoc(1, mustInt(9001))
			require.NoError(t, err)

			cnt, err := v.Count()
			assert.NoError(t, err)
			assert.Equal(t, 16, cnt)

			got, err := v.EntryAt(1)
			require.NoError(t, err)

			eq, err := core.Eq(mustInt(9001), got)
			require.NoError(t, err)
			assert.True(t, eq)
		})

		t.Run("Append", func(t *testing.T) {
			v, err := core.NewVector(capnp.SingleSegment(nil), valueRange(16)...)
			require.NoError(t, err)

			v, err = v.Assoc(16, mustInt(16))
			assert.NoError(t, err)

			cnt, err := v.Count()
			assert.NoError(t, err)
			assert.Equal(t, 17, cnt)
		})

		t.Run("Overflow", func(t *testing.T) {
			v, err := core.NewVector(capnp.SingleSegment(nil), mustInt(16))
			require.NoError(t, err)

			v, err = v.Assoc(9001, mustInt(9001))
			assert.EqualError(t, err, core.ErrIndexOutOfBounds.Error())
			assert.Nil(t, v)
		})
	})

	t.Run("EntryAt", func(t *testing.T) {
		t.Parallel()

		v, err := core.NewVector(capnp.SingleSegment(nil), valueRange(count)...)
		require.NoError(t, err)

		for i := 0; i < count; i++ {
			withParallelIndex(t, i, func(t *testing.T, i int) {
				item, err := v.EntryAt(i)
				assert.NoError(t, err)

				eq, err := core.Eq(item, mustInt(i))
				assert.NoError(t, err)
				assert.True(t, eq)
			})
		}

		t.Run("NegativeIndex", func(t *testing.T) {
			item, err := v.EntryAt(-1)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, core.ErrIndexOutOfBounds))
			assert.Nil(t, item)
		})

		t.Run("Overflow", func(t *testing.T) {
			item, err := v.EntryAt(33)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, core.ErrIndexOutOfBounds))
			assert.Nil(t, item)
		})

	})

	t.Run("Pop", func(t *testing.T) {
		t.Parallel()

		t.Run("ResultIsEmpty", func(t *testing.T) {
			t.Parallel()

			v, err := core.NewVector(capnp.SingleSegment(nil), valueRange(1)...)
			require.NoError(t, err)

			res, err := v.Pop()
			assert.NoError(t, err)
			assert.Equal(t, core.Vector(core.EmptyVector), res)
		})

		t.Run("ResultNotEmpty", func(t *testing.T) {
			t.Parallel()

			v, err := core.NewVector(capnp.SingleSegment(nil), valueRange(2)...)
			require.NoError(t, err)

			res, err := v.Pop()
			assert.NoError(t, err)
			assert.IsType(t, core.ShallowPersistentVector{}, res)

			t.Run("CountOK", func(t *testing.T) {
				t.Parallel()

				cnt, err := res.Count()
				assert.NoError(t, err)
				assert.Equal(t, 1, cnt)
			})

			t.Run("EntryAtOK", func(t *testing.T) {
				t.Parallel()

				item, err := res.EntryAt(0)
				assert.NoError(t, err)
				eq, err := core.Eq(mustInt(0), item)
				require.NoError(t, err)
				assert.True(t, eq)
			})

			t.Run("EqualityOK", func(t *testing.T) {
				t.Parallel()

				want, err := core.NewVector(capnp.SingleSegment(nil), valueRange(1)...)
				require.NoError(t, err)
				eq, err := core.Eq(want, res)
				assert.NoError(t, err)
				assert.True(t, eq, "expected %s, got %s", mustRender(want), mustRender(res))
			})
		})
	})

	t.Run("Cons", func(t *testing.T) {
		t.Parallel()

		for i := 1; i < count; i++ {
			withParallelIndex(t, i, func(t *testing.T, i int) {
				v, err := core.NewVector(capnp.SingleSegment(nil), valueRange(i)...)
				require.NoError(t, err)

				insert := mustInt(9001)

				res, err := v.Cons(insert)
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.IsType(t, core.ShallowPersistentVector{}, res)

				cnt, err := res.Count()
				assert.NoError(t, err)
				assert.Equal(t, i+1, cnt)

				item, err := res.EntryAt(i)
				assert.NoError(t, err)
				eq, err := core.Eq(insert, item)
				assert.NoError(t, err)
				assert.True(t, eq, "expected %s, got %s", insert, item)
			})
		}

		t.Run("Overflow", func(t *testing.T) {
			t.Parallel()

			v, err := core.NewVector(capnp.SingleSegment(nil), valueRange(count)...)
			require.NoError(t, err)

			insert := mustInt(9001)

			res, err := v.Cons(insert)
			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.IsType(t, core.DeepPersistentVector{}, res)

			cnt, err := res.Count()
			assert.NoError(t, err)
			assert.Equal(t, 33, cnt)

			item, err := res.EntryAt(count)
			assert.NoError(t, err)
			eq, err := core.Eq(insert, item)
			assert.NoError(t, err)
			assert.True(t, eq, "expected %s, got %s", insert, item)
		})
	})

	t.Run("Conj", func(t *testing.T) {
		t.Parallel()

		t.Run("InRange", func(t *testing.T) {
			t.Parallel()

			items := valueRange(count)

			v, err := core.NewVector(capnp.SingleSegment(nil), items[0])
			require.NoError(t, err)

			ctr, err := v.Conj(items[1:]...)
			assert.NoError(t, err)
			assert.NotNil(t, ctr)

			cnt, err := ctr.Count()
			assert.NoError(t, err)
			assert.Equal(t, count, cnt)

			require.IsType(t, core.ShallowPersistentVector{}, ctr)
			for i, want := range items {
				got, err := ctr.(core.ShallowPersistentVector).EntryAt(i)
				assert.NoError(t, err)

				eq, err := core.Eq(want, got)
				require.NoError(t, err)
				require.True(t, eq)
			}
		})

		t.Run("Overflow", func(t *testing.T) {
			t.Parallel()

			items := valueRange(33)

			v, err := core.NewVector(capnp.SingleSegment(nil), items[0])
			require.NoError(t, err)

			ctr, err := v.Conj(items[1:]...)
			assert.NoError(t, err)
			assert.NotNil(t, ctr)

			cnt, err := ctr.Count()
			assert.NoError(t, err)
			assert.Equal(t, 33, cnt)

			require.IsType(t, core.DeepPersistentVector{}, ctr)
			for i, want := range items {
				got, err := ctr.(core.DeepPersistentVector).EntryAt(i)
				assert.NoError(t, err)

				eq, err := core.Eq(want, got)
				require.NoError(t, err)
				require.True(t, eq)
			}
		})
	})

	t.Run("Seq", func(t *testing.T) {
		t.Parallel()

		items := valueRange(count)

		v, err := core.NewVector(capnp.SingleSegment(nil), items...)
		require.NoError(t, err)
		require.NotNil(t, v)

		seq, err := v.Seq()
		require.NoError(t, err)
		require.NotNil(t, seq)

		results, err := core.ToSlice(seq)
		require.NoError(t, err)
		assert.Len(t, results, len(items))

		for i, got := range results {
			eq, err := core.Eq(items[i], got)
			assert.NoError(t, err)
			assert.True(t, eq)
		}
	})
}

func TestDeepPersistentVector(t *testing.T) {
	t.Parallel()

	const count = 4096
	var ranges = [...]int{33, 64, 128, 256, 512, 1024, 2048, 4096}

	t.Run("New", func(t *testing.T) {
		t.Parallel()

		v, err := core.NewVector(capnp.SingleSegment(nil), valueRange(count)...)
		require.NoError(t, err)
		require.IsType(t, core.DeepPersistentVector{}, v)

		cnt, err := v.Count()
		require.NoError(t, err)
		assert.Equal(t, count, cnt)

		t.Run("AllocError", func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			arena := mock_vendor.NewMockArena(ctrl)
			arena.EXPECT().NumSegments().
				Return(int64(0)).
				Times(1)

			arena.EXPECT().
				Allocate(gomock.Any(), gomock.Any()).
				Return(capnp.SegmentID(0), nil, errors.New("mock arena allocation failure"))

			_, err := core.NewVector(arena, valueRange(count)...)
			require.Error(t, err)
		})
	})

	t.Run("Count", func(t *testing.T) {
		t.Parallel()

		for _, i := range ranges {
			withParallelIndex(t, i, func(t *testing.T, i int) {
				v, err := core.NewVector(capnp.SingleSegment(nil), valueRange(i)...)
				require.NoError(t, err)

				cnt, err := v.Count()
				assert.NoError(t, err)
				assert.Equal(t, i, cnt)
			})
		}
	})

	t.Run("Render", func(t *testing.T) {
		t.Parallel()

		for _, i := range ranges {
			withParallelIndex(t, i, func(t *testing.T, i int) {
				v, err := core.NewVector(capnp.SingleSegment(nil), valueRange(i)...)
				require.NoError(t, err)

				s, err := core.Render(v)
				assert.NoError(t, err)

				assert.Equal(t, renderRange(i), s)
			})
		}
	})

	t.Run("Assoc", func(t *testing.T) {
		t.Run("Update", func(t *testing.T) {
			t.Run("Tail", func(t *testing.T) {
				v, err := core.NewVector(capnp.SingleSegment(nil), valueRange(count)...)
				require.NoError(t, err)

				v, err = v.Assoc(count-1, mustInt(9001))
				require.NoError(t, err)

				cnt, err := v.Count()
				assert.NoError(t, err)
				assert.Equal(t, count, cnt)

				got, err := v.EntryAt(count - 1)
				require.NoError(t, err)

				eq, err := core.Eq(mustInt(9001), got)
				require.NoError(t, err)
				assert.True(t, eq)
			})

			t.Run("Trie", func(t *testing.T) {
				v, err := core.NewVector(capnp.SingleSegment(nil), valueRange(count)...)
				require.NoError(t, err)

				v, err = v.Assoc(1024, mustInt(9001))
				require.NoError(t, err)

				cnt, err := v.Count()
				assert.NoError(t, err)
				assert.Equal(t, count, cnt)

				got, err := v.EntryAt(1024)
				require.NoError(t, err)

				eq, err := core.Eq(mustInt(9001), got)
				require.NoError(t, err)
				assert.True(t, eq)
			})
		})

		t.Run("Append", func(t *testing.T) {
			v, err := core.NewVector(capnp.SingleSegment(nil), valueRange(count)...)
			require.NoError(t, err)

			v, err = v.Assoc(count, mustInt(count))
			assert.NoError(t, err)

			cnt, err := v.Count()
			assert.NoError(t, err)
			assert.Equal(t, count+1, cnt)
		})

		t.Run("Overflow", func(t *testing.T) {
			v, err := core.NewVector(capnp.SingleSegment(nil), mustInt(count))
			require.NoError(t, err)

			v, err = v.Assoc(9001, mustInt(9001))
			assert.EqualError(t, err, core.ErrIndexOutOfBounds.Error())
			assert.Nil(t, v)
		})
	})

	t.Run("EntryAt", func(t *testing.T) {
		t.Parallel()

		v, err := core.NewVector(capnp.SingleSegment(nil), valueRange(count)...)
		require.NoError(t, err)

		for _, i := range ranges {
			withParallelIndex(t, i, func(t *testing.T, i int) {
				item, err := v.EntryAt(i - 1)
				assert.NoError(t, err)

				eq, err := core.Eq(item, mustInt(i-1))
				assert.NoError(t, err)
				assert.True(t, eq)
			})
		}

		t.Run("NegativeIndex", func(t *testing.T) {
			item, err := v.EntryAt(-1)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, core.ErrIndexOutOfBounds))
			assert.Nil(t, item)
		})
	})

	t.Run("Pop", func(t *testing.T) {
		t.Parallel()

		t.Run("ResultIsShallow", func(t *testing.T) {
			t.Parallel()

			v, err := core.NewVector(capnp.SingleSegment(nil), valueRange(33)...)
			require.NoError(t, err)

			res, err := v.Pop()
			assert.NoError(t, err)
			assert.IsType(t, core.ShallowPersistentVector{}, res)

			cnt, err := res.Count()
			require.NoError(t, err)
			assert.Equal(t, cnt, width)
		})

		t.Run("ResultIsDeep", func(t *testing.T) {
			t.Parallel()

			t.Run("PopTail", func(t *testing.T) {
				t.Parallel()

				// with cnt=128 the vector is 'deep' and popping a value WILL NOT
				// cause a new tail to be retrieved from the trie.
				v, err := core.NewVector(capnp.SingleSegment(nil), valueRange(128)...)
				require.NoError(t, err)

				res, err := v.Pop()
				assert.NoError(t, err)
				assert.IsType(t, core.DeepPersistentVector{}, res)

				cnt, err := res.Count()
				require.NoError(t, err)
				assert.Equal(t, cnt, 127)
			})

			t.Run("PopTrie", func(t *testing.T) {
				t.Parallel()

				// with cnt=1025 the vector is 'deep' and popping a value WILL
				// cause a new tail to be retrieved from the trie.
				v, err := core.NewVector(capnp.SingleSegment(nil), valueRange(1025)...)
				require.NoError(t, err)

				res, err := v.Pop()
				assert.NoError(t, err)
				assert.IsType(t, core.DeepPersistentVector{}, res)

				cnt, err := res.Count()
				require.NoError(t, err)
				assert.Equal(t, cnt, 1024)
			})
		})

	})

	t.Run("Cons", func(t *testing.T) {
		t.Parallel()

		for _, i := range ranges {
			withParallelIndex(t, i, func(t *testing.T, i int) {
				v, err := core.NewVector(capnp.SingleSegment(nil), valueRange(i)...)
				require.NoError(t, err)

				insert := mustInt(9001)

				res, err := v.Cons(insert)
				require.NoError(t, err)
				require.NotNil(t, res)
				require.IsType(t, core.DeepPersistentVector{}, res)

				cnt, err := res.Count()
				require.NoError(t, err)
				require.Equal(t, i+1, cnt)

				item, err := res.EntryAt(i)
				require.NoError(t, err)
				eq, err := core.Eq(insert, item)
				require.NoError(t, err)
				require.True(t, eq, "expected %s, got %s", insert, item)
			})
		}
	})

	t.Run("Conj", func(t *testing.T) {
		t.Parallel()

		items := valueRange(count)

		v, err := core.NewVector(capnp.SingleSegment(nil), items[0])
		require.NoError(t, err)

		ctr, err := v.Conj(items[1:]...)
		assert.NoError(t, err)
		assert.NotNil(t, ctr)

		cnt, err := ctr.Count()
		assert.NoError(t, err)
		assert.Equal(t, count, cnt)

		require.IsType(t, core.DeepPersistentVector{}, ctr)
		for i, want := range items {
			got, err := ctr.(core.DeepPersistentVector).EntryAt(i)
			require.NoError(t, err)

			eq, err := core.Eq(want, got)
			require.NoError(t, err)
			require.True(t, eq)
		}
	})

	t.Run("Seq", func(t *testing.T) {
		t.Parallel()

		items := valueRange(count)

		v, err := core.NewVector(capnp.SingleSegment(nil), items...)
		require.NoError(t, err)
		require.NotNil(t, v)

		seq, err := v.Seq()
		require.NoError(t, err)
		require.NotNil(t, seq)

		results, err := core.ToSlice(seq)
		require.NoError(t, err)
		assert.Len(t, results, len(items))

		for i, got := range results {
			eq, err := core.Eq(items[i], got)
			assert.NoError(t, err)
			assert.True(t, eq)
		}
	})
}

func valueRange(n int) []ww.Any {
	vs := make([]ww.Any, n)
	for i := range vs {
		vs[i] = mustInt(i)
	}
	return vs
}

func renderRange(n int) string {
	var b strings.Builder
	b.WriteRune('[')

	for i, val := range valueRange(n) {
		s, err := core.Render(val)
		if err != nil {
			panic(err)
		}

		b.WriteString(s)

		if i < n-1 {
			b.WriteRune(' ')
		}
	}

	b.WriteRune(']')
	return b.String()
}

func withParallelIndex(t *testing.T, i int, f func(*testing.T, int)) {
	t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
		t.Parallel()
		f(t, i)
	})
}

func assertVectorTypeOK(t *testing.T, v core.Vector) bool {
	cnt, err := v.Count()
	require.NoError(t, err)

	switch {
	case cnt == 0:
		return assert.IsType(t, core.EmptyPersistentVector{}, v)

	case cnt <= width:
		return assert.IsType(t, core.ShallowPersistentVector{}, v)

	default:
		return assert.IsType(t, core.DeepPersistentVector{}, v)

	}
}
