# natpass

内网穿透工具

## 实现原理

基于tls链接，server端转发流量

## 编译

    CGO_ENABLED=0 go build -o bin/np-svr code/server/*.go
    CGO_ENABLED=0 go build -o bin/np-cli code/client/*.go

## 配置

server端配置如下：

    listen: 6154       # 监听端口号
    secret: 0123456789 # 预共享密钥
    tls:
    key: /dir/to/tls/key/file # tls密钥
    crt: /dir/to/tls/crt/file # tls证书

client端配置如下：

    id: this               # 客户端ID
    server: 127.0.0.1:6154 # 服务器地址
    secret: 0123456789     # 预共享密钥
    tunnel:                     # 远端tunnel列表可为空
      - name: rdp               # 链路名称
        target: that            # 目标客户端ID
        type: tcp               # 连接类型tcp或udp
        local_addr: 0.0.0.0     # 本地监听地址
        local_port: 3389        # 本地监听端口号
        remote_addr: 127.0.0.1  # 目标客户端连接地址
        remote_port: 3389       # 目标客户端连接端口号