CREATE TABLE ai_provider_configs (
    purpose varchar(64) PRIMARY KEY,
    provider varchar(64) NOT NULL,
    base_url text NOT NULL,
    model text NOT NULL,
    api_key_ciphertext text NOT NULL DEFAULT '',
    timeout_ms integer NOT NULL DEFAULT 30000,
    max_input_chars integer NOT NULL DEFAULT 6000,
    max_context_chars integer NOT NULL DEFAULT 4000,
    max_suggestions integer NOT NULL DEFAULT 3,
    enabled boolean NOT NULL DEFAULT true,
    updated_by uuid REFERENCES users (id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ai_provider_configs_timeout_ms_positive CHECK (timeout_ms > 0),
    CONSTRAINT ai_provider_configs_max_input_chars_positive CHECK (max_input_chars > 0),
    CONSTRAINT ai_provider_configs_max_context_chars_non_negative CHECK (max_context_chars >= 0),
    CONSTRAINT ai_provider_configs_max_suggestions_positive CHECK (max_suggestions > 0)
);
