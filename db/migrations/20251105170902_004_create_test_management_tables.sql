-- +goose Up
-- +goose StatementBegin
-- Таблица тест-планов
CREATE TABLE test_plans (
    test_plan_id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL UNIQUE,
    description TEXT NULL,
    priority_id INT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    acceptance_criteria TEXT NULL,
    owner_user_id INT NOT NULL,
    version_no INT NOT NULL DEFAULT 1,
    status VARCHAR(50) NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'ACTIVE', 'COMPLETED', 'ARCHIVED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NULL,
    CONSTRAINT chk_testplans_dates CHECK (start_date <= end_date)
);

COMMENT ON TABLE test_plans IS 'Планы тестирования';

CREATE INDEX idx_testplans_owner ON test_plans(owner_user_id);
CREATE INDEX idx_testplans_priority ON test_plans(priority_id);
CREATE INDEX idx_testplans_dates ON test_plans(start_date, end_date);
CREATE INDEX idx_testplans_status ON test_plans(status);

-- Таблица автотестов
CREATE TABLE autotests (
    autotest_id SERIAL PRIMARY KEY,
    test_case_id INT NOT NULL,
    name VARCHAR(200) NOT NULL UNIQUE,
    description TEXT NULL,
    owner_user_id INT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NULL
);

COMMENT ON TABLE autotests IS 'Автоматизированные тесты (код)';

CREATE INDEX idx_autotests_owner ON autotests(owner_user_id);
CREATE INDEX idx_autotests_active ON autotests(is_active) WHERE is_active = TRUE;

-- Таблица версий автотестов
CREATE TABLE autotest_versions (
    version_id SERIAL PRIMARY KEY,
    autotest_id INT NOT NULL,
    version_string VARCHAR(50) NOT NULL,
    commit_hash VARCHAR(64) NULL,
    change_description TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_autotest_version UNIQUE (autotest_id, version_string)
);

COMMENT ON TABLE autotest_versions IS 'Версии автотестов (версионирование кода)';
COMMENT ON COLUMN autotest_versions.commit_hash IS 'Хеш коммита из системы контроля версий (Git)';

CREATE INDEX idx_autotestversions_autotest ON autotest_versions(autotest_id, created_at DESC);

-- Таблица тест-кейсов
CREATE TABLE test_cases (
    test_case_id SERIAL PRIMARY KEY,
    priority_id INT NOT NULL,
    name VARCHAR(200) NOT NULL UNIQUE,
    description TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NULL,
    owner_user_id INT NOT NULL,
    is_automated BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    estimated_duration_minutes INT NULL CHECK (estimated_duration_minutes > 0)
);

COMMENT ON TABLE test_cases IS 'Тестовые случаи (ручные и автоматизированные)';

CREATE INDEX idx_testcases_priority ON test_cases(priority_id);
CREATE INDEX idx_testcases_owner ON test_cases(owner_user_id);
CREATE INDEX idx_testcases_automated ON test_cases(is_automated) WHERE is_automated = TRUE;
CREATE INDEX idx_testcases_active ON test_cases(is_active) WHERE is_active = TRUE;

-- Таблица шагов тест-кейсов
CREATE TABLE test_case_steps (
    step_id SERIAL PRIMARY KEY,
    test_case_id INT NOT NULL,
    step_order INT NOT NULL CHECK (step_order > 0),
    action_text TEXT NOT NULL,
    expected_result TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_testcasesteps UNIQUE (test_case_id, step_order)
);

COMMENT ON TABLE test_case_steps IS 'Шаги выполнения тестовых случаев';

CREATE INDEX idx_testcasesteps_testcase ON test_case_steps(test_case_id, step_order);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS autotest_versions;

DROP TABLE IF EXISTS autotests;

DROP TABLE IF EXISTS test_case_steps;

DROP TABLE IF EXISTS test_cases;

DROP TABLE IF EXISTS test_plans;
-- +goose StatementEnd
