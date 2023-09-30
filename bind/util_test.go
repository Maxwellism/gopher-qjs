package bind

import (
	"fmt"
	quickjs "github.com/Maxwellism/gopher-qjs/wrap"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestArgTest(t *testing.T) {
	rt := quickjs.NewRuntime()
	defer rt.Close()

	// Create a new context
	ctx := rt.NewContext()

	testParams := []paramData{
		{
			arg:     int(8),
			jsValue: ctx.Int32(32),
		},
		{
			arg:     3.123,
			jsValue: ctx.Float64(32.32),
		},
		{
			arg:     ExampleStruct{Name: "name", Age: 32},
			jsValue: ctx.ParseJSON(`{"Name":"ccc","Age":12}`),
		},
		{
			arg:     []byte{},
			jsValue: ctx.ArrayBuffer([]byte{1, 1, 1, 0}),
		},
		{
			arg:     []string{},
			jsValue: ctx.ParseJSON(`["apple", "banana", "orange"]`),
		},
		{
			arg:     map[string]interface{}{},
			jsValue: ctx.ParseJSON(`{"Name":"ccc","Age":12}`),
		},
	}

	for _, data := range testParams {
		res, err := JsValueToGoObject(reflect.TypeOf(data.arg), data.jsValue)
		defer data.jsValue.Free()

		assert.NoError(t, err)

		fmt.Println(res)
	}

	defer ctx.Close()
}

func TestArgErrorTest(t *testing.T) {
	rt := quickjs.NewRuntime()
	defer rt.Close()

	// Create a new context
	ctx := rt.NewContext()
	defer ctx.Close()

	testParams := []paramData{
		{
			arg:     int(8),
			jsValue: ctx.String("?????"),
		},
	}
	for _, data := range testParams {
		res, err := JsValueToGoObject(reflect.TypeOf(data.arg), data.jsValue)
		defer data.jsValue.Free()
		assert.NoError(t, err)
		fmt.Println(res)
	}
}
