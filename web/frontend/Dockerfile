# Stage 1: Build the Vue frontend
FROM node:21-alpine AS build-stage

WORKDIR /app

# Install dependencies
COPY frontend/package*.json ./
RUN npm install

# Copy the rest of the application code
COPY frontend/ .

# Build the Vue project for production
RUN npm run build

# Stage 2: Serve the frontend using Nginx
FROM nginx:alpine

# Copy the built files from the previous stage
COPY --from=build-stage /app/dist /usr/share/nginx/html

# Copy Nginx configuration file (if needed)
COPY frontend/nginx.conf /etc/nginx/nginx.conf

RUN mkdir -p /etc/nginx/logs && touch /etc/nginx/logs/error.log && touch /etc/nginx/logs/error.log && \
    chown -R nginx:nginx /etc/nginx/logs && chmod 755 /etc/nginx/logs

# Expose port 80 for serving the frontend
EXPOSE 80

# Start Nginx
CMD ["nginx", "-g", "daemon off;"]
