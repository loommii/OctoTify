#!/usr/bin/env bash
# OctoTify 后端服务启动脚本
# 功能：停止旧进程 → 构建最新代码 → 启动服务
# 用法：./run/start.sh
# 环境变量：PORT（可选，默认 34123）

set -euo pipefail

# 定位项目目录
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BACKEND_DIR="$PROJECT_DIR/backend"

# 端口配置，默认 34123
PORT="${PORT:-34123}"

# 停止占用端口的旧进程
if lsof -ti :"$PORT" >/dev/null 2>&1; then
  echo "正在停止端口 $PORT 上的旧进程..."
  lsof -ti :"$PORT" | xargs kill -9 2>/dev/null || true
  sleep 1
fi

# 编译最新代码
echo "正在编译..."
cd "$BACKEND_DIR"
go build -o server ./cmd/server

# 启动服务
echo "正在启动服务，监听端口 :$PORT ..."
cd "$BACKEND_DIR"
./server
