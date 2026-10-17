# greeting

基于 Kratos 的 Go 业务模板，提供 HTTP/gRPC 问候示例，以及按需启用的 WebSocket、MQTT、RabbitMQ、TCP、UDP 和 KCP。默认只启动 HTTP/gRPC，不要求数据库、Redis 或消息中间件。

Go module：`github.com/itsping999/greeting`。

## 文档与开发入口

Agent 从 [AGENTS.md](AGENTS.md) 开始，人和 Agent 共用 [文档总导航](docs/README.md) 与 [开发流程](docs/development/README.md#业务改动地图)。应用运行不依赖 AI 平台。

| 文档分类 | 入口 |
| --- | --- |
| 目标、业务规则与验收 | [需求](docs/requirements/README.md) |
| 模块、运行、数据与安全 | [架构](docs/architecture/README.md) |
| 上手、配置步骤、开发与测试 | [开发](docs/development/README.md) |
| API、协议、CLI 与配置输入 | [接口](docs/interfaces/README.md) |
| 构建、部署、排障与恢复 | [交付与运维](docs/operations/README.md) |
| 实施范围、进度和实际验收 | [任务计划](docs/plans/README.md) |

## 快速开始

需要 Go 1.27.1 或更高版本、make 和 Python 3.9 及以上。生成代码已包含在仓库中，普通构建无需安装 protoc 或生成器。

```sh
make doctor
make build
./bin/server -conf ./configs/config.yaml
```

另开终端：

```sh
./bin/client -protocol http
./bin/client -protocol grpc
curl -fsS http://127.0.0.1:8000/healthz
```

客户端创建问候后按返回 ID 查询同一记录；Ctrl+C 或 SIGTERM 停止。更多前提与阻塞处理见 [开发上手](docs/development/README.md#开发上手)。

## 修改与验证

```sh
# 复杂跨层任务才创建计划
make task NAME=order-create
# 按分层实现；proto / Wire 变更后 make all
make fmt
make check
```

复杂任务写入 docs/plans/ 并登记索引，简单修复不需要任务单。Go 改动执行上述检查；纯文档只运行 make check-docs 和 git diff --check。make check 保留完整代码检查，生成检查需要 protoc 和固定工具，在临时副本运行。各类 smoke 前提见 [验证选择](docs/development/README.md#验证选择)。

## 示例与交付边界

问候本地仓储是并发安全的内存实现，重启清空；历史 GET 创建接口保留保存副作用。启用远程 Greeter 时保存和查询均转发。数据库客户端初始化不表示业务已持久化，可选事件发布失败不回滚主操作。

HTTP/gRPC 字段、错误和兼容规则见 [问候契约](docs/interfaces/README.md#问候接口)。其他协议的地址、消息和客户端参数见 [协议参考](docs/interfaces/protocols.md)，健康/就绪/元信息/指标见 [健康契约](docs/interfaces/README.md#健康与观测)。

镜像、Kubernetes、Broker/KCP 示例及资源清理见 [部署说明](docs/operations/README.md#部署与协议示例)，各架构制品与授权边界见 [发布流程](docs/operations/README.md#构建与交付)。模板的本地验证不替代生产部署或恢复验收。

新模块优先使用 make scaffold-module，可选 CRUD=1；分层桩和接线生成后仍需实现业务与真实验收。贡献入口见 [CONTRIBUTING.md](CONTRIBUTING.md)，通用分类及文档完成门槛见 [文档规范](docs/documentation-standard.md)。
