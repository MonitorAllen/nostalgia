ALTER TABLE IF EXISTS comments DROP CONSTRAINT IF EXISTS comments_article_id_fkey;

ALTER TABLE IF EXISTS comments DROP CONSTRAINT IF EXISTS comments_from_user_id_fkey;

ALTER TABLE IF EXISTS comments DROP CONSTRAINT IF EXISTS comments_to_user_id_fkey;

DROP TABLE IF EXISTS comments;