package quickjsWrap_test

import (
	"fmt"
	quickjs "github.com/Maxwellism/gopher-qjs/bind"
	"github.com/Maxwellism/gopher-qjs/wrap"
	"testing"
)

var FnTest = func() {
	fmt.Println("FnTest")
}

func TestWrapFn(t *testing.T) {
	wrap := quickjsWrap.WrapFn(FnTest)
	rt := quickjs.NewRuntime()
	defer rt.Close()

	ctx := rt.NewContext()
	defer ctx.Close()

	ctx.Globals().Set("wrapTestFn", ctx.Function(func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		return wrap(ctx, this, args)
	}))

	ctx.Eval(`
wrapTestFn()
`)
}
