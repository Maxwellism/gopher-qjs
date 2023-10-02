# gopher-qjs

go语言中quickjs的友好包装器，该仓库的计划支持有：

- [x] 方法注册
  - [x] 普通方法注册
  - [x] 异步方法注册
- [ ] 类支持
  - [x] 类构造方法
  - [x] 类方法
  - [ ] 类静态属性
  - [x] 类属性get/set
  - [x] 类对象注销方法
  - [x] 模块类支持
- [x] 模块支持
  - [x] 方法
  - [x] 类
  - [x] 对象
- [ ] debug支持

相关使用方法请参考里面的test.go文件

## Thanks

|                                                       | About                                                        |
| ----------------------------------------------------- | ------------------------------------------------------------ |
| [buke/quickjs-go](https://github.com/buke/quickjs-go) | Go 语言的QuickJS绑定库。感谢buke的大量基础工作，本仓库基于buke/quickjs-go仓库进行大量的二次开发。 |
| [bellard/quickjs](https://github.com/bellard/quickjs) | QuickJS是一个小型并且可嵌入的Javascript引擎，它支持ES2020规范，包括模块，异步生成器和代理器。 |

