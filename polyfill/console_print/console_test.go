package console_print_test

import (
	quickjs "github.com/Maxwellism/gopher-qjs/bind"
	print2 "github.com/Maxwellism/gopher-qjs/polyfill/console_print"
	"testing"
)

func TestConsole(t *testing.T) {
	rt := quickjs.NewRuntime()
	defer rt.Close()

	ctx := rt.NewContext()
	defer ctx.Close()

	print2.ConsoleInjectTo(ctx)

	ret, err := ctx.Eval(`
	const obj = {"a": 1, "b": 2};
	console.error(obj);
	
	let c = Object.entries({foo: 'bar'});

	const map = new Map(c)
	console.log(map,obj);

	console.log('hello', 'world');
	//console.logg('ccc');
	`)
	//fmt.Println(ret)
	defer ret.Free()
	if err != nil {
		panic(err)
	}
}
