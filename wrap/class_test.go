package quickjs_test

import (
	"fmt"
	quickjs "github.com/Maxwellism/gopher-qjs/wrap"
	"testing"
)

func TestNewClass(t *testing.T) {
	class := quickjs.NewClass("classTest")
	class.SetConstructor(func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) interface{} {
		this.Set("aaa", ctx.Int32(32))
		//fmt.Println("=========start==========")
		return 32
	})
	class.SetFinalizer(func(obj interface{}) {
		fmt.Println(obj)
		fmt.Println("???????????")
	})

	//class.AddClassFn("testClassFn", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
	//	return ctx.Float64(3.21)
	//})
	rt := quickjs.NewRuntime()
	defer rt.Close()

	// Create a new context
	ctx := rt.NewContext()
	defer ctx.Close()

	//ctx.Globals().Set("classTest", ctx.Function(func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
	//	//this.Set("aaa", ctx.Int32(32))
	//	fmt.Println("????????????")
	//	return ctx.Undefined()
	//}))

	res, err := ctx.Eval(`
let c = new classTest();
console.log(c._goClassID)
console.log(c._goObjectID)
//console.log(c.testClassFn())
`)
	defer res.Free()
	rt.RunGC()
	if err != nil {
		fmt.Println(err)
	}
}
