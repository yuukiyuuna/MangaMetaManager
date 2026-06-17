# Stage 1: Build the frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package*.json ./
# 使用国内镜像源加速前端依赖安装
RUN npm config set registry https://registry.npmmirror.com && npm install
COPY web/ ./
RUN npm run build

# Stage 2: Build the backend
FROM golang:alpine AS backend-builder
# 设置 Go 代理加速下载
ENV GOPROXY=https://goproxy.cn,direct
# Install GCC and musl-dev for CGO (required by go-sqlite3)
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build the binary with CGO enabled and optimizations
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o mmm main.go

# Stage 3: Final lightweight image
FROM alpine:latest
# 安装运行时必要的依赖，包括 su-exec 用于降权
RUN apk add --no-cache ca-certificates tzdata su-exec
WORKDIR /app

# 创建 mmm 用户和组，预设 UID/GID 为 1000
RUN addgroup -g 1000 mmm && \
    adduser -u 1000 -G mmm -h /app -D mmm

# 从构建阶段拷贝产物
COPY --from=backend-builder /app/mmm .
# 拷贝前端产物到正确位置，以便 Gin 提供服务
COPY --from=frontend-builder /app/web/dist ./web/dist

# 拷贝并设置启动脚本
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# 创建必要目录并预设权限
# 同时创建 /app/data/logs 以确保日志能正常写入
RUN mkdir -p /app/data/logs /app/logs && \
    chown -R mmm:mmm /app/data /app/logs /app/web/dist /app

# 暴露后端服务端口
EXPOSE 8080

# 设置环境变量
ENV GIN_MODE=release

# 使用脚本作为入口点
ENTRYPOINT ["/app/entrypoint.sh"]

# 默认运行参数
CMD ["./mmm", "serve", "--db", "/app/data/mmm.db", "--port", "8080"]
