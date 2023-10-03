package window_test

import (
	quickjs "github.com/Maxwellism/gopher-qjs/bind"
	"github.com/Maxwellism/gopher-qjs/polyfill/window"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWindow(t *testing.T) {
	rt := quickjs.NewRuntime()
	defer rt.Close()

	ctx := rt.NewContext()
	defer ctx.Close()

	window.WindowInjectTo(ctx)

	ret, _ := ctx.Eval("Object.is(globalThis,globalThis.window)")
	defer ret.Free()
	require.EqualValues(t, true, ret.Bool())

}
