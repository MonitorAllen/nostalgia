ALTER TABLE ai_provider_configs
    ADD COLUMN prompt_templates jsonb NOT NULL DEFAULT '{}'::jsonb;
