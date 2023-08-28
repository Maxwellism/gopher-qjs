package register

import (
	"fmt"
	"github.com/buke/quickjs-go"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClass(t *testing.T) {
	rt := quickjs.NewRuntime()
	defer rt.Close()

	ctx := rt.NewContext()
	defer ctx.Close()

	SetConstructorClass("ClassTest", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) {
		this.Set("name", ctx.String(args[0].String()))
		this.Set("age", ctx.Int32(args[1].Int32()))
	})
	SetClassMethod("ClassTest", "sayHello", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		name := this.Get("name")
		assert.NotEmpty(t, name)
		fmt.Println(fmt.Sprintf("test name is :%s", name))
		return ctx.Null()
	})

	RuntimeWrapper(ctx)

	ret1, err := ctx.Eval(`
let c = new ClassTest("ccc",23);
let b = new ClassTest("bbb",43);
c.sayHello();
b.sayHello();
class ClassTest1{}
//b.constructor === ClassTest1
//b.constructor === ClassTest

class ClassExtendTest extends ClassTest {
    constructor(name, age, count) {
        super(name, age);
        this.count = count;
    }
    get_count() {
        return this.count;
    }
};
let a = new ClassExtendTest("aa",23,55);
a.get_count();
`)
	fmt.Println(ret1.String())
	assert.NoError(t, err)
	defer ret1.Free()
}
