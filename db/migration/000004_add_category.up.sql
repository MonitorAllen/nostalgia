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

INSERT INTO sys_menus (id, name, path, icon, is_active, type, parent_id, sort)
VALUES (3, '分类管理', '/manage/categories', 'pi pi-tag', true, 2, 1, (SELECT COALESCE(MAX(sort), 0) + 1 FROM sys_menus WHERE parent_id = 1));

INSERT INTO role_permissions (role_id, menu_id, created_by)
VALUES (1, 3, 1);