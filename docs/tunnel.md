# 隧道配置

所有隧道均为正向隧道，由连接发起方进行配置

## tcp隧道

tcp隧道用于反向代理远程的任意服务，如rdp、ssh、http等

    - name: rdp               # 链路名称
      target: that            # 目标客户端ID
      type: tcp               # 连接类型tcp或udp
      local_addr: 0.0.0.0     # 本地监听地址
      local_port: 3389        # 本地监听端口号
      remote_addr: 127.0.0.1  # 目标客户端连接地址
      remote_port: 3389       # 目标客户端连接端口号

1. `name`: 该隧道名称，必须全局唯一
2. `target`: 对端客户端ID
3. `type`: tcp
4. `local_addr`: 本地监听地址，如只允许局域网访问可绑定在局域网IP地址上
5. `local_port`: 本地监听端口号
6. `remote_addr`: 目标客户端连接地址，该地址为127.0.0.1时表示连接本机服务，也可连接局域网或广域网上的其他地址
7. `remote_port`: 目标客户端连接端口号

## shell隧道

shell隧道用于创建一个网页端的命令行操作页面

    - name: shell             # 链路名称
      target: that            # 目标客户端ID
      type: shell             # web shell
      local_addr: 0.0.0.0     # 本地监听地址
      local_port: 8080        # 本地监听端口号
      #exec: /bin/bash        # 运行命令
                              # windows默认powershell或cmd
                              # 其他系统bash或sh
      env:                    # 环境变量设置
        - TERM=xterm

1. `name`: 该隧道名称，必须全局唯一
2. `target`: 对端客户端ID
3. `type`: shell
4. `local_addr`: 本地监听地址，如只允许局域网访问可绑定在局域网IP地址上
5. `local_port`: 本地监听端口号
6. `exec`: 连接建立成功后的启动命令
    - 指定该参数：直接使用设定的命令运行
    - linux系统：优先查找bash命令，若没有则查找sh命令，否则报错
    - windows系统：优先查找powershell命令，若没有则查找cmd命令，否则报错
7. `env`: 进程启动时的环境变量设置

连接成功后即可使用浏览器访问`local_port`所对应的端口来创建shell隧道，如http://127.0.0.1:8080