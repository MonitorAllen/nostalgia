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
                        "created_at" timestamptz NOT NULL DEFAULT (now()),
                        "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
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

CREATE INDEX ON "articles" ("is_publish", "created_at");

COMMENT ON COLUMN "users"."about" IS '介绍';

COMMENT ON COLUMN "articles"."title" IS '标题';

COMMENT ON COLUMN "articles"."summary" IS '摘要';

COMMENT ON COLUMN "articles"."content" IS '内容';

COMMENT ON COLUMN "articles"."views" IS '浏览量';

COMMENT ON COLUMN "articles"."likes" IS '点赞数';

COMMENT ON COLUMN "articles"."is_publish" IS '是否发布';

COMMENT ON COLUMN "articles"."owner" IS '拥有者';

COMMENT ON COLUMN "tags"."name" IS '名称';

COMMENT ON COLUMN "tags"."article_id" IS '文章ID';

ALTER TABLE "articles" ADD FOREIGN KEY ("owner") REFERENCES "users" ("id");

ALTER TABLE "tags" ADD FOREIGN KEY ("article_id") REFERENCES "articles" ("id");

ALTER TABLE "sessions" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "verify_emails" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");
