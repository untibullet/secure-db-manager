-- +goose Up
-- +goose StatementBegin
-- Таблица версий приложения
CREATE TABLE app_versions (
    version_id SERIAL PRIMARY KEY,
    version_string VARCHAR(50) NOT NULL UNIQUE,
    build_date DATE NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_version_string CHECK (version_string ~ '^\d+\.\d+\.\d+')
);

COMMENT ON TABLE app_versions IS 'Версии тестируемого приложения';
COMMENT ON COLUMN app_versions.version_string IS 'Семантическая версия (формат: X.Y.Z)';

-- Таблица приоритетов
CREATE TABLE priorities (
    priority_id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    rank SMALLINT NOT NULL CHECK (rank BETWEEN 1 AND 5),
    color_hex VARCHAR(7) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE priorities IS 'Справочник приоритетов тестовых случаев и планов';
COMMENT ON COLUMN priorities.rank IS 'Уровень важности от 1 (низкий) до 5 (критический)';

-- Таблица статусов
CREATE TABLE statuses (
    status_id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50) NULL CHECK (category IN ('TEST_EXECUTION', 'PLAN', 'RUN', 'GENERAL')),
    is_final BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE statuses IS 'Универсальный справочник статусов';
COMMENT ON COLUMN statuses.is_final IS 'Признак финального статуса (не подлежит изменению)';

-- Таблица ролей пользователей
CREATE TABLE roles (
    role_id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT NULL,
    db_role_name VARCHAR(100) NOT NULL UNIQUE,
    access_level SMALLINT NOT NULL DEFAULT 1 CHECK (access_level BETWEEN 1 AND 10),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE roles IS 'Справочник ролей пользователей в системе';
COMMENT ON COLUMN roles.access_level IS 'Уровень доступа: 1-минимальный (Guest), 10-максимальный (Administrator)';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE roles;

DROP TABLE statuses;

DROP TABLE priorities;

DROP TABLE app_versions;
-- +goose StatementEnd
