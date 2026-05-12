#!/usr/bin/env bash
# OctoTify 前端开发服务器启动脚本
# 功能：安装依赖 → 启动 Vite 开发服务器
# 用法：./run/start-frontend.sh

set -euo pipefail

# 定位项目目录
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
FRONTEND_DIR="$PROJECT_DIR/frontend"

# 检查 Node.js 是否安装
if ! command -v node &> /dev/null; then
  echo "错误：未找到 Node.js，请先安装 Node.js"
  exit 1
fi

# 检查 pnpm 是否安装
if ! command -v pnpm &> /dev/null; then
  echo "错误：未找到 pnpm，请先安装 pnpm"
  exit 1
fi

echo "Node.js 版本: $(node -v)"
echo "pnpm 版本: $(pnpm -v)"

# 检查 node_modules 是否存在
if [ ! -d "$FRONTEND_DIR/node_modules" ]; then
  echo "正在安装依赖..."
  cd "$FRONTEND_DIR"
  pnpm install
fi

# 启动 Vite 开发服务器（端口 8080，由 vite.config.ts 配置）
echo "正在启动前端开发服务器..."
cd "$FRONTEND_DIR"
pnpm dev
