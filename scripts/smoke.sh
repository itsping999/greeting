#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
task_tmp=$(mktemp -d)
task_server_pid=
task_compose_project="greeting-smoke-$$"
task_brokers=false
if [[ ${1:-} == --brokers ]]; then task_brokers=true; elif [[ $# -gt 0 ]]; then echo "usage: $0 [--brokers]" >&2; exit 2; fi
cleanup() {
  if [[ -n $task_server_pid ]]; then
    kill -TERM "$task_server_pid" 2>/dev/null || true
    for _ in $(seq 1 100); do
      if ! kill -0 "$task_server_pid" 2>/dev/null; then break; fi
      sleep .1
    done
    if kill -0 "$task_server_pid" 2>/dev/null; then kill -KILL "$task_server_pid" 2>/dev/null || true; fi
    wait "$task_server_pid" 2>/dev/null || true
  fi
  if [[ $task_brokers == true ]]; then docker compose -p "$task_compose_project" --profile examples down --volumes >/dev/null 2>&1 || true; fi
  rm -rf "$task_tmp"
}
trap cleanup EXIT
export GREETING_KCP_KEY=$(openssl rand -hex 32)
export GREETING_KCP_SALT=$(openssl rand -hex 16)
python3 - "$task_tmp" "$task_brokers" <<'PYTHON'
import pathlib,socket,sys
root=pathlib.Path(sys.argv[1]);brokers=sys.argv[2]=='true'
sockets=[];ports={}
for name in ['http','grpc','websocket','tcp','udp','kcp','mqtt','rabbitmq','management']:
    sock=socket.socket(type=socket.SOCK_DGRAM if name in ['udp','kcp'] else socket.SOCK_STREAM)
    sock.bind(('127.0.0.1',0));sockets.append(sock);ports[name]=sock.getsockname()[1]
lines=['server:']
for name in ['http','grpc','websocket','tcp','udp','kcp']:
    key=name
    lines += [f'  {key}:','    enabled: true',f'    addr: 127.0.0.1:{ports[name]}','    timeout: 5s']
    if name=='websocket':lines.append('    path: /ws')
for name,scheme in [('mqtt','tcp'),('rabbitmq','amqp')]:
    addr=f'{scheme}://127.0.0.1:{ports[name]}' if name=='mqtt' else f'amqp://guest:guest@127.0.0.1:{ports[name]}/'
    lines += [f'  {name}:',f'    enabled: {str(brokers).lower()}',f'    addr: {addr}']
(root/'config.yaml').write_text('\n'.join(lines)+'\n')
(root/'ports.sh').write_text('\n'.join(f'export GREETING_{name.upper()}_PORT={port}' for name,port in ports.items())+'\n')
for sock in sockets:sock.close()
PYTHON
source "$task_tmp/ports.sh"
export GREETING_RABBITMQ_MANAGEMENT_PORT=$GREETING_MANAGEMENT_PORT
if [[ $task_brokers == true ]]; then
  docker compose -p "$task_compose_project" --profile examples up -d --wait --wait-timeout 90
  python3 - "$GREETING_MQTT_PORT" <<'PYTHON'
import socket,sys,time
end=time.monotonic()+20
while time.monotonic()<end:
    try:
        with socket.create_connection(('127.0.0.1',int(sys.argv[1])),timeout=.5):break
    except OSError:time.sleep(.2)
else:raise SystemExit('MQTT did not become ready')
PYTHON
fi
./bin/server -conf "$task_tmp/config.yaml" >"$task_tmp/server.log" 2>&1 &
task_server_pid=$!
for _ in $(seq 1 100); do
  if curl --fail --silent "http://127.0.0.1:$GREETING_HTTP_PORT/healthz" >/dev/null; then break; fi
  if ! kill -0 "$task_server_pid" 2>/dev/null; then cat "$task_tmp/server.log"; exit 1; fi
  sleep .1
done
curl --fail --silent "http://127.0.0.1:$GREETING_HTTP_PORT/healthz" >/dev/null
./bin/client -protocol http -addr "http://127.0.0.1:$GREETING_HTTP_PORT"
./bin/client -protocol grpc -addr "127.0.0.1:$GREETING_GRPC_PORT"
./bin/client -protocol websocket -addr "ws://127.0.0.1:$GREETING_WEBSOCKET_PORT/ws"
./bin/client -protocol tcp -addr "127.0.0.1:$GREETING_TCP_PORT"
./bin/client -protocol udp -addr "127.0.0.1:$GREETING_UDP_PORT"
./bin/client -protocol kcp -addr "127.0.0.1:$GREETING_KCP_PORT"
if [[ $task_brokers == true ]]; then
  ./bin/client -protocol mqtt -addr "tcp://127.0.0.1:$GREETING_MQTT_PORT"
  ./bin/client -protocol rabbitmq -addr "amqp://guest:guest@127.0.0.1:$GREETING_RABBITMQ_PORT/"
fi
status=$(curl --silent -o /dev/null -w '%{http_code}' "http://127.0.0.1:$GREETING_HTTP_PORT/greetings/999999")
[[ $status == 404 ]] || { echo "expected missing greeting to return 404, got $status" >&2; exit 1; }
kill -TERM "$task_server_pid"
for _ in $(seq 1 100); do
  if ! kill -0 "$task_server_pid" 2>/dev/null; then break; fi
  sleep .1
done
if kill -0 "$task_server_pid" 2>/dev/null; then echo "server did not stop in time" >&2; kill -KILL "$task_server_pid"; exit 1; fi
wait "$task_server_pid"
task_server_pid=
echo "Smoke checks passed, including graceful shutdown."
