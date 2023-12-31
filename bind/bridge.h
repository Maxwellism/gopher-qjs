#include <stdlib.h>
#include <string.h>
#include "quickjs-libc.h"

extern JSValue JS_NewNull();
extern JSValue JS_NewUndefined();
extern JSValue JS_NewUninitialized();
extern JSValue ThrowSyntaxError(JSContext *ctx, const char *fmt) ;
extern JSValue ThrowTypeError(JSContext *ctx, const char *fmt) ;
extern JSValue ThrowReferenceError(JSContext *ctx, const char *fmt) ;
extern JSValue ThrowRangeError(JSContext *ctx, const char *fmt) ;
extern JSValue ThrowInternalError(JSContext *ctx, const char *fmt) ;
int JS_DeletePropertyInt64(JSContext *ctx, JSValueConst obj, int64_t idx, int flags);

extern JSValue InvokeProxy(JSContext *ctx, JSValueConst this_val, int argc, JSValueConst *argv);
extern JSValue InvokeAsyncProxy(JSContext *ctx, JSValueConst this_val, int argc, JSValueConst *argv);

extern JSValue InvokeGoModFn(JSContext *ctx, JSValueConst this_val,int argc, JSValueConst *argv, int magic);
extern int InvokeGoModInit(JSContext *ctx, JSModuleDef *m);

extern JSValue InvokeGoClassSetFn(JSContext *ctx, JSValueConst this_val,int argc, JSValueConst *argv, int magic);
extern JSValue InvokeGoClassGetFn(JSContext *ctx, JSValueConst this_val,int argc, JSValueConst *argv, int magic);
extern JSValue InvokeGoClassFn(JSContext *ctx, JSValueConst this_val,int argc, JSValueConst *argv, int magic);
extern JSValue InvokeGoClassConstructor(JSContext *ctx, JSValueConst new_target, int argc, JSValueConst *argv, int magic);
extern void goFinalizer(JSRuntime *rt, JSValue val);
extern void registerGoClass(JSContext *ctx);
extern void JS_NewGlobalCConstructorHandle(JSContext *ctx,
                                             JSValue func_obj,
                                             const char *name,
                                             JSValueConst proto);

typedef struct {
    uintptr_t fn;
} handlerArgs;

typedef struct {
    int goClassID;
    char *objectId;
} GoClassObjectInfo;

extern void SetInterruptHandler(JSRuntime *rt, void *handlerArgs);

extern int getValTag(JSValueConst v);

//extern JSModuleDef *js_my_module_loader(JSContext *ctx,const char *module_name, void *opaque);

