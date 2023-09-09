#include "_cgo_export.h"
#include "quickjs-libc.h"
#include <string.h>


JSValue JS_NewNull() { return JS_NULL; }
JSValue JS_NewUndefined() { return JS_UNDEFINED; }
JSValue JS_NewUninitialized() { return JS_UNINITIALIZED; }

JSValue ThrowSyntaxError(JSContext *ctx, const char *fmt) { return JS_ThrowSyntaxError(ctx, "%s", fmt); }
JSValue ThrowTypeError(JSContext *ctx, const char *fmt) { return JS_ThrowTypeError(ctx, "%s", fmt); }
JSValue ThrowReferenceError(JSContext *ctx, const char *fmt) { return JS_ThrowReferenceError(ctx, "%s", fmt); }
JSValue ThrowRangeError(JSContext *ctx, const char *fmt) { return JS_ThrowRangeError(ctx, "%s", fmt); }
JSValue ThrowInternalError(JSContext *ctx, const char *fmt) { return JS_ThrowInternalError(ctx, "%s", fmt); }

JSValue InvokeProxy(JSContext *ctx, JSValueConst this_val, int argc, JSValueConst *argv) {
	 return goProxy(ctx, this_val, argc, argv);
}

JSValue InvokeAsyncProxy(JSContext *ctx, JSValueConst this_val, int argc, JSValueConst *argv) {
	return goAsyncProxy(ctx, this_val, argc, argv);
}

int interruptHandler(JSRuntime *rt, void *handlerArgs) {
	return goInterruptHandler(rt, handlerArgs);
}

void SetInterruptHandler(JSRuntime *rt, void *handlerArgs){
	JS_SetInterruptHandler(rt, &interruptHandler, handlerArgs);
}

JSCFunctionListEntry getJSCFunctionEntry(const char *fnName,int argLen,JSCFunction jsFn){
    JSCFunctionListEntry res;
    JSCFunctionType cfunc;
    cfunc.generic = jsFn;

    res.name = fnName;
    res.u.func.length = argLen;
    res.u.func.cfunc = cfunc;
    res.u.func.cproto = JS_CFUNC_generic;
    // res = JS_CFUNC_DEF(fnName, argLen, jsFn);
    return res;
}

int getValTag(JSValueConst v) {
	return JS_VALUE_GET_TAG(v);
}

JSModuleDef *js_my_module_loader(JSContext *ctx,
                              const char *module_name, void *opaque)
{
    JSModuleDef *m;


    size_t buf_len;
    uint8_t *buf;
    JSValue func_val;

    printf("模块名称:%s\n", module_name);

    buf = js_load_file(ctx, &buf_len, module_name);
    if (!buf) {
        JS_ThrowReferenceError(ctx, "could not load module filename '%s'",
                               module_name);
        return NULL;
    }

    /* compile the module */
    func_val = JS_Eval(ctx, (char *)buf, buf_len, module_name,
                       JS_EVAL_TYPE_MODULE | JS_EVAL_FLAG_COMPILE_ONLY);
    js_free(ctx, buf);
    if (JS_IsException(func_val))
        return NULL;
    /* XXX: could propagate the exception */
    js_module_set_import_meta(ctx, func_val, 1, 0);
    /* the module is already referenced, so we must free it */
    m = JS_VALUE_GET_PTR(func_val);
    JS_FreeValue(ctx, func_val);

    return m;
}