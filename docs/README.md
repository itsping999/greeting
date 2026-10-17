# greeting 文档入口

六类目录可复用于其他软件项目，按实质内容创建。示例项目的分类 README 同时承担正文和导航；较复杂的协议/配置才独立成页，规则见 [文档规范](documentation-standard.md)。

| 要完成的任务 | 阅读入口 |
| --- | --- |
| 了解目标和问候规则 | [需求](requirements/README.md) |
| 理解分层、运行、数据或安全 | [架构](architecture/README.md) |
| 上手、扩展业务、生成与测试 | [开发](development/README.md) |
| 调用问候、健康与观测接口 | [接口](interfaces/README.md) |
| 接入可选协议或配置输入 | [协议与 CLI](interfaces/protocols.md)、[配置参考](interfaces/configuration.md) |
| 构建、部署、排障和恢复 | [交付与运维](operations/README.md) |
| 复杂任务与交接 | [计划](plans/README.md) |

[AGENTS.md](../AGENTS.md) 保留操作边界与按任务必读入口，实际执行前读取相关章节。原有重要决定保留在 [决策索引](architecture/decisions/README.md)，迁移与轻量合并的实际验证见 [实施记录](plans/documentation-sync.md)。

纯文档修改运行 make check-docs 与 git diff --check；代码与工具变化按 [验证选择](development/README.md#验证选择) 执行。当前内存、远程转发和可选基础设施的边界由架构正文维护，测试与构建结果不代替目标环境交付。
