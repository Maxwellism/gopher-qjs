package quickjs_test

import (
	"fmt"
	"github.com/Maxwellism/gopher-qjs/wrap"
	"testing"
)

func TestModule(t *testing.T) {

	m := quickjs.NewMod("module_test")
	m.AddExportFn("fnTest", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		fmt.Println("3241")
		return ctx.Float64(3.123)
	})
	m.AddExportFn("fnTest1", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		val := ctx.Object()
		val.Set("Name", ctx.String("boy"))
		val.Set("Age", ctx.Int32(32))
		return val
	})

	m.AddExportFn("fnTest2", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) > 0 {
			//defer args[1].Free()
			fmt.Println(args[0].String())
		}
		return ctx.Null()
	})

	rt := quickjs.NewRuntime()
	defer rt.Close()
	rt.AddGoMod(m)

	// Create a new context
	ctx := rt.NewContext()
	defer ctx.Close()

	res, err := ctx.EvalFile("./examples/my_module_test.js")
	if err != nil {
		panic(err)
	}
	defer res.Free()
	//defer res.Free()
}
