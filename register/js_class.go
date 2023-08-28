package register

import "github.com/buke/quickjs-go"

type QjsClassConstructor func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value)

type JsClass struct {
	Constructor  QjsClassConstructor
	MountMethods map[string]QJSFnType
}

var jsClassList = map[string]*JsClass{}

func SetConstructorClass(className string, constructor QjsClassConstructor) {
	if jsClassList[className] == nil {
		jsClassList[className] = &JsClass{Constructor: constructor}
	} else {
		jsClassList[className].Constructor = constructor
	}
}

func SetClassMethod(className, method string, qjsFn QJSFnType) {
	if jsClassList[className] == nil {
		jsClassList[className] = &JsClass{
			MountMethods: map[string]QJSFnType{},
		}
	} else if jsClassList[className].MountMethods == nil {
		jsClassList[className].MountMethods = map[string]QJSFnType{}
	}
	jsClassList[className].MountMethods[method] = qjsFn
}
