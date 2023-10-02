package quickjsBind

import (
	"fmt"
	"runtime"
	"runtime/cgo"
	"sync"
	"sync/atomic"
	"unsafe"
)

/*
#include <stdint.h>
#include "bridge.h"
#include "cutils.h"
*/
import "C"

type funcEntry struct {
	ctx     *Context
	fn      func(ctx *Context, this Value, args []Value) Value
	asyncFn func(ctx *Context, this Value, promise Value, args []Value) Value
}

var funcPtrLen int64
var funcPtrLock sync.Mutex
var funcPtrStore = make(map[int64]funcEntry)
var funcPtrClassID C.JSClassID

func init() {
	C.JS_NewClassID(&funcPtrClassID)
}

func storeFuncPtr(v funcEntry) int64 {
	id := atomic.AddInt64(&funcPtrLen, 1) - 1
	funcPtrLock.Lock()
	defer funcPtrLock.Unlock()
	funcPtrStore[id] = v
	return id
}

func restoreFuncPtr(ptr int64) funcEntry {
	funcPtrLock.Lock()
	defer funcPtrLock.Unlock()
	return funcPtrStore[ptr]
}

//func freeFuncPtr(ptr int64) {
//	funcPtrLock.Lock()
//	defer funcPtrLock.Unlock()
//	delete(funcPtrStore, ptr)
//}

//export goProxy
func goProxy(ctx *C.JSContext, thisVal C.JSValueConst, argc C.int, argv *C.JSValueConst) C.JSValue {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			runtime.Stack(buf, false)
			fmt.Printf("Go panic: %v\n%s", err, buf)
		}
	}()
	// https://github.com/golang/go/wiki/cgo#turning-c-arrays-into-go-slices
	refs := unsafe.Slice(argv, argc) // Go 1.17 and later

	id := C.int64_t(0)
	C.JS_ToInt64(ctx, &id, refs[0])

	entry := restoreFuncPtr(int64(id))

	args := make([]Value, len(refs)-1)
	for i := 0; i < len(args); i++ {
		args[i].ctx = entry.ctx
		args[i].ref = refs[1+i]
	}

	result := entry.fn(entry.ctx, Value{ctx: entry.ctx, ref: thisVal}, args)

	return result.ref
}

//export goAsyncProxy
func goAsyncProxy(ctx *C.JSContext, thisVal C.JSValueConst, argc C.int, argv *C.JSValueConst) C.JSValue {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			runtime.Stack(buf, false)
			fmt.Printf("Go panic: %v\n%s", err, buf)
		}
	}()
	// https://github.com/golang/go/wiki/cgo#turning-c-arrays-into-go-slices
	refs := unsafe.Slice(argv, argc) // Go 1.17 and later

	id := C.int64_t(0)
	C.JS_ToInt64(ctx, &id, refs[0])

	entry := restoreFuncPtr(int64(id))

	args := make([]Value, len(refs)-1)
	for i := 0; i < len(args); i++ {
		args[i].ctx = entry.ctx
		args[i].ref = refs[1+i]
	}
	promise := args[0]

	result := entry.asyncFn(entry.ctx, Value{ctx: entry.ctx, ref: thisVal}, promise, args[1:])
	return result.ref

}

//export goInterruptHandler
func goInterruptHandler(rt *C.JSRuntime, handlerArgs unsafe.Pointer) C.int {
	handlerArgsStruct := (*C.handlerArgs)(handlerArgs)

	hFn := cgo.Handle(handlerArgsStruct.fn)
	hFnValue := hFn.Value().(InterruptHandler)
	// defer hFn.Delete()

	return C.int(hFnValue())
}

//export goModFnHandle
func goModFnHandle(ctx *C.JSContext, thisVal C.JSValueConst, argc C.int, argv *C.JSValueConst, magic int) C.JSValue {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			runtime.Stack(buf, false)
			fmt.Printf("Go panic: %v\n%s", err, buf)
		}
	}()
	refs := unsafe.Slice(argv, argc) // Go 1.17 and later

	id := int32(magic)

	entry := getModFnByID(id)

	if entry == nil {
		panic(fmt.Sprintf("not find magic id is %d func", id))
	}

	crt := C.JS_GetRuntime(ctx)

	goRuntime := &Runtime{
		ref:  crt,
		loop: NewLoop(),
	}

	goContext := &Context{
		ref:     ctx,
		runtime: goRuntime}

	args := make([]Value, len(refs))
	for i := 0; i < len(args); i++ {
		args[i].ctx = goContext
		args[i].ref = refs[i]
	}

	result := entry.fn(goContext, Value{ctx: goContext, ref: thisVal}, args)

	return result.ref
}

//export goClassFnHandle
func goClassFnHandle(ctx *C.JSContext, thisVal C.JSValueConst, argc C.int, argv *C.JSValueConst, magic int) C.JSValue {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			runtime.Stack(buf, false)
			fmt.Printf("Go panic: %v\n%s", err, buf)
		}
	}()
	refs := unsafe.Slice(argv, argc) // Go 1.17 and later

	goClassFnId := int32(magic)

	jsGoClassFn := getClassFnByID(goClassFnId)

	if jsGoClassFn == nil {
		return C.JS_NewUndefined()
	}

	args := make([]Value, len(refs))
	for i := 0; i < len(args); i++ {
		args[i].ctx = jsGoClassFn.ctx
		args[i].ref = refs[i]
	}

	v := jsGoClassFn.fn(jsGoClassFn.ctx, Value{ctx: jsGoClassFn.ctx, ref: thisVal}, args)

	return v.ref
}

//export goClassGetFnHandle
func goClassGetFnHandle(ctx *C.JSContext, thisVal C.JSValueConst, argc C.int, argv *C.JSValueConst, magic int) C.JSValue {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			runtime.Stack(buf, false)
			fmt.Printf("Go panic: %v\n%s", err, buf)
		}
	}()
	refs := unsafe.Slice(argv, argc) // Go 1.17 and later

	goClassFnId := int32(magic)

	jsGoClassFieldGetFn := jsClassFieldFnPtrStore[goClassFnId]

	if jsGoClassFieldGetFn == nil {
		return C.JS_NewUndefined()
	}

	args := make([]Value, len(refs))
	for i := 0; i < len(args); i++ {
		args[i].ctx = jsGoClassFieldGetFn.ctx
		args[i].ref = refs[i]
	}

	v := jsGoClassFieldGetFn.getFn(jsGoClassFieldGetFn.ctx, Value{ctx: jsGoClassFieldGetFn.ctx, ref: thisVal}, args)

	return v.ref
}

//export goClassSetFnHandle
func goClassSetFnHandle(ctx *C.JSContext, thisVal C.JSValueConst, argc C.int, argv *C.JSValueConst, magic int) C.JSValue {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			runtime.Stack(buf, false)
			fmt.Printf("Go panic: %v\n%s", err, buf)
		}
	}()
	refs := unsafe.Slice(argv, argc) // Go 1.17 and later

	goClassFnId := int32(magic)

	jsGoClassFieldSetFn := jsClassFieldFnPtrStore[goClassFnId]

	if jsGoClassFieldSetFn == nil {
		return C.JS_NewUndefined()
	}

	args := make([]Value, len(refs))
	for i := 0; i < len(args); i++ {
		args[i].ctx = jsGoClassFieldSetFn.ctx
		args[i].ref = refs[i]
	}

	v := jsGoClassFieldSetFn.setFn(jsGoClassFieldSetFn.ctx, Value{ctx: jsGoClassFieldSetFn.ctx, ref: thisVal}, args)

	return v.ref
}

//export goClassConstructorHandle
func goClassConstructorHandle(ctx *C.JSContext, newTarget C.JSValueConst, argc C.int, argv *C.JSValueConst, magic int) C.int32_t {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			runtime.Stack(buf, false)
			fmt.Printf("Go panic: %v\n%s", err, buf)
		}
	}()
	refs := unsafe.Slice(argv, argc) // Go 1.17 and later

	goClassId := uint32(magic)

	jsGoClass := getClassByID(goClassId)

	if jsGoClass == nil {
		return C.int32_t(-1)
	}

	args := make([]Value, len(refs))
	for i := 0; i < len(args); i++ {
		args[i].ctx = jsGoClass.ctx
		args[i].ref = refs[i]
	}

	v := jsGoClass.constructorFn(jsGoClass.ctx, Value{ctx: jsGoClass.ctx, ref: newTarget}, args)
	objectID := pushGoObjectByJS(v)

	return C.int32_t(objectID)
}

//export goFinalizerHandle
func goFinalizerHandle(goClassID C.int, goObjectID C.int32_t) {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			runtime.Stack(buf, false)
			fmt.Printf("Go panic: %v\n%s", err, buf)
		}
	}()

	objectID := int32(goObjectID)

	classID := uint32(goClassID)

	jClass := getClassByID(classID)
	jClass.finalizerFn(getGoObjectByID(objectID))

	deleteGoObjectByID(objectID)

}

//export registerGoClassHandle
func registerGoClassHandle(ctx *C.JSContext) {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			runtime.Stack(buf, false)
			fmt.Printf("Go panic: %v\n%s", err, buf)
		}
	}()
	jsGlobalClassLock.Lock()
	defer jsGlobalClassLock.Unlock()
	for _, jsClass := range jsGlobalClassIDMap {
		goClassBuild(ctx, nil, jsClass)
	}
}

func goClassBuild(ctx *C.JSContext, m *C.JSModuleDef, jsClass *JSClass) {
	cClassID := C.JSClassID(jsClass.goClassID)
	def := C.JSClassDef{
		finalizer: (*C.JSClassFinalizer)(C.goFinalizer),
	}

	goClassName := C.CString(jsClass.ClassName)
	defer C.free(unsafe.Pointer(goClassName))

	def.class_name = goClassName
	if ctx == nil {
		panic(fmt.Sprintf("go class %s ctx point is null", jsClass.ClassName))
	}
	C.JS_NewClass(C.JS_GetRuntime(ctx), cClassID, &def)

	goProto := C.JS_NewObject(ctx)

	goClassConstructor := C.JS_NewCFunctionMagic(
		ctx,
		(*C.JSCFunctionMagic)(unsafe.Pointer(C.goClassConstructor)),
		goClassName,
		0,
		C.JS_CFUNC_constructor_magic,
		C.int(cClassID))

	crt := C.JS_GetRuntime(ctx)

	goRuntime := &Runtime{
		ref:  crt,
		loop: NewLoop(),
	}
	jsClass.ctx = &Context{ref: ctx, runtime: goRuntime}
	jsClass.constructorFnObj = &Value{ctx: jsClass.ctx, ref: goClassConstructor}

	jsClass.fnLock.Lock()
	for _, fnID := range jsClass.fnIds {
		goFnInfo := getClassFnByID(fnID)

		goFnInfo.ctx = jsClass.ctx

		goClassFnName := C.CString(goFnInfo.fnName)
		defer C.free(unsafe.Pointer(goClassFnName))

		goFnObj := C.JS_NewCFunctionMagic(
			ctx,
			(*C.JSCFunctionMagic)(unsafe.Pointer(C.InvokeGoClassFn)),
			goClassFnName,
			0,
			C.JS_CFUNC_generic_magic,
			C.int(fnID))

		C.JS_SetPropertyStr(ctx, goProto, goClassFnName, goFnObj)
	}
	jsClass.fnLock.Unlock()

	jsClass.fieldLock.Lock()
	for fieldName, id := range jsClass.fieldFn {

		fieldInfo := jsClassFieldFnPtrStore[*id]
		fieldInfo.ctx = jsClass.ctx

		goClassFieldName := C.CString(fieldName)
		defer C.free(unsafe.Pointer(goClassFieldName))

		goGetFnObj := C.JS_NewCFunctionMagic(
			ctx,
			(*C.JSCFunctionMagic)(unsafe.Pointer(C.InvokeGoClassGetFn)),
			goClassFieldName,
			0,
			C.JS_CFUNC_generic_magic,
			C.int(*id))
		goSetFnObj := C.JS_NewCFunctionMagic(
			ctx,
			(*C.JSCFunctionMagic)(unsafe.Pointer(C.InvokeGoClassSetFn)),
			goClassFieldName,
			0,
			C.JS_CFUNC_generic_magic,
			C.int(*id))

		fieldNameAtom := C.JS_NewAtom(ctx, goClassFieldName)
		C.JS_DefinePropertyGetSet(ctx, goProto, fieldNameAtom, goGetFnObj, goSetFnObj, C.JS_PROP_CONFIGURABLE)
	}
	jsClass.fieldLock.Unlock()

	if m != nil {
		C.JS_SetClassProto(ctx, cClassID, goProto)
		C.JS_SetConstructor(ctx, goClassConstructor, goProto)
		C.JS_SetModuleExport(ctx, m, goClassName, goClassConstructor)
	} else {
		C.JS_SetClassProto(ctx, cClassID, goProto)
		C.JS_NewGlobalCConstructorHandle(ctx, goClassConstructor, goClassName, goProto)
	}
}

//export GoInitModule
func GoInitModule(ctx *C.JSContext, m *C.JSModuleDef) C.int {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			runtime.Stack(buf, false)
			fmt.Printf("Go panic: %v\n%s", err, buf)
		}
	}()
	jsAtom := C.JS_GetModuleName(ctx, m)
	a := Atom{ctx: &Context{ref: ctx}, ref: jsAtom}
	jsMod := moduleMap[a.String()]

	jsMod.fnLock.Lock()
	for _, id := range jsMod.fnIDList {

		fnInfo := getModFnByID(id)

		goStr := fnInfo.fnName

		cStr := C.CString(goStr)
		defer C.free(unsafe.Pointer(cStr)) // 释放内存

		val := C.JS_NewCFunctionMagic(
			ctx,
			(*C.JSCFunctionMagic)(unsafe.Pointer(C.InvokeGoModFn)),
			cStr,
			0,
			C.JS_CFUNC_generic_magic,
			C.int(id))

		C.JS_SetModuleExport(ctx, m, cStr, val)
	}
	jsMod.fnLock.Unlock()

	jsMod.classLock.Lock()

	for _, goClassID := range jsMod.classIDList {
		jsClass := getClassByID(goClassID)
		goClassBuild(ctx, m, jsClass)
	}

	jsMod.classLock.Unlock()
	return C.int(0)
}
