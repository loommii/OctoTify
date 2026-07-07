#!/bin/bash

# OctoTify Web E2E 测试一键启动脚本
# 用法: ./run/test_web.sh [选项]
# 选项:
#   --headed    可视化模式运行（显示浏览器窗口）
#   --ui        UI 模式运行（Playwright UI）
#   --report    仅查看上次测试报告
#   --install   仅安装浏览器依赖
#   --file=     运行指定的测试文件（如 --file=specs/auth.spec.ts）
#   --grep=     运行匹配指定名称的用例（如 --grep="C-01"）
#   无参数      默认无头模式运行所有测试

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
E2E_DIR="$PROJECT_DIR/e2e"

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  OctoTify Web E2E 测试${NC}"
echo -e "${GREEN}========================================${NC}"

# 检查 Node.js
if ! command -v node &> /dev/null; then
    echo -e "${RED}错误: 未找到 Node.js${NC}"
    exit 1
fi

# 检查 pnpm
if ! command -v pnpm &> /dev/null; then
    echo -e "${RED}错误: 未找到 pnpm${NC}"
    exit 1
fi

# 解析参数
MODE="headless"
TEST_FILE=""
TEST_GREP=""
for arg in "$@"; do
    case $arg in
        --headed)
            MODE="headed"
            shift
            ;;
        --ui)
            MODE="ui"
            shift
            ;;
        --report)
            MODE="report"
            shift
            ;;
        --install)
            MODE="install"
            shift
            ;;
        --file=*)
            TEST_FILE="${arg#*=}"
            shift
            ;;
        --grep=*)
            TEST_GREP="${arg#*=}"
            shift
            ;;
        *)
            echo -e "${RED}未知参数: $arg${NC}"
            echo "用法: $0 [--headed|--ui|--report|--install|--file=xxx|--grep=xxx]"
            exit 1
            ;;
    esac
done

# 进入 e2e 目录
cd "$E2E_DIR"

# 安装依赖
if [ ! -d "node_modules" ]; then
    echo -e "${YELLOW}正在安装 E2E 测试依赖...${NC}"
    pnpm install
fi

# 检查 Playwright 浏览器
if [ "$MODE" != "install" ] && [ "$MODE" != "report" ]; then
    if [ ! -d "$HOME/.cache/ms-playwright" ]; then
        echo -e "${YELLOW}检测到 Playwright 浏览器未安装，正在安装...${NC}"
        pnpm run install:browsers
    fi
fi

# 检查前后端服务（仅在实际运行测试时）
if [ "$MODE" != "install" ] && [ "$MODE" != "report" ]; then
    echo -e "${BLUE}正在检查服务状态...${NC}"
    
    # 检查前端服务
    if curl -s -o /dev/null -w "%{http_code}" http://localhost:5777 | grep -q "200\|301\|302"; then
        echo -e "${GREEN}✓ 前端服务运行中 (http://localhost:5777)${NC}"
    else
        echo -e "${YELLOW}⚠ 前端服务可能未运行 (http://localhost:5777)${NC}"
        echo -e "${YELLOW}  请确保前端服务已启动，否则测试可能失败${NC}"
    fi
    
    # 检查后端服务
    if curl -s -o /dev/null -w "%{http_code}" http://localhost:34123/api/health | grep -q "200\|404"; then
        echo -e "${GREEN}✓ 后端服务运行中 (http://localhost:34123)${NC}"
    else
        echo -e "${YELLOW}⚠ 后端服务可能未运行 (http://localhost:34123)${NC}"
        echo -e "${YELLOW}  请确保后端服务已启动，否则测试可能失败${NC}"
    fi
    echo ""
fi

# 构建测试命令
TEST_CMD="pnpm run test:e2e"
if [ -n "$TEST_FILE" ]; then
    # 如果未包含路径前缀，自动添加 specs/
    if [[ "$TEST_FILE" != specs/* ]] && [[ "$TEST_FILE" != tests/* ]]; then
        TEST_FILE="specs/$TEST_FILE"
    fi
    TEST_CMD="$TEST_CMD $TEST_FILE"
fi
if [ -n "$TEST_GREP" ]; then
    TEST_CMD="$TEST_CMD --grep '$TEST_GREP'"
fi

# 根据模式执行
case $MODE in
    install)
        echo -e "${YELLOW}正在安装 Playwright 浏览器...${NC}"
        pnpm run install:browsers
        echo -e "${GREEN}浏览器安装完成!${NC}"
        ;;
    report)
        echo -e "${YELLOW}正在打开测试报告...${NC}"
        pnpm run test:e2e:report
        ;;
    headed)
        echo -e "${YELLOW}正在启动可视化 E2E 测试...${NC}"
        if [ -n "$TEST_FILE" ] || [ -n "$TEST_GREP" ]; then
            $TEST_CMD --headed
        else
            pnpm run test:e2e:headed
        fi
        ;;
    ui)
        echo -e "${YELLOW}正在启动 E2E 测试 UI...${NC}"
        pnpm run test:e2e:ui
        ;;
    *)
        echo -e "${YELLOW}正在运行无头模式 E2E 测试...${NC}"
        if [ -n "$TEST_FILE" ] || [ -n "$TEST_GREP" ]; then
            eval $TEST_CMD
        else
            pnpm run test:e2e
        fi
        ;;
esac

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  测试完成!${NC}"
echo -e "${GREEN}========================================${NC}"
