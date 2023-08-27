package register

import (
	quickjs "github.com/Maxwellism/gopher-qjs/wrapper_qjs"
)

type QJSFnType func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value

var GlobalsFn = map[string]QJSFnType{}

var NameSpaceFnList = map[string]map[string]QJSFnType{}

func QJSRegisterSpaceNameFn(namespace string, qjsFnName string, qjsFn QJSFnType) {
	if namespace == "" {
		panic("namespace is empty!")
	}
	if namespaceInfo := NameSpaceFnList[namespace]; namespaceInfo != nil {
		namespaceInfo[qjsFnName] = qjsFn
	} else {
		namespaceInfo = map[string]QJSFnType{}
		namespaceInfo[qjsFnName] = qjsFn

		NameSpaceFnList[namespace] = namespaceInfo
	}
}

func QJSRegisterGlobalFn(qjsFnName string, qjsFn QJSFnType) {
	GlobalsFn[qjsFnName] = qjsFn
}

func RuntimeWrapper(ctx *quickjs.Context) {
	//polyfill.InjectAll(ctx)
	for jsFnName, qjsFn := range GlobalsFn {
		ctx.Globals().Set(jsFnName, ctx.Function(qjsFn))
	}
	for namespace, namespaceInfo := range NameSpaceFnList {
		register := ctx.Object()
		for qjsFnName, namespaceQJSFn := range namespaceInfo {
			register.Set(qjsFnName, ctx.Function(namespaceQJSFn))
		}
		ctx.Globals().Set(namespace, register)
	}
}
