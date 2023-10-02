package quickjsBind

/*
#include "bridge.h"
*/
import "C"
import (
	"io"
	"runtime"
	"time"
	"unsafe"
)

// Runtime represents a Javascript runtime corresponding to an object heap. Several runtimes can exist at the same time but they cannot exchange objects. Inside a given runtime, no multi-threading is supported.
type Runtime struct {
	ref       *C.JSRuntime
	loop      *Loop // only one loop per runtime
	goModList []*JSModule
}

// NewRuntime creates a new quickjs runtime.
func NewRuntime() Runtime {
	runtime.LockOSThread() // prevent multiple quickjs runtime from being created
	rt := Runtime{ref: C.JS_NewRuntime(), loop: NewLoop()}
	C.JS_SetCanBlock(rt.ref, C.int(1))
	return rt
}

// RunGC will call quickjs's garbage collector.
func (r Runtime) RunGC() {
	C.JS_RunGC(r.ref)
}

// Close will free the runtime pointer.
func (r Runtime) Close() {
	C.JS_FreeRuntime(r.ref)
}

// SetMemoryLimit the runtime memory limit; if not set, it will be unlimit.
func (r Runtime) SetMemoryLimit(limit uint32) {
	C.JS_SetMemoryLimit(r.ref, C.size_t(limit))
}

// SetGCThreshold the runtime's GC threshold; use -1 to disable automatic GC.
func (r Runtime) SetGCThreshold(threshold int64) {
	C.JS_SetGCThreshold(r.ref, C.size_t(threshold))
}

// SetMaxStackSize will set max runtime's stack size; default is 255
func (r Runtime) SetMaxStackSize(stack_size uint32) {
	C.JS_SetMaxStackSize(r.ref, C.size_t(stack_size))
}

// NewModuleContext creates a new JavaScript context.
// enable BigFloat/BigDecimal support and enable .
// enable operator overloading.
func (r Runtime) NewModuleContext() *Context {
	ref := C.JS_NewContext(r.ref)

	C.js_std_init_handlers(r.ref)

	C.JS_SetModuleLoaderFunc(
		r.ref,
		(*C.JSModuleNormalizeFunc)(unsafe.Pointer(nil)),
		(*C.JSModuleLoaderFunc)(C.js_module_loader),
		unsafe.Pointer(nil),
	)

	C.JS_AddIntrinsicBigFloat(ref)
	C.JS_AddIntrinsicBigDecimal(ref)
	C.JS_AddIntrinsicOperators(ref)
	C.JS_EnableBignumExt(ref, C.int(1))
	loadPreludeModules(ref)

	C.registerGoClass(ref)

	for _, m := range r.goModList {
		m.buildModule(ref)
	}

	//
	//cStr := C.CString("ModuleNameTest")
	//defer C.free(unsafe.Pointer(cStr))
	//// JSModuleInitFunc
	//C.JS_NewCModule(
	//	ref,
	//	cStr,
	//	(*C.JSModuleInitFunc)(unsafe.Pointer(C.InvokeGoInitModule)))

	return &Context{ref: ref, runtime: &r}
}

// NewSimpleContext create a simple context with no modules and custom class
func (r Runtime) NewSimpleContext() *Context {
	ref := C.JS_NewContext(r.ref)

	C.JS_AddIntrinsicBigFloat(ref)
	C.JS_AddIntrinsicBigDecimal(ref)
	C.JS_AddIntrinsicOperators(ref)
	C.JS_EnableBignumExt(ref, C.int(1))

	return &Context{ref: ref, runtime: &r}
}

func loadPreludeModules(ctx *C.JSContext) {

	stdModulePtr := C.CString("std")
	defer C.free(unsafe.Pointer(stdModulePtr))

	C.js_init_module_std(ctx, stdModulePtr)

	osModulePtr := C.CString("os")
	defer C.free(unsafe.Pointer(osModulePtr))
	C.js_init_module_os(ctx, osModulePtr)

	C.js_std_add_helpers(ctx, -1, (**C.char)(unsafe.Pointer(nil)))
	// C.JS_AddIntrinsicProxy(ctx)
}

// ExecutePendingJob will execute all pending jobs.
func (r Runtime) ExecutePendingJob() (Context, error) {
	var ctx Context

	err := C.JS_ExecutePendingJob(r.ref, &ctx.ref)
	if err <= 0 {
		if err == 0 {
			return ctx, io.EOF
		}
		return ctx, ctx.Exception()
	}

	return ctx, nil
}

// IsJobPending returns true if there is a pending job.
func (r Runtime) IsJobPending() bool {
	return C.JS_IsJobPending(r.ref) == 1
}

// IsLoopJobPending returns true if there is a pending loop job.
func (r Runtime) IsLoopJobPending() bool {
	return r.loop.isLoopPending()
}

func (r Runtime) ExecuteAllPendingJobs() error {
	var err error
	for r.loop.isLoopPending() || r.IsJobPending() {
		// execute loop job
		r.loop.run()

		// excute promiIs
		_, err := r.ExecutePendingJob()
		if err == io.EOF {
			err = nil
		}
		time.Sleep(time.Millisecond * 1) // prevent 100% CPU
	}
	return err
}

//// AddGoModule add go module
//func (r *Runtime) AddGoModule(m *JSModule) {
//	r.goModList = append(r.goModList, m)
//}

func (r *Runtime) CreateModule(moduleName string) *JSModule {
	m := newMod(moduleName)
	r.goModList = append(r.goModList, m)
	return m
}

// CreateGlobalClass add context global class
func (r *Runtime) CreateGlobalClass(className string) *JSClass {
	jsClass := newGlobalClass(className)
	return jsClass
}
