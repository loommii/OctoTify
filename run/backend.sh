#!/bin/bash
# OctoTify 后端单独启动脚本
# 用法: bash run/backend.sh

set -e

PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

echo "启动 OctoTify 后端服务..."
cd "$PROJECT_DIR/backend"
go run ./cmd/server
