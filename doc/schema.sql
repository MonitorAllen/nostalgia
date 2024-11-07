-- SQL dump generated using DBML (dbml.dbdiagram.io)
-- Database: PostgreSQL
-- Generated at: 2024-09-04T16:29:25.618Z

CREATE TABLE "users" (
  "id" uuid PRIMARY KEY,
  "username" varchar UNIQUE NOT NULL,
  "hashed_password" varchar NOT NULL,
  "full_name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "is_email_verified" boolean NOT NULL DEFAULT false,
  "about" varchar NOT NULL DEFAULT '',
  "role" varchar NOT NULL DEFAULT 'visitor',
  "create_at" timestamptz NOT NULL DEFAULT (now()),
  "update_at" timestamptz NOT NULL DEFAULT (0001-01-01 00:00:00Z),
  "delete_at" timestamptz NOT NULL DEFAULT (0001-01-01 00:00:00Z)
);

CREATE TABLE "post" (
  "id" uuid PRIMARY KEY,
  "title" varchar NOT NULL,
  "summary" varchar NOT NULL DEFAULT '',
  "content" text NOT NULL,
  "views" int NOT NULL DEFAULT 0,
  "likes" int NOT NULL DEFAULT 0,
  "is_publish" boolean NOT NULL DEFAULT false,
  "owner" uuid NOT NULL,
  "create_at" timestamptz NOT NULL DEFAULT (now()),
  "update_at" timestamptz NOT NULL DEFAULT (0001-01-01 00:00:00Z),
  "delete_at" timestamptz NOT NULL DEFAULT (0001-01-01 00:00:00Z)
);

CREATE TABLE "tag" (
  "id" bigserial PRIMARY KEY,
  "name" varchar NOT NULL,
  "post_id" uuid NOT NULL,
  "create_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "sessions" (
  "id" uuid PRIMARY KEY,
  "username" varchar NOT NULL,
  "refresh_token" varchar NOT NULL,
  "user_agent" varchar NOT NULL,
  "client_ip" varchar NOT NULL,
  "is_blocked" boolean NOT NULL DEFAULT false,
  "expires_at" timestamptz NOT NULL,
  "create_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "verify_emails" (
  "id" bigserial PRIMARY KEY,
  "user_id" varchar NOT NULL,
  "email" varchar NOT NULL,
  "secret_code" varchar NOT NULL,
  "is_used" bool NOT NULL DEFAULT false,
  "create_at" timestamptz NOT NULL DEFAULT (now()),
  "expired_at" timestamptz NOT NULL DEFAULT (now() + interval '15 minutes')
);

CREATE INDEX ON "post" ("is_publish", "create_at");

COMMENT ON COLUMN "users"."about" IS '介绍';

COMMENT ON COLUMN "post"."title" IS '标题';

COMMENT ON COLUMN "post"."summary" IS '摘要';

COMMENT ON COLUMN "post"."content" IS '内容';

COMMENT ON COLUMN "post"."views" IS '浏览量';

COMMENT ON COLUMN "post"."likes" IS '点赞数';

COMMENT ON COLUMN "post"."is_publish" IS '是否发布';

COMMENT ON COLUMN "post"."owner" IS '拥有者';

COMMENT ON COLUMN "tag"."name" IS '名称';

COMMENT ON COLUMN "tag"."post_id" IS '文章ID';

ALTER TABLE "post" ADD FOREIGN KEY ("owner") REFERENCES "users" ("id");

ALTER TABLE "tag" ADD FOREIGN KEY ("post_id") REFERENCES "post" ("id");

ALTER TABLE "sessions" ADD FOREIGN KEY ("username") REFERENCES "users" ("id");

ALTER TABLE "verify_emails" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");
