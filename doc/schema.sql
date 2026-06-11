-- Current PostgreSQL schema after migrations 000001 through 000009.

CREATE EXTENSION IF NOT EXISTS pgroonga;

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
  "deleted_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  CONSTRAINT users_role_check CHECK (role IN ('admin', 'visitor'))
);

CREATE TABLE "categories" (
  "id" bigserial PRIMARY KEY,
  "name" varchar UNIQUE NOT NULL DEFAULT '',
  "is_system" bool NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
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
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "deleted_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "category_id" bigint NOT NULL DEFAULT 1,
  "slug" varchar(100),
  "cover" varchar NOT NULL DEFAULT '',
  "last_updated" timestamptz NOT NULL DEFAULT now(),
  "check_outdated" bool NOT NULL DEFAULT true,
  "read_time" varchar(20) NOT NULL DEFAULT ''
);

CREATE TABLE "tags" (
  "id" bigserial PRIMARY KEY,
  "name" varchar NOT NULL,
  "article_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "sessions" (
  "id" uuid PRIMARY KEY,
  "user_id" uuid NOT NULL,
  "refresh_token" varchar NOT NULL,
  "user_agent" varchar NOT NULL,
  "client_ip" varchar NOT NULL,
  "is_blocked" boolean NOT NULL DEFAULT false,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "verify_emails" (
  "id" bigserial PRIMARY KEY,
  "user_id" uuid NOT NULL,
  "email" varchar NOT NULL,
  "secret_code" varchar NOT NULL,
  "is_used" bool NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "expired_at" timestamptz NOT NULL DEFAULT (now() + interval '15 minutes')
);

CREATE TABLE "comments" (
  "id" bigserial PRIMARY KEY,
  "content" varchar NOT NULL,
  "article_id" uuid NOT NULL,
  "parent_id" bigint NOT NULL DEFAULT 0,
  "likes" int NOT NULL DEFAULT 0,
  "from_user_id" uuid NOT NULL,
  "to_user_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "deleted_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE INDEX ON "articles" ("is_publish", "created_at");

CREATE INDEX ON "articles" ("category_id", "is_publish", "created_at");

CREATE UNIQUE INDEX ON "articles" ("slug");

CREATE INDEX articles_search_pgroonga_idx ON articles
  USING pgroonga ((title || ' ' || summary || ' ' || content));

CREATE INDEX IF NOT EXISTS users_role_idx ON users(role);

CREATE UNIQUE INDEX IF NOT EXISTS users_single_admin_idx ON users(role) WHERE role = 'admin';

COMMENT ON TABLE "users" IS '用户表';

COMMENT ON COLUMN "users"."id" IS '主键ID';

COMMENT ON COLUMN "users"."username" IS '用户名';

COMMENT ON COLUMN "users"."full_name" IS '全名';

COMMENT ON COLUMN "users"."email" IS '邮箱';

COMMENT ON COLUMN "users"."is_email_verified" IS '邮箱是否验证';

COMMENT ON COLUMN "users"."about" IS '介绍';

COMMENT ON COLUMN "users"."role" IS 'admin 或 visitor';

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

COMMENT ON COLUMN "articles"."slug" IS '短标识';

COMMENT ON COLUMN "articles"."cover" IS '封面';

COMMENT ON COLUMN "articles"."last_updated" IS '最后更新时间';

COMMENT ON COLUMN "articles"."check_outdated" IS '检查过时';

COMMENT ON COLUMN "articles"."read_time" IS '阅读时间';

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

COMMENT ON COLUMN "comments"."likes" IS '点赞数';

COMMENT ON COLUMN "comments"."from_user_id" IS '评论人ID';

COMMENT ON COLUMN "comments"."to_user_id" IS '被评论人ID';

COMMENT ON TABLE "tags" IS '标签表';

COMMENT ON COLUMN "tags"."id" IS '主键ID';

COMMENT ON COLUMN "tags"."name" IS '名称';

COMMENT ON COLUMN "tags"."article_id" IS '文章ID';

COMMENT ON COLUMN "tags"."created_at" IS '创建时间';

COMMENT ON TABLE "sessions" IS '用户会话表';

COMMENT ON COLUMN "sessions"."id" IS '主键ID';

COMMENT ON COLUMN "sessions"."user_id" IS '用户ID';

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

ALTER TABLE "articles" ADD FOREIGN KEY ("owner") REFERENCES "users" ("id");

ALTER TABLE "articles" ADD FOREIGN KEY ("category_id") REFERENCES "categories" ("id");

ALTER TABLE "tags" ADD FOREIGN KEY ("article_id") REFERENCES "articles" ("id");

ALTER TABLE "sessions" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "verify_emails" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "comments" ADD FOREIGN KEY ("article_id") REFERENCES "articles" ("id");

ALTER TABLE "comments" ADD FOREIGN KEY ("from_user_id") REFERENCES "users" ("id");

ALTER TABLE "comments" ADD FOREIGN KEY ("to_user_id") REFERENCES "users" ("id");
