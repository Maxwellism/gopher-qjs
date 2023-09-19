# gopher-qjs

go语言中quickjs的友好包装器，短期内无开发计划，在此充当一个占位仓库，该仓库的计划支持有：

- [ ] 方法注册
  - [x] 普通方法注册
  - [ ] 异步方法注册
- [ ] 类支持
  - [x] 类构造方法
  - [x] 类方法
  - [ ] 类静态属性
  - [x] 类属性get/set
  - [x] 类对象注销方法
  - [x] 模块类支持
- [ ] 模块支持
  - [x] 方法
  - [x] 类
  - [ ] 对象
- [ ] debug支持
现在这个仓库还未开放，组织里的人要是想使用这些api，请设置一下GOPRIVATE：

go env -w GOPRIVATE=github.com/Maxwellism/gopher-qjs
