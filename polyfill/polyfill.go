package polyfill

import (
	quickjs "github.com/Maxwellism/gopher-qjs/bind"
	"github.com/Maxwellism/gopher-qjs/polyfill/console_print"
	"github.com/Maxwellism/gopher-qjs/polyfill/window"
)

func Polyfill() {
	quickjs.BuiltPolyfill["print"] = console_print.PrintInjectTo
	quickjs.BuiltPolyfill["console"] = console_print.ConsoleInjectTo
	quickjs.BuiltPolyfill["window"] = window.WindowInjectTo
}
