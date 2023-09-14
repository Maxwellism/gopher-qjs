package quickjs

/*
#include <stdint.h>
#include "bridge.h"
*/
import "C"
import (
	"sync"
	"sync/atomic"
	"unsafe"
)

type moduleFuncEntry struct {
	fnName string
	fn     func(ctx *Context, this Value, args []Value) Value
}

var moduleMap = make(map[string]*JSModule)

var goModFnPtrLen int32
var goModLock sync.Mutex
var goModFnLock sync.Mutex
var goModFnPtrStore = make(map[int32]*moduleFuncEntry)

type JSModule struct {
	modName string
	ids     []int32
}

func NewMod(modName string) *JSModule {
	m := &JSModule{
		ids:     []int32{},
		modName: modName,
	}
	goModLock.Lock()
	moduleMap[modName] = m
	goModLock.Unlock()
	return m
}

func (m *JSModule) storeFuncModPtr(v *moduleFuncEntry) int32 {
	id := atomic.AddInt32(&goModFnPtrLen, 1) - 1
	goModFnLock.Lock()
	defer goModFnLock.Unlock()
	goModFnPtrStore[id] = v
	m.ids = append(m.ids, id)
	return id
}

func restoreFuncModPtr(ptr int32) *moduleFuncEntry {
	goModFnLock.Lock()
	defer goModFnLock.Unlock()
	return goModFnPtrStore[ptr]
}

func (m *JSModule) AddExportFn(fnName string, fn func(ctx *Context, this Value, args []Value) Value) {
	mFnEntry := &moduleFuncEntry{
		fn:     fn,
		fnName: fnName,
	}
	m.storeFuncModPtr(mFnEntry)
}

func (m *JSModule) buildModule(ctx *C.JSContext) {
	cStr := C.CString(m.modName)
	defer C.free(unsafe.Pointer(cStr))
	// JSModuleInitFunc
	if ctx == nil {
		panic("quickjs JSContext is null")
	}
	cmod := C.JS_NewCModule(
		ctx,
		cStr,
		(*C.JSModuleInitFunc)(unsafe.Pointer(C.InvokeGoModInit)))

	for _, id := range m.ids {
		fnInfo := restoreFuncModPtr(id)
		goStr := fnInfo.fnName
		cStr1 := C.CString(goStr)
		defer C.free(unsafe.Pointer(cStr1))
		C.JS_AddModuleExport(ctx, cmod, cStr1)
	}

}
