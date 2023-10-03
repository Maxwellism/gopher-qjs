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

var BuiltPolyfill = map[string]func(ctx *Context){}

type ContextOpt func(*contextConfig)

type contextConfig struct {
	polyfillNames      []string
	isBuildAllPolyfill bool
	moduleNames        map[string]bool
	isBuildAllModule   bool
}

func newContextConfig() *contextConfig {
	return &contextConfig{
		isBuildAllModule:   true,
		isBuildAllPolyfill: true,
	}
}

func SetIsBuildAllPolyfill(flag bool) ContextOpt {
	return func(config *contextConfig) {
		config.isBuildAllPolyfill = flag
	}
}

func SetIsBuildAllModule(flag bool) ContextOpt {
	return func(config *contextConfig) {
		config.isBuildAllModule = flag
	}
}

func WithBuildPolyfill(polyName string) ContextOpt {
	return func(config *contextConfig) {
		config.polyfillNames = append(config.polyfillNames, polyName)
	}
}

func WithBuildModule(moduleName string) ContextOpt {
	return func(config *contextConfig) {
		if config.moduleNames == nil {
			config.moduleNames = make(map[string]bool)
		}
		config.moduleNames[moduleName] = true
	}
}

// Runtime represents a Javascript runtime corresponding to an object heap. Several runtimes can exist at the same time but they cannot exchange objects. Inside a given runtime, no multi-threading is supported.
type Runtime struct {
	ref          *C.JSRuntime
	loop         *Loop // only one loop per runtime
	goModuleList []*JSModule
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

// NewContext creates a new JavaScript context.
// enable BigFloat/BigDecimal support and enable .
// enable operator overloading.
func (r Runtime) NewContext(opts ...ContextOpt) *Context {

	config := newContextConfig()

	for _, opt := range opts {
		opt(config)
	}

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
	//loadPreludeModules(ref)

	C.registerGoClass(ref)

	//
	//cStr := C.CString("ModuleNameTest")
	//defer C.free(unsafe.Pointer(cStr))
	//// JSModuleInitFunc
	//C.JS_NewCModule(
	//	ref,
	//	cStr,
	//	(*C.JSModuleInitFunc)(unsafe.Pointer(C.InvokeGoInitModule)))

	resContext := &Context{ref: ref, runtime: &r}

	r.buildPolyfill(config, resContext)
	r.buildModule(config, resContext)

	return resContext
}

func (r *Runtime) buildPolyfill(config *contextConfig, ctx *Context) {
	if config.isBuildAllPolyfill {
		for _, value := range BuiltPolyfill {
			value(ctx)
		}
	} else {
		for _, name := range config.polyfillNames {
			BuiltPolyfill[name](ctx)
		}
	}
}

func (r *Runtime) buildModule(config *contextConfig, ctx *Context) {
	if config.isBuildAllModule {
		for _, m := range r.goModuleList {
			m.buildModule(ctx)
		}
	} else {
		for _, m := range r.goModuleList {
			if config.moduleNames[m.modName] {
				m.buildModule(ctx)
			}
		}
	}
}

func loadPreludeModules(ctx *C.JSContext) {

	stdModulePtr := C.CString("std")
	defer C.free(unsafe.Pointer(stdModulePtr))

	C.js_init_module_std(ctx, stdModulePtr)

	osModulePtr := C.CString("os")
	defer C.free(unsafe.Pointer(osModulePtr))
	C.js_init_module_os(ctx, osModulePtr)

	//C.js_std_add_helpers(ctx, -1, (**C.char)(unsafe.Pointer(nil)))
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
//	r.goModuleList = append(r.goModuleList, m)
//}

func (r *Runtime) CreateModule(moduleName string) *JSModule {
	m := newMod(moduleName)
	r.goModuleList = append(r.goModuleList, m)
	return m
}

// CreateGlobalClass add context global class
func (r *Runtime) CreateGlobalClass(className string) *JSClass {
	jsClass := newGlobalClass(className)
	return jsClass
}
