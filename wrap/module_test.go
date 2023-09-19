package quickjs_test

import (
	"encoding/json"
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

func TestModClass(t *testing.T) {
	var classObject *ExampleObject

	m := quickjs.NewMod("module_test")

	class := m.CreateExportClass("classTest")

	class.SetConstructor(func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) interface{} {
		fmt.Println("=========start Constructor==========")
		if len(args) < 2 {
			panic("Constructor arg len is < 1")
		}
		if classObject == nil {
			classObject = &ExampleObject{}
		}
		classObject.Name = args[0].String()
		classObject.Age = args[1].Int32()
		return classObject
	})
	class.SetFinalizer(func(obj interface{}) {
		fmt.Println("=========finalizer=======")
		data, err := json.Marshal(obj)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		fmt.Println("go object value is:", string(data))
	})

	//class.AddClassGetFn("Name", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
	//	return ctx.String(classObject.Name)
	//})
	//
	//class.AddClassSetFn("Name", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
	//	classObject.Name = args[0].String() + "323"
	//	return ctx.Undefined()
	//})
	//
	//class.AddClassGetFn("Age", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
	//	return ctx.Int32(classObject.Age)
	//})

	rt := quickjs.NewRuntime()
	defer rt.Close()
	rt.AddGoMod(m)

	// Create a new context
	ctx := rt.NewContext()
	defer ctx.Close()

	res, err := ctx.EvalFile("./examples/my_module_class_test.js")
	if err != nil {
		panic(err)
	}
	defer res.Free()
}
