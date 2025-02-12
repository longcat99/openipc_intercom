
# OpenIPC 音频端点接口

此存储库提供工具，可直接从 Web 浏览器或 Home Assistant 卡片向 OpenIPC `/play_audio` 端点发送音频。

## 兼容性

- 主要为 ssc338q 设备设计。
- 可直接部署在设备上或外部服务器上。
- 采用 Golang 编写，可交叉编译至大多数处理器架构。
- 已包含 amd64 和 mipsle 版本的二进制文件。
- 修改自 https://github.com/gtxaspec/openipc_intercom
## 设置与安装

### 编译

使用提供的脚本编译源代码：

```sh
./compile.sh
```

**注意**：为了获得最佳效果并减少二进制文件大小，确保已安装 UPX（mipsle 需要 UPX 4.0.2 或更高版本），然后运行：

```sh
./compile.sh upx
```

### 配置

1. 重命名示例配置文件：

```sh
mv config.json.example config.json
```

2. 根据你的具体需求修改 `config.json`。

### 运行服务器

启动服务器：

```sh
./intercom
```

如需调试输出，请使用 `--debug` 选项：

```sh
./intercom --debug
```

### 访问界面

已在 Chrome 浏览器上测试。  
要启用麦克风功能（由于服务器未使用 HTTPS），请将服务器 URL 添加到 Chrome 选项 `chrome://flags/#unsafely-treat-insecure-origin-as-secure`。
请使用ngins代理
推荐 https://github.com/xiaoxinpro/nginx-proxy-manager-zh
服务器运行后，可在浏览器访问：

```sh
http://<ip-address>:3333/index.html
```

## 致谢

本项目受 [simple-recorderjs-demo](https://github.com/addpipe/simple-recorderjs-demo/tree/master) 启发，并基于其代码进行开发。


本项目受 [openipc_intercom](https://github.com/gtxaspec/openipc_intercom) 启发，并基于其代码进行开发。

