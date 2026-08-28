#!/usr/bin/env bash
# XiaoLong-Ridy 一键启动全部服务（Linux / macOS）
# 前置：
#   1. 已安装 Go 1.21+ 且 go 在 PATH
#   2. 先启动中间件： docker compose -f deploy/docker/infra.yml up -d
#   3. 可选： export AMAP_API_KEY="你的高德Key" 后 locationsvc 才会真正调用高德
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
mkdir -p logs

run_svc() {
  local dir="$1" yaml="$2"
  echo "=> 启动 $yaml (日志: logs/$yaml.log)"
  go run "./$dir" -f "./$dir/etc/$yaml.yaml" > "logs/$yaml.log" 2>&1 &
}

run_svc rpc/usersvc              usersvc
run_svc rpc/pricesvc             pricesvc
run_svc rpc/paysvc               paysvc
run_svc rpc/locationsvc          locationsvc
run_svc rpc/pushesvc             pushesvc
run_svc rpc/ordersvc             ordersvc
run_svc rpc/dispatchsvc          dispatchsvc
run_svc mq-consumer/location-consumer location-consumer
run_svc job                      job

echo "已全部后台启动，日志在 logs/ 目录。停止： pkill -f 'go run' "
wait
