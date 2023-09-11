package quickjs_test

import (
	"fmt"
	"github.com/Maxwellism/gopher-qjs/wrap"
	"testing"
)

func TestModule(t *testing.T) {
	rt := quickjs.NewRuntime()
	defer rt.Close()

	// Create a new context
	ctx := rt.NewContext()
	defer ctx.Close()

	m := ctx.CreateModule("module_test")
	m.AddExportFn("fnTest", 0, func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		fmt.Println("3241")
		return ctx.Undefined()
	})

	m.BuildModule()

	res, err := ctx.EvalFile("./examples/my_module_test.js")
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
	//defer res.Free()
}
