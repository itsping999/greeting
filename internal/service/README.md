# Service

将 protobuf 请求和协议消息映射到问候业务。HTTP/gRPC 使用生成的接口；MQTT/RabbitMQ 返回带 request_id 的问候回复；WebSocket 返回标准 JSON 广播消息。传输连接及资源管理由 `internal/server` 负责。
