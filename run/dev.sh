#!/bin/bash
# OctoTify 开发环境启动脚本
# 用法: bash run/dev.sh

set -e

PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

echo "============================================="
echo "  OctoTify 开发环境启动"
echo "============================================="
echo ""

# 1. 启动后端
echo "[1/2] 启动后端服务..."
cd "$PROJECT_DIR/backend"
go run ./cmd/server &
BACKEND_PID=$!
echo "  后端 PID: $BACKEND_PID"
echo "  后端地址: http://localhost:34123"
echo ""

# 等待后端启动
echo "  等待后端启动..."
sleep 2

# 2. 启动前端
echo "[2/2] 启动前端开发服务器..."
cd "$PROJECT_DIR/frontend"
pnpm dev:ele &
FRONTEND_PID=$!
echo "  前端 PID: $FRONTEND_PID"
echo "  前端地址: http://localhost:5555"
echo ""

echo "============================================="
echo "  全部服务已启动"
echo "  后端: http://localhost:34123"
echo "  前端: http://localhost:5555"
echo "  按 Ctrl+C 停止所有服务"
echo "============================================="

# 捕获退出信号
cleanup() {
    echo ""
    echo "正在停止所有服务..."
    kill $BACKEND_PID 2>/dev/null
    kill $FRONTEND_PID 2>/dev/null
    echo "所有服务已停止"
    exit 0
}
trap cleanup INT TERM

# 等待子进程
wait
