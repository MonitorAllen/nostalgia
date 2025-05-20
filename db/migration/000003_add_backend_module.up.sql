CREATE TABLE "roles" (
                         "id" bigserial PRIMARY KEY,
                         "role_name" varchar UNIQUE NOT NULL,
                         "description" text NOT NULL DEFAULT '',
                         "is_system" bool NOT NULL DEFAULT false,
                         "created_at" timestamptz NOT NULL DEFAULT (now())
);

-- 系统内置超级管理员角色
INSERT INTO roles (role_name, description, is_system)
VALUES ('super', '系统内置超级管理员，拥有全部权限', true);

-- 普通管理员角色
INSERT INTO roles (role_name, description, is_system)
VALUES ('admin', '普通管理员，拥有基本管理权限', true);

CREATE TABLE "admins" (
                          "id" bigserial PRIMARY KEY,
                          "username" varchar UNIQUE NOT NULL,
                          "hashed_password" varchar NOT NULL,
                          "is_active" bool NOT NULL DEFAULT false,
                          "role_id" bigint NOT NULL DEFAULT 2,
                          "created_at" timestamptz NOT NULL DEFAULT (now()),
                          "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

INSERT INTO admins (username, hashed_password, is_active, role_id)
VALUES (
           'super',
           '$2a$10$pXJDlyeVCvQrGu.sPsxS4.I6xgPM1ewS1K8pjfo9OTKYs4VZxYp62',
           true,
           1
       );

CREATE TABLE "sys_menus" (
                             "id" bigserial PRIMARY KEY,
                             "name" varchar NOT NULL,
                             "path" varchar NOT NULL DEFAULT '',
                             "icon" varchar NOT NULL DEFAULT '',
                             "is_active" bool NOT NULL DEFAULT false,
                             "type" int NOT NULL DEFAULT 2,
                             "sort" int NOT NULL DEFAULT 0,
                             "parent_id" bigint DEFAULT null,
                             "created_at" timestamptz NOT NULL DEFAULT (now()),
                             "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

INSERT INTO sys_menus (name, path, icon, is_active, type, sort)
VALUES ('系统管理', '', 'pi pi-cog', true, 1, 1);

INSERT INTO sys_menus (name, path, icon, is_active, type, parent_id, sort)
VALUES ('仪表盘', '/dashboard', 'pi pi-home', true, 2, 1, 2);

INSERT INTO sys_menus (name, path, icon, is_active, type, sort)
VALUES ('基本管理', '', 'pi pi-sliders-h', true, 1, 3);

INSERT INTO sys_menus (name, path, icon, is_active, type, parent_id, sort)
VALUES ('用户管理', '/manage/users', 'pi pi-user', true, 2, 3, 4);

INSERT INTO sys_menus (name, path, icon, is_active, type, parent_id, sort)
VALUES ('文章管理', '/manage/articles', 'pi pi-book', true, 2, 3, 5);

CREATE TABLE "role_permissions" (
                                    "id" bigserial PRIMARY KEY,
                                    "role_id" bigint NOT NULL,
                                    "menu_id" bigint NOT NULL,
                                    "created_by" bigint NOT NULL,
                                    "created_at" timestamptz NOT NULL DEFAULT (now())
);

INSERT INTO role_permissions (role_id, menu_id, created_by)
VALUES (1, 1, 1);

INSERT INTO role_permissions (role_id, menu_id, created_by)
VALUES (1, 2, 1);

INSERT INTO role_permissions (role_id, menu_id, created_by)
VALUES (1, 3, 1);

INSERT INTO role_permissions (role_id, menu_id, created_by)
VALUES (1, 4, 1);

INSERT INTO role_permissions (role_id, menu_id, created_by)
VALUES (1, 5, 1);

CREATE UNIQUE INDEX ON "sys_menus" ("name", "parent_id");

CREATE UNIQUE INDEX ON "role_permissions" ("role_id", "menu_id");

COMMENT ON TABLE "roles" IS '角色表';

COMMENT ON COLUMN "roles"."id" IS '主键ID';

COMMENT ON COLUMN "roles"."role_name" IS '角色名称';

COMMENT ON COLUMN "roles"."description" IS '角色描述';

COMMENT ON COLUMN "roles"."is_system" IS '是否为系统角色';

COMMENT ON COLUMN "roles"."created_at" IS '创建时间';

COMMENT ON TABLE "admins" IS '管理员账号表';

COMMENT ON COLUMN "admins"."id" IS '主键ID';

COMMENT ON COLUMN "admins"."username" IS '管理员名称';

COMMENT ON COLUMN "admins"."hashed_password" IS '密码';

COMMENT ON COLUMN "admins"."is_active" IS '是否激活，默认：否';

COMMENT ON COLUMN "admins"."role_id" IS '角色ID，默认：2';

COMMENT ON COLUMN "admins"."created_at" IS '创建时间';

COMMENT ON COLUMN "admins"."updated_at" IS '更新时间';

COMMENT ON TABLE "sys_menus" IS '后台系统菜单表';

COMMENT ON COLUMN "sys_menus"."id" IS '主键ID';

COMMENT ON COLUMN "sys_menus"."name" IS '菜单名称';

COMMENT ON COLUMN "sys_menus"."path" IS '菜单路径';

COMMENT ON COLUMN "sys_menus"."icon" IS '菜单图标';

COMMENT ON COLUMN "sys_menus"."is_active" IS '是否激活，默认：否';

COMMENT ON COLUMN "sys_menus"."type" IS '1：目录；2：菜单；3：按钮（事件）';

COMMENT ON COLUMN "sys_menus"."sort" IS '排序编号';

COMMENT ON COLUMN "sys_menus"."parent_id" IS '父菜单ID';

COMMENT ON COLUMN "sys_menus"."created_at" IS '创建时间';

COMMENT ON TABLE "role_permissions" IS '角色权限表';

COMMENT ON COLUMN "role_permissions"."id" IS '主键ID';

COMMENT ON COLUMN "role_permissions"."role_id" IS '角色ID';

COMMENT ON COLUMN "role_permissions"."menu_id" IS '菜单ID';

COMMENT ON COLUMN "role_permissions"."created_by" IS '创建人ID';

COMMENT ON COLUMN "role_permissions"."created_at" IS '创建时间';

ALTER TABLE "admins" ADD FOREIGN KEY ("role_id") REFERENCES "roles" ("id");

ALTER TABLE "sys_menus" ADD FOREIGN KEY ("parent_id") REFERENCES "sys_menus" ("id");

ALTER TABLE "role_permissions" ADD FOREIGN KEY ("role_id") REFERENCES "roles" ("id");

ALTER TABLE "role_permissions" ADD FOREIGN KEY ("menu_id") REFERENCES "sys_menus" ("id");

ALTER TABLE "role_permissions" ADD FOREIGN KEY ("created_by") REFERENCES "admins" ("id");