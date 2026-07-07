#!/bin/bash
# OctoTify 开发环境启动脚本
# 用法: bash run/dev.sh

set -e

PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BACKEND_PORT=34123
FRONTEND_PORT=5555

echo "============================================="
echo "  OctoTify 开发环境启动"
echo "============================================="
echo ""

# 颜色输出
info()  { echo -e "\033[0;32m[INFO]\033[0m $1"; }
warn()  { echo -e "\033[0;33m[WARN]\033[0m $1"; }
error() { echo -e "\033[0;31m[ERROR]\033[0m $1"; }

# 清理函数：停止所有后台子进程
CLEANUP_RAN=0
cleanup() {
    if [ "$CLEANUP_RAN" -ne 0 ]; then
        return
    fi
    CLEANUP_RAN=1
    echo ""
    info "正在停止所有服务..."
    # 终止整个进程组，确保后台任务被清理
    kill -- -$$ 2>/dev/null || true
    info "所有服务已停止"
    exit 0
}
trap cleanup INT TERM

# 检查端口是否已被占用
check_port() {
    local port=$1 name=$2
    if lsof -i :"$port" >/dev/null 2>&1; then
        warn "$name 端口 $port 已被占用，跳过启动"
        return 1
    fi
    return 0
}

# 等待后端就绪（最多 30 秒）
wait_for_backend() {
    local url="http://localhost:$BACKEND_PORT/ping"
    info "等待后端就绪..."
    for i in $(seq 1 30); do
        if curl -sf "$url" >/dev/null 2>&1; then
            info "后端就绪（第 ${i}s）"
            return 0
        fi
        sleep 1
    done
    error "后端启动超时（30s），请检查日志"
    return 1
}

# 等待前端就绪（最多 30 秒）
wait_for_frontend() {
    info "等待前端就绪..."
    for i in $(seq 1 30); do
        if curl -sf "http://localhost:$FRONTEND_PORT" >/dev/null 2>&1; then
            info "前端就绪（第 ${i}s）"
            return 0
        fi
        sleep 1
    done
    warn "前端启动超时（30s），可能仍在编译中，请稍后手动刷新"
    return 0
}

# 1. 启动后端
BACKEND_PID=""
if check_port "$BACKEND_PORT" "后端"; then
    info "[1/2] 启动后端服务..."
    cd "$PROJECT_DIR/backend"
    go run ./cmd/server &
    BACKEND_PID=$!
    info "  后端 PID: $BACKEND_PID"
    info "  后端地址: http://localhost:$BACKEND_PORT"
    echo ""

    # 等待后端就绪
    wait_for_backend || { cleanup; exit 1; }
else
    echo ""
fi

# 2. 启动前端
FRONTEND_PID=""
if check_port "$FRONTEND_PORT" "前端"; then
    info "[2/2] 启动前端开发服务器..."
    cd "$PROJECT_DIR/frontend"
    pnpm dev:ele &
    FRONTEND_PID=$!
    info "  前端 PID: $FRONTEND_PID"
    info "  前端地址: http://localhost:$FRONTEND_PORT"
    echo ""

    # 等待前端就绪（非阻塞，超时不中断）
    wait_for_frontend
else
    echo ""
fi

echo "============================================="
info "  全部服务已启动"
info "  后端: http://localhost:$BACKEND_PORT"
info "  前端: http://localhost:$FRONTEND_PORT"
info "  按 Ctrl+C 停止所有服务"
echo "============================================="

# 等待任意子进程退出（保持脚本前台运行）
wait
