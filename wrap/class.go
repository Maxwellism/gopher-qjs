package quickjs

/*
#include <stdint.h>
#include "bridge.h"
*/
import "C"
import (
	"sync"
	"sync/atomic"
)

var jsClassLock sync.Mutex
var jsClassIDMap = make(map[uint32]*JSClass)

var jsGlobalClassIDMap = make(map[uint32]*JSClass)

var jsClassMapGoObjectPtrLen int32
var jsClassMapGoObjectLock sync.Mutex
var jsClassMapGoObject = make(map[int32]interface{})

func storeGoObjectPtr(v interface{}) int32 {
	id := atomic.AddInt32(&jsClassMapGoObjectPtrLen, 1) - 1
	jsClassMapGoObjectLock.Lock()
	defer jsClassMapGoObjectLock.Unlock()
	jsClassMapGoObject[id] = v
	return id
}

var jsClassFnPtrLen int32
var jsClassFnLock sync.Mutex
var jsClassFnPtrStore = make(map[int32]*jsClassFnEntry)

var jsClassFieldFnPtrLen int32
var jsClassFieldFnLock sync.Mutex
var jsClassFieldFnPtrStore = make(map[int32]*jsClassFieldFnEntry)

type jsClassFnEntry struct {
	fnName string
	ctx    *Context
	fn     func(ctx *Context, this Value, args []Value) Value
}

type jsClassFieldFnEntry struct {
	fieldName string
	ctx       *Context
	getFn     func(ctx *Context, this Value, args []Value) Value
	setFn     func(ctx *Context, this Value, args []Value) Value
}

type JSClass struct {
	className        string
	goClassID        uint32
	fnIds            []int32
	fieldFn          map[string]*int32
	constructorFn    func(ctx *Context, this Value, args []Value) interface{}
	constructorFnObj *Value
	ctx              *Context
	finalizerFn      func(goObject interface{})
}

func newGlobalClass(className string) *JSClass {
	jsClass := &JSClass{
		fnIds:     []int32{},
		fieldFn:   make(map[string]*int32),
		className: className,
	}
	jsClassLock.Lock()

	cGoClassID := C.JS_NewClassID(new(C.JSClassID))
	jsClass.goClassID = uint32(cGoClassID)
	jsClassIDMap[jsClass.goClassID] = jsClass

	jsGlobalClassIDMap[jsClass.goClassID] = jsClass

	defer jsClassLock.Unlock()
	return jsClass
}

func newModClass(className string) *JSClass {
	jsClass := &JSClass{
		fnIds:     []int32{},
		fieldFn:   make(map[string]*int32),
		className: className,
	}
	jsClassLock.Lock()

	cGoClassID := C.JS_NewClassID(new(C.JSClassID))
	jsClass.goClassID = uint32(cGoClassID)
	jsClassIDMap[jsClass.goClassID] = jsClass
	defer jsClassLock.Unlock()
	return jsClass
}

func (j *JSClass) storeFuncClassPtr(v *jsClassFnEntry) int32 {
	id := atomic.AddInt32(&jsClassFnPtrLen, 1) - 1
	jsClassFnLock.Lock()
	defer jsClassFnLock.Unlock()
	jsClassFnPtrStore[id] = v
	j.fnIds = append(j.fnIds, id)
	return id
}

func storeClassFieldFnPtr(v *jsClassFieldFnEntry) int32 {
	id := atomic.AddInt32(&jsClassFieldFnPtrLen, 1) - 1
	jsClassFieldFnLock.Lock()
	defer jsClassFieldFnLock.Unlock()
	jsClassFieldFnPtrStore[id] = v
	return id
}

func (j *JSClass) SetConstructor(fn func(ctx *Context, this Value, args []Value) interface{}) {
	//j.constructorID = j.storeConstructorPtr(getFn)
	j.constructorFn = fn
}

func (j *JSClass) SetFinalizer(fn func(obj interface{})) {
	//j.finalizerID = j.storeFinalizerPtr(getFn)
	j.finalizerFn = fn
}

func (j *JSClass) CreateGoJsClassObject(args ...Value) Value {
	if j.ctx == nil {
		panic("[CreateGoJsClassObject] the corresponding class is not initialized.If it is a global class, it cannot be called until the ctx has been created; If it is a module, it needs to be initialized before it can be called")
	}
	cargs := []C.JSValue{}
	for _, x := range args {
		cargs = append(cargs, x.ref)
	}
	if len(args) == 0 {
		return Value{ctx: j.ctx, ref: C.JS_CallConstructor(j.ctx.ref, j.constructorFnObj.ref, 0, nil)}
	}
	return Value{ctx: j.ctx, ref: C.JS_CallConstructor(j.ctx.ref, j.constructorFnObj.ref, C.int(len(args)), &cargs[0])}
}

func (j *JSClass) AddClassFn(fnName string, fn func(ctx *Context, this Value, args []Value) Value) {
	classFnEntry := &jsClassFnEntry{
		fn:     fn,
		fnName: fnName,
	}
	j.storeFuncClassPtr(classFnEntry)
}

func (j *JSClass) AddClassGetFn(fieldName string, fn func(ctx *Context, this Value, args []Value) Value) {
	if id := j.fieldFn[fieldName]; id == nil {
		classFnEntry := &jsClassFieldFnEntry{
			getFn:     fn,
			fieldName: fieldName,
		}
		fnId := storeClassFieldFnPtr(classFnEntry)
		j.fieldFn[fieldName] = &fnId
	} else {
		classFnEntry := jsClassFieldFnPtrStore[*id]
		classFnEntry.getFn = fn
	}

}

func (j *JSClass) AddClassSetFn(fieldName string, fn func(ctx *Context, this Value, args []Value) Value) {
	if id := j.fieldFn[fieldName]; id == nil {
		classFnEntry := &jsClassFieldFnEntry{
			setFn:     fn,
			fieldName: fieldName,
		}
		fnId := storeClassFieldFnPtr(classFnEntry)
		j.fieldFn[fieldName] = &fnId
	} else {
		classFnEntry := jsClassFieldFnPtrStore[*id]
		classFnEntry.setFn = fn
	}
}
