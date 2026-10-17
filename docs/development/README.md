# 开发与验证

- 适用范围：greeting
- 维护责任：开发与工具维护者
- 核对日期：2026-10-18（依赖与安全验证入口；现场环境另行验收）

## 开发上手

需要 Go 1.27.2 或更高版本、make、Python 3.9 及以上；先 make doctor。已有生成物支持普通构建，proto/Wire 变化或生成一致性检查需要 protoc 和 make init 安装的固定工具；工具在忽略的 .tools/bin，安装需要网络。

默认只启用 HTTP/gRPC，不要求数据库或 Broker。最小运行与另一终端验证：

```sh
make doctor
make build
./bin/server -conf ./configs/config.yaml
# 另一终端
./bin/client -protocol http
./bin/client -protocol grpc
curl -fsS http://127.0.0.1:8000/healthz
```

客户端创建后按返回 ID 查询同一问候，输出包含 Hello World 与保存/查询 ID。Ctrl+C 或 SIGTERM 停止；make run 可直接运行源码。健康接口仅证明 HTTP 存活。

## 本地配置

默认文件是 configs/config.yaml，容器使用 configs/container.yaml，全协议使用 configs/examples.yaml；选择一个明确文件，按任务启用能力。字段、默认行为、加载顺序与环境覆盖由 [配置参考](../interfaces/configuration.md) 唯一维护。

只准备启用能力所需的依赖与输入，不假定任意 APP_* 自动生效。KCP 客户端/服务端共享临时受控 key/salt，不保存真实值；敏感 YAML 不进入日志或计划。验证实际监听、健康/就绪与业务链路，配置保存成功不代表生效。启动阻塞见 [排障](../operations/README.md#排障与恢复)。

## 常用命令

| 命令 | 用途 |
| --- | --- |
| make doctor | 检查基础工具与按需生成/烟测工具 |
| make task NAME=orders | 创建并索引 docs/plans/orders.md，拒绝覆盖 |
| make init | 安装固定版本生成器 |
| make all | 顺序生成 API、运行配置提示与 Wire |
| make fmt | 格式化手写 Go |
| make check | 文档、格式、vet、race、生成一致性和构建 |
| make check-generated | 在临时副本生成比较，不修改当前源码 |
| make check-docs | 文档结构/链接、隔离回归与任务工具验证 |
| make init-security | 安装固定版本的官方 govulncheck 到 .tools/bin |
| make vuln-check | 扫描含测试的调用路径；有漏洞记录时返回非零，需结合实际防护审阅 |
| make smoke | 临时端口验证六种无中间件协议和退出 |
| make examples-smoke | 增加 MQTT/RabbitMQ，创建并清理专属 Docker 资源 |
| make scaffold-smoke | 临时副本真实生成普通/CRUD 模块、接线、Wire、编译与生成一致性 |
| make e2e-smoke | HTTP 健康、就绪、元信息与业务入口 |

make verify 等价于 make check；make generated-check 等价于 make check-generated。构建与制品命令见 [交付指南](../operations/README.md#构建与交付)。

生成检查使用当前工作区，不要求已提交；不复制 .git、bin、本地工具或 configs/config.yaml，不启动应用。新 proto 先 make all；生成器版本变化重新 make init 并审查差异。配置使用普通 Go 结构体，不生成配置 protobuf。

## 业务改动地图

先明确契约、规则、存储、错误与验收。新业务优先 make scaffold-module PROTO=order/v1/order.proto，CRUD=1 增加 Create/Get/List/Update/Delete、标准字段、业务模型和 repo/usecase 桩。scaffold-proto 生成 proto、service/biz/data 与接线、remote_grpc 配置入口；scaffold-module 再生成 Wire。普通未标注方法沿 Kratos RPC 路由，CRUD 使用显式 HTTP 映射；桩仍需真实实现与测试。

手工实现沿 HTTP/gRPC → service → biz/仓储接口 → data：

| 改动 | 接线与核对位置 |
| --- | --- |
| 仓储 | internal/data/data.go ProviderSet、NewData、连接 cleanup 与入口 |
| usecase | internal/biz/biz.go ProviderSet |
| service | internal/service/service.go ProviderSet |
| HTTP/gRPC | internal/server/http.go、grpc.go 构造与 Register；cmd/server/main.go |
| Wire | cmd/server/wire.go；生成后审查 wire_gen.go |
| 配置 | internal/conf/config.go、所属模块结构、cmd/server/env.go、校验/安全示例 |
| 契约 | `api/<domain>/v1/*.proto`、生成物、OpenAPI、调用示例与真实入口测试 |

复杂跨层任务在 [计划](../plans/README.md) 写责任模块、源码证据、入口到副作用链路、受影响状态与验收，中断前更新进度和下一步。简单修复无需任务单。

## 开发约定

- proto 是契约源，新业务有自己的 domain，字段号不复用。不手改 `*.pb.go`、`*.pb.validate.go`、`*_http.pb.go`、`*_grpc.pb.go`、`wire_gen.go` 或 `openapi.yaml`；proto/构造依赖变化先 make all，生成物随源审查。
- protobuf 类型留在 service；biz 拥有模型、业务规则和仓储接口；data 做具体存储/外部适配，遵守 [架构边界](../architecture/README.md#分层与接线)。不用自动发现、反射注册或隐藏副作用框架。
- 新可选基础设施禁用时沿现有构造契约返回 nil、安全 cleanup 和无错误，禁用传输返回 nil；启用能力明确超时、取消和资源关闭。
- 参数校验覆盖 proto 与直接调用 usecase；保留 HTTP int64 字符串、稳定 reason 及兼容例外，具体错误见 [契约](../interfaces/README.md#问候接口)。
- internal/conf/config.go 聚合配置，internal/server/config.go/适配器拥有传输项，internal/data/config.go 聚合基础设施，环境覆盖显式定义于 cmd/server/env.go；新增字段同步消费、校验、示例、覆盖和测试。
- 优先既有 Kratos/GORM/标准工具。单体先按 domain 扩展，持久化与拆分遵守 [存储演进](../architecture/README.md#存储与演进)，不因配置或客户端存在声称保障已完成。
- 不记录秘密或完整敏感载荷；取消沿 context 传递，关闭资源，保留无关工作区改动。

## 验证选择

| 改动 | 必需验证 |
| --- | --- |
| 纯 Markdown、目录或导航 | make check-docs、git diff --check；核对正文归属与事实 |
| 文档/任务工具 | make check-docs 的隔离回归及实际错误/防覆盖行为；影响生成/构建时增加相关检查 |
| Go 业务、契约、配置或接线 | 受影响包 → make fmt → make check；proto/构造变化先 make all |
| 启动、协议或生命周期 | 上述检查及 make smoke |
| MQTT/RabbitMQ | 相应检查及 make examples-smoke，需要 Docker |
| 模块脚手架或注册逻辑 | 相应检查及 make scaffold-smoke |
| Docker/Kubernetes | 构建或清单渲染、目标环境验收按 [交付范围](../operations/README.md) 明确 |

规则测试有意义地覆盖成功、失败、边界；新入口证明请求到仓储，不能只测常量 stub。仓储验证并发、取消、远程一致性与错误，持久化使用真实存储集成测试。外部 smoke 仅在 RUN_SMOKE=1 且依赖就绪时执行，默认跳过如实记录；纯业务不拉起全部中间件。

文档检查不补写文件、不联网、不读取运行配置或生成缓存，文档工具与 [任务工具](../../scripts/dev.py) 使用临时隔离样例。入口包括 [检查器](check_docs.py)、[检查回归](test_check_docs.py)、[任务回归](../../scripts/test_dev.py)，具体规则见 [文档规范](../documentation-standard.md)。

完成时记录实际命令、结果、失败原因和未验证影响，不以单元测试或构建代替部署/恢复。计划回写正式文档；提交、推送、部署和发送消息仍需用户授权。

## 依赖安全检查

依赖变化运行 go mod tidy、完整检查、相应协议烟测和 make vuln-check。需要模块级对照时执行 .tools/bin/govulncheck -C cmd/server -scan=module；模块记录、导入包、实际调用与 GitHub 告警是不同口径。JSON 格式即使有发现也可能退出成功，不能只看退出码。

同时核对 GitHub/OSV 的已审阅告警，避免数据库收录时间不同造成漏项；例如 RabbitMQ 的 [协商前帧限制漏洞](https://github.com/advisories/GHSA-w6r9-248c-frg8) 要求 1.14.0，1.13.0 的早期修复并不完整。本项目保留安全的伪连接回归验证提前分配受控小内存的问题，并以本地 Broker 验证正常请求/回复。

当前 Kratos v2.9.2 的 [GO-2026-5471](https://pkg.go.dev/vuln/GO-2026-5471) 尚无上游修复版本。本项目通过显式 404/405 兜底处理器阻止全局 ServeMux 暴露，并以真实 HTTP 回归证明；扫描仍会报告其版本/调用记录，不标为扫描清零、不关闭检测，也不自动忽略。详见 [修复记录](../plans/dependency-security.md) 与 [架构边界](../architecture/README.md#访问与资源边界)。
