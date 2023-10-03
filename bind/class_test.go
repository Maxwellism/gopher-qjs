package quickjsBind_test

import (
	"fmt"
	quickjs "github.com/Maxwellism/gopher-qjs/bind"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClass(t *testing.T) {
	rt := quickjs.NewRuntime()
	defer rt.Close()

	class := rt.CreateGlobalClass("ClassTest")
	class.SetConstructor(func(ctx *quickjs.Context, args []quickjs.Value) interface{} {
		res := make(map[string]interface{})
		res["name"] = args[0].String()
		res["age"] = args[1].Int32()
		return res
	})

	class.AddClassGetFn("name", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		res, _ := quickjs.GetGoObject[map[string]interface{}](this)
		name, _ := (*res)["name"].(string)
		return ctx.String(name)
	})

	class.AddClassSetFn("name", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		res, _ := quickjs.GetGoObject[map[string]interface{}](this)
		(*res)["name"] = args[0].String()
		return ctx.Undefined()
	})

	class.AddClassGetFn("age", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		res, _ := quickjs.GetGoObject[map[string]interface{}](this)
		age, _ := (*res)["age"].(int32)
		return ctx.Int32(age)
	})
	class.AddClassSetFn("age", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		res, _ := quickjs.GetGoObject[map[string]interface{}](this)
		(*res)["age"] = args[0].Int32()
		return ctx.Undefined()
	})

	class.AddClassFn("sayHello", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		name := this.Get("name").String()
		assert.NotEmpty(t, name)
		fmt.Println(fmt.Sprintf("test name is :%s", name))
		return ctx.Null()
	})

	ctx := rt.NewContext()
	defer ctx.Close()

	// static set
	val := class.GetClassValue()
	if val != nil {
		(*val).Set("Name", ctx.String("class static name test"))
	}

	ret1, err := ctx.Eval(`
let c = new ClassTest("ccc",23);
let b = new ClassTest("bbb",43);

c.sayHello();
b.sayHello();
ClassTest.Name
`)
	fmt.Println(ret1.String())
	assert.NoError(t, err)
	defer ret1.Free()
}
