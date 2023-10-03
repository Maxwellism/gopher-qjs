package console_print

import (
	"fmt"
	quickjs "github.com/Maxwellism/gopher-qjs/bind"
	"os"
)

func printJsValue(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
	for _, arg := range args {
		if arg.IsObject() {
			jsonStr := arg.JSONStringify()
			if arg.IsMap() {
				serializeObj := ctx.Object()
				defer serializeObj.Free()

				serializeObj.Set("dataType", ctx.String("Map"))
				array := ctx.Globals().Get("Array")

				defer array.Free()

				Iterator := arg.Call("entries")

				defer Iterator.Free()

				serializeObj.Set("value", array.Call("from", Iterator))
				jsonStr = serializeObj.JSONStringify()
			}
			fmt.Fprintf(os.Stdout, "%s %s", arg.String(), jsonStr)
		} else {
			fmt.Print(arg.String())
		}
		fmt.Print(" ")
	}
	fmt.Println()
	return ctx.Null()
}

func ConsoleInjectTo(ctx *quickjs.Context) {
	consoleObj := ctx.Object()
	consoleObj.Set("trace", ctx.Function(printJsValue))
	consoleObj.Set("debug", ctx.Function(printJsValue))
	consoleObj.Set("info", ctx.Function(printJsValue))
	consoleObj.Set("log", ctx.Function(printJsValue))
	consoleObj.Set("warn", ctx.Function(printJsValue))
	consoleObj.Set("error", ctx.Function(printJsValue))

	ctx.Globals().Set("console", consoleObj)
}

func PrintInjectTo(ctx *quickjs.Context) {
	ctx.Globals().Set("print", ctx.Function(printJsValue))
}
