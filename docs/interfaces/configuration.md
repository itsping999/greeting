# 配置来源、字段与环境输入

- 类型：配置与启动输入参考
- 适用范围：greeting 服务端与示例客户端
- 维护责任：配置、传输与基础设施维护者
- 核对日期：2026-10-13（输入定义与校验；不读取本机配置值）

## 来源与默认边界

服务端 -conf 默认 ./configs/config.yaml，可指定单个 YAML；传目录时只读取其中 config.yaml。先解析文件，再应用下表显式环境覆盖，最后校验启用的传输。不存在任意 APP_* 自动覆盖，也不自动合并多个配置文件。

默认仓库配置只启用 HTTP/gRPC；可选协议与外部基础设施默认关闭。这是示例配置的选择，bool 字段缺省为 false，不表示未配置字段会自动启用能力。至少需要一种传输。

| 配置组/字段 | 类型与含义 | 约束 |
| --- | --- | --- |
| server.http / grpc | enabled(bool)、network(string)、addr(string)、timeout(duration) | 启用时 addr 为 host:port，timeout 不得为负 |
| server.websocket | enabled、addr、path、timeout | path 以 / 开头；受读写超时与消息上限约束 |
| server.tcp / udp / kcp | enabled、addr、timeout；KCP 另有 key/salt | KCP 启用需非空 key/salt，客户端一致 |
| server.mqtt / rabbitmq | enabled 与 Broker/请求回复配置 | MQTT URL 为 tcp/ssl，RabbitMQ 为 amqp/amqps |
| server.middleware.ratelimit_enabled | bool，启用现有限流中间件 | 不代表业务认证或授权 |
| data.postgres / redis / mongodb | enabled、连接输入及相应池/超时配置 | 启用能力才初始化；不改变问候内存仓储 |
| data.remote_grpc.greeter | enabled、target、timeout、circuitbreaker_enabled | 启用需目标和正超时，保存/查询均转发 |
| data.discovery.kubernetes | enabled、mode、连接/命名空间与超时等 | 按部署模式准备访问配置 |
| logger | level、format、caller、filename、轮转与输出配置 | 敏感信息不进入日志 |
| tracing | enabled、exporter、endpoint、sampler/ratio、headers、timeout 等 | 按实际 exporter 配置，不公开认证 headers |

duration 使用带单位的时长，例如 5s；WS/TCP/UDP/KCP 的无正 timeout 情况使用现有 5s 兜底，HTTP/gRPC 不指定正 timeout 时沿其传输行为。各字段的业务含义与边界由所属能力维护，本表不宣称一份配置保证全部依赖可用。

## 显式环境覆盖

| 环境变量 | 覆盖字段 |
| --- | --- |
| APP_SERVER_HTTP_ADDR / APP_SERVER_GRPC_ADDR | server.http.addr / server.grpc.addr |
| APP_SERVER_HTTP_ENABLED / APP_SERVER_GRPC_ENABLED | 对应 enabled |
| APP_SERVER_MIDDLEWARE_RATELIMIT_ENABLED | server.middleware.ratelimit_enabled |
| APP_DATA_POSTGRES_ENABLED / APP_DATA_POSTGRES_SOURCE | data.postgres.enabled / source |
| APP_DATA_REDIS_ENABLED / APP_DATA_REDIS_ADDR / APP_DATA_REDIS_PASSWORD | data.redis.enabled / addr / password |
| APP_DATA_MONGODB_ENABLED / APP_DATA_MONGODB_URI | data.mongodb.enabled / uri |
| APP_DATA_DISCOVERY_KUBERNETES_ENABLED | data.discovery.kubernetes.enabled |
| APP_DATA_REMOTE_GRPC_GREETER_ENABLED / APP_DATA_REMOTE_GRPC_GREETER_TARGET | data.remote_grpc.greeter.enabled / target |
| APP_TRACING_ENABLED / APP_TRACING_ENDPOINT | tracing.enabled / endpoint |
| APP_LOGGER_LEVEL | logger.level |
| GREETING_KCP_KEY / GREETING_KCP_SALT | server.kcp.key / salt |

已存在的环境变量覆盖文件值，包括空字符串；布尔字段按布尔解析，非法值使启动失败且错误不回显原值。不向文档复制实际运行值。

本地准备见 [配置步骤](../development/README.md#本地配置)，交付、重启与生效验证见 [部署说明](../operations/README.md#部署与协议示例)，客户端参数见 [协议与 CLI](protocols.md)。
