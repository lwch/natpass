# S3 - shell



shell功能的主要流程如下：

![流程图](../.gitbook/assets/shell.main.drawio.png)

1. 当用户点击终端页面的连接按钮后，会调用`/new`接口创建一个到远端的连接
2. 在`/new`接口中会优先发送一个connect\_req消息来创建一个shell的连接
3. 等待远端返回connect\_rep消息，若连接创建成功则返回该链接的ID
4. 创建`/ws/<linkid>`的websocket连接
5. 调用`/resize`接口调整pty设备的大小
6. 开始转发数据
