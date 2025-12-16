DROP INDEX IF EXISTS  articles_slug_idx;
ALTER TABLE articles DROP COLUMN IF EXISTS slug;

ALTER TABLE articles DROP COLUMN IF EXISTS cover;

ALTER TABLE articles DROP COLUMN IF EXISTS last_updated;

ALTER TABLE articles DROP COLUMN IF EXISTS check_outdated;

ALTER TABLE articles DROP COLUMN IF EXISTS read_time;
