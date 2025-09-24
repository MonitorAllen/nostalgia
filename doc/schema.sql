-- SQL dump generated using DBML (dbml.dbdiagram.io)
-- Database: PostgreSQL
-- Generated at: 2025-09-08T09:02:41.246Z

CREATE TABLE "users" (
  "id" uuid PRIMARY KEY,
  "username" varchar UNIQUE NOT NULL,
  "hashed_password" varchar NOT NULL,
  "full_name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "is_email_verified" boolean NOT NULL DEFAULT false,
  "about" varchar NOT NULL DEFAULT '',
  "role" varchar NOT NULL DEFAULT 'visitor',
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "deleted_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "articles" (
  "id" uuid PRIMARY KEY,
  "title" varchar NOT NULL,
  "summary" varchar NOT NULL DEFAULT '',
  "content" text NOT NULL,
  "views" int NOT NULL DEFAULT 0,
  "likes" int NOT NULL DEFAULT 0,
  "is_publish" boolean NOT NULL DEFAULT false,
  "owner" uuid NOT NULL,
  "category_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "deleted_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "categories" (
  "id" bigserial PRIMARY KEY,
  "name" varchar UNIQUE NOT NULL DEFAULT '',
  "is_system" bool NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "comments" (
  "id" bigserial PRIMARY KEY,
  "content" varchar NOT NULL,
  "article_id" uuid NOT NULL,
  "parent_id" int NOT NULL DEFAULT 0,
  "from_user_id" uuid NOT NULL,
  "to_user_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "deleted_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "tags" (
  "id" bigserial PRIMARY KEY,
  "name" varchar NOT NULL,
  "article_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "sessions" (
  "id" uuid PRIMARY KEY,
  "username" varchar NOT NULL,
  "refresh_token" varchar NOT NULL,
  "user_agent" varchar NOT NULL,
  "client_ip" varchar NOT NULL,
  "is_blocked" boolean NOT NULL DEFAULT false,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "verify_emails" (
  "id" bigserial PRIMARY KEY,
  "user_id" varchar NOT NULL,
  "email" varchar NOT NULL,
  "secret_code" varchar NOT NULL,
  "is_used" bool NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "expired_at" timestamptz NOT NULL DEFAULT (now() + interval '15 minutes')
);

CREATE TABLE "roles" (
  "id" bigserial PRIMARY KEY,
  "role_name" varchar UNIQUE NOT NULL,
  "description" varchar NOT NULL DEFAULT '',
  "is_system" bool NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "admins" (
  "id" bigserial PRIMARY KEY,
  "username" varchar UNIQUE NOT NULL,
  "hashed_password" varchar NOT NULL,
  "is_active" bool NOT NULL DEFAULT false,
  "role_id" bigint NOT NULL DEFAULT 2,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
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

CREATE TABLE "role_permissions" (
  "id" bigserial PRIMARY KEY,
  "role_id" bigint NOT NULL,
  "menu_id" bigint NOT NULL,
  "created_by" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE INDEX ON "articles" ("is_publish", "created_at");

CREATE INDEX ON "articles" ("category_id", "is_publish", "created_at");

CREATE UNIQUE INDEX ON "sys_menus" ("name", "parent_id");

CREATE UNIQUE INDEX ON "role_permissions" ("role_id", "menu_id");

COMMENT ON TABLE "users" IS '用户表';

COMMENT ON COLUMN "users"."id" IS '主键ID';

COMMENT ON COLUMN "users"."username" IS '用户名';

COMMENT ON COLUMN "users"."full_name" IS '全名';

COMMENT ON COLUMN "users"."email" IS '邮箱';

COMMENT ON COLUMN "users"."is_email_verified" IS '邮箱是否验证';

COMMENT ON COLUMN "users"."about" IS '介绍';

COMMENT ON COLUMN "users"."created_at" IS '创建时间';

COMMENT ON COLUMN "users"."updated_at" IS '更新时间';

COMMENT ON COLUMN "users"."deleted_at" IS '删除时间';

COMMENT ON TABLE "articles" IS '文章表';

COMMENT ON COLUMN "articles"."id" IS '主键ID';

COMMENT ON COLUMN "articles"."title" IS '标题';

COMMENT ON COLUMN "articles"."summary" IS '摘要';

COMMENT ON COLUMN "articles"."content" IS '内容';

COMMENT ON COLUMN "articles"."views" IS '浏览量';

COMMENT ON COLUMN "articles"."likes" IS '点赞数';

COMMENT ON COLUMN "articles"."is_publish" IS '是否发布';

COMMENT ON COLUMN "articles"."owner" IS '拥有者';

COMMENT ON COLUMN "articles"."category_id" IS '分类ID';

COMMENT ON COLUMN "articles"."updated_at" IS '更新时间';

COMMENT ON COLUMN "articles"."deleted_at" IS '删除时间';

COMMENT ON TABLE "categories" IS '文章分类表';

COMMENT ON COLUMN "categories"."name" IS '分类名称';

COMMENT ON COLUMN "categories"."is_system" IS '是否为系统分类';

COMMENT ON TABLE "comments" IS '文章评论表';

COMMENT ON COLUMN "comments"."id" IS '主键ID';

COMMENT ON COLUMN "comments"."content" IS '评论内容';

COMMENT ON COLUMN "comments"."article_id" IS '文章ID';

COMMENT ON COLUMN "comments"."parent_id" IS '父评论ID';

COMMENT ON COLUMN "comments"."from_user_id" IS '评论人ID';

COMMENT ON COLUMN "comments"."to_user_id" IS '被评论人ID';

COMMENT ON TABLE "tags" IS '标签表';

COMMENT ON COLUMN "tags"."id" IS '主键ID';

COMMENT ON COLUMN "tags"."name" IS '名称';

COMMENT ON COLUMN "tags"."article_id" IS '文章ID';

COMMENT ON COLUMN "tags"."created_at" IS '创建时间';

COMMENT ON TABLE "sessions" IS '用户会话表';

COMMENT ON COLUMN "sessions"."id" IS '主键ID';

COMMENT ON COLUMN "sessions"."username" IS '用户名';

COMMENT ON COLUMN "sessions"."refresh_token" IS '刷新token';

COMMENT ON COLUMN "sessions"."user_agent" IS '用户浏览器';

COMMENT ON COLUMN "sessions"."client_ip" IS '用户IP';

COMMENT ON COLUMN "sessions"."is_blocked" IS '是否阻止登陆';

COMMENT ON COLUMN "sessions"."expires_at" IS '到期时间';

COMMENT ON COLUMN "sessions"."created_at" IS '创建时间';

COMMENT ON TABLE "verify_emails" IS '邮箱验证表';

COMMENT ON COLUMN "verify_emails"."id" IS '主键ID';

COMMENT ON COLUMN "verify_emails"."user_id" IS '用户ID';

COMMENT ON COLUMN "verify_emails"."email" IS '用户邮箱';

COMMENT ON COLUMN "verify_emails"."secret_code" IS '验证密钥';

COMMENT ON COLUMN "verify_emails"."is_used" IS '是否使用';

COMMENT ON COLUMN "verify_emails"."created_at" IS '创建时间';

COMMENT ON COLUMN "verify_emails"."expired_at" IS '到期时间';

COMMENT ON TABLE "roles" IS '角色表';

COMMENT ON COLUMN "roles"."id" IS '主键ID';

COMMENT ON COLUMN "roles"."role_name" IS '角色名称';

COMMENT ON COLUMN "roles"."description" IS '角色描述';

COMMENT ON COLUMN "roles"."is_system" IS '是否为系统角色';

COMMENT ON COLUMN "roles"."created_at" IS '创建时间';

COMMENT ON COLUMN "roles"."updated_at" IS '更新时间';

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

ALTER TABLE "articles" ADD FOREIGN KEY ("owner") REFERENCES "users" ("id");

ALTER TABLE "articles" ADD FOREIGN KEY ("category_id") REFERENCES "categories" ("id");

ALTER TABLE "comments" ADD FOREIGN KEY ("article_id") REFERENCES "articles" ("id") ON DELETE CASCADE;

ALTER TABLE "tags" ADD FOREIGN KEY ("article_id") REFERENCES "articles" ("id") ON DELETE CASCADE;

ALTER TABLE "sessions" ADD FOREIGN KEY ("username") REFERENCES "users" ("id");

ALTER TABLE "verify_emails" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "admins" ADD FOREIGN KEY ("role_id") REFERENCES "roles" ("id");

ALTER TABLE "sys_menus" ADD FOREIGN KEY ("parent_id") REFERENCES "sys_menus" ("id");

ALTER TABLE "role_permissions" ADD FOREIGN KEY ("role_id") REFERENCES "roles" ("id");

ALTER TABLE "role_permissions" ADD FOREIGN KEY ("menu_id") REFERENCES "sys_menus" ("id");

ALTER TABLE "role_permissions" ADD FOREIGN KEY ("created_by") REFERENCES "admins" ("id");
