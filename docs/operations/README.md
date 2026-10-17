# 构建、部署与排障

- 适用范围：greeting 制品与运行示例
- 维护责任：发布、部署与运维维护者
- 核对日期：2026-10-18（Go 制品基线；未执行现场部署或恢复演练）

## 构建与交付

普通构建使用已有生成物。生成变化先准备 protoc 与 make init 的固定工具，运行 make all 后审查 API、OpenAPI 和 Wire，运行配置使用 Go 结构体。代码交付按 [验证选择](../development/README.md#验证选择) 执行 make fmt、make check 及相关 smoke；纯文档使用文档检查。

| 命令 | 制品 |
| --- | --- |
| make build | bin/server、bin/client |
| make docker-build-linux-amd64 | bin/amd64/app |
| make docker-build-linux-arm | bin/arm/app，ARMv7 |
| make docker-build-linux-arm64 | bin/arm64/app |
| make docker-build | greeting:local 镜像 |

记录源码版本、构建命令、目标 OS/架构、版本参数与实际制品。构建成功不授权提交、推送、部署或发送消息，不代表已经交付目标环境。生成物与源一起审查，保留无关改动。

当前最低工具链、CI 和 Docker 构建镜像统一为 Go 1.27.2；制品须重新构建才能使用修复后的标准库与依赖。依赖扫描与上游未修复项按 [安全检查](../development/README.md#依赖安全检查) 和任务记录独立确认，不以版本号变更或本地通过声明 GitHub 告警已关闭。

## 部署与协议示例

镜像使用 configs/container.yaml，默认 HTTP/gRPC，以非 root 用户运行；本机 configs/config.yaml 不进入构建上下文。

```sh
make docker-build
docker run --rm -p 127.0.0.1:8000:8000 -p 127.0.0.1:9000:9000 greeting:local
# 或使用 app profile
docker compose --profile app up --build greeting
# 明确配置只读挂载
docker run --rm -p 127.0.0.1:8000:8000 -v "$PWD/configs/container.yaml:/data/config.yaml:ro" greeting:local -conf /data/config.yaml
```

配置字段与覆盖见 [配置参考](../interfaces/configuration.md)。交付确认实际目标、制品、配置来源、运行用户、权限、监听、代理/TLS、凭据渠道及验收。Kubernetes 示例见 [deploy/k8s](../../deploy/k8s/)，用 kubectl kustomize deploy/k8s 渲染；默认不要求 ServiceMonitor CRD，另行启用。渲染不等于部署，应用需明确目标与授权。

全协议使用 configs/examples.yaml，MQTT/RabbitMQ 需要 Docker Compose，KCP 需要 OpenSSL。以下命令创建本地资源并启动应用：

```sh
docker compose --profile examples up -d --wait mqtt rabbitmq
export GREETING_KCP_KEY=$(openssl rand -hex 32)
export GREETING_KCP_SALT=$(openssl rand -hex 16)
./bin/server -conf ./configs/examples.yaml
```

客户端使用同一 KCP 环境输入，命令与消息见 [协议参考](../interfaces/protocols.md)。WS 验证广播、Broker 验证关联回复、字节传输验证回显。中间件只绑定本机回环地址，匿名访问/默认账户用于本地示例。退出后仅清理属于本次示例的资源：

```sh
docker compose --profile examples down --volumes
```

自动验收优先 make smoke / make examples-smoke；后者创建专属 Compose 项目并清理其容器、网络和卷。分别确认进程、健康/就绪、业务与启用协议，不用 /readyz 代替业务验收。

## 排障与恢复

先确认目标、版本、配置来源、启用能力与授权范围，优先只读取证，不打印密钥、连接串或完整载荷。

| 现象 | 检查重点 |
| --- | --- |
| 启动失败 | -conf、环境覆盖格式、至少一种传输、地址/端口与依赖初始化 |
| /healthz 失败 | 进程、HTTP 开关、监听及入口连通 |
| /readyz 返回 503 | checks 中 down 的 PostgreSQL/Redis/MongoDB，其他能力另查 |
| 问候查询 404 | ID、实例、是否重启与内存/远程模式 |
| Broker 无回复 | 先订阅、关联 ID、请求/回复目标、连接与消费/发布失败 |
| WS/TCP/KCP 空闲断开 | 读写超时、同源、消息大小与协议输入 |
| /metrics 返回 503 | exporter 初始化及启动日志 |

详细响应见 [接口契约](../interfaces/README.md)。交接记录时间、版本、目标、实际命令、首次错误、可公开请求标识和期望/实际行为，不提供完整敏感配置。日志、指标、追踪存在不代表现场告警和责任人已配置。

本地内存重启丢失，没有业务快照或迁移体系；数据库客户端不改变此边界，远程模式恢复由目标服务承担。新增持久化设计一致性备份、结构迁移、回退触发与验收；运行中目录复制不能视为一致性备份，旧二进制不能自动保证数据兼容。

重启或回退后重新检查健康、就绪、选定业务和相关协议，按实际通过范围记录。源码、测试、构建、部署与恢复证据分开，失败和未验证范围写明；操作规则与权限见 [开发指南](../development/README.md)。
