ALTER TABLE ai_provider_configs
    ADD COLUMN api_protocol text NOT NULL DEFAULT 'chat/completions';
