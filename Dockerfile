# 使用官方 Golang 镜像作为构建环境
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制源代码
COPY . .

# 检查 go.mod 是否存在，如果不存在则初始化一个新的模块
RUN if [ ! -f go.mod ]; then \
    go mod init myapp; \
    fi

# 下载依赖
RUN go mod tidy

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 使用轻量级的 alpine 镜像
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 从 builder 阶段复制构建的二进制文件
COPY --from=builder /app/main .
COPY .env.example .env

# 复制 SSL 证书文件夹
COPY ssl /root/ssl

# 暴露端口
EXPOSE 1188
EXPOSE 443

# 运行应用
CMD ["./main"]