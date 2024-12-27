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


ALTER TABLE "comments" ADD FOREIGN KEY ("article_id") REFERENCES "articles" ("id");
ALTER TABLE "comments" ADD FOREIGN KEY ("from_user_id") REFERENCES "users" ("id");
ALTER TABLE "comments" ADD FOREIGN KEY ("to_user_id") REFERENCES "users" ("id");

COMMENT ON COLUMN "comments"."content" IS '评论内容';

COMMENT ON COLUMN "comments"."article_id" IS '文章ID';

COMMENT ON COLUMN "comments"."parent_id" IS '父评论ID';

COMMENT ON COLUMN "comments"."from_user_id" IS '评论人ID';

COMMENT ON COLUMN "comments"."to_user_id" IS '被评论人ID';