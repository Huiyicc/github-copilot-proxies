# 使用官方 Golang 镜像作为构建环境
FROM golang:alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go mod 和 sum 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 使用轻量级的 alpine 镜像
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 从 builder 阶段复制构建的二进制文件
COPY --from=builder /app/main .

# 暴露端口（如果你的应用需要的话）
EXPOSE 11110

# 运行应用
CMD ["./main"]