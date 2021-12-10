# 开始使用

首先需要准备一张tls证书，推荐使用*阿里云*或*腾讯云*的免费证书。

注：以下示例均在*debian*系统下进行，其他系统请自行查找相关系统命令，
    windows系统可通过services.msc命令进入系统服务管理面板进行服务管理。

## server端部署

1. 在服务器上[下载](https://github.com/lwch/natpass/releases)并解压到任意目录
2. 修改server.yaml配置文件，设置*key*和*crt*参数到你的证书所在路径
3. 修改conf.d/conn.yaml配置文件，修改*secret*密钥，建议使用以下命令生成随机16位密钥

        tr -dc A-Za-z0-9 < /dev/urandom|dd bs=16 count=1 2>/dev/null
4. 修改conf.d/log.yaml配置文件，将log目录设置为一个绝对路径
5. 使用以下命令将np-svr注册为系统服务，其中-conf参数后跟配置文件所在路径，-user参数后为程序启动身份（建议使用nobody身份启动）

        sudo ./np-svr -conf server.yaml -action install -user nobody
6. 手动创建第4步中设置的log目录，并将其属主设置为nobody

        sudo mkdir -p <dir>
        sudo chown -R nobody <dir>
7. 将该服务设置为系统启动项，并启动服务

        sudo systemctl enable np-svr
        sudo systemctl start np-svr

## 受控端部署

1. 在受控端机器上[下载](https://github.com/lwch/natpass/releases)并解压到任意目录
2. 修改client.yaml配置文件，设置*id*为remote，设置*server*地址，并删除以下配置

        #include rule.d/*.yaml
3. 修改conf.d/conn.yaml配置文件，修改*secret*密钥，该密钥必须与服务器端保持一致
4. 修改conf.d/log.yaml配置文件，将log目录设置为一个绝对路径
5. 使用以下命令将np-cli注册为系统服务，其中-conf参数后跟配置文件所在路径，-user参数后为程序启动身份（建议使用nobody身份启动）

        sudo ./np-cli -conf client.yaml -action install -user nobody
6. 手动创建第4步中设置的log目录，并将其属主设置为nobody

        sudo mkdir -p <dir>
        sudo chown -R nobody <dir>
7. 将该服务设置为系统启动项，并启动服务

        sudo systemctl enable np-cli
        sudo systemctl start np-cli

**注：当受控端为linux操作系统时，请使用np-cli.vnc程序进行启动，暂不支持系统服务方式启动，手工启动命令如下**

        ./np-cli -conf client.yaml

## 控制端部署

1. 在本地控制机上[下载](https://github.com/lwch/natpass/releases)并解压到任意目录
2. 修改client.yaml配置文件，设置*id*为local，设置*server*地址
3. 修改conf.d/conn.yaml配置文件，修改*secret*密钥，该密钥必须与服务器端保持一致
4. 修改conf.d/log.yaml配置文件，将log目录设置为一个绝对路径
5. 修改rule.d目录下的规则配置文件，[rule配置方法](rules.md)
6. 使用以下命令将np-cli注册为系统服务，其中-conf参数后跟配置文件所在路径，-user参数后为程序启动身份（建议使用nobody身份启动）

        sudo ./np-cli -conf client.yaml -action install -user nobody
7. 手动创建第4步中设置的log目录，并将其属主设置为nobody

        sudo mkdir -p <dir>
        sudo chown -R nobody <dir>
8. 将该服务设置为系统启动项，并启动服务

        sudo systemctl enable np-cli
        sudo systemctl start np-cli
9. 在以上操作成功后即可在浏览器中通过client.yaml中配置的端口号进行访问，默认地址：

        http://127.0.0.1:8080

## 注册系统服务

1. 在命令行中使用`-action install`参数即可将程序注册为系统服务，使用参数`-user`可设置该服务的启动身份
2. linux系统使用systemd管理系统服务，windows系统可用services.msc面板启动或停止服务