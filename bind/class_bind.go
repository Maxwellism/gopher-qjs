package bind

import (
	"errors"
	"fmt"
	quickjs "github.com/Maxwellism/gopher-qjs/wrap"
	"reflect"
)

var defaultFinalizer = func(obj interface{}) {}

type classOpts struct {
	finalizer     func(obj interface{})
	methodBindMap map[string]string
	fieldBindMap  map[string]string
}

type ClassOpt func(*classOpts)

func WithFinalizer(finalizer func(obj interface{})) ClassOpt {
	return func(opt *classOpts) {
		opt.finalizer = finalizer
	}
}

func WithMethodBindList(methodBindMap map[string]string) ClassOpt {
	return func(opt *classOpts) {
		opt.methodBindMap = methodBindMap
	}
}

func WithFieldBindList(fieldBindMap map[string]string) ClassOpt {
	return func(opt *classOpts) {
		opt.fieldBindMap = fieldBindMap
	}
}

func WrapClass(class *quickjs.JSClass, constructorFn interface{}, opts ...ClassOpt) *quickjs.JSClass {
	fn := reflect.ValueOf(constructorFn)
	if fn.Type().NumOut() != 1 {
		panic(fmt.Sprintf("class[%s] constructor func there must be only one return parameter", class.ClassName))
	}
	// constructor
	class.SetConstructor(func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) interface{} {
		res, err := bindConstructor(fn, args)
		if err != nil {
			panic(err)
		}
		return res
	})
	cOpts := &classOpts{}
	for _, opt := range opts {
		opt(cOpts)
	}

	// finalizer
	if cOpts.finalizer != nil {
		class.SetFinalizer(cOpts.finalizer)
	} else {
		class.SetFinalizer(defaultFinalizer)
	}

	res := fn.Type().Out(0)
	if res.Kind() == reflect.Ptr {
		res = res.Elem()
	}

	// bind field
	bindField(class, res, cOpts)

	// methods
	bindJsClassMethod(class, cOpts)

	return class
}

func bindConstructor(constructorFnType reflect.Value, args []quickjs.Value) (interface{}, error) {
	if len(args) != constructorFnType.Type().NumIn() {
		return nil, errors.New("the number of parameters passed by js is inconsistent with that of go constructor")
	}
	var callArgs []reflect.Value
	for i := 0; i < constructorFnType.Type().NumIn(); i++ {
		val, err := JsValueToGoObject(constructorFnType.Type().In(i), args[i])
		if err != nil {
			return nil, err
		}
		callArgs = append(callArgs, reflect.ValueOf(val))
	}
	res := constructorFnType.Call(callArgs)
	return res[0].Interface(), nil
}

func bindJsClassMethod(class *quickjs.JSClass, opt *classOpts) {
	if opt.methodBindMap != nil {
		for goMethodName, jsMethodName := range opt.methodBindMap {
			class.AddClassFn(jsMethodName, func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
				val, err := this.GetGoClassObject()
				if err != nil {
					panic(err)
				}

				goValue := reflect.ValueOf(val)

				method := goValue.MethodByName(goMethodName)

				if method.Kind() == reflect.Invalid {
					panic("no method found:" + goMethodName)
				}

				callArgs, err := getBindFnArgs(method.Type(), args)
				if err != nil {
					return ctx.Error(err)
				}

				res := goValue.MethodByName(goMethodName).Call(callArgs)
				if len(res) == 0 {
					return ctx.Undefined()
				}
				return GoObjectToJsValue(res[0].Interface(), ctx)
			})
		}
	}

}

func getBindFnArgs(fnType reflect.Type, args []quickjs.Value) ([]reflect.Value, error) {
	if len(args) != fnType.NumIn() {
		return nil, errors.New("the number of parameters passed by js is inconsistent with that of go fn")
	}
	var callArgs []reflect.Value
	for i := 0; i < fnType.NumIn(); i++ {
		val, err := JsValueToGoObject(fnType.In(i), args[i])
		if err != nil {
			return nil, err
		}
		callArgs = append(callArgs, reflect.ValueOf(val))
	}
	return callArgs, nil
}

func bindField(class *quickjs.JSClass, structType reflect.Type, opt *classOpts) {
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		goFieldName := field.Name

		jsFieldName := goFieldName

		if opt.fieldBindMap != nil && opt.fieldBindMap[goFieldName] != "" {
			jsFieldName = opt.fieldBindMap[goFieldName]
		}

		// field set
		class.AddClassSetFn(jsFieldName, func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
			goObject, err := this.GetGoClassObject()
			if err != nil {
				panic(err)
			}

			var structValue reflect.Value

			objType := reflect.TypeOf(goObject)
			if objType.Kind() == reflect.Ptr {
				structValue = reflect.ValueOf(goObject).Elem()
			} else {
				structValue = reflect.ValueOf(&goObject).Elem()
			}

			err = bindFieldSet(structValue, goFieldName, args[0])
			if err != nil {
				return ctx.Error(err)
			}
			return ctx.Undefined()
		})

		// field get
		class.AddClassGetFn(jsFieldName, func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
			goObject, err := this.GetGoClassObject()
			if err != nil {
				panic(err)
			}

			var structValue reflect.Value

			objType := reflect.TypeOf(goObject)
			if objType.Kind() == reflect.Ptr {
				structValue = reflect.ValueOf(goObject).Elem()
			} else {
				structValue = reflect.ValueOf(&goObject).Elem()
			}

			return bindFieldGet(structValue, goFieldName, ctx)
		})
	}
}

func bindFieldGet(structValue reflect.Value, fieldName string, ctx *quickjs.Context) quickjs.Value {
	fieldValue := structValue.FieldByName(fieldName)
	return GoObjectToJsValue(fieldValue.Interface(), ctx)
}

// According to the field name, convert the quickjs.Value to the corresponding go instance
func bindFieldSet(structValue reflect.Value, fieldName string, jsValue quickjs.Value) error {
	fieldValue := structValue.FieldByName(fieldName)
	if !fieldValue.CanSet() {
		return errors.New(fmt.Sprintf("[%s] field is not set", fieldName))
	}
	val, err := JsValueToGoObject(fieldValue.Type(), jsValue)
	if err != nil {
		return err
	}

	fieldValue.Set(reflect.ValueOf(val))

	return nil
}
