package quickjsBind

/*
#include <stdint.h>
#include "bridge.h"
*/
import "C"
import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

type moduleFuncEntry struct {
	fnName string
	fn     func(ctx *Context, this Value, args []Value) Value
}

var goModLock sync.Mutex
var moduleMap = make(map[string]*JSModule)

var goModFnPtrLen int32
var goModFnLock sync.Mutex
var goModFnPtrStore = make(map[int32]*moduleFuncEntry)

func putModFn(v *moduleFuncEntry) int32 {
	goModFnLock.Lock()
	defer goModFnLock.Unlock()
	id := atomic.AddInt32(&goModFnPtrLen, 1) - 1
	goModFnPtrStore[id] = v
	return id
}

func getModFnByID(id int32) *moduleFuncEntry {
	goModFnLock.Lock()
	defer goModFnLock.Unlock()
	return goModFnPtrStore[id]
}

type JSModule struct {
	modName     string
	fnIDList    []int32
	classIDList []uint32
}

func NewMod(modName string) *JSModule {
	m := &JSModule{
		fnIDList: []int32{},
		modName:  modName,
	}
	goModLock.Lock()
	moduleMap[modName] = m
	goModLock.Unlock()
	return m
}

func (m *JSModule) storeFuncModPtr(v *moduleFuncEntry) int32 {
	id := putModFn(v)
	m.fnIDList = append(m.fnIDList, id)
	return id
}

func (m *JSModule) AddExportFn(fnName string, fn func(ctx *Context, this Value, args []Value) Value) {
	mFnEntry := &moduleFuncEntry{
		fn:     fn,
		fnName: fnName,
	}
	m.storeFuncModPtr(mFnEntry)
}

func (m *JSModule) buildModule(ctx *C.JSContext) {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			runtime.Stack(buf, false)
			fmt.Printf("Go panic: %v\n%s", err, buf)
		}
	}()
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

	for _, id := range m.fnIDList {
		fnInfo := getModFnByID(id)
		goStr := fnInfo.fnName
		cStr1 := C.CString(goStr)
		defer C.free(unsafe.Pointer(cStr1))
		C.JS_AddModuleExport(ctx, cmod, cStr1)
	}

	for _, classID := range m.classIDList {
		jsClass := getClassByID(classID)
		goStr := jsClass.ClassName
		cStr1 := C.CString(goStr)
		defer C.free(unsafe.Pointer(cStr1))
		C.JS_AddModuleExport(ctx, cmod, cStr1)
	}

}

func (m *JSModule) CreateExportClass(className string) *JSClass {
	jsClass := newModClass(className)
	m.classIDList = append(m.classIDList, jsClass.goClassID)
	return jsClass
}