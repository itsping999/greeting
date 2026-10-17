# 0001：采用通用软件项目文档结构

- 类型：技术决策记录
- 状态：Accepted
- 适用范围：greeting 文档组织与开发工具
- 维护责任：项目文档与工具维护者
- 决定与核对日期：2026-10-13

## 背景与选择

用户要求按 s-series 当前 docs 结果同步本项目并修改 AGENTS。源项目已采用 requirements、architecture、development、interfaces、operations、plans 六类通用责任目录，并将允许项与必需文件分开登记。

原 greeting 将开发指南放在 docs 根、任务放在 tasks，AGENTS 同时维护详细架构、开发与交付规则。保留旧布局不符合通用归属；直接复制源项目正文又会引入与 greeting 不符的业务、数据和部署事实。

## 采用决定与后果

采用源项目的通用分类、规范约束和版本 2 检查工具，正文依据 greeting 自身契约与实现维护。开发指南首次迁入 development/workflow.md，任务与模板迁入 plans；make task 保持原命令，创建新计划并登记目录索引，拒绝覆盖。

AGENTS 保留项目身份、全局操作边界和按任务必读路由；长期规范迁入对应责任正文。make check-docs 接入本项目 Make 与已有 GitHub Actions，实际远端执行仍需要其运行证据。

六类名称可以复用于前后端、CLI、SDK、设备等软件项目，按内容创建所需类别；二级目录按规范登记，当前只使用决策目录。既有 Go 服务、proto、Wire、配置和业务行为保持原范围；本决定不授权提交、推送或部署。

长期规则见 [文档规范](../../documentation-standard.md)，实际实施和验收见 [同步计划](../../plans/documentation-sync.md)。

## 轻量化补充

2026-10-13，用户进一步确认采用合并后的轻量方案。六类名称与机器规则继续使用，分类 README 兼作正文；首次独立的短说明合并到对应 README，协议/配置与既有记录保留。纯文档验证按文档范围执行。此为原分类的规模裁剪，不为普通合并新建决策；当前入口和验收以规范及实施记录为准。
