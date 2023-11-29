# 规则配置

所有链接均为正向配置，由连接发起方进行配置

## shell规则

shell规则用于创建一个网页端的命令行操作页面

    - name: shell             # 链路名称
      target: that            # 目标客户端ID
      type: shell             # web shell
      local_addr: 0.0.0.0     # 本地监听地址
      #local_port: 8080        # 本地监听端口号
      #exec: /bin/bash        # 运行命令
                              # windows默认powershell或cmd
                              # 其他系统bash或sh
      env:                    # 环境变量设置
        - TERM=xterm

1. `name`: 该规则名称，必须全局唯一
2. `target`: 对端客户端ID
3. `type`: shell
4. `local_addr`: 本地监听地址，如只允许局域网访问可绑定在局域网IP地址上
5. `local_port`: 本地监听端口号，可选
6. `exec`: 连接建立成功后的启动命令
    - 指定该参数：直接使用设定的命令运行
    - linux系统：优先查找bash命令，若没有则查找sh命令，否则报错
    - windows系统：优先查找powershell命令，若没有则查找cmd命令，否则报错
7. `env`: 进程启动时的环境变量设置

## vnc规则

vnc规则用于创建一个网页端的远程桌面操作页面

    - name: vnc            # 链路名称
      target: that         # 目标客户端ID
      type: vnc            # web vnc
      local_addr: 0.0.0.0  # 本地监听地址
      #local_port: 5900     # 本地监听端口号
      fps: 10              # 刷新频率

1. `name`: 该规则名称，必须全局唯一
2. `target`: 对端客户端ID
3. `type`: shell
4. `local_addr`: 本地监听地址，如只允许局域网访问可绑定在局域网IP地址上
5. `local_port`: 本地监听端口号，可选
6. `fps`: 每秒钟截屏多少次，最高50

注意：

1. 创建vnc连接后远端服务会创建一个子进程进行截屏和键鼠操作，
   主进程会在`6155~6955`之间选一个端口进行监听用于与子进程通信
2. 使用rdp连接的windows主机，需要将np-cli.exe[注册为系统服务](startup.md#注册系统服务（可选）)，
   否则在rdp窗口最小化或者rdp连接关闭后将无法刷新
3. windows2008系统下需要启用sas策略才可使用ctrl+alt+del按钮进行解锁登录页面，配置方法如下：

    1. 运行gpedit.msc打开组策略编辑器
    2. 找到计算机配置 => 管理模板 => Windows组件 => Windows登录选项 => 禁用或启用软件安全注意序列
    3. 在详情中设置为已启用，设置允许哪个软件生成软件安全注意序列为*服务*

## code-server规则

vnc规则用于创建一个网页端的code-server页面，主要用于远程开发

    - name: code-server    # 链路名称
      target: remote       # 目标客户端ID
      type: code-server    # code-server
      local_addr: 0.0.0.0  # 本地监听地址
      #local_port: 8000     # 本地监听端口号

1. `name`: 该规则名称，必须全局唯一
2. `target`: 对端客户端ID
3. `type`: code-server
4. `local_addr`: 本地监听地址，如只允许局域网访问可绑定在局域网IP地址上
5. `local_port`: 本地监听端口号，可选