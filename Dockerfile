FROM golang:1.21-alpine AS builder
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.18
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai
ENV PORT=80
ENV SERVER_PORT=80
ENV GIN_MODE=release
ENV CLOUDBASE_INSTANCE=(default)
ENV CLOUDBASE_DATABASE=(default)
WORKDIR /app
COPY --from=builder /server ./
# 含 config.yaml；若存在 apikey 会一并打入镜像供运行时读取
COPY backend/config/ ./config/
EXPOSE 80
CMD ["./server"]
