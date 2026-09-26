# greeting

基于 Kratos 的 Go 服务示例，覆盖 HTTP、gRPC、WebSocket、MQTT、RabbitMQ、TCP、UDP 和 KCP。默认只启动 HTTP/gRPC，不需要数据库、Redis 或消息中间件。

Go module：`github.com/itsping999/greeting`。

## Agent 开发入口

让 Agent 从 [AGENTS.md](AGENTS.md) 开始；业务改动地图和验收流程见 [开发指南](docs/development.md)。它们适用于不同 Agent，应用不依赖 AI 平台。

```bash
make doctor                    # 检查本机工具
make task NAME=order-create     # 创建需求、契约、验收及交接任务单
# 按业务分层实现；proto / Wire 变更后 make all
make fmt
make check                     # 格式、vet、race、生成一致性、构建
```

`make check` 的生成检查需要 protoc 和 `make init` 安装的工具；在临时副本生成，不改动当前工作区。协议或启动行为变化再运行 `make smoke`。任务单记录真实验证结果与未完成事项，便于 Agent 接续工作。

## 快速开始

需要 Go 1.27.1 或更高版本。生成代码已包含在仓库里，首次运行无需安装 protoc 或生成器。

```bash
git clone git@github.com:itsping999/greeting.git
cd greeting
make build
./bin/server -conf ./configs/config.yaml
```

另开终端验证 HTTP 和 gRPC：

```bash
./bin/client -protocol http
./bin/client -protocol grpc
curl -fsS http://127.0.0.1:8000/healthz
```

客户端会调用问候接口，再按返回 ID 查询同一条记录。输出包含 `Hello World` 和 `saved and retrieved id=...`。`/healthz` 返回 `{"status":"ok"}`，表示 HTTP 入口存活。

`make run` 可直接运行源码。使用 Ctrl+C 或 SIGTERM 退出；服务停止接收连接，并关闭已建立的 WebSocket、TCP 和 KCP 连接。

配置中每个协议都有 `enabled` 开关。只校验、组装和启动启用的协议；至少需要启用一种。`-conf` 默认指向一个文件，避免把基础配置、容器配置和全协议示例配置一起加载。

## 问候与数据示例

| HTTP 接口 | 行为 |
|---|---|
| `GET /helloworld/{name}` | 生成问候，并在进程内保存记录；保留原示例路由 |
| `GET /greetings/{id}` | 查询已保存记录 |
| `GET /healthz` | HTTP 存活检查 |

```bash
curl -fsS http://127.0.0.1:8000/helloworld/World
# {"message":"Hello World","id":"1"}
curl -fsS http://127.0.0.1:8000/greetings/1
# 返回同一条问候
```

ID 是 protobuf `int64`，在 HTTP JSON 中编码为字符串。名称去除首尾空白后须包含 1–256 个字符；非法名称返回 400，非法 ID 返回 400，不存在的记录返回 404、原因 `GREETING_NOT_FOUND`。gRPC 提供同名的 `SayHello` 和 `GetGreeting` 方法。

数据层是带并发保护的**内存实现**，支持保存和按 ID 查询，返回数据副本。重启清空数据。需要持久化时替换 `GreeterRepo` 实现。

## 其他协议

独立的 `configs/examples.yaml` 启用全部八种协议。示例客户端有超时，并检查回复内容；任何验证失败都会以非零状态退出。

| 协议 | 本地地址 | 验证命令 | 预期结果 |
|---|---|---|---|
| HTTP | `127.0.0.1:8000` | `./bin/client -protocol http` | 问候保存后按 ID 查询成功 |
| gRPC | `127.0.0.1:9000` | `./bin/client -protocol grpc` | 同上 |
| WebSocket | `ws://127.0.0.1:8001/ws` | `./bin/client -protocol websocket` | 两个客户端均收到广播 |
| MQTT | `tcp://127.0.0.1:1883` | `./bin/client -protocol mqtt` | 请求与回复的 ID 一致 |
| RabbitMQ | `127.0.0.1:5672` | `./bin/client -protocol rabbitmq` | 请求与回复的 ID 一致 |
| TCP | `127.0.0.1:9001` | `./bin/client -protocol tcp` | 字节回显一致 |
| UDP | `127.0.0.1:9002` | `./bin/client -protocol udp` | 数据报回显一致 |
| KCP | `127.0.0.1:9003` | `./bin/client -protocol kcp` | 字节回显一致 |

启动全协议示例需要 Docker Compose 和 OpenSSL：

```bash
docker compose --profile examples up -d --wait mqtt rabbitmq
export GREETING_KCP_KEY=$(openssl rand -hex 32)
export GREETING_KCP_SALT=$(openssl rand -hex 16)
./bin/server -conf ./configs/examples.yaml
```

客户端在另一个终端运行；KCP 客户端和服务端需要一致的 `GREETING_KCP_KEY`、`GREETING_KCP_SALT`。可以在同一终端后台运行服务来共享环境变量。退出示例后清理本地中间件：

```bash
docker compose --profile examples down --volumes
```

`-addr` 可覆盖客户端地址，`-name` 可覆盖名称，`-timeout` 可覆盖整体请求超时。WS/TCP/KCP 的服务端 timeout 是读写超时，空闲连接会关闭；UDP 的 timeout 用于读写等待，读超时后继续接收。

### WebSocket 消息

使用标准文本 JSON。发送：

```json
{"type":1,"name":"World"}
```

所有已连接客户端收到：

```json
{"type":1,"message":"Hello World"}
```

不支持的类型或非法名称仅向发送者返回 `error`。服务限制消息大小为 4 KiB，并采用 gorilla/websocket 默认的同源检查。客户端命令会连接两个会话来验证广播。

### MQTT / RabbitMQ 消息

请求 JSON：

```json
{"request_id":"demo-1","name":"World"}
```

回复 JSON：

```json
{"request_id":"demo-1","message":"Hello World","id":1}
```

`request_id` 必须是 1–64 位字母、数字、下划线或连字符。先订阅回复，再发送请求：

| 协议 | 请求目标 | 回复目标 |
|---|---|---|
| MQTT | `greeting/requests` | `greeting/replies/{request_id}` |
| RabbitMQ | exchange `greeting`，routing key `greeting.request` | 同一 exchange，routing key `greeting.reply.{request_id}` |

RabbitMQ 服务消费持久队列 `greeting.requests`，处理及发布回复成功后确认消息；客户端使用独占、自动删除的回复队列。非法名称返回带同一 request_id 的 `error` 回复；非法 request_id 拒绝处理。MQTT 客户端以 QoS 1 发送请求。示例展示请求与回复，不提供持久化业务处理或幂等保证。

Compose 的中间件仅绑定本机回环地址。Mosquitto 匿名访问和 RabbitMQ 默认账户用于本地示例。

## 构建与容器

```bash
make build                         # bin/server 和 bin/client
make docker-build-linux-amd64       # bin/amd64/app
make docker-build-linux-arm         # bin/arm/app，ARMv7
make docker-build-linux-arm64       # bin/arm64/app

docker build -t greeting:local .
docker run --rm -p 127.0.0.1:8000:8000 -p 127.0.0.1:9000:9000 greeting:local
# 或：
docker compose --profile app up --build greeting
```

镜像使用独立的 `configs/container.yaml`，默认只启用 HTTP/gRPC，以非 root 用户运行。构建上下文排除本机 `configs/config.yaml`，避免把本机配置放进镜像。自定义配置可只读挂载并通过 `-conf` 指定：

```bash
docker run --rm -p 127.0.0.1:8000:8000 \
  -v "$PWD/configs/container.yaml:/data/config.yaml:ro" \
  greeting:local -conf /data/config.yaml
```

## 生成与验证

修改 proto 或依赖组装后重新生成。需要 protoc；本次生成使用 protoc 36.2。生成器固定版本，安装到仓库内忽略的 `.tools/bin`。

```bash
make init
make all          # API 和 Wire；运行配置使用 Go 结构体
make test         # go test -race ./...
make smoke        # 临时端口验证六种无中间件协议、404 和 SIGTERM 退出
make examples-smoke  # 临时本地容器验证全部八种协议
```

冒烟脚本使用临时配置和自动分配的回环端口，不读取本机业务配置。`examples-smoke` 为本次运行创建独立 Compose 项目，结束时清理所属容器、网络及卷。HTTP/gRPC 验证保存后查询；WebSocket 验证广播；MQTT/RabbitMQ 验证请求与回复；TCP/UDP/KCP 验证回显。单元测试覆盖配置、并发存储、HTTP 错误、连接超时和退出。

业务实现放在 `internal/biz`，传输适配放在 `internal/server`，入口与返回值映射放在 `internal/service`，数据访问放在 `internal/data`。单体可以组装多个业务模块，独立服务可以只组装所需模块；当前默认入口是一个示例进程。

## 基础设施与脚手架

合并保留的 `internal/pkg` 提供可选 PostgreSQL、Redis、MongoDB、Kubernetes 服务发现、远程 gRPC、日志、指标和链路追踪能力。运行配置使用普通 Go 结构体；结构和部署环境变量入口见 `internal/conf/config.go`、各模块 config.go 和 `cmd/server/env.go`。默认外部依赖全部关闭。启用数据库只创建客户端，问候仓储仍是内存实现；启用 `data.remote_grpc.greeter` 时保存和查询转发到远程 API。

HTTP 额外提供 `/readyz`（已启用的数据库客户端检查）、`/meta` 和 `/metrics`。部署示例在 `deploy/k8s`，可用 `kubectl kustomize deploy/k8s` 渲染，ServiceMonitor 另行启用。

```bash
make init
make scaffold-module PROTO=order/v1/order.proto CRUD=1
make verify            # 等同 make check
make generated-check   # 等同 make check-generated
make e2e-smoke          # HTTP 健康、就绪、元信息和业务入口
```

脚手架自动生成 proto / service / biz / data 桩并接通注册和 provider，仍需实现业务规则、存储和验收测试。详细贡献流程见 [CONTRIBUTING.md](CONTRIBUTING.md)。
