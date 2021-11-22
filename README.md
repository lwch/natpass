# natpass

[![natpass](https://github.com/lwch/natpass/actions/workflows/build.yml/badge.svg)](https://github.com/lwch/natpass/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/lwch/natpass)](https://goreportcard.com/report/github.com/lwch/natpass)
[![go-mod](https://img.shields.io/github/go-mod/go-version/lwch/natpass)](https://github.com/lwch/natpass)
[![license](https://img.shields.io/github/license/lwch/natpass)](https://opensource.org/licenses/MIT)
[![platform](https://img.shields.io/badge/platform-linux%20%7C%20windows%20%7C%20macos-lightgrey.svg)](https://github.com/lwch/natpass)
[![QQ群711086098](https://img.shields.io/badge/QQ%E7%BE%A4-711086098-success)](https://jq.qq.com/?_wv=1027&k=6Fz2vkVE)

新一代主机管理工具，支持shell管理，支持远程桌面管理[实现原理](docs/desc.md)

1. [如何部署](docs/startup.md)
2. [规则配置](docs/rules.md)

功能与特性：

1. 较小的内存占用（约20M左右）
2. 支持tls安全连接
3. 支持多路异步IO
4. 支持虚拟链路层
5. 支持链路和终端会话监控
6. protobuf数据编码
7. 支持web shell
   - linux和mac系统支持创建pty设备和颜色输出
   - windows系统支持powershell
8. 支持web vnc
   - 支持基本的键盘鼠标操作
   - 支持全屏显示
   - 支持滚动
   - 支持远程剪贴板设置与读取
9. 支持多种操作系统
   - [x] linux
   - [x] windows
   - [x] macos

## 效果图

dashboard页面

![dashboard](docs/imgs/dashboard.png)

linux命令行效果

![linux-shell](docs/imgs/shell_linux.png)

windows命令行效果

![windows-shell](docs/imgs/shell_win.png)

windows2008远程桌面

![win2008-vnc](docs/imgs/vnc_win2008.png)

windows10远程桌面

![win10-vnc](docs/imgs/vnc_win10.png)

windows11远程桌面

![win11-vnc](docs/imgs/vnc_win11.png)

windows读取剪贴板内容

![vnc-clipboard](docs/imgs/vnc_clipboard.png)

## TODO

1. ~~支持include的yaml配置文件~~
2. ~~通用的connect、connect_response、disconnect消息~~
3. ~~dashboard页面~~
4. 文件传输
5. ~web远程桌面~
6. ~~流量监控统计页面，server还是client?~~
7. web端管理规则
8. 支持录屏

## stars

![stars](https://starchart.cc/lwch/natpass.svg)

## 免责声明

本软件仅用于个人研究学习，包括但不限于以下条款：

1. 严禁用于黑客攻击、远程控制他人计算机等违法违规行为
2. 软件使用者使用该软件造成的任何损失均与软件作者无关，
   一切后果由使用者自己负责
3. 严禁用于一切商业用途，包括但不限于提供云桌面、云主机等