# S1 - 代码结构



1. 使用[go](https://go.dev/)语言进行开发
2. 使用[bindata](https://github.com/go-bindata/go-bindata)进行前端页面的打包和嵌入
3. build\_all脚本用于打包[release](https://github.com/lwch/natpass/releases)下的可执行文件，依赖于docker
   * windows版本编译需要依赖mingw32
   * linux版本编译时需要依赖libx11

目录结构如下：

```
.
├── code // 后端代码
│   ├── client // np-cli代码
│   │   ├── app       // 项目主入口
│   │   ├── dashboard // dashboard页面代码
│   │   ├── global    // 配置文件
│   │   ├── pool      // 连接池
│   │   └── rule      // 规则处理模块
│   │       ├── shell // shell模块
│   │       └── vnc   // vnc模块
│   │           ├── define     // windows下的API和常量定义
│   │           ├── process    // vnc进程管理模块
│   │           ├── vncnetwork // vnc父进程与子进程通信协议
│   │           └── worker     // 子进程处理模块
│   ├── network // server => client通信协议
│   ├── server  // np-svr代码
│   │   ├── global  // 配置文件
│   │   └── handler // 处理逻辑
│   └── utils // 公共代码
├── conf      // 默认配置文件
├── contrib   // 打包脚本等
│   ├── bindata // bindata封装程序，同https://github.com/go-bindata/go-bindata/blobmaster/go-bindata/main.go
│   └── build   // 打包脚本
├── docs // 文档
└── html // 前端代码，dashboard、shell、vnc为对应功能的页面，其他目录均为第三方库
    ├── AdminLTE-3.1.0
    ├── bootstrap-4.6.1
    ├── dashboard
    │   ├── AdminLTE -> ../AdminLTE-3.1.0 // 软链到AdminLTE库，下同
    │   ├── bootstrap -> ../bootstrap-4.6.1
    │   ├── fontawesome -> ../fontawesome-free-5.15.4
    │   ├── jquery -> ../jquery
    │   ├── js
    │   └── templates
    ├── fontawesome-free-5.15.4
    ├── jquery
    ├── jquery-mousewheel
    ├── js // 公共js代码
    ├── shell
    │   ├── jquery -> ../jquery
    │   └── xterm.js -> ../xterm.js-4.14.1
    ├── vnc
    │   ├── AdminLTE -> ../AdminLTE-3.1.0
    │   ├── bootstrap -> ../bootstrap-4.6.1
    │   ├── fontawesome -> ../fontawesome-free-5.15.4
    │   ├── jquery -> ../jquery
    │   └── jquery-mousewheel -> ../jquery-mousewheel
    └── xterm.js-4.14.1
```
