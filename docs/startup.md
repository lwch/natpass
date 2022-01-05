# 开始使用

以下示例均在*debian*系统下进行，其他系统请自行查找相关系统命令，
    windows系统可通过services.msc命令进入系统服务管理面板进行服务管理。

## 服务器端部署

1. 在服务器上[下载](https://github.com/lwch/natpass/releases)并解压到任意目录
2. 使用以下命令将np-svr注册为系统服务，其中-conf参数后跟配置文件所在路径，-user参数后为程序启动身份（建议使用nobody身份启动）

        sudo ./np-svr -conf server.yaml -action install -user nobody
3. 启动服务

        sudo systemctl start np-svr

4. （可选）开放外网防火墙，默认端口6154

## 受控端部署

1. 在受控端机器上[下载](https://github.com/lwch/natpass/releases)并解压到任意目录
2. （可选）修改remote.yaml配置文件，修改*server*地址
3. 使用以下命令将np-cli注册为系统服务，其中-conf参数后跟配置文件所在路径，-user参数后为程序启动身份

        sudo ./np-cli -conf remote.yaml -action install -user `whoami`
4. 启动服务

        sudo systemctl start np-cli

## 控制端部署

1. 在本地控制机上[下载](https://github.com/lwch/natpass/releases)并解压到任意目录
2. （可选）修改local.yaml配置文件，修改*server*地址
3. （可选）修改rule.d目录下的规则配置文件，[rule配置方法](rules.md)
4. 使用以下命令将np-cli注册为系统服务，其中-conf参数后跟配置文件所在路径，-user参数后为程序启动身份（建议使用nobody身份启动）

        sudo ./np-cli -conf local.yaml -action install -user nobody
5. 启动服务

        sudo systemctl start np-cli
6. 在以上操作成功后即可在浏览器中通过local.yaml中配置的端口号进行访问，默认地址：

        http://127.0.0.1:8080

## 安全连接（可选）

1. 建议使用tls加密连接，使用方式如下

        - 修改服务器端的server.yaml文件，配置tls相关文件路径，并重启服务
        - 修改受控端的remote.yaml配置，将ssl设置为true，并重启服务
        - 修改控制端的local.yaml配置，将ssl设置为true，并重启服务

2. 修改默认连接密钥，修改方式如下

        - 使用以下命令生成一个16位随机串
          tr -dc A-Za-z0-9 < /dev/urandom|dd bs=16 count=1 2>/dev/null
        - 修改服务器端的common.yaml文件，将secret设置为新的密钥，并重启服务
        - 修改受控端的common.yaml文件，将secret设置为新的密钥，并重启服务
        - 修改控制端的common.yaml文件，将secret设置为新的密钥，并重启服务

## 注册系统服务（可选）

1. 在命令行中使用`-action install`参数即可将程序注册为系统服务，使用参数`-user`可设置该服务的启动身份
2. linux系统使用systemd管理系统服务，windows系统可用services.msc面板启动或停止服务