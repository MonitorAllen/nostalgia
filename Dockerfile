# syntax=docker/dockerfile:1.7

# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go env -w GO111MODULE=on && \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main main.go

# Run stage
FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

ENV TZ=Asia/Shanghai

COPY --from=builder /app/main .
COPY --from=builder /app/db/migration ./db/migration

COPY start.sh .
COPY wait-for.sh .

RUN chmod +x /app/start.sh /app/wait-for.sh

EXPOSE 8080 9091

ENTRYPOINT [ "/app/start.sh" ]
CMD [ "/app/main" ]
