# Agent 开发入口

本仓库是 Go / Kratos 业务模板。先读本文，再读 `docs/development.md`；按需读取相关层和测试，不需要遍历整个仓库。所有 Agent 使用同一套命令，不依赖特定模型、编辑器或外部 AI 服务。

## 先定位，再修改

1. 查看 `git status --short`，保留已有改动；不要清理、覆盖、提交或推送不属于本次任务的内容。
2. 运行 `make doctor` 检查本机条件。普通构建无需生成工具；修改 proto / Wire 后需要 `protoc` 和 `make init`。
3. 把需求转成接口、业务规则、存储语义、错误和验收场景。多文件业务用 `make task NAME=orders` 创建任务单，随实现更新。简单修复不必创建。
4. 遵循 `docs/development.md` 的业务改动地图；不确定的业务决策明确写出，不能把猜测当作验收结果。

## 边界与约定

- `api/<domain>/v1/*.proto` 是 HTTP/gRPC 契约源；新增业务使用自己的 domain。不能手改 `*.pb.go`、`*.pb.validate.go`、`*_http.pb.go`、`*_grpc.pb.go`、`wire_gen.go` 或 `openapi.yaml`。
- `internal/service` 负责协议输入输出；`internal/biz` 负责业务规则及仓储接口；`internal/data` 实现仓储。业务规则不能藏在协议回调或数据层。
- 依赖显式构造并通过 Wire 注入。三个层的 `ProviderSet`、`NewServers` 的依赖和 HTTP/gRPC 注册都必须接通。禁止仅增加未注册的 service。
- 单体先用业务模块和进程内调用；拆服务时沿相同 proto / usecase / repo 边界，不提前引入注册中心、消息队列或新的框架。
- 默认只有 HTTP/gRPC；新增业务不需要同时扩展所有示例协议。需要消息协议时保持关联 ID、明确重试/幂等和失败语义。
- 当前仓储是内存示例，重启丢失。不能声称已有生产持久化或事务保障。
- 现有 `GET /helloworld/{name}` 有保存副作用，是兼容示例。新建资源用 POST；不要复制这个历史行为。保留已有接口兼容性，protobuf 字段号不复用。
- 不读取或输出不必要的本地密钥、配置值、令牌；示例和任务单不放秘密。取消与超时沿 context 传递，关闭资源，不记录完整敏感载荷。
- 用户只要求调查或方案时不执行修改、生成或验证构建；任务授权范围优先。部署、发送消息、提交和推送需符合用户明确指令。

## 交付门槛

- 修改 proto 或构造依赖后运行 `make all`；它按 API → 配置 → Wire 顺序生成。
- 运行 `make fmt`，再运行 `make check`：格式、vet、race 测试、生成一致性、二进制构建。
- 业务规则增加有意义的成功、失败、边界测试；新增入口至少证明路由注册和请求到仓储的真实链路。不要只测 stub 能返回常量。
- 涉及启动、协议、生命周期时运行 `make smoke`；涉及 MQTT/RabbitMQ 时运行 `make examples-smoke`（需要 Docker）。纯业务改动无需拉起全部中间件。
- 未完成的验证明确记录命令、失败原因和影响；不允许把未运行的检查写成通过。
- 最终说明实际行为变化、验证结果和剩余限制。维护契约、README 与任务单；不要仅交付实现文件。
