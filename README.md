## Goasyncsvr
Goasyncsvr 是一个用Go语言实现的全异步的后台服务框架。它是支持自定义协议的，可以自定义各种协议包。例子中它是基于protobuf 3.0协议的。

## Requirements
- Go version >= 1.8 
- Beego

## Installation
git clone https://github.com/ssucc/goasyncsvr.git

go get github.com/astaxie/beego

## 代码功能和亮点
- 生成，增加协议，只需要实现协议的handle
- 采用反射的机制，只需要注册命令号与处理此命令号的handler。开发快速，简单可靠
- 引入自动化测试例子
