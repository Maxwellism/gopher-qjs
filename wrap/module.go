package quickjs

/*
#include <stdint.h>
#include "bridge.h"
*/
import "C"
import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

type modFuncEntry struct {
	ctx *Context
	fn  func(ctx *Context, this Value, args []Value) Value
}

var funcModPtrLen int32
var fnModPtrStore = make(map[int32]modFuncEntry)
var moduleList = make(map[string]*JSModule)

func storeFuncModPtr(v modFuncEntry) int32 {
	id := atomic.AddInt32(&funcModPtrLen, 1) - 1
	funcPtrLock.Lock()
	defer funcPtrLock.Unlock()
	fnModPtrStore[id] = v
	return id
}

func restoreFuncModPtr(ptr int32) modFuncEntry {
	funcPtrLock.Lock()
	defer funcPtrLock.Unlock()
	return fnModPtrStore[ptr]
}

type JSModule struct {
	ctx        *Context
	ModuleName string
	//fnIds      []int32
	fnList      []C.JSCFunctionListEntry
	exportFnLen int
}

func (m *JSModule) AddExportFn(fnName string, argLen int, fn func(ctx *Context, this Value, args []Value) Value) {
	moduleFnEntry := modFuncEntry{
		ctx: m.ctx,
		fn:  fn,
	}
	m.exportFnLen += 1
	id := storeFuncModPtr(moduleFnEntry)

	cStr := C.CString(fnName)
	defer C.free(unsafe.Pointer(cStr))
	fmt.Println(id)
	jsFn := C.getJSCFunctionMagicEntry(
		cStr,
		C.int(argLen),
		C.int(id),
		(*C.JSCFunctionMagic)(unsafe.Pointer(C.InvokeGoFn)))

	m.fnList = append(m.fnList, jsFn)
	//m.fnIds = append(m.fnIds, id)
}

func (m *JSModule) BuildModule() {
	cStr := C.CString(m.ModuleName)
	defer C.free(unsafe.Pointer(cStr))
	// JSModuleInitFunc
	jsMod := C.JS_NewCModule(
		m.ctx.ref,
		cStr,
		(*C.JSModuleInitFunc)(unsafe.Pointer(C.InvokeGoInitModule)))

	funcs := (*C.JSCFunctionListEntry)(unsafe.Pointer(&m.fnList[0]))

	C.JS_SetModuleExportList(
		m.ctx.ref,
		jsMod,
		funcs,
		C.int(m.exportFnLen))
}
