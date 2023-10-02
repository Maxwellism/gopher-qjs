package quickjsWrap_test

import (
	"fmt"
	quickjs "github.com/Maxwellism/gopher-qjs/bind"
	"github.com/Maxwellism/gopher-qjs/wrap"
	"testing"
)

type paramData struct {
	arg     interface{}
	jsValue quickjs.Value
}

type ExampleStruct struct {
	Name string `json:"name,omitempty"`
	Age  int32  `json:"age,omitempty"`
}

// test
func (s *ExampleStruct) QJSGetTest() {
	fmt.Println(fmt.Sprintf("Method Test\nGet test name: %s\nGet test age: %d\n", s.Name, s.Age))
}

func TestClassBind(t *testing.T) {
	// Create a new runtime
	rt := quickjs.NewRuntime()
	defer rt.Close()

	var goObject *ExampleStruct

	class := rt.CreateGlobalClass("ClassTest")
	class = quickjsWrap.WrapClass(
		class,
		quickjsWrap.WithBindConstructorFn(
			func() *ExampleStruct {
				goObject = &ExampleStruct{}
				return goObject
			}),
		quickjsWrap.WithExportMethodBindList(map[string]string{
			"QJSGetTest": "GetTest",
		}),
		quickjsWrap.WithExportFieldBindList(map[string]string{
			"Name": "name",
			"Age":  "Age",
		}),
	)
	// Create a new context
	ctx := rt.NewContext()
	defer ctx.Close()

	ret, err := ctx.Eval(`
let c = new ClassTest();
console.log("name:",c.name)
console.log("age:",c.Age)
c.name = "class Name Test"
c.Age = 23
c.GetTest()
`)
	if err != nil {
		panic(err)
	}
	//fmt.Println(ret.String())
	defer ret.Free()
	fmt.Println(goObject.Name)
	fmt.Println(goObject.Age)
}

func TestClassConstructor(t *testing.T) {
	// Create a new runtime
	rt := quickjs.NewRuntime()
	defer rt.Close()

	class1 := rt.CreateGlobalClass("ClassTest1")
	class1 = quickjsWrap.WrapClass(
		class1,
		quickjsWrap.WithBindConstructorFn(
			func() *ExampleStruct {
				fmt.Println("ClassTest1")
				return &ExampleStruct{}
			}),
		quickjsWrap.WithExportMethodBindList(map[string]string{
			"QJSGetTest": "GetTest",
		}),
		quickjsWrap.WithExportFieldBindList(map[string]string{
			"Name": "name",
			"Age":  "Age",
		}),
	)

	class2 := rt.CreateGlobalClass("ClassTest2")
	class2 = quickjsWrap.WrapClass(
		class2,
		quickjsWrap.WithQjsConstructorFn(func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) ExampleStruct {
			fmt.Println("ClassTest2")
			return ExampleStruct{}
		}),
		quickjsWrap.WithExportMethodBindList(map[string]string{
			"QJSGetTest": "GetTest",
		}),
		quickjsWrap.WithExportFieldBindList(map[string]string{
			"Name": "name",
			"Age":  "Age",
		}),
	)

	// Create a new context
	ctx := rt.NewContext()
	defer ctx.Close()

	ret, err := ctx.Eval(`
new ClassTest1();
new ClassTest2();
`)
	if err != nil {
		panic(err)
	}
	//fmt.Println(ret.String())
	defer ret.Free()
}
