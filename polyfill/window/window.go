package window

import quickjs "github.com/Maxwellism/gopher-qjs/bind"

func WindowInjectTo(ctx *quickjs.Context) {
	ctx.Globals().Set("window", ctx.Globals().Get("globalThis"))
}
