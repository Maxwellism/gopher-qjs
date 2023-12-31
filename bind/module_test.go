package quickjsBind_test

import (
	"fmt"
	"github.com/Maxwellism/gopher-qjs/bind"
	"github.com/Maxwellism/gopher-qjs/polyfill"
	json "github.com/json-iterator/go"
	"testing"
)

type ExampleObject struct {
	Name string `json:"name,omitempty"`
	Age  int32  `json:"age,omitempty"`
}

func TestModule(t *testing.T) {

	polyfill.Polyfill()

	rt := quickjsBind.NewRuntime()
	defer rt.Close()

	m := rt.CreateModule("module_test")
	m.AddExportFn("fnTest", func(ctx *quickjsBind.Context, this quickjsBind.Value, args []quickjsBind.Value) quickjsBind.Value {
		fmt.Println("3241")
		return ctx.Float64(3.123)
	})
	m.AddExportFn("fnTest1", func(ctx *quickjsBind.Context, this quickjsBind.Value, args []quickjsBind.Value) quickjsBind.Value {
		val := ctx.Object()
		val.Set("Name", ctx.String("boy"))
		val.Set("Age", ctx.Int32(32))
		return val
	})

	m.AddExportFn("fnTest2", func(ctx *quickjsBind.Context, this quickjsBind.Value, args []quickjsBind.Value) quickjsBind.Value {
		if len(args) > 0 {
			//defer args[1].Free()
			fmt.Println(args[0].String())
		}
		return ctx.Null()
	})

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
	rt := quickjsBind.NewRuntime()
	defer rt.Close()

	m := rt.CreateModule("module_test")

	class := m.CreateExportClass("classTest")

	class.SetConstructor(func(ctx *quickjsBind.Context, args []quickjsBind.Value) interface{} {
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

	// Create a new context
	ctx := rt.NewContext()
	defer ctx.Close()

	res, err := ctx.EvalFile("./examples/my_module_class_test.js")
	if err != nil {
		panic(err)
	}
	defer res.Free()
}

func TestCreateModClassObject(t *testing.T) {
	var classObject *ExampleObject

	rt := quickjsBind.NewRuntime()
	defer rt.Close()

	m := rt.CreateModule("module_test")

	class := m.CreateExportClass("classTest")

	class.SetConstructor(func(ctx *quickjsBind.Context, args []quickjsBind.Value) interface{} {
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

	// Create a new context
	ctx := rt.NewContext()
	defer ctx.Close()

	res, err := ctx.EvalFile("./examples/my_module_class_test.js")
	if err != nil {
		panic(err)
	}
	defer res.Free()

	goClassObjectValue := class.CreateGoJsClassObject(ctx.String("test Name 1"), ctx.Int32(23))
	defer goClassObjectValue.Free()

	goVal, err := quickjsBind.GetGoObject[*ExampleObject](goClassObjectValue)

	if err != nil {
		panic(err)
	}

	data, err := json.Marshal(goVal)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}

func TestModuleObject(t *testing.T) {

	polyfill.Polyfill()

	rt := quickjsBind.NewRuntime()
	defer rt.Close()

	sCtx := rt.NewContext()
	defer sCtx.Close()

	m := rt.CreateModule("module_test")

	//obj := sCtx.Object()
	//defer obj.Free()

	m.AddExportObject("object1", sCtx.Int32(32))
	//m.AddExportObject("object2", obj)

	mCtx := rt.NewContext()

	_, err := mCtx.EvalMode(`
import * as m from "module_test";
console.log("js console1:",m.object1)
`, quickjsBind.JS_EVAL_TYPE_MODULE)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	mCtx.Close()

	mCtx = rt.NewContext()

	_, err = mCtx.EvalMode(`
	import * as m from "module_test";
	console.log("js console2:",m.object1)
	//m.object2.name = "change object test1"
	//console.log("js console3:",m.object2.name)
	`, quickjsBind.JS_EVAL_TYPE_MODULE)
	if err != nil {
		fmt.Println(err)
	}
	defer mCtx.Close()

}
