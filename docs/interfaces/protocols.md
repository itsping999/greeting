# 可选协议与示例客户端

- 类型：协议与 CLI 参考
- 适用范围：greeting 可选传输示例
- 维护责任：协议与客户端维护者
- 核对日期：2026-10-13

默认只启用 HTTP/gRPC，其他协议按配置开启。configs/examples.yaml 是全协议示例。示例展示问候或回显，不预置生产持久化、幂等或跨服务事务。

## 地址与 CLI

| 协议 | 示例地址 | 客户端 |
| --- | --- | --- |
| HTTP | 127.0.0.1:8000 | ./bin/client -protocol http |
| gRPC | 127.0.0.1:9000 | ./bin/client -protocol grpc |
| WebSocket | ws://127.0.0.1:8001/ws | ./bin/client -protocol websocket |
| MQTT | tcp://127.0.0.1:1883 | ./bin/client -protocol mqtt |
| RabbitMQ | 127.0.0.1:5672 | ./bin/client -protocol rabbitmq |
| TCP | 127.0.0.1:9001 | ./bin/client -protocol tcp |
| UDP | 127.0.0.1:9002 | ./bin/client -protocol udp |
| KCP | 127.0.0.1:9003 | ./bin/client -protocol kcp |

- -protocol 选择协议，默认 http；-addr 覆盖协议默认地址。
- -name 指定名称，默认 World；-timeout 是整体请求超时，默认 10s。
- HTTP/gRPC 客户端验证创建后查询；WS 验证两个会话广播；消息协议核对关联 ID；字节协议核对回显。验证失败返回非零状态。

## WebSocket

标准文本 JSON 请求为 {"type":1,"name":"World"}；所有连接收到 {"type":1,"message":"Hello World"}。不支持的类型或非法名称只向发送者返回 error。消息上限 4 KiB，使用默认同源检查；空闲连接受服务端读写超时限制。

## MQTT 与 RabbitMQ

先订阅回复再发送请求：

```json
{"request_id":"demo-1","name":"World"}
```

回复：

```json
{"request_id":"demo-1","message":"Hello World","id":1}
```

request_id 必须是 1–64 位字母、数字、下划线或连字符。非法名称返回同一 request_id 的 error，非法 request_id 拒绝处理。

| 协议 | 请求 | 回复 |
| --- | --- | --- |
| MQTT | greeting/requests | greeting/replies/{request_id} |
| RabbitMQ | exchange greeting、routing key greeting.request | 同一 exchange、routing key greeting.reply.{request_id} |

MQTT 示例使用 QoS 1。RabbitMQ 服务消费持久队列 greeting.requests，处理与回复发布成功后确认；客户端使用独占、自动删除的回复队列。关联 ID 不等于业务幂等或持久化处理保证。

## 字节协议与环境

TCP/KCP 回显字节流，UDP 回显数据报。TCP、WS、KCP timeout 为读写超时，空闲连接会关闭；UDP 读超时后继续接收。KCP 客户端与服务端使用一致的 GREETING_KCP_KEY 和 GREETING_KCP_SALT。

Broker/KCP 的启动步骤见 [部署与示例](../operations/README.md#部署与协议示例)，临时端口、停止和隔离容器验证见 [测试说明](../development/README.md#验证选择)。需要新消息业务时额外明确顺序、重试、幂等、失败与兼容规则。
