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

JSValue InvokeGoModFn(JSContext *ctx, JSValueConst this_val,int argc, JSValueConst *argv, int magic) {
    return goModFnHandle(ctx, this_val, argc, argv, magic);
}

JSValue InvokeGoClassSetFn(JSContext *ctx, JSValueConst this_val,int argc, JSValueConst *argv, int magic) {
    return goClassSetFnHandle(ctx, this_val, argc, argv, magic);
}

JSValue InvokeGoClassGetFn(JSContext *ctx, JSValueConst this_val,int argc, JSValueConst *argv, int magic) {
    return goClassGetFnHandle(ctx, this_val, argc, argv, magic);
}

JSValue InvokeGoClassFn(JSContext *ctx, JSValueConst this_val,int argc, JSValueConst *argv, int magic) {
    return goClassFnHandle(ctx, this_val, argc, argv, magic);
}

JSValue InvokeGoClassConstructor(JSContext *ctx, JSValueConst new_target, int argc, JSValueConst *argv, int magic) {
    JSValue obj = JS_UNDEFINED;
    JSValue proto;

    char *goObjectId = goClassConstructorHandle(ctx,new_target,argc,argv,magic);

    if (strncmp(goObjectId,"",1)==0){
        return JS_EXCEPTION;
    }

    GoClassObjectInfo *goClassObjectInfo;

    goClassObjectInfo = js_mallocz(ctx, sizeof(*goClassObjectInfo));
    if (!goClassObjectInfo)
        return JS_EXCEPTION;

    goClassObjectInfo->objectId = goObjectId;
    goClassObjectInfo->goClassID = magic;

    proto = JS_GetPropertyStr(ctx, new_target, "prototype");

    obj = JS_NewObjectProtoClass(ctx, proto, magic);

//    JS_SetPropertyStr(ctx, proto, "_goClassID", JS_NewInt32(ctx, (int32_t)magic));
    JS_SetPropertyStr(ctx, proto, "_goObjectID", JS_NewString(ctx, goObjectId));

    JS_FreeValue(ctx, proto);
    if (JS_IsException(obj)){
        JS_FreeValue(ctx, obj);
        js_free(ctx, goClassObjectInfo);
        return JS_EXCEPTION;
    }
    JS_SetOpaque(obj, goClassObjectInfo);
    return obj;
}

void goFinalizer(JSRuntime *rt, JSValue val) {

    GoClassObjectInfo *goClassObjectInfo = JS_UnsafeGetOpaque(val);
//    printf("go object id:%d\n", goClassObject->objectId);
    goFinalizerHandle(goClassObjectInfo->goClassID, goClassObjectInfo->objectId);

    js_free_rt(rt,goClassObjectInfo);
}

void registerGoClass(JSContext *ctx) {
    registerGoClassHandle(ctx);
}

int interruptHandler(JSRuntime *rt, void *handlerArgs) {
	return goInterruptHandler(rt, handlerArgs);
}

void SetInterruptHandler(JSRuntime *rt, void *handlerArgs){
	JS_SetInterruptHandler(rt, &interruptHandler, handlerArgs);
}

int getValTag(JSValueConst v) {
	return JS_VALUE_GET_TAG(v);
}

int InvokeGoModInit(JSContext *ctx, JSModuleDef *m) {
    return GoInitModule(ctx,m);
}

void JS_NewGlobalCConstructorHandle(JSContext *ctx,
                                      JSValue func_obj,
                                      const char *name,
                                      JSValueConst proto)
{
    JSValue global =  JS_GetGlobalObject(ctx);
    JS_DefinePropertyValueStr(ctx, global, name,
                           JS_DupValue(ctx, func_obj),
                           JS_PROP_WRITABLE | JS_PROP_CONFIGURABLE);
    JS_FreeValue(ctx, global);
    JS_SetConstructor(ctx, func_obj, proto);
    JS_FreeValue(ctx, func_obj);
}

//JSModuleDef *js_my_module_loader(JSContext *ctx,
//                              const char *module_name, void *opaque)
//{
//    JSModuleDef *m;
//
//
//    size_t buf_len;
//    uint8_t *buf;
//    JSValue func_val;
//
//    printf("模块名称:%s\n", module_name);
//
//    buf = js_load_file(ctx, &buf_len, module_name);
//    if (!buf) {
//        JS_ThrowReferenceError(ctx, "could not load module filename '%s'",
//                               module_name);
//        return NULL;
//    }
//
//    /* compile the module */
//    func_val = JS_Eval(ctx, (char *)buf, buf_len, module_name,
//                       JS_EVAL_TYPE_MODULE | JS_EVAL_FLAG_COMPILE_ONLY);
//    js_free(ctx, buf);
//    if (JS_IsException(func_val))
//        return NULL;
//    /* XXX: could propagate the exception */
//    js_module_set_import_meta(ctx, func_val, 1, 0);
//    /* the module is already referenced, so we must free it */
//    m = JS_VALUE_GET_PTR(func_val);
//    JS_FreeValue(ctx, func_val);
//
//    return m;
//}