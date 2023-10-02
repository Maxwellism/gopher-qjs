package quickjsBind

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

func putJsClass(id uint32, jsClass *JSClass) {
	jsClassLock.Lock()
	defer jsClassLock.Unlock()
	jsClassIDMap[id] = jsClass
}
func getClassByID(id uint32) *JSClass {
	jsClassLock.Lock()
	defer jsClassLock.Unlock()
	return jsClassIDMap[id]
}

var jsGlobalClassLock sync.Mutex
var jsGlobalClassIDMap = make(map[uint32]*JSClass)

func putGlobalJsClass(id uint32, jsClass *JSClass) {
	jsGlobalClassLock.Lock()
	defer jsGlobalClassLock.Unlock()
	jsGlobalClassIDMap[id] = jsClass
}

var jsClassMapGoObjectPtrLen int32
var jsClassMapGoObjectLock sync.Mutex
var jsClassMapGoObject = make(map[int32]interface{})

func pushGoObjectByJS(v interface{}) int32 {
	id := atomic.AddInt32(&jsClassMapGoObjectPtrLen, 1) - 1
	jsClassMapGoObjectLock.Lock()
	defer jsClassMapGoObjectLock.Unlock()
	jsClassMapGoObject[id] = v
	return id
}
func getGoObjectByID(id int32) interface{} {
	jsClassMapGoObjectLock.Lock()
	defer jsClassMapGoObjectLock.Unlock()
	return jsClassMapGoObject[id]
}
func putGoObject(id int32, v interface{}) {
	jsClassMapGoObjectLock.Lock()
	defer jsClassMapGoObjectLock.Unlock()
	jsClassMapGoObject[id] = v
}
func deleteGoObjectByID(id int32) {
	jsClassMapGoObjectLock.Lock()
	defer jsClassMapGoObjectLock.Unlock()
	delete(jsClassMapGoObject, id)
}

var jsClassFnPtrLen int32
var jsClassFnLock sync.Mutex
var jsClassFnPtrStore = make(map[int32]*jsClassFnEntry)

func pushClassFn(v *jsClassFnEntry) int32 {
	jsClassFnLock.Lock()
	defer jsClassFnLock.Unlock()
	id := atomic.AddInt32(&jsClassFnPtrLen, 1) - 1
	jsClassFnPtrStore[id] = v
	return id
}
func getClassFnByID(id int32) *jsClassFnEntry {
	jsClassFnLock.Lock()
	defer jsClassFnLock.Unlock()
	return jsClassFnPtrStore[id]
}

var jsClassFieldFnPtrLen int32
var jsClassFieldFnLock sync.Mutex
var jsClassFieldFnPtrStore = make(map[int32]*jsClassFieldFnEntry)

func pushClassFieldFn(v *jsClassFieldFnEntry) int32 {
	id := atomic.AddInt32(&jsClassFieldFnPtrLen, 1) - 1
	jsClassFieldFnLock.Lock()
	defer jsClassFieldFnLock.Unlock()
	jsClassFieldFnPtrStore[id] = v
	return id
}

func getClassFieldFnByID(id int32) *jsClassFieldFnEntry {
	jsClassFieldFnLock.Lock()
	defer jsClassFieldFnLock.Unlock()
	return jsClassFieldFnPtrStore[id]
}

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
	ClassName        string
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
		ClassName: className,
	}
	cGoClassID := C.JS_NewClassID(new(C.JSClassID))
	jsClass.goClassID = uint32(cGoClassID)
	putJsClass(jsClass.goClassID, jsClass)
	putGlobalJsClass(jsClass.goClassID, jsClass)
	return jsClass
}

func newModClass(className string) *JSClass {
	jsClass := &JSClass{
		fnIds:     []int32{},
		fieldFn:   make(map[string]*int32),
		ClassName: className,
	}
	cGoClassID := C.JS_NewClassID(new(C.JSClassID))
	jsClass.goClassID = uint32(cGoClassID)

	putJsClass(jsClass.goClassID, jsClass)
	return jsClass
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
	pushClassFn(classFnEntry)
}

func (j *JSClass) AddClassGetFn(fieldName string, fn func(ctx *Context, this Value, args []Value) Value) {
	if id := j.fieldFn[fieldName]; id == nil {
		classFnEntry := &jsClassFieldFnEntry{
			getFn:     fn,
			fieldName: fieldName,
		}
		fnId := pushClassFieldFn(classFnEntry)
		j.fieldFn[fieldName] = &fnId
	} else {
		classFnEntry := getClassFieldFnByID(*id)
		classFnEntry.getFn = fn
	}

}

func (j *JSClass) AddClassSetFn(fieldName string, fn func(ctx *Context, this Value, args []Value) Value) {
	if id := j.fieldFn[fieldName]; id == nil {
		classFnEntry := &jsClassFieldFnEntry{
			setFn:     fn,
			fieldName: fieldName,
		}
		fnId := pushClassFieldFn(classFnEntry)
		j.fieldFn[fieldName] = &fnId
	} else {
		classFnEntry := getClassFieldFnByID(*id)
		classFnEntry.setFn = fn
	}
}
