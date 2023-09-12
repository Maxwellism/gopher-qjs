package quickjs

import (
	"fmt"
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
	refs := unsafe.Slice(argv, argc) // Go 1.17 and later

	id := int32(magic)

	entry := restoreFuncModPtr(id)

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
	return C.JS_NewNull()
}

//export goClassGetFnHandle
func goClassGetFnHandle(ctx *C.JSContext, thisVal C.JSValueConst, argc C.int, argv *C.JSValueConst, magic int) C.JSValue {
	return C.JS_NewNull()
}

//export goClassSetFnHandle
func goClassSetFnHandle(ctx *C.JSContext, thisVal C.JSValueConst, argc C.int, argv *C.JSValueConst, magic int) C.JSValue {
	return C.JS_NewNull()
}

//export GoInitModule
func GoInitModule(ctx *C.JSContext, m *C.JSModuleDef) C.int {
	jsAtom := C.JS_GetModuleName(ctx, m)
	a := Atom{ctx: &Context{ref: ctx}, ref: jsAtom}
	jsMod := moduleMap[a.String()]

	for _, id := range jsMod.ids {

		fnInfo := restoreFuncModPtr(id)

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
	return C.int(0)
}
