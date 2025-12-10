# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go env -w GO111MODULE=on && \
    go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main main.go

# Run stage
FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

ENV TZ=Asia/Shanghai

COPY --from=builder /app/main .
COPY --from=builder /app/.env .
COPY --from=builder /app/db/migration ./db/migration

COPY start.sh .
COPY wait-for.sh .

RUN chmod +x /app/start.sh /app/wait-for.sh

EXPOSE 8080

ENTRYPOINT [ "/app/start.sh" ]
CMD [ "/app/main" ]
