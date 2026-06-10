# Stage 1: Build unified frontend
FROM oven/bun:1-alpine AS build-frontend
WORKDIR /app/frontend

COPY frontend/package.json frontend/bun.lock* ./
RUN bun install --frozen-lockfile

COPY frontend/ ./
RUN bun run build

# Stage 2: Final nginx server
FROM nginx:alpine

# 设置时区
RUN apk add --no-cache tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 清理默认文件
RUN rm -rf /usr/share/nginx/html/*

# 拷贝构建后的静态资源
COPY --from=build-frontend /app/frontend/dist /usr/share/nginx/html/

# 拷贝 nginx 配置
COPY nginx.conf /etc/nginx/nginx.conf

# 日志文件夹
RUN mkdir -p /etc/nginx/logs && touch /etc/nginx/logs/error.log && \
    chmod 755 /etc/nginx/logs

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
