package quickjsBind

import (
	"errors"
	quickjs "github.com/Maxwellism/gopher-qjs/wrap"
	"reflect"
)

func WrapFn(goFn interface{}) func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
	goFnValue := reflect.ValueOf(goFn)
	return func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		if len(args) != goFnValue.Type().NumIn() {
			return ctx.ThrowError(errors.New("the number of parameters passed by js is inconsistent with that of go constructor"))
		}
		var callArgs []reflect.Value
		for i := 0; i < goFnValue.Type().NumIn(); i++ {
			val, err := JsValueToGoObject(goFnValue.Type().In(i), args[i])
			if err != nil {
				return ctx.ThrowError(err)
			}
			callArgs = append(callArgs, reflect.ValueOf(val))
		}
		res := goFnValue.Call(callArgs)
		if len(res) == 0 {
			return ctx.Undefined()
		}
		return GoObjectToJsValue(res[0], ctx)
	}
}
