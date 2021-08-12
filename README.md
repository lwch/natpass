# natpass

内网穿透工具

## 实现原理

基于tls链接，server端转发流量

## 编译

    ./build

## 配置

server端配置如下：

    listen: 6154       # 监听端口号
    secret: 0123456789 # 预共享密钥
    log:
      dir: ./logs # 路径
      size: 50M   # 单个文件大小
      rotate: 7   # 保留数量
    tls:
      key: /dir/to/tls/key/file # tls密钥
      crt: /dir/to/tls/crt/file # tls证书

client端配置如下：

    id: this               # 客户端ID
    server: 127.0.0.1:6154 # 服务器地址
    secret: 0123456789     # 预共享密钥，必须与server端相同，否则握手失败
    log:
      dir: ./logs # 路径
      size: 50M   # 单个文件大小
      rotate: 7   # 保留数量
    tunnel:                     # 远端tunnel列表可为空
      - name: rdp               # 链路名称
        target: that            # 目标客户端ID
        type: tcp               # 连接类型tcp或udp
        local_addr: 0.0.0.0     # 本地监听地址
        local_port: 3389        # 本地监听端口号
        remote_addr: 127.0.0.1  # 目标客户端连接地址
        remote_port: 3389       # 目标客户端连接端口号

## 运行

server端运行：

    ./np-svr -conf server.yaml

client端运行：

    ./np-cli -conf client.yaml