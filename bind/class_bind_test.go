package bind

import (
	"fmt"
	quickjs "github.com/Maxwellism/gopher-qjs/wrap"
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
	class = WrapClass(
		class,
		func() *ExampleStruct {
			goObject = &ExampleStruct{}
			return goObject
		},
		WithExportMethodBindList(map[string]string{
			"QJSGetTest": "GetTest",
		}),
		WithExportFieldBindList(map[string]string{
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
