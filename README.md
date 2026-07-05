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
- `teslamate.state_file`：选车状态文件路径，默认 `car_state.json`

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

2. 编辑 `config.toml` 文件（确保 `state_file` 与下方挂载路径一致，默认 `car_state.json`）

3. 创建选车状态持久化文件（**首次部署必须**，否则容器内无法写入状态）：

```bash
mkdir -p data
echo '{"selections":{}}' > data/car_state.json
```

4. 使用 Docker Compose 运行：

```bash
docker compose up -d
```

#### Docker 持久化说明

`docker-compose.yml` 将以下路径挂载到容器内，请确保宿主机对应文件/目录存在：

| 宿主机路径 | 容器内路径 | 说明 |
|-----------|-----------|------|
| `./config.toml` | `/app/config.toml` | 配置文件（只读） |
| `./data/car_state.json` | `/app/car_state.json` | 各会话选车状态（**需可写**） |

`config.toml` 中 `teslamate.state_file` 须与容器内路径一致，即：

```toml
state_file = "car_state.json"
```

若修改 `state_file` 为其他路径，需同步调整 `docker-compose.yml` 中的 volume 挂载。

## 许可证

MIT License

## 鸣谢

- [TeslaMate](https://github.com/teslamate-org/teslamate) - Tesla数据记录器
- [TeslaMateApi](https://github.com/tobiasehlert/teslamateapi) - TeslaMate RESTful API
