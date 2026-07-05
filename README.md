# Tesla TeslaMate Telegram Bot

一个基于Go语言开发的Telegram Bot，用于通过TeslaMate API查看和监控Tesla车辆信息。

## 功能特性

- 🚗 **多车支持** - 从 TeslaMate API 读取车辆列表，按会话选择/切换车辆
- 🚗 **车辆信息查询** - 查看车辆详细信息（型号、VIN、外观等）
- ⚡ **实时状态监控** - 查看电量、温度、车门/车窗状态等
- 🔋 **电池健康度** - 监控电池容量和健康状态
- 🔌 **充电记录** - 查看最新的充电记录详情
- 🔐 **白名单机制** - 只允许授权用户使用Bot
- 🔄 **一键刷新** - 所有信息页面支持实时刷新
- 📱 **双重交互** - 支持命令和内联键盘两种操作方式

## 多车与选车状态

Bot 支持监控多辆 Tesla。每个 Telegram 会话（`chat_id`）可独立选择当前车辆，选择结果会持久化到本地文件，重启后仍保留。

- `/cars` 或主菜单「切换车辆」：查看车辆列表并切换
- `teslamate.car_id`：默认车辆 ID（可选）；未手动选车时作为 fallback；设为 `0` 则使用 API 返回的第一辆车
- 选车状态固定保存至 `data/car_state.json`（无需配置，首次运行自动创建）

状态文件格式示例：

```json
{
  "selections": {
    "123456789": 1,
    "987654321": 2
  }
}
```

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

#### Docker 持久化说明

选车状态通过**目录挂载**持久化（勿挂载单个文件：宿主机文件不存在时 Docker 会错误地创建同名目录）。

| 宿主机路径 | 容器内路径 | 说明 |
|-----------|-----------|------|
| `./config.toml` | `/app/config.toml` | 配置文件（只读） |
| `./data` | `/app/data` | 数据目录；`car_state.json` 由 Bot 首次运行时自动创建 |

宿主机上的状态文件路径为 `./data/car_state.json`，无需手动创建。若曾误将 `./data/car_state.json` 挂载为单文件导致生成了目录，请先删除该目录再重新 `docker compose up`。

## 许可证

MIT License

## 鸣谢

- [TeslaMate](https://github.com/teslamate-org/teslamate) - Tesla数据记录器
- [TeslaMateApi](https://github.com/tobiasehlert/teslamateapi) - TeslaMate RESTful API
