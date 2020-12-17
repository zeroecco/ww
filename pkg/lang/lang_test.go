package lang_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mock_ww "github.com/wetware/ww/internal/test/mock/pkg"
	ww "github.com/wetware/ww/pkg"
	"github.com/wetware/ww/pkg/lang"
	"github.com/wetware/ww/pkg/lang/core"
	"github.com/wetware/ww/pkg/lang/reader"
)

// TODO:  replace this with testdata/lang_test.ww
const src = `
;; call built-in function
(= [:head] (pop [:head :tail]))

;; define nilary function
(= 'nop (def nop (fn nop [] nil)))

;; call nilary function
(nil? (nop))

;; define unary function
(= 'id (def id (fn id [x] x)))

;; call unary function
(= :value (id :value))

;; define multi-arity function
(= 'maybe (def maybe (fn maybe
						 ([] nil)
						 ([x] x))))

;; call multi-arity function
(nil? (maybe))
(= :value (maybe :value))
`

func TestLang(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	root := mock_ww.NewMockAnchor(ctrl)
	vm := lang.VM{
		Analyzer: &lang.Analyzer{Root: root},
		Env:      lang.NewEnv(),
	}

	require.NoError(t, vm.Init())

	forms, err := reader.New(strings.NewReader(src)).All()
	require.NoError(t, err)

	var res interface{}
	for i, form := range forms {
		s, err := core.Render(form.(ww.Any))
		require.NoError(t, err, "error rendering form for debugging")

		assert.NotPanics(t, func() {
			if res, err = vm.Eval(form); err != nil {
				require.NoError(t, err, "line %d: %s", i+1, s)
			}
		}, "line %d: %s", i+1, s)
	}

	b, ok := res.(core.Bool)
	require.True(t, ok, "test returned non-boolean type '%s'", reflect.TypeOf(res))

	assert.True(t, b.Bool(), "test failed")
}
