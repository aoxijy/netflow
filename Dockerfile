# 构建阶段：使用 Go 官方轻量镜像
FROM golang:1.21.5-alpine3.19 AS builder

# 设置工作目录
WORKDIR /go/src/

# 将当前目录内容复制到容器中
COPY ./ /go/src/

# 编译 Go 程序
RUN go mod tidy && \
    go build -o netflow && \
    chmod +x netflow

# 运行阶段：使用轻量级 alpine 镜像
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译后的二进制文件
COPY --from=builder /go/src/netflow /app/netflow

# 优化：更改 alpine 源（中国用户加速）并设置时区为 Asia/Shanghai
RUN set -eux && \
    sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && \
    apk --update add --no-cache tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 定义容器启动时运行的命令
ENTRYPOINT ["/app/netflow"]
