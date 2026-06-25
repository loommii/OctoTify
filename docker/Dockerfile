# 第一阶段：构建前端
FROM node:20-alpine AS frontend-builder

WORKDIR /frontend

# 先复制依赖文件，利用 Docker 缓存层
COPY frontend/package*.json ./
RUN npm ci

# 复制前端源码并构建
COPY frontend/ ./
RUN npm run build

# 第二阶段：构建后端
FROM golang:1.26-alpine AS backend-builder

WORKDIR /build

# 安装 CGO 依赖（SQLite 驱动需要 C 编译器）
RUN apk add --no-cache git ca-certificates gcc musl-dev

# 先复制依赖文件，利用 Docker 缓存层
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 复制后端源码并构建
COPY backend/ ./
RUN CGO_ENABLED=1 GOOS=linux go build -tags=sonic -ldflags='-s -w' -o octotify ./cmd/server

# 第三阶段：最终运行镜像
FROM alpine:latest

# 安装运行依赖（Nginx、时区数据）
RUN apk add --no-cache ca-certificates tzdata nginx \
    && addgroup -g 1000 app \
    && adduser -D -G app -u 1000 app

WORKDIR /app

# 创建数据目录并设置权限
RUN mkdir -p /app/data /app/data/log /var/log/nginx /run/nginx \
    && chown -R app:app /app /var/log/nginx /var/lib/nginx

# 从构建阶段复制产物
COPY --from=backend-builder /build/octotify .
COPY --from=backend-builder /build/config ./config
COPY --from=frontend-builder /frontend/dist /usr/share/nginx/html

# 复制配置文件和启动脚本
COPY docker/nginx.conf /etc/nginx/http.d/default.conf
COPY docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

USER app

EXPOSE 5233

ENTRYPOINT ["/entrypoint.sh"]
