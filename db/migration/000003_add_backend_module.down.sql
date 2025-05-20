-- 1. 首先删除角色权限表（依赖role和sys_menu）
DROP TABLE IF EXISTS role_permissions CASCADE;

-- 2. 删除菜单表（自引用外键需要CASCADE）
DROP TABLE IF EXISTS sys_menus CASCADE;

-- 3. 删除管理员表（依赖role表）
DROP TABLE IF EXISTS admins CASCADE;

-- 4. 最后删除角色表
DROP TABLE IF EXISTS roles CASCADE;
