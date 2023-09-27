package quickjs_test

import (
	"encoding/json"
	"fmt"
	quickjs "github.com/Maxwellism/gopher-qjs/wrap"
	"testing"
)

type ExampleObject struct {
	Name string
	Age  int32
}

func TestNewClass(t *testing.T) {
	rt := quickjs.NewRuntime()
	defer rt.Close()

	var classObject *ExampleObject
	class := rt.CreateGlobalClass("classTest")
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

	class.AddClassGetFn("Name", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		return ctx.String(classObject.Name)
	})

	class.AddClassSetFn("Name", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		classObject.Name = args[0].String() + "323"
		return ctx.Undefined()
	})

	class.AddClassGetFn("Age", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		return ctx.Int32(classObject.Age)
	})

	// Create a new context
	ctx := rt.NewContext()
	defer ctx.Close()

	ctx.Globals().Set("getGoObject", ctx.Function(func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		arg := args[0]
		goValue, err := arg.GetGoObject()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(goValue)
		return ctx.Undefined()
	}))

	res, err := ctx.Eval(`
let c = new classTest("class test",32);
console.log(c.Name)
c.Name = "class test1"
console.log(c.Name)
getGoObject(c)
//console.log(c.testClassFn())
`)
	defer res.Free()
	rt.RunGC()
	if err != nil {
		fmt.Println(err)
	}
}

func TestNewClassCallConstructor(t *testing.T) {
	rt := quickjs.NewRuntime()
	defer rt.Close()

	var classObject *ExampleObject
	class := rt.CreateGlobalClass("classTest")
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

	class.AddClassGetFn("Name", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		return ctx.String(classObject.Name)
	})

	class.AddClassSetFn("Name", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		classObject.Name = args[0].String() + "323"
		return ctx.Undefined()
	})

	class.AddClassGetFn("Age", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		return ctx.Int32(classObject.Age)
	})

	// Create a new context
	ctx := rt.NewContext()
	defer ctx.Close()

	goClassObjectValue := class.CreateGoJsClassObject(ctx.String("test Name 1"), ctx.Int32(23))
	defer goClassObjectValue.Free()

	goVal, err := goClassObjectValue.GetGoObject()

	if err != nil {
		panic(err)
	}

	data, err := json.Marshal(goVal)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}
