syncsvr


Welcome to the goasyncsvr wiki!

go 语言实现的一个全异步的，自定义协议的框架svr，可以自定义各种协议。目前模版中是基于protobuf 3.0的协议的。

代码特点：
1. 代码自动生成，增加协议，只需要实现协议的handle

2. 采用反射的机制，只需要注册命令号与处理此命令号的handler。开发快速，简单可靠

3. 引入自动化测试例子

使用：
1. 安装go开发环境
2. checkout beego
3. checkout 本项目

