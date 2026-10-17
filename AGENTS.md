# Agent 开发入口

本仓库是 greeting Go / Kratos 示例模板，仓库根是 Go module 根。人和 Agent 使用同一套命令，应用不依赖 AI 平台。文档按六类通用责任组织，分类 README 兼作正文，入口见 [docs/README.md](docs/README.md)。

## 操作边界

- 修改前检查 git status --short，保留无关改动；实施前 make doctor。
- 只调查或提方案时不修改、生成或构建。提交、推送、部署和发送消息遵循用户明确授权。
- 不读取或输出无关密钥、令牌和完整敏感配置，日志与计划不放秘密。
- 明确事实、假设和未验证项，未执行或失败的检查不能记为通过。
- 按任务实际读取下列相关章节，规范移入 docs 后仍是适用任务的验收要求；必读内容不可读或冲突时先报告，用户指令优先。

## 按任务阅读

| 任务 | 必读入口 |
| --- | --- |
| 目标、示例业务与验收 | [需求](docs/requirements/README.md) |
| 修改业务、存储、配置、接线或生命周期 | [架构](docs/architecture/README.md)、[开发](docs/development/README.md) 的相关章节 |
| HTTP/gRPC、消息、CLI 或配置输入 | [接口](docs/interfaces/README.md)，按需读 [协议](docs/interfaces/protocols.md)、[配置参考](docs/interfaces/configuration.md) |
| 构建、部署、Kubernetes、日志与恢复 | [交付与运维](docs/operations/README.md) |
| 复杂跨层任务与交接 | [计划](docs/plans/README.md) 与明确选定的计划；简单修复不建任务单 |
| 文档、目录或工具维护 | [文档规范](docs/documentation-standard.md)、所属分类 README |

## 验证与维护

纯文档执行 make check-docs、git diff --check。Go 变化先相关包测试，再 make fmt、make check；proto/构造变化先 make all。启动/协议、Broker 或模块脚手架变化按 [验证选择](docs/development/README.md#验证选择) 增加对应 smoke。

生成、分层、完整注册、内存/远程语义、兼容与资源要求在对应正文维护，不能因合并而省略。只同步受影响正文与入口，复杂任务完成时回写长期规则；重要技术选择按 [决策规则](docs/architecture/decisions/README.md) 记录，普通文档合并不另建决定。

交付说明实际变化、验证结果与剩余影响。贡献入口见 [CONTRIBUTING.md](CONTRIBUTING.md)。
