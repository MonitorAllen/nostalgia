ALTER TABLE articles ADD COLUMN slug varchar(100);
CREATE UNIQUE INDEX ON "articles" ("slug");
COMMENT ON COLUMN "articles"."slug" IS '短标识';

ALTER TABLE articles ADD COLUMN cover varchar NOT NULL default '';
COMMENT ON COLUMN "articles"."cover" IS '封面';

ALTER TABLE articles ADD COLUMN last_updated timestamptz NOT NULL default now();
COMMENT ON COLUMN "articles"."last_updated" IS '最后更新时间';

ALTER TABLE articles ADD COLUMN "check_outdated" bool NOT NULL DEFAULT true;
COMMENT ON COLUMN "articles"."check_outdated" IS '检查过时';

ALTER TABLE articles ADD COLUMN "read_time" varchar(20) NOT NULL DEFAULT '';
COMMENT ON COLUMN "articles"."read_time" IS '阅读时间';

