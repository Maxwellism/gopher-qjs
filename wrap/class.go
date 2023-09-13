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

var jsClassGetFnPtrLen int32
var jsClassGetFnLock sync.Mutex
var jsClassGetFieldFnPtrStore = make(map[int32]*jsClassFieldGetFnEntry)

var jsClassSetFnPtrLen int32
var jsClassSetFnLock sync.Mutex
var jsClassSetFieldFnPtrStore = make(map[int32]*jsClassFieldSetFnEntry)

type jsClassFnEntry struct {
	fnName string
	fn     func(ctx *Context, this Value, args []Value) Value
}

type jsClassFieldGetFnEntry struct {
	fieldName string
	fn        func(ctx *Context, this Value, args []Value) Value
}

type jsClassFieldSetFnEntry struct {
	fieldName string
	fn        func(ctx *Context, this Value, args []Value) Value
}

type JSClass struct {
	className     string
	goClassID     uint32
	fnIds         []int32
	getFnIds      []int32
	setFnIds      []int32
	constructorFn func(ctx *Context, this Value, args []Value) interface{}
	finalizerFn   func(goObject interface{})
}

func NewClass(className string) *JSClass {
	jsClass := &JSClass{
		fnIds:     []int32{},
		getFnIds:  []int32{},
		setFnIds:  []int32{},
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

func (j *JSClass) storeSetClassFieldFnPtr(v *jsClassFieldSetFnEntry) int32 {
	id := atomic.AddInt32(&jsClassSetFnPtrLen, 1) - 1
	jsClassSetFnLock.Lock()
	defer jsClassSetFnLock.Unlock()
	jsClassSetFieldFnPtrStore[id] = v
	j.setFnIds = append(j.setFnIds, id)
	return id
}

func (j *JSClass) storeGetClassFieldFnPtr(v *jsClassFieldGetFnEntry) int32 {
	id := atomic.AddInt32(&jsClassGetFnPtrLen, 1) - 1
	jsClassGetFnLock.Lock()
	defer jsClassGetFnLock.Unlock()
	jsClassGetFieldFnPtrStore[id] = v
	j.getFnIds = append(j.getFnIds, id)
	return id
}

func (j *JSClass) SetConstructor(fn func(ctx *Context, this Value, args []Value) interface{}) {
	//j.constructorID = j.storeConstructorPtr(fn)
	j.constructorFn = fn
}

func (j *JSClass) SetFinalizer(fn func(obj interface{})) {
	//j.finalizerID = j.storeFinalizerPtr(fn)
	j.finalizerFn = fn
}

func (j *JSClass) AddClassFn(fnName string, fn func(ctx *Context, this Value, args []Value) Value) {
	classFnEntry := &jsClassFnEntry{
		fn:     fn,
		fnName: fnName,
	}
	j.storeFuncClassPtr(classFnEntry)
}

func (j *JSClass) AddClassGetFn(fieldName string, fn func(ctx *Context, this Value, args []Value) Value) {
	classFnEntry := &jsClassFieldGetFnEntry{
		fn:        fn,
		fieldName: fieldName,
	}
	j.storeGetClassFieldFnPtr(classFnEntry)
}

func (j *JSClass) AddClassSetFn(fieldName string, fn func(ctx *Context, this Value, args []Value) Value) {
	classFnEntry := &jsClassFieldSetFnEntry{
		fn:        fn,
		fieldName: fieldName,
	}
	j.storeSetClassFieldFnPtr(classFnEntry)
}
