package quickjsWrap

import (
	"errors"
	"fmt"
	quickjs "github.com/Maxwellism/gopher-qjs/bind"
	"reflect"
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

type classConstructorOpts struct {
	constructorFn *reflect.Value
	bindFn        *reflect.Value
}

type ConstructorOpt func(opts *classConstructorOpts)

func WithBindConstructorFn(constructorFn interface{}) ConstructorOpt {
	return func(opts *classConstructorOpts) {
		fn := reflect.ValueOf(constructorFn)

		if fn.Type().Kind() != reflect.Func {
			panic("constructorFn parameters are not func")
		}
		if fn.Type().NumOut() != 1 {
			panic(fmt.Sprintf("constructor func there must be only one return parameter"))
		}

		opts.bindFn = &fn
	}
}

func WithQjsConstructorFn(constructorFn interface{}) ConstructorOpt {
	return func(opts *classConstructorOpts) {

		fn := reflect.ValueOf(constructorFn)

		if fn.Type().Kind() != reflect.Func {
			panic("constructorFn parameters are not func")
		}
		if fn.Type().NumOut() != 1 {
			panic(fmt.Sprintf("constructor func there must be only one return parameter"))
		}

		if fn.Type().NumIn() != 2 {
			panic(
				fmt.Sprintf(
					"constructor the input parameters of the method must be two, and the first is [%s] and the second is [%s]",
					"*github.com/Maxwellism/gopher-qjs/bind.Context",
					"github.com/Maxwellism/gopher-qjs/bind.Value",
				))
		}
		theFirstArg := fn.Type().In(0)
		if theFirstArg.Kind() != reflect.Ptr {
			panic("constructor the first parameter type is not *github.com/Maxwellism/gopher-qjs/bind.Context")
		}
		theFirstArg = theFirstArg.Elem()
		if fmt.Sprintf("%s.%s", theFirstArg.PkgPath(), theFirstArg.Name()) != "github.com/Maxwellism/gopher-qjs/bind.Context" {
			panic("constructor the first parameter type is not *github.com/Maxwellism/gopher-qjs/bind.Context")
		}
		theSecondArg := fn.Type().In(1).Elem()
		//fmt.Println(theSecondArg.Elem().PkgPath())
		if fmt.Sprintf("%s.%s", theSecondArg.PkgPath(), theSecondArg.Name()) != "github.com/Maxwellism/gopher-qjs/bind.Value" {
			panic("constructor the second parameter type is not github.com/Maxwellism/gopher-qjs/bind.Value")
		}
		opts.constructorFn = &fn
	}
}

func WithSetCheckInArgCount() ClassOpt {
	return func(opt *classOpts) {
		opt.isCheckArgCount = true
	}
}

func WrapClass(class *quickjs.JSClass, classConstructorOptFn ConstructorOpt, opts ...ClassOpt) *quickjs.JSClass {
	cOpts := &classOpts{}
	for _, opt := range opts {
		opt(cOpts)
	}

	classConstructor := &classConstructorOpts{}

	classConstructorOptFn(classConstructor)

	var fn reflect.Value

	if classConstructor.bindFn != nil {
		fn = *classConstructor.bindFn
	} else {
		fn = *classConstructor.constructorFn
	}

	// classConstructor
	class.SetConstructor(func(ctx *quickjs.Context, args []quickjs.Value) interface{} {
		if classConstructor.bindFn != nil {
			res, err := bindConstructor(fn, args, cOpts)
			if err != nil {
				panic(err)
			}
			return res
		} else {
			res := classConstructor.constructorFn.Call(
				[]reflect.Value{
					reflect.ValueOf(ctx),
					reflect.ValueOf(args),
				},
			)
			return res[0].Interface()
		}
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
	bindJsClassMethods(class, cOpts)

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

func bindJsClassMethods(class *quickjs.JSClass, opt *classOpts) {
	for goMethodName, jsMethodName := range opt.methodBindMap {
		if jsMethodName == "" {
			jsMethodName = goMethodName
		}
		bindJsClassMethod(class, opt, jsMethodName, goMethodName)
	}
}

func bindJsClassMethod(class *quickjs.JSClass, opt *classOpts, jsFnName, goFnName string) {
	class.AddClassFn(jsFnName, func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		val, err := this.GetBindGoObject()
		if err != nil {
			panic(err)
		}

		goValue := reflect.ValueOf(val)

		method := goValue.MethodByName(goFnName)

		if method.Kind() == reflect.Invalid {
			panic("[" + class.ClassName + "]no method found:" + goFnName)
		}

		callArgs, err := getBindFnArgs(method.Type(), args, opt)
		if err != nil {
			return ctx.ThrowError(err)
		}

		res := goValue.MethodByName(goFnName).Call(callArgs)
		if len(res) == 0 {
			return ctx.Undefined()
		}
		return GoObjectToJsValue(res[0].Interface(), ctx)
	})
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
		}(jsFieldName, goFieldName)

		// field get
		go func(jsField, goField string) {
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
		}(jsFieldName, goFieldName)
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
