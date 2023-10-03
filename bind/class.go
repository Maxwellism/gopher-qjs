package quickjsBind

/*
#include <stdint.h>
#include "bridge.h"
*/
import "C"
import (
	"fmt"
	"reflect"
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
	putJsClass(jsClass.goClassID, jsClass)
	jsGlobalClassIDMap[id] = jsClass
}

var jsClassMapGoObjectLock sync.Mutex
var jsClassMapGoObject = make(map[string]interface{})

func pushGoObjectByJS(v interface{}) string {
	jsClassMapGoObjectLock.Lock()
	defer jsClassMapGoObjectLock.Unlock()
	id := getObjectPtr(v)
	jsClassMapGoObject[id] = v
	return id
}
func getGoObjectByID(id string) interface{} {
	jsClassMapGoObjectLock.Lock()
	defer jsClassMapGoObjectLock.Unlock()
	return jsClassMapGoObject[id]
}
func putGoObject(id string, v interface{}) {
	jsClassMapGoObjectLock.Lock()
	defer jsClassMapGoObjectLock.Unlock()
	jsClassMapGoObject[id] = v
}
func deleteGoObjectByID(id string) {
	jsClassMapGoObjectLock.Lock()
	defer jsClassMapGoObjectLock.Unlock()
	delete(jsClassMapGoObject, id)
}

func getObjectPtr(v interface{}) string {
	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Ptr {
		value = reflect.ValueOf(&v)
	}
	return fmt.Sprintf("%v", value.Pointer())
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

func setClassFn(id int32, v *jsClassFnEntry) {
	jsClassFnLock.Lock()
	defer jsClassFnLock.Unlock()
	jsClassFnPtrStore[id] = v
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
	ClassName      string
	goClassID      uint32
	fnList         sync.Map
	fieldFnList    sync.Map
	constructorFn  func(ctx *Context, args []Value) interface{}
	classStaticVal *Value
	ctx            *Context
	finalizerFn    func(goObject interface{})
}

func newGlobalClass(className string) *JSClass {
	jsClass := &JSClass{
		ClassName: className,
	}
	cGoClassID := C.JS_NewClassID(new(C.JSClassID))
	jsClass.goClassID = uint32(cGoClassID)
	putGlobalJsClass(jsClass.goClassID, jsClass)
	return jsClass
}

func newModClass(className string) *JSClass {
	jsClass := &JSClass{
		ClassName: className,
	}
	cGoClassID := C.JS_NewClassID(new(C.JSClassID))
	jsClass.goClassID = uint32(cGoClassID)

	putJsClass(jsClass.goClassID, jsClass)
	return jsClass
}

func (j *JSClass) SetConstructor(fn func(ctx *Context, args []Value) interface{}) {
	//j.constructorID = j.storeConstructorPtr(getFn)
	j.constructorFn = fn
}

func (j *JSClass) SetFinalizer(fn func(obj interface{})) {
	//j.finalizerID = j.storeFinalizerPtr(getFn)
	j.finalizerFn = fn
}

func (j *JSClass) CreateGoJsClassObject(args ...Value) Value {
	if j.ctx == nil {
		panic("[" + j.ClassName + "] the corresponding class is not initialized.If it is a global class, it cannot be called until the ctx has been created; If it is a module, it needs to be initialized before it can be called")
	}
	cargs := []C.JSValue{}
	for _, x := range args {
		cargs = append(cargs, x.ref)
	}
	if len(args) == 0 {
		return Value{ctx: j.ctx, ref: C.JS_CallConstructor(j.ctx.ref, j.classStaticVal.ref, 0, nil)}
	}
	return Value{ctx: j.ctx, ref: C.JS_CallConstructor(j.ctx.ref, j.classStaticVal.ref, C.int(len(args)), &cargs[0])}
}

func (j *JSClass) AddClassFn(fnName string, fn func(ctx *Context, this Value, args []Value) Value) {
	classFnEntry := &jsClassFnEntry{
		fn:     fn,
		fnName: fnName,
	}
	if idValue, ok := j.fnList.Load(fnName); !ok {
		fnId := pushClassFn(classFnEntry)
		j.fnList.Store(fnName, fnId)
	} else {
		id, _ := idValue.(int32)
		setClassFn(id, classFnEntry)
	}
}

func (j *JSClass) GetClassValue() *Value {
	return j.classStaticVal
}

func (j *JSClass) AddClassGetFn(fieldName string, fn func(ctx *Context, this Value, args []Value) Value) {
	if id, ok := j.fieldFnList.Load(fieldName); !ok {
		classFnEntry := &jsClassFieldFnEntry{
			getFn:     fn,
			fieldName: fieldName,
		}
		fnId := pushClassFieldFn(classFnEntry)
		j.fieldFnList.Store(fieldName, fnId)
	} else {
		fnId, _ := id.(int32)
		classFnEntry := getClassFieldFnByID(fnId)
		classFnEntry.getFn = fn
	}

}

func (j *JSClass) AddClassSetFn(fieldName string, fn func(ctx *Context, this Value, args []Value) Value) {
	if id, ok := j.fieldFnList.Load(fieldName); !ok {
		classFnEntry := &jsClassFieldFnEntry{
			setFn:     fn,
			fieldName: fieldName,
		}
		fnId := pushClassFieldFn(classFnEntry)
		j.fieldFnList.Store(fieldName, fnId)
	} else {
		fnId, _ := id.(int32)
		classFnEntry := getClassFieldFnByID(fnId)
		classFnEntry.setFn = fn
	}
}
