package bind

import (
	"errors"
	"fmt"
	quickjs "github.com/Maxwellism/gopher-qjs/wrap"
	"reflect"
	"sync"
)

var defaultFinalizer = func(obj interface{}) {}

type classOpts struct {
	isCheckArgCount    bool
	finalizer          func(obj interface{})
	methodBindMap      map[string]string
	exportFieldBindMap map[string]string
}

type ClassOpt func(*classOpts)

func WithFinalizer(finalizer func(obj interface{})) ClassOpt {
	return func(opt *classOpts) {
		opt.finalizer = finalizer
	}
}

func WithExportMethodBindList(methodBindMap map[string]string) ClassOpt {
	return func(opt *classOpts) {
		opt.methodBindMap = methodBindMap
	}
}

func WithExportFieldBindList(fieldBindMap map[string]string) ClassOpt {
	return func(opt *classOpts) {
		opt.exportFieldBindMap = fieldBindMap
	}
}

func WithCheckInArgCount(isCheck bool) ClassOpt {
	return func(opt *classOpts) {
		opt.isCheckArgCount = isCheck
	}
}

func WrapClass(class *quickjs.JSClass, constructorFn interface{}, opts ...ClassOpt) *quickjs.JSClass {
	fn := reflect.ValueOf(constructorFn)
	if fn.Type().NumOut() != 1 {
		panic(fmt.Sprintf("class[%s] constructor func there must be only one return parameter", class.ClassName))
	}

	cOpts := &classOpts{}
	for _, opt := range opts {
		opt(cOpts)
	}
	// constructor
	class.SetConstructor(func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) interface{} {
		res, err := bindConstructor(fn, args, cOpts)
		if err != nil {
			panic(err)
		}
		return res
	})

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

func bindConstructor(constructorFnType reflect.Value, args []quickjs.Value, opt *classOpts) (interface{}, error) {
	if opt.isCheckArgCount {
		if len(args) != constructorFnType.Type().NumIn() {
			return nil, errors.New("the number of parameters passed by js is inconsistent with that of go constructor")
		}
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
	var wg sync.WaitGroup
	for goMethodName, jsMethodName := range opt.methodBindMap {
		wg.Add(1)
		if jsMethodName == "" {
			jsMethodName = goMethodName
		}
		go func(jsFnName, goFnName string) {
			class.AddClassFn(jsFnName, func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
				val, err := this.GetBindGoObject()
				if err != nil {
					panic(err)
				}

				goValue := reflect.ValueOf(val)

				method := goValue.MethodByName(goFnName)

				if method.Kind() == reflect.Invalid {
					panic("no method found:" + goFnName)
				}

				callArgs, err := getBindFnArgs(method.Type(), args, opt)
				if err != nil {
					return ctx.Error(err)
				}

				res := goValue.MethodByName(goFnName).Call(callArgs)
				if len(res) == 0 {
					return ctx.Undefined()
				}
				return GoObjectToJsValue(res[0].Interface(), ctx)
			})
		}(jsMethodName, goMethodName)

		wg.Done()
	}
	wg.Wait()
}

func getBindFnArgs(fnType reflect.Type, args []quickjs.Value, opt *classOpts) ([]reflect.Value, error) {
	if opt.isCheckArgCount {
		if len(args) != fnType.NumIn() {
			return nil, errors.New("the number of parameters passed by js is inconsistent with that of go fn")
		}
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
	var wg sync.WaitGroup
	for goFieldName, jsFieldName := range opt.exportFieldBindMap {
		_, ok := structType.FieldByName(goFieldName)
		if !ok {
			panic("not find field by:" + goFieldName)
		}

		if jsFieldName == "" {
			jsFieldName = goFieldName
		}
		// field set
		go func(jsField, goField string) {
			wg.Add(1)
			class.AddClassSetFn(jsField, func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
				goObject, err := this.GetBindGoObject()
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

				err = bindFieldSet(structValue, goField, args[0])
				if err != nil {
					return ctx.Error(err)
				}
				return ctx.Undefined()
			})
			wg.Done()
		}(jsFieldName, goFieldName)

		// field get
		go func(jsField, goField string) {
			wg.Add(1)
			class.AddClassGetFn(jsField, func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
				goObject, err := this.GetBindGoObject()
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

				return bindFieldGet(structValue, goField, ctx)
			})
			wg.Done()
		}(jsFieldName, goFieldName)
	}
	wg.Wait()
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
