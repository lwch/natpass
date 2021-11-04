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

## vnc隧道

vnc隧道用于创建一个网页端的远程桌面操作页面，目前仅支持*windows*操作系统

    - name: vnc            # 链路名称
      target: that         # 目标客户端ID
      type: vnc            # web vnc
      local_addr: 0.0.0.0  # 本地监听地址
      local_port: 5900     # 本地监听端口号
      fps: 10              # 刷新频率

1. `name`: 该隧道名称，必须全局唯一
2. `target`: 对端客户端ID
3. `type`: shell
4. `local_addr`: 本地监听地址，如只允许局域网访问可绑定在局域网IP地址上
5. `local_port`: 本地监听端口号
6. `fps`: 每秒钟截屏多少次，最高50

连接成功后即可使用浏览器访问`local_port`所对应的端口来创建vnc隧道，如http://127.0.0.1:5900

注意：

1. 创建vnc连接后远端服务会创建一个子进程进行截屏和键鼠操作，
   主进程会在`6155~6955`之间选一个端口进行监听用于与子进程通信
2. 使用rdp连接的windows主机，需要将np-cli.exe[注册为系统服务](startup.md#注册系统服务)，
   否则在rdp窗口最小化或者rdp连接关闭后将无法刷新
3. windows2008系统下需要启用sas策略才可使用ctrl+alt+del按钮进行解锁登录页面，配置方法如下：

    1. 运行gpedit.msc打开组策略编辑器
    2. 找到计算机配置 => 管理模板 => Windows组件 => Windows登录选项 => 禁用或启用软件安全注意序列
    3. 在详情中设置为已启用，设置允许哪个软件生成软件安全注意序列为*服务*