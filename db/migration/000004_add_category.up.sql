CREATE TABLE "categories" (
  "id" bigserial PRIMARY KEY,
  "name" varchar UNIQUE NOT NULL DEFAULT '',
  "is_system" bool NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

-- 内置系统分类
INSERT INTO categories (name, is_system)
VALUES ('其他', true)
ON CONFLICT DO NOTHING;

COMMENT ON TABLE "categories" IS '文章分类表';

COMMENT ON COLUMN "categories"."name" IS '分类名称';

COMMENT ON COLUMN "categories"."is_system" IS '是否为系统分类';

ALTER TABLE articles ADD COLUMN category_id bigint NOT NULL DEFAULT 1;

ALTER TABLE "articles" ADD FOREIGN KEY ("category_id") REFERENCES "categories" ("id");

CREATE INDEX ON "articles" ("category_id", "is_publish", "created_at");
