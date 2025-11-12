-- +goose Up
-- +goose StatementBegin
-- Таблица конфигураций окружения
CREATE TABLE env_configs (
    env_config_id SERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL UNIQUE,
    description VARCHAR(500) NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NULL
);

COMMENT ON TABLE env_configs IS 'Конфигурации тестовых окружений (DEV, QA, STAGING, PROD)';

-- Таблица параметров конфигурации окружения
CREATE TABLE env_config_params (
    env_config_param_id SERIAL PRIMARY KEY,
    env_config_id INT NOT NULL,
    key VARCHAR(100) NOT NULL,
    value TEXT NULL,
    is_sensitive BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_envconfigparams UNIQUE (env_config_id, key)
);

COMMENT ON TABLE env_config_params IS 'Параметры конфигурации окружения (ключ-значение)';
COMMENT ON COLUMN env_config_params.is_sensitive IS 'Признак конфиденциальных данных (пароли, API-ключи)';

CREATE INDEX idx_envconfigparams_key ON env_config_params(key);
CREATE INDEX idx_envconfigparams_sensitive ON env_config_params(env_config_id) WHERE is_sensitive = TRUE;

-- Таблица ошибок конфигурации
CREATE TABLE config_errors (
    config_error_id SERIAL PRIMARY KEY,
    env_config_id INT NOT NULL,
    description TEXT NOT NULL,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    severity_code VARCHAR(20) NOT NULL CHECK (severity_code IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    resolved_at TIMESTAMPTZ NULL,
    resolved_by INT NULL
);

COMMENT ON TABLE config_errors IS 'Журнал ошибок конфигурации окружений';

CREATE INDEX idx_configerrors_unresolved ON config_errors(env_config_id, detected_at) WHERE resolved_at IS NULL;

-- Таблица конфигураций инструментов
CREATE TABLE tool_configs (
    tool_config_id SERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL UNIQUE,
    description VARCHAR(500) NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NULL
);

COMMENT ON TABLE tool_configs IS 'Конфигурации инструментов автотестирования (Selenium, Playwright, Jest)';

-- Таблица параметров конфигурации инструментов
CREATE TABLE tool_config_params (
    tool_config_param_id SERIAL PRIMARY KEY,
    tool_config_id INT NOT NULL,
    key VARCHAR(100) NOT NULL,
    value TEXT NULL,
    is_sensitive BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_toolconfigparams UNIQUE (tool_config_id, key)
);

COMMENT ON TABLE tool_config_params IS 'Параметры конфигурации инструментов (ключ-значение)';

CREATE INDEX idx_toolconfigparams_key ON tool_config_params(key);

-- Таблица шаблонов отчетов
CREATE TABLE report_templates (
    template_id SERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL UNIQUE,
    description TEXT NULL,
    template_content TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE report_templates IS 'Шаблоны отчетов о качестве';

-- Таблица отчетов
CREATE TABLE reports (
    report_id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    template_id INT NOT NULL,
    owner_user_id INT NOT NULL,
    report_content TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE reports IS 'Отчеты о качестве продукта';

CREATE INDEX idx_reports_owner ON reports(owner_user_id);
CREATE INDEX idx_reports_template ON reports(template_id);
CREATE INDEX idx_reports_created ON reports(created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS reports;

DROP TABLE IF EXISTS report_templates;

DROP TABLE IF EXISTS tool_config_params;

DROP TABLE IF EXISTS tool_configs;

DROP TABLE IF EXISTS config_errors;

DROP TABLE IF EXISTS env_config_params;

DROP TABLE IF EXISTS env_configs;
-- +goose StatementEnd
