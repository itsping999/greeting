# 问候、健康与接口参考

- 适用范围：greeting HTTP/gRPC 与运行输入
- 维护责任：接口、协议与配置维护者
- 核对日期：2026-10-13（当前实现）

## 问候接口

先创建问候，再使用返回 ID 查询同一记录。默认本地数据重启丢失，远程模式的存储能力由目标服务决定。模板未预置用户认证，部署入口的鉴权按实际业务明确。

| 操作 | HTTP | gRPC |
| --- | --- | --- |
| 创建问候 | GET /helloworld/{name} | helloworld.v1.Greeter/SayHello |
| 查询记录 | GET /greetings/{id} | helloworld.v1.Greeter/GetGreeting |

| 字段 | 位置 | 类型 | 必填与约束 |
| --- | --- | --- | --- |
| name | 创建路径 / gRPC 请求 | string | 是；去除首尾空白后 1–256 个字符，路径 URL 编码 |
| id | 查询路径 / gRPC 请求 | int64 | 是；正数，HTTP JSON 中 int64 使用字符串 |

成功返回 HTTP 200 / gRPC OK，包含 message 与 id：

```json
{"message":"Hello World","id":"1"}
```

message 是问候文本，id 是记录标识。GET 创建有保存副作用，是兼容示例；不能按纯读取缓存，网络失败重试可能产生新记录，不承诺幂等。新增资源使用 POST。

| 失败条件 | HTTP | 原因与调用方处理 |
| --- | --- | --- |
| 规范化名称为空或超过 256 字符 | 400 | INVALID_NAME；修正名称 |
| ID 非正数或不能绑定为整数 | 400 | 参数错误；修正 ID，绑定错误按实际响应处理 |
| 记录不存在 | 404 | GREETING_NOT_FOUND；确认实例、ID 和是否重启 |

gRPC 参数错误对应 InvalidArgument，缺失对应 NotFound；远程/基础设施失败按实际结果处理，不假定保存已成功。

```sh
curl -fsS http://127.0.0.1:8000/helloworld/World
curl -fsS http://127.0.0.1:8000/greetings/1
```

查询使用实际创建返回的 ID。错误示例：不存在时 GET /greetings/999 返回 404、GREETING_NOT_FOUND；GET /helloworld/%20 返回 400、INVALID_NAME。

## 健康与观测

| 方法与路径 | 成功 | 失败/边界 |
| --- | --- | --- |
| GET /healthz | HTTP 200，{"status":"ok"} | 只说明 HTTP 存活 |
| GET /readyz | HTTP 200，status=ready、checks 对象 | 启用的数据库客户端不可用时 503、status=not_ready |
| GET /meta | HTTP 200，status=running、started_at | started_at 是本次启动 UTC RFC3339 时间，不是版本 |
| GET /metrics | Prometheus 格式指标 | exporter 未初始化时 503 |

checks 含 postgres、redis、mongodb，值为 up/down/skipped；禁用客户端 skipped，快照缓存约 2 秒。不验证问候持久化、Broker、远程 Greeter 或全部业务。探针分别验证存活与就绪，业务另行验收，暴露面由部署入口明确。

## 独立参考

| 文档 | 使用边界 |
| --- | --- |
| [可选协议与 CLI](protocols.md) | WS、MQTT/RabbitMQ、TCP/UDP/KCP、示例客户端 |
| [配置输入](configuration.md) | 配置文件、字段、默认行为与显式环境覆盖 |

proto 源仍在仓库根 api/；interfaces/ 是文档分类。新增契约写清调用时机、认证、参数位置/类型/必填/约束、成功/错误与兼容。实现边界见 [架构](../architecture/README.md)，操作步骤见 [开发](../development/README.md) 和 [运维](../operations/README.md)，维护要求见 [文档规范](../documentation-standard.md)。

安全维护补充（2026-10-18）：未注册 HTTP 路径返回 404，已注册路径的方法不匹配返回 405，均不会转到全局 HTTP 处理器。已有问候、健康、配置及消息契约保持各自定义。
