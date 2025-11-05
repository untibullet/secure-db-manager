-- +goose Up
-- +goose StatementBegin
-- Таблица пользователей
CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(200) NULL UNIQUE,
    full_name VARCHAR(200) NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_login TIMESTAMPTZ NULL,
    failed_login_attempts INT NOT NULL DEFAULT 0,
    account_locked_until TIMESTAMPTZ NULL,
    password_changed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NULL,
    CONSTRAINT chk_users_email CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT chk_users_failed_attempts CHECK (failed_login_attempts >= 0)
);

COMMENT ON TABLE users IS 'Пользователи системы тестирования';
COMMENT ON COLUMN users.password_hash IS 'Хеш пароля (bcrypt/argon2/scram-sha-256)';
COMMENT ON COLUMN users.failed_login_attempts IS 'Счетчик неудачных попыток входа';

-- Индексы для пользователей
CREATE INDEX idx_users_email ON users(email) WHERE email IS NOT NULL;
CREATE INDEX idx_users_active ON users(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_users_username_lower ON users(LOWER(username));

-- Таблица связи пользователей и ролей
CREATE TABLE user_roles (
    user_id INT NOT NULL,
    role_id INT NOT NULL,
    granted_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    granted_by INT NULL,
    valid_until TIMESTAMPTZ NULL,
    PRIMARY KEY (user_id, role_id)
);

COMMENT ON TABLE user_roles IS 'Связь пользователей с ролями (многие-ко-многим)';
COMMENT ON COLUMN user_roles.valid_until IS 'Дата окончания действия роли (NULL = бессрочно)';

CREATE TABLE audit_log (
    audit_id BIGSERIAL PRIMARY KEY,
    table_name VARCHAR(100) NOT NULL,
    operation VARCHAR(10) NOT NULL CHECK (operation IN ('INSERT', 'UPDATE', 'DELETE')),
    record_id INT NOT NULL,
    user_id INT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    old_values JSONB NULL,
    new_values JSONB NULL,
    ip_address INET NULL
);

COMMENT ON TABLE audit_log IS 'Журнал аудита изменений данных';

CREATE INDEX idx_auditlog_table ON audit_log(table_name, changed_at DESC);
CREATE INDEX idx_auditlog_user ON audit_log(user_id, changed_at DESC);
CREATE INDEX idx_auditlog_record ON audit_log(table_name, record_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_roles;

DROP TABLE IF EXISTS audit_log;

DROP TABLE IF EXISTS users;
-- +goose StatementEnd
