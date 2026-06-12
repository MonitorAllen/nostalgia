ALTER TABLE articles
  DROP CONSTRAINT IF EXISTS articles_automation_request_id_fkey;

DROP INDEX IF EXISTS automation_article_requests_status_created_at_idx;

ALTER TABLE articles
  DROP COLUMN IF EXISTS automation_request_id,
  DROP COLUMN IF EXISTS automation_status,
  DROP COLUMN IF EXISTS created_by_automation;

DROP TABLE IF EXISTS automation_article_requests;
