ALTER TABLE articles
  ADD COLUMN created_by_automation boolean NOT NULL DEFAULT false,
  ADD COLUMN automation_status varchar(32) NOT NULL DEFAULT '',
  ADD COLUMN automation_request_id bigint;

CREATE TABLE automation_article_requests (
  id bigserial PRIMARY KEY,
  idempotency_key varchar(160) UNIQUE NOT NULL,
  request_hash varchar(64) NOT NULL,
  key_id varchar(100) NOT NULL,
  status varchar(32) NOT NULL,
  article_id uuid,
  title varchar NOT NULL DEFAULT '',
  source_topic varchar NOT NULL DEFAULT '',
  source_prompt text NOT NULL DEFAULT '',
  generation_model varchar NOT NULL DEFAULT '',
  error_message text NOT NULL DEFAULT '',
  client_ip varchar NOT NULL DEFAULT '',
  user_agent varchar NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE INDEX automation_article_requests_status_created_at_idx
  ON automation_article_requests (status, created_at);

ALTER TABLE articles
  ADD CONSTRAINT articles_automation_request_id_fkey
  FOREIGN KEY (automation_request_id) REFERENCES automation_article_requests(id);
