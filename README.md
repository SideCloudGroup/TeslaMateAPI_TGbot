# Tesla TeslaMate Telegram Bot

一个基于Go语言开发的Telegram Bot，用于通过TeslaMate API查看和监控Tesla车辆信息。

## 功能特性

- 🚗 **车辆信息查询** - 查看车辆详细信息（型号、VIN、外观等）
- ⚡ **实时状态监控** - 查看电量、温度、车门/车窗状态等
- 🔋 **电池健康度** - 监控电池容量和健康状态
- 🔌 **充电记录** - 查看最新的充电记录详情
- 🔐 **白名单机制** - 只允许授权用户使用Bot
- 🔄 **一键刷新** - 所有信息页面支持实时刷新
- 📱 **双重交互** - 支持命令和内联键盘两种操作方式

## 部署

### Docker (推荐)

#### 使用 Docker Compose (推荐)

1. 下载 `docker-compose.yml` 和 `config.example.toml`：

```bash
wget https://github.com/SideCloudGroup/TeslaMateAPI_TGbot/raw/refs/heads/main/docker-compose.yml
wget https://github.com/SideCloudGroup/TeslaMateAPI_TGbot/raw/refs/heads/main/config.example.toml -O config.toml
```

2. 编辑 `config.toml` 文件

3. 使用 Docker Compose 运行：

```bash
docker compose up -d
```

## 许可证

MIT License

## 鸣谢

- [TeslaMate](https://github.com/teslamate-org/teslamate) - Tesla数据记录器
- [TeslaMateApi](https://github.com/tobiasehlert/teslamateapi) - TeslaMate RESTful API
