# 通用文档结构与轻量维护

- 适用范围：greeting；分类可复用于其他软件项目
- 维护责任：项目与对应主题维护者
- 核对日期：2026-10-13

围绕读者任务维护当前事实和规则，按实际内容创建文档。示例默认一类一页，内容明显增长、受众或维护周期独立时再拆分；不用文件数量衡量完成。

## 固定目录与命名规范

| 目录 | 内容责任 | 本项目入口 |
| --- | --- | --- |
| requirements/ | 目标、范围、业务规则与验收 | [需求](requirements/README.md) |
| architecture/ | 分层、运行、数据、安全与设计 | [架构](architecture/README.md) |
| development/ | 上手、本地配置、实施与验证 | [开发](development/README.md) |
| interfaces/ | API、协议、CLI/SDK、配置与文件格式契约 | [接口](interfaces/README.md) |
| operations/ | 制品、部署、排障与恢复 | [运维](operations/README.md) |
| plans/ | 复杂任务的范围、进度与实际结果 | [计划](plans/README.md) |

六类名称固定，按内容创建，不预建空目录；其他项目沿用责任并更新本项目链接。根目录只放总导航与本规范，每个已创建目录有 README。**分类 README 可以同时承担正文和导航**，已经回答的问题不另建同义页面。

默认平铺；稳定主题需要二级目录时说明理由、边界及引用影响并登记机器规则，最多两层。当前保留已有 architecture/decisions/，重要选择才记录决定，普通修复或文档合并无需另建决策。

新增、合并或迁移同步父/分类索引、受影响引用和 AGENTS 阅读入口；不给前后端、平台、业务模块或历史状态各建一套分类。主题文件优先 kebab-case，版本/替代关系明确，不用临时、最新或副本命名。附件跟随正文，不能存日志、数据库、备份、制品或未登记工具。

## 唯一正文与适用范围

每项规则或契约维护一个责任版本，其他资料概述并引用。配置字段、默认值和输入约束由 interfaces 维护，本地操作由 development、交付操作由 operations 引用。新业务需求留 requirements，设计留 architecture，计划只留任务范围、步骤和结果。

正文标明适用范围、维护角色与核对日期；类型或读者不清晰时补充。计划/决策声明状态，草案和历史记录不自动生效，核对日期不代表现场部署通过。示例不放真实凭据、令牌或完整敏感配置，接口页面向调用方，不夹带内部实现推理。

## 强制约束与完成门槛

- 目录、层级、根文件和工具符合机器规则，不以隐藏目录、符号链接或修改忽略项绕过。
- 每份独立正文从所属 README 直接可达，子目录从父 README 可达；本地目标与标题锚点有效。合并正文直接位于 README 时无需第二份入口页。
- 不保留重复正文，替代资料标明状态与去向；类型、事实、业务归属、敏感信息及语义重复人工核对。
- 纯 Markdown、目录与导航改动完成前运行 make check-docs、git diff --check。工具变化执行相关回归；Go、proto、接线与构建变化按 [验证选择](development/README.md#验证选择) 增加检查。
- 复杂计划完成前按 [正式文档回写](plans/README.md#完成与正式文档回写) 更新长期内容，登记实际结果、失败与未验证影响。检查失败不能报告完成，不用旧结果代替本轮验证。

目录归属和允许项变化先在本规范或已有实施记录写明理由与影响，再同步机器规则及导航；重要技术选择另留决定。不得只放宽脚本名单、删检查或复制源码项目的事实来通过。

## 唯一机器规则

版本 2 的 allowed_* 声明允许出现的内容，required_files 声明必须存在的文件；已创建目录的 README 总是必需。轻量合并不改变以下允许项与检查逻辑。

<!-- docs-check-policy:start -->
```json
{
  "version": 2,
  "allowed_directories": [
    "requirements",
    "architecture",
    "development",
    "interfaces",
    "operations",
    "plans"
  ],
  "allowed_subdirectories": [
    "architecture/decisions"
  ],
  "allowed_root_files": [
    "README.md",
    "documentation-standard.md"
  ],
  "allowed_tool_files": [
    "development/check_docs.py",
    "development/test_check_docs.py"
  ],
  "required_files": [
    "README.md",
    "documentation-standard.md"
  ]
}
```
<!-- docs-check-policy:end -->

[检查器](development/check_docs.py) 与 [隔离回归](development/test_check_docs.py) 仅依赖 Python 3.9+ 标准库。make check-docs 同时验证任务工具，只读检查不联网、不读取配置、不补写正文或生成缓存；回归使用系统临时目录。结构变化同步此处规则，失败先修正文档，不临时跳过。

本地 Make 与 [CI](../.github/workflows/ci.yml) 使用同一文档检查入口，远端运行结果另行验证。实现、测试、构建、部署、恢复和性能证据各自记录；不因客户端初始化、单元测试或示例运行声明生产保障。
