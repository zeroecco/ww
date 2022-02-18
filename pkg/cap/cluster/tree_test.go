package cluster

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var paths = [][]string{
	{"alpha"},
	{"alpha", "bravo"},
	{"alpha", "bravo", "charlie"},
	{"alpha", "bravo", "charlie", "delta"},
	{"alpha", "bravo", "charlie", "delta", "echo"},
	{"alpha", "bravo", "charlie", "delta", "echo", "fox"},
}

func TestWalk(t *testing.T) {
	t.Parallel()

	for _, path := range paths {
		t.Run(fmt.Sprintf("Depth-%d", len(path)), func(t *testing.T) {
			t.Parallel()

			var root node
			cs, release := root.Children()
			require.Empty(t, cs, "should not have children initially")
			require.Empty(t, root.Path(), "root anchor should have empty path")
			release()

			test := root.Walk(path)

			cs, release = root.Children()
			require.NotEmpty(t, cs, "root should have child")
			release()

			test.Release()

			cs, release = root.Children()
			assert.Empty(t, cs, "root should not have children")
			release()
		})
	}
}

var n *node

func BenchmarkTreeWalk(b *testing.B) {
	b.ReportAllocs()

	// Test performance when creating new anchor paths of various depth.
	b.Run("Alloc", func(b *testing.B) {
		var root node
		for _, path := range paths {
			b.Run(fmt.Sprintf("Depth-%d", len(path)), func(b *testing.B) {
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					n = root.Walk(path)
					n.Release()
				}
			})
		}
	})

	// Test performance on existing anchor paths of various depth.
	b.Run("Exists", func(b *testing.B) {
		var root node
		for _, path := range paths {
			b.Run(fmt.Sprintf("Depth-%d", len(path)), func(b *testing.B) {

				// pre-alloc
				n = root.Walk(path)
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					n = root.Walk(path)
				}
			})
		}
	})
}
