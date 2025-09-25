# Stage 1: Build frontend
FROM node:23-alpine AS build-frontend
WORKDIR /app
COPY frontend/package*.json ./frontend/
RUN cd frontend && npm install
COPY frontend ./frontend
RUN cd frontend && npm run build

# Stage 2: Build backend
FROM node:23-alpine AS build-backend
WORKDIR /app
COPY backend/package*.json ./backend/
RUN cd backend && npm install
COPY backend ./backend
RUN cd backend && npm run build

# Stage 3: Final nginx server
FROM nginx:alpine

# 拷贝构建后的静态资源
COPY --from=build-frontend /app/frontend/dist /usr/share/nginx/html/
COPY --from=build-backend /app/backend/dist /usr/share/nginx/html/backend/

# 拷贝 nginx 配置
COPY nginx.conf /etc/nginx/nginx.conf

# 日志文件夹
RUN mkdir -p /etc/nginx/logs && touch /etc/nginx/logs/error.log && \
    chmod 755 /etc/nginx/logs

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]