# third_party

仅保留当前 proto 的传递依赖：Google HTTP annotations、http、protobuf descriptor / duration / timestamp，以及 protoc-gen-validate 的规则定义。各文件保留上游许可证头。

新增 API 引入额外 proto 时，将对应上游文件及其传递依赖放到同名路径；不要为尚未使用的能力预置整套定义。生成入口见根目录 Makefile。
