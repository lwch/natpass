---
title: "开发文档"
linkTitle: "开发文档"
type: docs
weight: 1
menu:
  main:
    weight: 1
---

最新版本: ![version](https://img.shields.io/github/v/release/lwch/natpass)

natpass主要用于居家办公，远程开发等使用场景，目前已支持以下功能：

1. web shell: 通过shell的方式控制远程机器
2. web vnc: 对于桌面类的操作系统，可实现远程桌面的控制
3. [code-server](https://github.com/coder/code-server): 可通过web的方式来访问远端机器上已安装好的[code-server](https://github.com/coder/code-server)服务

## 架构图

![架构图](../images/%E6%9E%B6%E6%9E%84%E5%9B%BE.png)

注: 其中虚线部分表示暂未实现的功能