# 从需求到可验证业务

开发入口是根目录 `AGENTS.md`。目标是让人与 Agent 都能从明确契约出发，在同一个反馈回路里实现、生成、验证和交付。应用运行不依赖 AI 服务。

## 常用命令

| 命令 | 作用 |
|---|---|
| `make doctor` | 报告 Go、make、Python3、生成工具和可选烟测工具；缺少基础工具时失败 |
| `make task NAME=orders` | 从任务模板创建 `docs/tasks/orders.md`，拒绝覆盖已有任务 |
| `make init` | 安装固定版本生成器到 `.tools/bin`；需要网络 |
| `make all` | 顺序生成 API 和 Wire（运行配置是 Go 结构体）；修改生成文件所在工作区 |
| `make fmt` | 格式化手写 Go 文件 |
| `make check` | 格式检查、vet、race、生成一致性、构建；失败立即停止 |
| `make check-generated` | 在临时副本生成并比较，不修改当前源码；首次需要生成器 |
| `make smoke` | 动态端口下验证六种无中间件协议和退出行为 |
| `make examples-smoke` | 增加 MQTT/RabbitMQ；创建并清理本次专属 Docker 资源 |

`make check-generated` 使用当前工作区内容，不依赖工作区干净或已提交。不会复制 `.git`、构建产物、本地工具和 `configs/config.yaml`，也不会启动服务。新 proto 应先 `make all`，否则缺失的生成物会被报告。生成工具版本由 Makefile 固定；改变版本时重新 `make init` 并审核生成差异。

## 业务改动地图

以新增订单为例，可先运行 `make scaffold-module PROTO=order/v1/order.proto CRUD=1` 创建分层桩和接线。普通脚手架未标注 HTTP 映射的方法使用 Kratos 默认 RPC 路由；CRUD 脚手架生成显式 HTTP 映射。业务落地时须在 proto 明确接口契约，再实现业务及真实存储，并补齐验收测试。手工创建时先写 `api/order/v1/order.proto`，明确 CreateOrder / GetOrder、HTTP 路由、校验和错误语义，再实现下列链路：

```text
HTTP / gRPC
  → internal/service/order.go（DTO 映射）
  → internal/biz/order.go（规则 + OrderRepo 接口）
  → internal/data/order.go（存储实现）
```

| 改动 | 必须检查的接线位置 |
|---|---|
| 新仓储 | `internal/data/data.go` 的 ProviderSet；NewData 是否需扩展；新增外部连接时增加 cleanup，并接回 Wire / main |
| 新用例 | `internal/biz/biz.go` 的 ProviderSet |
| 新 service | `internal/service/service.go` 的 ProviderSet |
| HTTP/gRPC 暴露 | `http.go` / `grpc.go` 的构造参数和 Register 方法；`cmd/server/main.go` 的传输组装 |
| 注入 | `cmd/server/wire.go`；`make all` 后审查 `wire_gen.go` |
| 配置 | `internal/conf/config.go`、所属模块的配置结构体、`cmd/server/env.go`、校验和不含秘密的配置示例 |
| 契约 | `make all` 更新 API 生成物与 OpenAPI；补调用示例和测试 |

默认继续使用普通 Go 构造函数和显式注册，方便搜索与审查。不要实现自动发现 service、反射注册或隐藏副作用的插件框架。

可参考 `greeter.proto`、`biz/greeter.go`、`data/greeter.go`、`service/greeter.go` 的保存→查询链路，以及 `internal/server/server_test.go` 的真实 HTTP 验证。参考现有错误风格：参数非法为 400、缺失资源为 404，reason 稳定可供调用方判断；HTTP 的 protobuf int64 是字符串。输入校验须兼顾 proto 边界与绕过协议直接调用 usecase 的场景。

## 持久化与拆分

开发原型可使用独立的内存 repo，明确重启丢失。业务需要持久化时实现同一仓储接口、显式迁移与事务边界、cleanup 和超时；用真实存储集成测试验证。仅增加配置字段不代表完成数据库接入。

单体按 domain 组织文件、用构造函数组装，在 biz 中定义所需的依赖接口。未来拆分时以 domain 为边界单独启动服务，将跨域接口替换为客户端适配器。跨服务一致性、幂等、鉴权和持久化需按具体业务设计；模板不预置这些保证。

## 验收与上下文管理

开始多文件功能时生成任务单，写最少必要的事实：目标、范围、接口、规则、失败行为、验收场景。中断前更新“进度与证据”，记录下一步，使其他 Agent 可以继续，不需要猜测聊天历史。任务单留在仓库随代码审查，不存凭据或生产载荷。

验证由近到远：受影响包测试 → `make check` → 与本次变化相关的 smoke / 集成验证。测试证明规则或链路，不复制实现细节。交付记录真实命令结果；环境不足时写未验证项及影响。模板的内存示例和本地 smoke 不能替代生产环境验收。

脚手架或注册逻辑变化时运行 `make scaffold-smoke`：在临时副本真实生成普通与 CRUD 模块，验证注册、Wire、编译和生成一致性，不修改当前源码。需要 `make init` 安装的工具。
