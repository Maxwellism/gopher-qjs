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
	ctx    *Context
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
	modName      string
	fnLock       sync.Mutex
	fnIDList     []int32
	classLock    sync.Mutex
	classIDList  []uint32
	exportObject sync.Map
	ctx          *Context
}

func newMod(modName string) *JSModule {
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
	m.fnLock.Lock()
	defer m.fnLock.Unlock()

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

func (m *JSModule) AddExportObject(name string, object Value) {
	m.exportObject.Store(name, &object)
}

//func (m *JSModule) SetExportObject(name string, object Value) {
//	m.exportObject.Store(name, &object)
//}

func (m *JSModule) GetExportObject(name string) *Value {
	if value, ok := m.exportObject.Load(name); ok {
		jsVal, _ := value.(Value)
		return &jsVal
	}
	return nil
}

func (m *JSModule) CreateExportClass(className string) *JSClass {
	m.classLock.Lock()
	defer m.classLock.Unlock()

	jsClass := newModClass(className)
	m.classIDList = append(m.classIDList, jsClass.goClassID)
	return jsClass
}

func (m *JSModule) buildModule(ctx *Context) {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			runtime.Stack(buf, false)
			fmt.Printf("Go panic: %v\n%s", err, buf)
		}
	}()

	if m.ctx == nil {
		m.ctx = ctx
	}

	cStr := C.CString(m.modName)
	defer C.free(unsafe.Pointer(cStr))

	cModule := C.JS_NewCModule(
		ctx.ref,
		cStr,
		(*C.JSModuleInitFunc)(unsafe.Pointer(C.InvokeGoModInit)))

	m.fnLock.Lock()
	for _, id := range m.fnIDList {
		fnInfo := getModFnByID(id)
		goStr := fnInfo.fnName
		cStr1 := C.CString(goStr)
		defer C.free(unsafe.Pointer(cStr1))
		C.JS_AddModuleExport(ctx.ref, cModule, cStr1)
	}
	m.fnLock.Unlock()

	m.classLock.Lock()
	for _, classID := range m.classIDList {
		jsClass := getClassByID(classID)
		goStr := jsClass.ClassName
		cStr1 := C.CString(goStr)
		defer C.free(unsafe.Pointer(cStr1))
		C.JS_AddModuleExport(ctx.ref, cModule, cStr1)
	}
	m.classLock.Unlock()

	m.exportObject.Range(func(key, value any) bool {
		if name, ok := key.(string); ok {
			cStr1 := C.CString(name)
			defer C.free(unsafe.Pointer(cStr1))
			C.JS_AddModuleExport(ctx.ref, cModule, cStr1)
		}
		return true
	})
}
