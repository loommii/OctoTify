#!/bin/bash
# OctoTify 前端单独启动脚本
# 用法: bash run/frontend.sh

set -e

PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

echo "启动 OctoTify 前端开发服务器..."
cd "$PROJECT_DIR/frontend"
pnpm dev:ele
