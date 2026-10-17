# 架构、运行与数据边界

- 适用范围：greeting Go / Kratos 模板
- 维护责任：项目与对应模块维护者
- 核对日期：2026-10-18（HTTP 安全边界；不代表目标环境已部署）

## 分层与接线

HTTP/gRPC → server 构造与注册 → service 协议映射 → biz/usecase → biz 仓储接口 → data 实现。消息协议经既有 server/service 适配进入业务，业务规则不隐藏在连接回调或数据访问层。

| 位置 | 责任 |
| --- | --- |
| `api/<domain>/v1/` | proto 契约源、HTTP/gRPC 映射与 protobuf 字段编号 |
| internal/server/ | 传输配置、协议生命周期、注册和中间件 |
| internal/service/ | 协议输入输出、protobuf/消息与业务模型映射 |
| internal/biz/ | 业务模型、规则、usecase、仓储及所需外部接口 |
| internal/data/ | 存储与外部 API/事件适配，明确错误、并发、取消和 cleanup |
| internal/conf/ | 聚合各模块拥有的配置 Go 结构体 |
| cmd/server/ | 文件/环境加载、Wire 根、应用启停 |
| internal/pkg/ | 可选数据库、发现、远程客户端、日志、指标与追踪能力 |

参考 [proto](../../api/helloworld/v1/greeter.proto)、[service](../../internal/service/greeter.go)、[usecase](../../internal/biz/greeter.go)、[repository](../../internal/data/greeter.go) 和 [HTTP 注册](../../internal/server/http.go) 定位纵向链路，按任务读取相关层和测试。

依赖使用普通构造函数与 Wire。新增业务接通 service/biz/data ProviderSet、HTTP/gRPC 构造参数与注册、cmd/server/wire.go；仅增加未注册 service 不算完成。禁止反射自动注册、隐藏副作用的插件框架或平行注入入口。

单体先按 domain 和进程内调用组织，拆服务沿 proto/usecase/repo 边界替换外部适配；不提前引入注册中心、消息队列或新框架。新业务默认 HTTP/gRPC，需要其他协议时明确关联 ID、重试、幂等和失败语义。实施地图见 [开发指南](../development/README.md#业务改动地图)。

## 运行与生命周期

[服务入口](../../cmd/server/main.go) 读取 -conf 文件，应用显式环境覆盖，校验启用协议，再初始化日志、指标与追踪，并经 [Wire 根](../../cmd/server/wire.go) 装配。Kratos App 收集非 nil 的八类 Server；关闭的传输不要求其地址、凭据或外部服务，至少启用一种协议。

Ctrl+C 或 SIGTERM 触发停止，关闭监听和已建立的 WS/TCP/KCP 等连接，执行构造返回的 cleanup 与指标/追踪清理。取消和超时沿 context 传递，不创建绕过 App 的独立常驻流程。配置、依赖初始化与启动错误如实上报，退出和协议超时按 [验证选择](../development/README.md#验证选择) 证明。

## 存储与演进

当前 [Data](../../internal/data/data.go) 用内存 map、递增 ID 和锁存储问候；保存/查询返回副本并遵守 context 取消，重启清空。没有业务实体表、生产持久化或问候迁移入口。

启用 data.remote_grpc.greeter 时，Save 调用远程 SayHello，FindByID 调用远程 GetGreeting，两种操作保持同一契约。远程错误、持久化与身份能力由目标服务决定。PostgreSQL/Redis/MongoDB 客户端初始化或 /readyz 通过不代表问候使用数据库。

创建事件在主保存成功后可选发布，失败记日志，不回滚保存，不提供跨服务事务。原型可用独立内存 repo；需要持久化时实现同一 biz 接口，明确实体、索引、显式迁移、事务、并发、幂等、删除与恢复语义，增加配置、初始化、超时、cleanup 和真实存储集成测试。事务边界由 usecase 明确，数据库访问由 data 实现；仅增加配置或客户端不算接入完成。

## 访问与资源边界

模板未预置业务账号、角色或资源授权。实际业务确认认证、授权、资源隔离与外部入口；健康或就绪结果不能替代权限保障。biz 同时校验直接调用 usecase 的输入，协议约束见 [接口目录](../interfaces/README.md)。

HTTP 入口显式设置未注册路径的 404 和方法不匹配的 405，均不回落 net/http.DefaultServeMux；即使其他包注册了诊断处理器也不会自动暴露。此为 Kratos 上游尚未发版漏洞的应用侧防护，不能当成依赖版本告警已消除。追踪初始化只记录 exporter 类型，不记录可能带凭据的 endpoint。

不读取或记录无关凭据、完整敏感配置和载荷，配置覆盖错误不回显原值。本机 configs/config.yaml 不进入生成副本或镜像构建上下文。WS 同源/大小限制、KCP 输入及本地 Broker 默认设置是示例边界，部署时确认监听、代理/TLS、凭据渠道和实际权限。

数据升级、重启和回退步骤见 [交付与运维](../operations/README.md)。已有重要选择保留在 [决策记录](decisions/README.md)；普通修复和文档合并无需逐项新建决定。
