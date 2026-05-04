#!/usr/bin/env bash
# OctoTify 前端开发服务器启动脚本
# 功能：安装依赖 → 启动 Vite 开发服务器
# 用法：./run/start-frontend.sh
# 环境变量：PORT（可选，默认 3000）

set -euo pipefail

# 定位项目目录
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
FRONTEND_DIR="$PROJECT_DIR/frontend"

# 端口配置，默认 3000
PORT="${PORT:-3000}"

# 检查 Node.js 是否安装
if ! command -v node &> /dev/null; then
  echo "错误：未找到 Node.js，请先安装 Node.js"
  exit 1
fi

# 检查 npm 是否安装
if ! command -v npm &> /dev/null; then
  echo "错误：未找到 npm，请先安装 npm"
  exit 1
fi

echo "Node.js 版本: $(node -v)"
echo "npm 版本: $(npm -v)"

# 检查 node_modules 是否存在
if [ ! -d "$FRONTEND_DIR/node_modules" ]; then
  echo "正在安装依赖..."
  cd "$FRONTEND_DIR"
  npm install
fi

# 启动 Vite 开发服务器
echo "正在启动前端开发服务器，监听端口 :$PORT ..."
cd "$FRONTEND_DIR"
PORT=$PORT npm run dev
