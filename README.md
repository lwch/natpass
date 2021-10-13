# natpass

[![natpass](https://github.com/lwch/natpass/actions/workflows/build.yml/badge.svg)](https://github.com/lwch/natpass/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/lwch/natpass)](https://goreportcard.com/report/github.com/lwch/natpass)
[![go-mod](https://img.shields.io/github/go-mod/go-version/lwch/natpass)](https://github.com/lwch/natpass)
[![license](https://img.shields.io/github/license/lwch/natpass)](https://opensource.org/licenses/MIT)
[![platform](https://img.shields.io/badge/platform-linux%20%7C%20windows%20%7C%20macos-lightgrey.svg)](https://github.com/lwch/natpass)
[![QQ群711086098](https://img.shields.io/badge/QQ%E7%BE%A4-711086098-success)](https://jq.qq.com/?_wv=1027&k=6Fz2vkVE)

NAT内网穿透工具，支持tcp隧道、shell隧道，[实现原理](docs/desc.md)

1. [如何部署](docs/startup.md)
2. [隧道配置](docs/tunnel.md)

## shell隧道

linux命令行效果

![linux-shell](docs/shell_linux.png)

windows命令行效果

![windows-shell](docs/shell_win.png)

## iperf3压测对比

使用相同参数，iperf3压测1分钟

    # natpass10路复用，读写均为1s超时
    [ ID] Interval           Transfer     Bitrate         Retr
    [  5]   0.00-60.00  sec  70.0 MBytes  9.79 Mbits/sec   22             sender
    [  5]   0.00-60.02  sec  57.9 MBytes  8.10 Mbits/sec                  receiver

    # frp10路复用stcp，tls
    [ ID] Interval           Transfer     Bitrate         Retr
    [  5]   0.00-60.00  sec  66.2 MBytes  9.26 Mbits/sec   31             sender
    [  5]   0.00-60.10  sec  57.7 MBytes  8.05 Mbits/sec                  receiver

## TODO

1. ~~支持include的yaml配置文件~~
2. ~~通用的connect、connect_response、disconnect消息~~
3. 所有隧道的portal页面
4. 文件传输
5. web远程桌面
6. 流量监控统计页面，server还是client?