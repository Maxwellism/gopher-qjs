package quickjsBind_test

import (
	"fmt"
	"github.com/Maxwellism/gopher-qjs/bind"
	json "github.com/json-iterator/go"
	"testing"
)

func TestModule(t *testing.T) {

	m := quickjsBind.NewMod("module_test")
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

	rt := quickjsBind.NewRuntime()
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

	m := quickjsBind.NewMod("module_test")

	class := m.CreateExportClass("classTest")

	class.SetConstructor(func(ctx *quickjsBind.Context, this quickjsBind.Value, args []quickjsBind.Value) interface{} {
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

	rt := quickjsBind.NewRuntime()
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

func TestCreateModClassObject(t *testing.T) {
	var classObject *ExampleObject

	m := quickjsBind.NewMod("module_test")

	class := m.CreateExportClass("classTest")

	class.SetConstructor(func(ctx *quickjsBind.Context, this quickjsBind.Value, args []quickjsBind.Value) interface{} {
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

	rt := quickjsBind.NewRuntime()
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
