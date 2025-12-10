DROP INDEX IF EXISTS articles_category_id_is_publish_created_at_idx;

ALTER TABLE articles DROP CONSTRAINT IF EXISTS articles_category_id_fkey;

ALTER TABLE articles DROP COLUMN IF EXISTS category_id;

DROP TABLE IF EXISTS categories;

DELETE FROM role_permissions WHERE menu_id = 3;

DELETE FROM sys_menus WHERE id = 3;

