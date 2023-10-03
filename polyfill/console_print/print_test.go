package console_print_test

import (
	quickjs "github.com/Maxwellism/gopher-qjs/bind"
	print2 "github.com/Maxwellism/gopher-qjs/polyfill/console_print"
	"testing"
)

func TestPrint(t *testing.T) {
	rt := quickjs.NewRuntime()
	defer rt.Close()

	ctx := rt.NewContext()
	defer ctx.Close()

	print2.PrintInjectTo(ctx)

	ret, err := ctx.Eval(`
	print("test console_print");
	`)
	//fmt.Println(ret)
	defer ret.Free()
	if err != nil {
		panic(err)
	}
}
