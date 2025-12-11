#!/bin/bash
# ================= 环境变量 =================
# export DB_USER=your db user

if [ -z "${DB_USER}" ]; then
  echo "❌ 致命错误: 环境变量 [DB_USER] 未设置或为空！"
  exit 1
fi

echo "✅ 环境变量检查通过: DB_USER=${DB_USER}"

# ================= 配置区 =================
# 定义服务名称 (docker-compose.yaml 中的 service key)
DB_SERVICE="postgres"
APP_SERVICES="api web caddy redis"

# 定义文件和卷名
BACKUP_FILE="nostalgia_data_$(date +%Y%m%d_%H%M%S).sql"
# ⚠️ 注意：这里最好用 docker volume ls 确认一下真实卷名
DATA_VOLUME_NAME="nostalgia_data-volume"

echo "--- 🚀 PGroonga 升级部署脚本启动 ---"

# ---------------- 1. 验证阶段 ----------------
# 检查是否在 docker-compose 目录下
if [ ! -f "docker-compose.yaml" ]; then
    echo "❌ 错误: 未找到 docker-compose.yaml，请在项目根目录运行此脚本。"
    exit 1
fi

# ---------------- 2. 更新阶段 ----------------
echo "⬇️  1. 拉取最新代码..."
git pull
if [ $? -ne 0 ]; then
    echo "❌ Git 拉取失败，脚本终止。"
    exit 1
fi

echo "⬇️  2. 预拉取新数据库镜像 (节省停机时间)..."
# 手动拉取新镜像，防止 down 之后再拉取导致服务中断时间过长
# 使用新的 groonga/pgroonga 镜像（确保在 docker-compose.yaml 中已更新 DB 镜像名）
docker pull groonga/pgroonga:3.2.3-alpine-16 || { echo "Error: 无法拉取新的 PGroonga 镜像."; exit 1; }
echo "   -> 新镜像拉取成功."

echo "⬇️  3. 拉取应用镜像..."
docker-compose pull api web

# ---------------- 3. 备份阶段 (关键) ----------------
echo "🛑 4. 停止应用服务 (保留数据库运行以备份)..."
docker-compose stop ${APP_SERVICES}

echo "💾 5. 正在执行数据导出..."
# 使用 docker-compose exec -T
# -T: 禁用伪终端分配 (脚本模式必需)
# --user postgres: 确保以 postgres 用户身份运行 pg_dump，避免权限问题
# pg_dump 参数解释:
# --column-inserts: 以 INSERT INTO (col) VALUES (...) 形式导出 (兼容性更好)
if docker-compose exec -T --user postgres ${DB_SERVICE} \
    pg_dump -U ${DB_USER} -d nostalgia --clean --if-exists --column-inserts \
    > ${BACKUP_FILE}; then

    if grep -q "PostgreSQL database dump" ${BACKUP_FILE}; then
        echo "✅ 数据备份成功: ${BACKUP_FILE} (大小: $(du -h ${BACKUP_FILE} | cut -f1))"
    else
        # 即使 Exit Code 是 0，如果文件里没有 SQL 特征，也视为失败
        echo "❌ 备份文件校验失败: 文件内容看似不是有效的 SQL Dump。"
        echo "   文件内容预览: $(head -n 5 ${BACKUP_FILE})"
        exit 1
    fi
else
    # 命令执行失败 (Exit Code 非 0)
    echo "❌ 备份命令执行出错！"
    echo "   错误原因可能已写入文件或输出到屏幕。"

    # 打印文件里的内容（通常是报错信息，比如 unable to find user...）
    if [ -f ${BACKUP_FILE} ]; then
        echo "👇 错误日志内容:"
        cat ${BACKUP_FILE}
    fi
    exit 1
fi

# ---------------- 4. 清理旧环境 ----------------
echo "🗑️  6. 移除旧数据库容器和数据卷..."
# 停止旧数据库
docker-compose stop ${DB_SERVICE}
# 移除旧数据库容器
docker-compose rm -f ${DB_SERVICE}
# 移除旧数据卷 (高危操作！)
docker volume rm ${DATA_VOLUME_NAME}
if [ $? -eq 0 ]; then
    echo "✅ 旧数据卷已清理."
else
    echo "⚠️  警告: 数据卷清理失败或不存在，尝试继续..."
fi

# ---------------- 5. 启动新环境 ----------------
echo "🚀 7. 启动新数据库容器..."
docker-compose up -d ${DB_SERVICE}

echo "⏳ 等待数据库初始化 (2秒)..."

sleep 2

echo "📥 8. 导入数据..."
# 【技巧】使用 docker-compose exec -T 将本地文件通过管道传给 psql
# 这比启动临时容器更简单，直接利用新容器的 psql 工具
cat ${BACKUP_FILE} | docker-compose exec -T --user postgres ${DB_SERVICE} \
    psql -U ${DB_USER} -d nostalgia

if [ $? -eq 0 ]; then
    echo "✅ 数据导入成功."
else
    echo "❌ 数据导入失败，请查看上方错误日志。"
    exit 1
fi

# ---------------- 6. 收尾 ----------------
echo "🚀 9. 启动所有应用服务..."
docker-compose up -d "${APP_SERVICES}"

echo "🧹 10. 清理备份文件..."
# 建议先保留，确认无误后再手动删，或者移动到 /tmp
mv ${BACKUP_FILE} /tmp/
echo "备份文件已移动到 /tmp/${BACKUP_FILE}"

echo "🎉 部署完成！请访问网站验证搜索功能。"