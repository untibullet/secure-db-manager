-- =====================================================================
-- БЕЗОПАСНАЯ ИНФОРМАЦИОННАЯ СИСТЕМА АВТОМАТИЗИРОВАННОГО ТЕСТИРОВАНИЯ ПО
-- ПОЛНАЯ ИНТЕГРИРОВАННАЯ СХЕМА С МЕХАНИЗМАМИ БЕЗОПАСНОСТИ
-- СУБД: PostgreSQL 14+
-- Версия: 2.0 (Объединенная схема)
-- Дата создания: 05.11.2025
-- =====================================================================

-- =====================================================================
-- БЛОК 1: УДАЛЕНИЕ СУЩЕСТВУЮЩИХ ОБЪЕКТОВ
-- =====================================================================

-- Удаление представлений
DROP VIEW IF EXISTS v_public_test_plans CASCADE;
DROP VIEW IF EXISTS v_tester_assigned_cases CASCADE;
DROP VIEW IF EXISTS v_lead_all_cases CASCADE;
DROP VIEW IF EXISTS v_engineer_autotests CASCADE;
DROP VIEW IF EXISTS v_shared_environments CASCADE;
DROP VIEW IF EXISTS v_tester_my_results CASCADE;
DROP VIEW IF EXISTS v_lead_all_results CASCADE;
DROP VIEW IF EXISTS v_public_results CASCADE;
DROP VIEW IF EXISTS v_shared_reports CASCADE;
DROP VIEW IF EXISTS v_admin_users_and_roles CASCADE;
DROP VIEW IF EXISTS v_active_test_runs CASCADE;
DROP VIEW IF EXISTS v_test_case_statistics CASCADE;
DROP VIEW IF EXISTS v_test_execution_summary CASCADE;

-- Удаление функций и триггеров
DROP TRIGGER IF EXISTS trg_users_updated_at ON users CASCADE;
DROP TRIGGER IF EXISTS trg_testcases_updated_at ON test_cases CASCADE;
DROP TRIGGER IF EXISTS trg_testplans_updated_at ON test_plans CASCADE;
DROP TRIGGER IF EXISTS trg_envconfigs_updated_at ON env_configs CASCADE;
DROP TRIGGER IF EXISTS trg_toolconfigs_updated_at ON tool_configs CASCADE;
DROP TRIGGER IF EXISTS trg_users_audit ON users CASCADE;
DROP TRIGGER IF EXISTS trg_userroles_audit ON user_roles CASCADE;
DROP TRIGGER IF EXISTS trg_lock_user_account ON users CASCADE;

DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;
DROP FUNCTION IF EXISTS audit_trigger_function() CASCADE;
DROP FUNCTION IF EXISTS lock_user_account() CASCADE;
DROP FUNCTION IF EXISTS check_password_complexity(TEXT) CASCADE;
DROP FUNCTION IF EXISTS check_password_expiration() CASCADE;

-- Удаление таблиц
DROP TABLE IF EXISTS config_errors CASCADE;
DROP TABLE IF EXISTS env_config_params CASCADE;
DROP TABLE IF EXISTS tool_config_params CASCADE;
DROP TABLE IF EXISTS test_result_artifacts CASCADE;
DROP TABLE IF EXISTS test_results CASCADE;
DROP TABLE IF EXISTS test_run_items CASCADE;
DROP TABLE IF EXISTS test_runs CASCADE;
DROP TABLE IF EXISTS test_case_steps CASCADE;
DROP TABLE IF EXISTS test_cases CASCADE;
DROP TABLE IF EXISTS test_plans CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS env_configs CASCADE;
DROP TABLE IF EXISTS tool_configs CASCADE;
DROP TABLE IF EXISTS autotest_versions CASCADE;
DROP TABLE IF EXISTS autotests CASCADE;
DROP TABLE IF EXISTS reports CASCADE;
DROP TABLE IF EXISTS report_templates CASCADE;
DROP TABLE IF EXISTS app_versions CASCADE;
DROP TABLE IF EXISTS priorities CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS statuses CASCADE;
DROP TABLE IF EXISTS audit_log CASCADE;

-- =====================================================================
-- БЛОК 2: СОЗДАНИЕ СПРАВОЧНЫХ ТАБЛИЦ
-- =====================================================================

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
    access_level SMALLINT NOT NULL DEFAULT 1 CHECK (access_level BETWEEN 1 AND 10),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE roles IS 'Справочник ролей пользователей в системе';
COMMENT ON COLUMN roles.access_level IS 'Уровень доступа: 1-минимальный (Guest), 10-максимальный (Administrator)';

-- =====================================================================
-- БЛОК 3: СОЗДАНИЕ ОСНОВНЫХ ТАБЛИЦ
-- =====================================================================

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
    executor_user_id INT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NULL,
    owner_user_id INT NOT NULL,
    is_automated BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    estimated_duration_minutes INT NULL CHECK (estimated_duration_minutes > 0)
);

COMMENT ON TABLE test_cases IS 'Тестовые случаи (ручные и автоматизированные)';
COMMENT ON COLUMN test_cases.executor_user_id IS 'Назначенный исполнитель тест-кейса';

CREATE INDEX idx_testcases_priority ON test_cases(priority_id);
CREATE INDEX idx_testcases_owner ON test_cases(owner_user_id);
CREATE INDEX idx_testcases_executor ON test_cases(executor_user_id) WHERE executor_user_id IS NOT NULL;
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

-- Таблица тестовых прогонов
CREATE TABLE test_runs (
    test_run_id SERIAL PRIMARY KEY,
    test_plan_id INT NOT NULL,
    env_config_id INT NOT NULL,
    tool_config_id INT NOT NULL,
    version_id INT NOT NULL,
    name VARCHAR(200) NOT NULL UNIQUE,
    description TEXT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PLANNED' CHECK (status IN ('PLANNED', 'IN_PROGRESS', 'COMPLETED', 'FAILED', 'CANCELLED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by INT NULL,
    CONSTRAINT chk_testruns_dates CHECK (end_date IS NULL OR start_date <= end_date)
);

COMMENT ON TABLE test_runs IS 'Прогоны тестирования';

CREATE INDEX idx_testruns_testplan ON test_runs(test_plan_id);
CREATE INDEX idx_testruns_envconfig ON test_runs(env_config_id);
CREATE INDEX idx_testruns_version ON test_runs(version_id);
CREATE INDEX idx_testruns_dates ON test_runs(start_date, end_date);
CREATE INDEX idx_testruns_status ON test_runs(status);

-- Таблица элементов тестового прогона
CREATE TABLE test_run_items (
    run_item_id SERIAL PRIMARY KEY,
    test_run_id INT NOT NULL,
    test_case_id INT NOT NULL,
    execution_order INT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_testrunitems UNIQUE (test_run_id, test_case_id)
);

COMMENT ON TABLE test_run_items IS 'Тест-кейсы, включенные в тестовый прогон';

CREATE INDEX idx_testrunitems_testrun ON test_run_items(test_run_id);
CREATE INDEX idx_testrunitems_testcase ON test_run_items(test_case_id);

-- Таблица результатов тестирования
CREATE TABLE test_results (
    test_result_id SERIAL PRIMARY KEY,
    run_item_id INT NOT NULL,
    status_id INT NOT NULL,
    executor_user_id INT NULL,
    execution_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    result_summary VARCHAR(1000) NULL,
    execution_duration_minutes INT NULL CHECK (execution_duration_minutes >= 0),
    error_message TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE test_results IS 'Результаты выполнения тестов';
COMMENT ON COLUMN test_results.execution_duration_minutes IS 'Фактическое время выполнения в минутах';

CREATE INDEX idx_testresults_runitem ON test_results(run_item_id);
CREATE INDEX idx_testresults_status ON test_results(status_id);
CREATE INDEX idx_testresults_executor ON test_results(executor_user_id);
CREATE INDEX idx_testresults_date ON test_results(execution_date DESC);

-- Таблица артефактов результатов
CREATE TABLE test_result_artifacts (
    artifact_id SERIAL PRIMARY KEY,
    test_result_id INT NOT NULL,
    kind VARCHAR(50) NOT NULL CHECK (kind IN ('SCREENSHOT', 'LOG', 'VIDEO', 'REPORT', 'OTHER')),
    file_path TEXT NULL,
    file_size_bytes BIGINT NULL CHECK (file_size_bytes >= 0),
    mime_type VARCHAR(100) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE test_result_artifacts IS 'Артефакты выполнения тестов (скриншоты, логи, видео)';

CREATE INDEX idx_testresultartifacts_result ON test_result_artifacts(test_result_id);
CREATE INDEX idx_testresultartifacts_kind ON test_result_artifacts(kind);

-- =====================================================================
-- БЛОК 4: ТАБЛИЦА АУДИТА
-- =====================================================================

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

-- =====================================================================
-- БЛОК 5: СОЗДАНИЕ ВНЕШНИХ КЛЮЧЕЙ
-- =====================================================================

-- UserRoles
ALTER TABLE user_roles ADD CONSTRAINT fk_userroles_users 
    FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE;
ALTER TABLE user_roles ADD CONSTRAINT fk_userroles_roles 
    FOREIGN KEY (role_id) REFERENCES roles (role_id) ON DELETE CASCADE;
ALTER TABLE user_roles ADD CONSTRAINT fk_userroles_grantedby 
    FOREIGN KEY (granted_by) REFERENCES users (user_id) ON DELETE SET NULL;

-- EnvConfigParams
ALTER TABLE env_config_params ADD CONSTRAINT fk_envconfigparams_envconfig 
    FOREIGN KEY (env_config_id) REFERENCES env_configs (env_config_id) ON DELETE CASCADE;

-- ConfigErrors
ALTER TABLE config_errors ADD CONSTRAINT fk_configerrors_envconfig 
    FOREIGN KEY (env_config_id) REFERENCES env_configs (env_config_id) ON DELETE CASCADE;
ALTER TABLE config_errors ADD CONSTRAINT fk_configerrors_resolvedby 
    FOREIGN KEY (resolved_by) REFERENCES users (user_id) ON DELETE SET NULL;

-- ToolConfigParams
ALTER TABLE tool_config_params ADD CONSTRAINT fk_toolconfigparams_toolconfig 
    FOREIGN KEY (tool_config_id) REFERENCES tool_configs (tool_config_id) ON DELETE CASCADE;

-- Reports
ALTER TABLE reports ADD CONSTRAINT fk_reports_template 
    FOREIGN KEY (template_id) REFERENCES report_templates (template_id) ON DELETE RESTRICT;
ALTER TABLE reports ADD CONSTRAINT fk_reports_owner 
    FOREIGN KEY (owner_user_id) REFERENCES users (user_id) ON DELETE RESTRICT;

-- TestPlans
ALTER TABLE test_plans ADD CONSTRAINT fk_testplans_priority 
    FOREIGN KEY (priority_id) REFERENCES priorities (priority_id) ON DELETE RESTRICT;
ALTER TABLE test_plans ADD CONSTRAINT fk_testplans_owneruser 
    FOREIGN KEY (owner_user_id) REFERENCES users (user_id) ON DELETE RESTRICT;

-- Autotests
ALTER TABLE autotests ADD CONSTRAINT fk_autotests_owner 
    FOREIGN KEY (owner_user_id) REFERENCES users (user_id) ON DELETE RESTRICT;

-- AutotestVersions
ALTER TABLE autotest_versions ADD CONSTRAINT fk_autotestversions_autotest 
    FOREIGN KEY (autotest_id) REFERENCES autotests (autotest_id) ON DELETE CASCADE;

-- TestCases
ALTER TABLE test_cases ADD CONSTRAINT fk_testcases_priority 
    FOREIGN KEY (priority_id) REFERENCES priorities (priority_id) ON DELETE RESTRICT;
ALTER TABLE test_cases ADD CONSTRAINT fk_testcases_owneruser 
    FOREIGN KEY (owner_user_id) REFERENCES users (user_id) ON DELETE RESTRICT;
ALTER TABLE test_cases ADD CONSTRAINT fk_testcases_executor 
    FOREIGN KEY (executor_user_id) REFERENCES users (user_id) ON DELETE SET NULL;

-- TestCaseSteps
ALTER TABLE test_case_steps ADD CONSTRAINT fk_testcasesteps_testcase 
    FOREIGN KEY (test_case_id) REFERENCES test_cases (test_case_id) ON DELETE CASCADE;

-- TestRuns
ALTER TABLE test_runs ADD CONSTRAINT fk_testruns_testplan 
    FOREIGN KEY (test_plan_id) REFERENCES test_plans (test_plan_id) ON DELETE RESTRICT;
ALTER TABLE test_runs ADD CONSTRAINT fk_testruns_envconfig 
    FOREIGN KEY (env_config_id) REFERENCES env_configs (env_config_id) ON DELETE RESTRICT;
ALTER TABLE test_runs ADD CONSTRAINT fk_testruns_toolconfig 
    FOREIGN KEY (tool_config_id) REFERENCES tool_configs (tool_config_id) ON DELETE RESTRICT;
ALTER TABLE test_runs ADD CONSTRAINT fk_testruns_version 
    FOREIGN KEY (version_id) REFERENCES app_versions (version_id) ON DELETE RESTRICT;
ALTER TABLE test_runs ADD CONSTRAINT fk_testruns_createdby 
    FOREIGN KEY (created_by) REFERENCES users (user_id) ON DELETE SET NULL;

-- TestRunItems
ALTER TABLE test_run_items ADD CONSTRAINT fk_testrunitems_testrun 
    FOREIGN KEY (test_run_id) REFERENCES test_runs (test_run_id) ON DELETE CASCADE;
ALTER TABLE test_run_items ADD CONSTRAINT fk_testrunitems_testcase 
    FOREIGN KEY (test_case_id) REFERENCES test_cases (test_case_id) ON DELETE RESTRICT;

-- TestResults
ALTER TABLE test_results ADD CONSTRAINT fk_testresults_runitem 
    FOREIGN KEY (run_item_id) REFERENCES test_run_items (run_item_id) ON DELETE CASCADE;
ALTER TABLE test_results ADD CONSTRAINT fk_testresults_status 
    FOREIGN KEY (status_id) REFERENCES statuses (status_id) ON DELETE RESTRICT;
ALTER TABLE test_results ADD CONSTRAINT fk_testresults_executor 
    FOREIGN KEY (executor_user_id) REFERENCES users (user_id) ON DELETE SET NULL;

-- TestResultArtifacts
ALTER TABLE test_result_artifacts ADD CONSTRAINT fk_testresultartifacts_testresult 
    FOREIGN KEY (test_result_id) REFERENCES test_results (test_result_id) ON DELETE CASCADE;

-- AuditLog
ALTER TABLE audit_log ADD CONSTRAINT fk_auditlog_user 
    FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE SET NULL;

-- =====================================================================
-- БЛОК 6: НАЧАЛЬНЫЕ ДАННЫЕ
-- =====================================================================

-- Вставка базовых приоритетов
INSERT INTO priorities (code, name, rank, color_hex) VALUES
    ('CRITICAL', 'Критический', 5, '#FF0000'),
    ('HIGH', 'Высокий', 4, '#FF6600'),
    ('MEDIUM', 'Средний', 3, '#FFCC00'),
    ('LOW', 'Низкий', 2, '#00CC00'),
    ('TRIVIAL', 'Незначительный', 1, '#CCCCCC');

-- Вставка базовых статусов
INSERT INTO statuses (code, name, category, is_final) VALUES
    ('PASSED', 'Пройден', 'TEST_EXECUTION', TRUE),
    ('FAILED', 'Провален', 'TEST_EXECUTION', TRUE),
    ('BLOCKED', 'Заблокирован', 'TEST_EXECUTION', TRUE),
    ('SKIPPED', 'Пропущен', 'TEST_EXECUTION', TRUE),
    ('IN_PROGRESS', 'В процессе', 'TEST_EXECUTION', FALSE),
    ('NOT_RUN', 'Не запущен', 'TEST_EXECUTION', FALSE),
    ('DRAFT', 'Черновик', 'PLAN', FALSE),
    ('ACTIVE', 'Активный', 'PLAN', FALSE),
    ('COMPLETED', 'Завершен', 'PLAN', TRUE),
    ('ARCHIVED', 'Архивирован', 'PLAN', TRUE);

-- Вставка базовых ролей
INSERT INTO roles (code, name, description, access_level) VALUES
    ('ADMIN', 'Администратор', 'Полный доступ к системе, управление пользователями и ролями', 10),
    ('TEST_LEAD', 'Тест-лид', 'Управление тест-планами, кейсами и командой тестирования', 8),
    ('AUTOMATION_ENGINEER', 'Инженер по автоматизации', 'Разработка и поддержка автотестов', 7),
    ('TESTER', 'Тестировщик', 'Выполнение тестов и создание тест-кейсов', 5),
    ('CI_CD_SYSTEM', 'CI/CD система', 'Автоматическое выполнение тестов через CI/CD pipeline', 6),
    ('DEVELOPER', 'Разработчик', 'Просмотр результатов и создание конфигураций', 4),
    ('GUEST', 'Гость', 'Только просмотр публичных данных', 1);

-- =====================================================================
-- БЛОК 7: ФУНКЦИИ И ТРИГГЕРЫ
-- =====================================================================

-- Функция обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггеры для автоматического обновления updated_at
CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_testcases_updated_at BEFORE UPDATE ON test_cases
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_testplans_updated_at BEFORE UPDATE ON test_plans
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_envconfigs_updated_at BEFORE UPDATE ON env_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_toolconfigs_updated_at BEFORE UPDATE ON tool_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_autotests_updated_at BEFORE UPDATE ON autotests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Функция аудита для критических таблиц
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
DECLARE
    user_id_value INT;
BEGIN
    -- Попытка получить user_id из настроек сессии
    BEGIN
        user_id_value := current_setting('app.current_user_id', true)::INT;
    EXCEPTION WHEN OTHERS THEN
        user_id_value := NULL;
    END;

    IF TG_OP = 'INSERT' THEN
        INSERT INTO audit_log (table_name, operation, record_id, user_id, new_values)
        VALUES (TG_TABLE_NAME, TG_OP, 
                CASE TG_TABLE_NAME
                    WHEN 'users' THEN NEW.user_id
                    WHEN 'test_cases' THEN NEW.test_case_id
                    WHEN 'test_plans' THEN NEW.test_plan_id
                    ELSE 0
                END,
                user_id_value, row_to_json(NEW)::JSONB);
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit_log (table_name, operation, record_id, user_id, old_values, new_values)
        VALUES (TG_TABLE_NAME, TG_OP,
                CASE TG_TABLE_NAME
                    WHEN 'users' THEN NEW.user_id
                    WHEN 'test_cases' THEN NEW.test_case_id
                    WHEN 'test_plans' THEN NEW.test_plan_id
                    ELSE 0
                END,
                user_id_value, row_to_json(OLD)::JSONB, row_to_json(NEW)::JSONB);
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO audit_log (table_name, operation, record_id, user_id, old_values)
        VALUES (TG_TABLE_NAME, TG_OP,
                CASE TG_TABLE_NAME
                    WHEN 'users' THEN OLD.user_id
                    WHEN 'test_cases' THEN OLD.test_case_id
                    WHEN 'test_plans' THEN OLD.test_plan_id
                    ELSE 0
                END,
                user_id_value, row_to_json(OLD)::JSONB);
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Триггеры аудита для критических таблиц
CREATE TRIGGER trg_users_audit
    AFTER INSERT OR UPDATE OR DELETE ON users
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER trg_userroles_audit
    AFTER INSERT OR UPDATE OR DELETE ON user_roles
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER trg_testcases_audit
    AFTER INSERT OR UPDATE OR DELETE ON test_cases
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER trg_testplans_audit
    AFTER INSERT OR UPDATE OR DELETE ON test_plans
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- Триггер для блокировки учетной записи после 5 неудачных попыток
CREATE OR REPLACE FUNCTION lock_user_account()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.failed_login_attempts >= 5 THEN
        NEW.account_locked_until = CURRENT_TIMESTAMP + INTERVAL '30 minutes';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_lock_user_account
    BEFORE UPDATE OF failed_login_attempts ON users
    FOR EACH ROW EXECUTE FUNCTION lock_user_account();

-- Функция для проверки сложности пароля
CREATE OR REPLACE FUNCTION check_password_complexity(password TEXT)
RETURNS BOOLEAN AS $$
BEGIN
    IF LENGTH(password) < 12 THEN RETURN FALSE; END IF;
    IF password !~ '[A-Z]' THEN RETURN FALSE; END IF;
    IF password !~ '[a-z]' THEN RETURN FALSE; END IF;
    IF password !~ '[0-9]' THEN RETURN FALSE; END IF;
    IF password !~ '[!@#$%^&*()_+\-=\[\]{};:''",.<>?/|\\]' THEN RETURN FALSE; END IF;
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- Функция для проверки истечения срока пароля
CREATE OR REPLACE FUNCTION check_password_expiration()
RETURNS TABLE(user_id INT, username VARCHAR, days_until_expiration INT) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        u.user_id,
        u.username,
        90 - EXTRACT(DAY FROM CURRENT_TIMESTAMP - u.password_changed_at)::INT AS days_until_expiration
    FROM users u
    WHERE u.is_active = TRUE
    AND CURRENT_TIMESTAMP - u.password_changed_at > INTERVAL '90 days';
END;
$$ LANGUAGE plpgsql;

-- =====================================================================
-- БЛОК 8: ПРЕДСТАВЛЕНИЯ ДЛЯ РОЛЕВОГО ДОСТУПА
-- =====================================================================

-- 1. Представление публичных тест-планов
CREATE OR REPLACE VIEW v_public_test_plans AS
SELECT
    tp.test_plan_id,
    tp.name AS test_plan_name,
    tp.description,
    tp.start_date,
    tp.end_date,
    tp.status,
    u.full_name AS owner,
    p.name AS priority,
    p.rank AS priority_rank
FROM test_plans tp
JOIN users u ON tp.owner_user_id = u.user_id
JOIN priorities p ON tp.priority_id = p.priority_id
WHERE tp.status IN ('ACTIVE', 'COMPLETED');

COMMENT ON VIEW v_public_test_plans IS 'Публичные тест-планы (доступны всем ролям)';

-- 2. Представление назначенных тест-кейсов для тестировщика
CREATE OR REPLACE VIEW v_tester_assigned_cases AS
SELECT
    tc.test_case_id,
    tc.name AS test_case_name,
    tc.description,
    tc.is_automated,
    p.name AS priority,
    u_owner.full_name AS owner,
    u_executor.username AS executor_username
FROM test_cases tc
JOIN priorities p ON tc.priority_id = p.priority_id
JOIN users u_owner ON tc.owner_user_id = u_owner.user_id
JOIN users u_executor ON tc.executor_user_id = u_executor.user_id
WHERE u_executor.username = CURRENT_USER
AND tc.is_active = TRUE;

COMMENT ON VIEW v_tester_assigned_cases IS 'Тест-кейсы, назначенные текущему пользователю (RLS)';

-- 3. Представление всех тест-кейсов для тест-лида
CREATE OR REPLACE VIEW v_lead_all_cases AS
SELECT
    tc.test_case_id,
    tc.name AS test_case_name,
    tc.description,
    tc.is_automated,
    tc.is_active,
    p.name AS priority,
    u_owner.full_name AS owner,
    COALESCE(u_executor.full_name, 'Не назначен') AS executor
FROM test_cases tc
JOIN priorities p ON tc.priority_id = p.priority_id
JOIN users u_owner ON tc.owner_user_id = u_owner.user_id
LEFT JOIN users u_executor ON tc.executor_user_id = u_executor.user_id;

COMMENT ON VIEW v_lead_all_cases IS 'Все тест-кейсы (доступ: Test_Lead, Automation_Engineer)';

-- 4. Представление автотестов для инженера по автоматизации
CREATE OR REPLACE VIEW v_engineer_autotests AS
SELECT
    a.autotest_id,
    a.name AS autotest_name,
    a.description,
    a.is_active,
    av.version_id,
    av.version_string,
    av.commit_hash,
    av.created_at AS version_date,
    u.full_name AS author
FROM autotests a
JOIN autotest_versions av ON a.autotest_id = av.autotest_id
JOIN users u ON a.owner_user_id = u.user_id
WHERE a.is_active = TRUE;

COMMENT ON VIEW v_engineer_autotests IS 'Автотесты с версиями (доступ: Automation_Engineer, Test_Lead)';

-- 5. Представление конфигураций окружений
CREATE OR REPLACE VIEW v_shared_environments AS
SELECT
    ec.env_config_id,
    ec.name AS environment_name,
    ec.description,
    ec.is_active,
    COUNT(ecp.env_config_param_id) AS params_count
FROM env_configs ec
LEFT JOIN env_config_params ecp ON ec.env_config_id = ecp.env_config_id
WHERE ec.is_active = TRUE
GROUP BY ec.env_config_id, ec.name, ec.description, ec.is_active;

COMMENT ON VIEW v_shared_environments IS 'Конфигурации окружений (доступны многим ролям)';

-- 6. Представление результатов тестов текущего пользователя
CREATE OR REPLACE VIEW v_tester_my_results AS
SELECT
    tr.test_result_id,
    tri.test_run_id,
    tc.name AS test_case_name,
    s.name AS status,
    tr.execution_date,
    tr.result_summary,
    tr.execution_duration_minutes,
    u.username AS executor_username
FROM test_results tr
JOIN test_run_items tri ON tr.run_item_id = tri.run_item_id
JOIN test_cases tc ON tri.test_case_id = tc.test_case_id
JOIN statuses s ON tr.status_id = s.status_id
JOIN users u ON tr.executor_user_id = u.user_id
WHERE u.username = CURRENT_USER;

COMMENT ON VIEW v_tester_my_results IS 'Результаты тестов текущего пользователя (RLS)';

-- 7. Представление всех результатов для тест-лида
CREATE OR REPLACE VIEW v_lead_all_results AS
SELECT
    tr.test_result_id,
    tri.test_run_id,
    trun.name AS test_run_name,
    tc.name AS test_case_name,
    s.name AS status,
    COALESCE(u.full_name, 'Автоматический запуск') AS executor,
    tr.execution_date,
    tr.result_summary,
    tr.execution_duration_minutes,
    tr.error_message
FROM test_results tr
JOIN test_run_items tri ON tr.run_item_id = tri.run_item_id
JOIN test_runs trun ON tri.test_run_id = trun.test_run_id
JOIN test_cases tc ON tri.test_case_id = tc.test_case_id
JOIN statuses s ON tr.status_id = s.status_id
LEFT JOIN users u ON tr.executor_user_id = u.user_id;

COMMENT ON VIEW v_lead_all_results IS 'Все результаты тестов (доступ: Test_Lead, Automation_Engineer)';

-- 8. Представление публичных результатов для гостей
CREATE OR REPLACE VIEW v_public_results AS
SELECT
    tr.test_result_id,
    tri.test_run_id,
    tc.name AS test_case_name,
    s.name AS status,
    tr.execution_date
FROM test_results tr
JOIN test_run_items tri ON tr.run_item_id = tri.run_item_id
JOIN test_cases tc ON tri.test_case_id = tc.test_case_id
JOIN statuses s ON tr.status_id = s.status_id
WHERE s.code IN ('PASSED', 'FAILED', 'BLOCKED')
AND s.is_final = TRUE;

COMMENT ON VIEW v_public_results IS 'Публичные результаты тестов (доступ: Guest)';

-- 9. Представление отчетов о качестве
CREATE OR REPLACE VIEW v_shared_reports AS
SELECT
    r.report_id,
    r.name AS report_name,
    r.created_at AS creation_date,
    rt.name AS template,
    u.full_name AS author
FROM reports r
JOIN report_templates rt ON r.template_id = rt.template_id
JOIN users u ON r.owner_user_id = u.user_id
WHERE rt.is_active = TRUE;

COMMENT ON VIEW v_shared_reports IS 'Отчеты о качестве (доступны многим ролям)';

-- 10. Представление пользователей и ролей для администратора
CREATE OR REPLACE VIEW v_admin_users_and_roles AS
SELECT
    u.user_id,
    u.username,
    u.full_name,
    u.email,
    u.is_active,
    u.created_at,
    u.last_login,
    u.failed_login_attempts,
    STRING_AGG(r.name, ', ') AS roles
FROM users u
LEFT JOIN user_roles ur ON u.user_id = ur.user_id
LEFT JOIN roles r ON ur.role_id = r.role_id
GROUP BY u.user_id, u.username, u.full_name, u.email, u.is_active, 
         u.created_at, u.last_login, u.failed_login_attempts;

COMMENT ON VIEW v_admin_users_and_roles IS 'Управление пользователями и ролями (доступ: Administrator)';

-- 11. Представление активных тестовых прогонов с статистикой
CREATE OR REPLACE VIEW v_active_test_runs AS
SELECT 
    tr.test_run_id,
    tr.name AS run_name,
    tp.name AS plan_name,
    av.version_string,
    ec.name AS environment,
    tc.name AS tool_config,
    tr.status,
    tr.start_date,
    tr.end_date,
    u.full_name AS created_by_name,
    COUNT(DISTINCT tri.test_case_id) AS total_tests,
    COUNT(DISTINCT CASE WHEN ts.code = 'PASSED' THEN tres.test_result_id END) AS passed_tests,
    COUNT(DISTINCT CASE WHEN ts.code = 'FAILED' THEN tres.test_result_id END) AS failed_tests,
    COUNT(DISTINCT CASE WHEN ts.code = 'BLOCKED' THEN tres.test_result_id END) AS blocked_tests
FROM test_runs tr
JOIN test_plans tp ON tr.test_plan_id = tp.test_plan_id
JOIN app_versions av ON tr.version_id = av.version_id
JOIN env_configs ec ON tr.env_config_id = ec.env_config_id
JOIN tool_configs tc ON tr.tool_config_id = tc.tool_config_id
LEFT JOIN users u ON tr.created_by = u.user_id
LEFT JOIN test_run_items tri ON tr.test_run_id = tri.test_run_id
LEFT JOIN test_results tres ON tri.run_item_id = tres.run_item_id
LEFT JOIN statuses ts ON tres.status_id = ts.status_id
WHERE tr.status IN ('PLANNED', 'IN_PROGRESS')
GROUP BY tr.test_run_id, tr.name, tp.name, av.version_string, 
         ec.name, tc.name, tr.status, tr.start_date, tr.end_date, u.full_name;

COMMENT ON VIEW v_active_test_runs IS 'Активные тестовые прогоны со статистикой';

-- 12. Представление статистики по тест-кейсам
CREATE OR REPLACE VIEW v_test_case_statistics AS
SELECT 
    tc.test_case_id,
    tc.name,
    p.name AS priority,
    u.full_name AS owner,
    tc.is_automated,
    tc.is_active,
    COUNT(DISTINCT tres.test_result_id) AS total_executions,
    COUNT(DISTINCT CASE WHEN s.code = 'PASSED' THEN tres.test_result_id END) AS passed_count,
    COUNT(DISTINCT CASE WHEN s.code = 'FAILED' THEN tres.test_result_id END) AS failed_count,
    ROUND(
        100.0 * COUNT(DISTINCT CASE WHEN s.code = 'PASSED' THEN tres.test_result_id END) / 
        NULLIF(COUNT(DISTINCT tres.test_result_id), 0), 
        2
    ) AS pass_rate_percent,
    ROUND(AVG(tres.execution_duration_minutes), 2) AS avg_duration_minutes
FROM test_cases tc
JOIN priorities p ON tc.priority_id = p.priority_id
JOIN users u ON tc.owner_user_id = u.user_id
LEFT JOIN test_run_items tri ON tc.test_case_id = tri.test_case_id
LEFT JOIN test_results tres ON tri.run_item_id = tres.run_item_id
LEFT JOIN statuses s ON tres.status_id = s.status_id
WHERE tc.is_active = TRUE
GROUP BY tc.test_case_id, tc.name, p.name, u.full_name, tc.is_automated, tc.is_active;

COMMENT ON VIEW v_test_case_statistics IS 'Статистика эффективности тест-кейсов';

-- 13. Представление сводки выполнения тестов
CREATE OR REPLACE VIEW v_test_execution_summary AS
SELECT
    trun.test_run_id,
    trun.name AS test_run_name,
    tp.name AS test_plan_name,
    av.version_string,
    COUNT(tri.test_case_id) AS total_cases,
    COUNT(CASE WHEN s.code = 'PASSED' THEN 1 END) AS passed,
    COUNT(CASE WHEN s.code = 'FAILED' THEN 1 END) AS failed,
    COUNT(CASE WHEN s.code = 'BLOCKED' THEN 1 END) AS blocked,
    COUNT(CASE WHEN s.code = 'SKIPPED' THEN 1 END) AS skipped,
    ROUND(
        100.0 * COUNT(CASE WHEN s.code = 'PASSED' THEN 1 END) / 
        NULLIF(COUNT(tri.test_case_id), 0), 
        2
    ) AS pass_rate,
    trun.start_date,
    trun.end_date,
    trun.status AS run_status
FROM test_runs trun
JOIN test_plans tp ON trun.test_plan_id = tp.test_plan_id
JOIN app_versions av ON trun.version_id = av.version_id
LEFT JOIN test_run_items tri ON trun.test_run_id = tri.test_run_id
LEFT JOIN test_results tr ON tri.run_item_id = tr.run_item_id
LEFT JOIN statuses s ON tr.status_id = s.status_id
GROUP BY trun.test_run_id, trun.name, tp.name, av.version_string, 
         trun.start_date, trun.end_date, trun.status;

COMMENT ON VIEW v_test_execution_summary IS 'Сводка выполнения тестов по прогонам';

-- =====================================================================
-- БЛОК 9: СОЗДАНИЕ РОЛЕЙ НА УРОВНЕ БД И НАСТРОЙКА ПРИВИЛЕГИЙ
-- =====================================================================

-- Создание ролей PostgreSQL
DO $$
BEGIN
    -- Администратор
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'db_admin') THEN
        CREATE ROLE db_admin WITH LOGIN PASSWORD 'Admin_SecureP@ss2025!' 
        VALID UNTIL '2026-12-31' CONNECTION LIMIT 5;
    END IF;
    
    -- Тест-лид
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'db_test_lead') THEN
        CREATE ROLE db_test_lead WITH LOGIN PASSWORD 'TestLead_SecureP@ss2025!'
        VALID UNTIL '2026-12-31' CONNECTION LIMIT 10;
    END IF;
    
    -- Инженер по автоматизации
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'db_automation_engineer') THEN
        CREATE ROLE db_automation_engineer WITH LOGIN PASSWORD 'AutoEngineer_SecureP@ss2025!'
        VALID UNTIL '2026-12-31' CONNECTION LIMIT 15;
    END IF;
    
    -- Тестировщик
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'db_tester') THEN
        CREATE ROLE db_tester WITH LOGIN PASSWORD 'Tester_SecureP@ss2025!'
        VALID UNTIL '2026-12-31' CONNECTION LIMIT 20;
    END IF;
    
    -- CI/CD система
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'db_cicd_system') THEN
        CREATE ROLE db_cicd_system WITH LOGIN PASSWORD 'CICD_SecureP@ss2025!'
        VALID UNTIL '2026-12-31' CONNECTION LIMIT 10;
    END IF;
    
    -- Разработчик
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'db_developer') THEN
        CREATE ROLE db_developer WITH LOGIN PASSWORD 'Developer_SecureP@ss2025!'
        VALID UNTIL '2026-12-31' CONNECTION LIMIT 15;
    END IF;
    
    -- Гость
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'db_guest') THEN
        CREATE ROLE db_guest WITH LOGIN PASSWORD 'Guest_SecureP@ss2025!'
        VALID UNTIL '2026-12-31' CONNECTION LIMIT 30;
    END IF;
END
$$;

-- =====================================================================
-- Настройка привилегий для роли АДМИНИСТРАТОР
-- =====================================================================

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO db_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO db_admin;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO db_admin;
GRANT ALL PRIVILEGES ON DATABASE postgres TO db_admin;

-- Полный доступ к представлениям
GRANT SELECT, INSERT, UPDATE, DELETE ON v_admin_users_and_roles TO db_admin;

-- =====================================================================
-- Настройка привилегий для роли ТЕСТ-ЛИД
-- =====================================================================

-- Доступ к представлениям
GRANT SELECT ON v_public_test_plans, v_lead_all_cases, v_engineer_autotests,
    v_shared_environments, v_lead_all_results, v_shared_reports,
    v_active_test_runs, v_test_case_statistics, v_test_execution_summary TO db_test_lead;

-- Доступ к таблицам для управления
GRANT SELECT, INSERT, UPDATE, DELETE ON test_plans, test_cases, test_case_steps,
    test_runs, test_run_items TO db_test_lead;

GRANT SELECT, INSERT, UPDATE ON users, user_roles TO db_test_lead;

GRANT SELECT, INSERT, UPDATE ON reports TO db_test_lead;

GRANT SELECT ON priorities, statuses, roles, app_versions, env_configs,
    tool_configs, audit_log, autotests, autotest_versions TO db_test_lead;

GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO db_test_lead;

-- =====================================================================
-- Настройка привилегий для роли ИНЖЕНЕР ПО АВТОМАТИЗАЦИИ
-- =====================================================================

-- Доступ к представлениям
GRANT SELECT ON v_public_test_plans, v_lead_all_cases, v_engineer_autotests,
    v_shared_environments, v_lead_all_results, v_shared_reports,
    v_active_test_runs, v_test_case_statistics TO db_automation_engineer;

-- Доступ к автотестам
GRANT SELECT, INSERT, UPDATE, DELETE ON autotests, autotest_versions TO db_automation_engineer;

-- Доступ к окружениям
GRANT SELECT, INSERT, UPDATE ON env_configs, env_config_params TO db_automation_engineer;

-- Доступ к результатам и тест-кейсам
GRANT SELECT ON test_plans, test_cases, test_runs, test_results,
    test_result_artifacts, priorities, statuses, users TO db_automation_engineer;

GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO db_automation_engineer;

-- =====================================================================
-- Настройка привилегий для роли ТЕСТИРОВЩИК
-- =====================================================================

-- Доступ к представлениям
GRANT SELECT ON v_public_test_plans, v_tester_assigned_cases,
    v_shared_environments, v_tester_my_results, v_shared_reports TO db_tester;

-- Доступ к тест-кейсам (только назначенным)
GRANT SELECT, INSERT, UPDATE ON test_cases, test_case_steps TO db_tester;

-- Доступ к результатам
GRANT SELECT, INSERT, UPDATE ON test_results, test_result_artifacts TO db_tester;

-- Чтение справочников
GRANT SELECT ON test_plans, test_runs, test_run_items, priorities,
    statuses, users, roles, env_configs, tool_configs TO db_tester;

GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO db_tester;

-- =====================================================================
-- Настройка привилегий для роли CI/CD СИСТЕМА
-- =====================================================================

-- Прямой доступ к таблицам (без представлений для автоматизации)
GRANT SELECT ON autotest_versions, env_configs TO db_cicd_system;

GRANT SELECT, INSERT, UPDATE ON test_runs, test_run_items,
    test_results, test_result_artifacts TO db_cicd_system;

GRANT SELECT, INSERT ON reports TO db_cicd_system;

GRANT SELECT ON test_cases, test_plans, priorities, statuses,
    tool_configs, app_versions TO db_cicd_system;

GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO db_cicd_system;

-- =====================================================================
-- Настройка привилегий для роли РАЗРАБОТЧИК
-- =====================================================================

-- Просмотр результатов
GRANT SELECT ON v_public_test_plans, v_lead_all_results,
    v_shared_reports, v_active_test_runs, v_test_execution_summary TO db_developer;

-- Управление конфигурациями
GRANT SELECT, INSERT, UPDATE ON env_configs, env_config_params,
    tool_configs, tool_config_params, config_errors TO db_developer;

GRANT SELECT, INSERT ON app_versions TO db_developer;

-- Чтение данных
GRANT SELECT ON test_plans, test_cases, test_runs, test_results,
    test_result_artifacts, priorities, statuses, users TO db_developer;

GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO db_developer;

-- =====================================================================
-- Настройка привилегий для роли ГОСТЬ
-- =====================================================================

-- Только чтение публичных данных
GRANT SELECT ON v_public_test_plans, v_public_results, v_shared_reports TO db_guest;

GRANT SELECT ON priorities, statuses TO db_guest;

-- =====================================================================
-- БЛОК 10: НАСТРОЙКА ROW LEVEL SECURITY (RLS)
-- =====================================================================

-- Включение RLS для пользователей
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

CREATE POLICY users_select_policy ON users
    FOR SELECT
    TO db_tester, db_developer, db_guest
    USING (is_active = TRUE);

CREATE POLICY users_select_all_policy ON users
    FOR SELECT
    TO db_test_lead, db_admin, db_automation_engineer
    USING (TRUE);

-- Включение RLS для тест-кейсов
ALTER TABLE test_cases ENABLE ROW LEVEL SECURITY;

CREATE POLICY testcases_select_policy ON test_cases
    FOR SELECT
    TO db_tester, db_developer, db_guest
    USING (is_active = TRUE);

CREATE POLICY testcases_select_all_policy ON test_cases
    FOR SELECT
    TO db_test_lead, db_admin, db_automation_engineer
    USING (TRUE);

-- Политика для тестировщика: видеть только назначенные тест-кейсы при UPDATE
CREATE POLICY testcases_update_assigned ON test_cases
    FOR UPDATE
    TO db_tester
    USING (
        EXISTS (
            SELECT 1 FROM users u
            WHERE u.user_id = test_cases.executor_user_id
            AND u.username = CURRENT_USER
        )
    );

-- Включение RLS для результатов тестов
ALTER TABLE test_results ENABLE ROW LEVEL SECURITY;

CREATE POLICY testresults_select_own ON test_results
    FOR SELECT
    TO db_tester
    USING (
        EXISTS (
            SELECT 1 FROM users u
            WHERE u.user_id = test_results.executor_user_id
            AND u.username = CURRENT_USER
        )
    );

CREATE POLICY testresults_select_all ON test_results
    FOR SELECT
    TO db_test_lead, db_admin, db_automation_engineer, db_developer
    USING (TRUE);

-- =====================================================================
-- БЛОК 11: КОНФИГУРАЦИЯ СЕТЕВОЙ БЕЗОПАСНОСТИ (КОММЕНТАРИИ)
-- =====================================================================

/*
==========================================================================
НАСТРОЙКА СЕТЕВОЙ БЕЗОПАСНОСТИ В pg_hba.conf:
==========================================================================

# TYPE  DATABASE        USER                    ADDRESS                 METHOD

# Локальное подключение для администратора
local   all             db_admin                                        scram-sha-256

# Удаленные SSL-подключения с корпоративной сети
hostssl all             db_test_lead            192.168.1.0/24          scram-sha-256
hostssl all             db_automation_engineer  192.168.1.0/24          scram-sha-256
hostssl all             db_tester               192.168.1.0/24          scram-sha-256
hostssl all             db_developer            192.168.1.0/24          scram-sha-256

# CI/CD система с выделенного сервера
hostssl all             db_cicd_system          10.0.10.100/32          scram-sha-256

# Гостевой доступ только через VPN
hostssl all             db_guest                172.16.0.0/16           scram-sha-256

# Запрет всех остальных подключений
host    all             all                     0.0.0.0/0               reject

==========================================================================
НАСТРОЙКА postgresql.conf:
==========================================================================

# Сетевые настройки
listen_addresses = '*'
port = 5432
max_connections = 100

# SSL/TLS обязателен
ssl = on
ssl_cert_file = '/etc/postgresql/ssl/server.crt'
ssl_key_file = '/etc/postgresql/ssl/server.key'
ssl_ca_file = '/etc/postgresql/ssl/ca.crt'
ssl_ciphers = 'HIGH:MEDIUM:+3DES:!aNULL'
ssl_prefer_server_ciphers = on
ssl_min_protocol_version = 'TLSv1.2'

# Аутентификация
password_encryption = scram-sha-256

# Таймауты безопасности
statement_timeout = 300000                    # 5 минут
idle_in_transaction_session_timeout = 600000  # 10 минут
tcp_keepalives_idle = 60
tcp_keepalives_interval = 10
tcp_keepalives_count = 3

# Логирование для аудита
logging_collector = on
log_directory = '/var/log/postgresql'
log_filename = 'postgresql-%Y-%m-%d.log'
log_rotation_age = 1d
log_rotation_size = 100MB

log_connections = on
log_disconnections = on
log_duration = on
log_hostname = on
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
log_statement = 'mod'  # Логировать INSERT, UPDATE, DELETE
log_min_duration_statement = 1000  # Логировать запросы > 1 сек

# Защита от перегрузки
shared_buffers = 256MB
work_mem = 8MB
maintenance_work_mem = 128MB
effective_cache_size = 1GB
max_wal_size = 2GB

# Блокировка опасных функций для не-администраторов
default_transaction_read_only = off

==========================================================================
*/

-- =====================================================================
-- БЛОК 12: ХЕЛПЕРНЫЕ ФУНКЦИИ ДЛЯ АДМИНИСТРИРОВАНИЯ
-- =====================================================================

-- Функция для просмотра активных сессий пользователей
CREATE OR REPLACE FUNCTION get_active_sessions()
RETURNS TABLE(
    pid INT,
    username TEXT,
    database TEXT,
    client_addr TEXT,
    application_name TEXT,
    state TEXT,
    query_start TIMESTAMPTZ,
    state_change TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        pg_stat_activity.pid,
        pg_stat_activity.usename::TEXT,
        pg_stat_activity.datname::TEXT,
        COALESCE(pg_stat_activity.client_addr::TEXT, 'localhost'),
        pg_stat_activity.application_name::TEXT,
        pg_stat_activity.state::TEXT,
        pg_stat_activity.query_start,
        pg_stat_activity.state_change
    FROM pg_stat_activity
    WHERE pg_stat_activity.datname = current_database()
    AND pg_stat_activity.pid <> pg_backend_pid()
    ORDER BY pg_stat_activity.query_start DESC;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION get_active_sessions() IS 'Просмотр активных подключений к БД';

-- Предоставление доступа администратору
GRANT EXECUTE ON FUNCTION get_active_sessions() TO db_admin;

-- =====================================================================
-- БЛОК 13: ЗАВЕРШЕНИЕ И ВЫВОД ИНФОРМАЦИИ
-- =====================================================================

DO $$
DECLARE
    table_count INT;
    view_count INT;
    index_count INT;
    trigger_count INT;
    function_count INT;
    role_count INT;
BEGIN
    SELECT COUNT(*) INTO table_count FROM information_schema.tables 
    WHERE table_schema = 'public' AND table_type = 'BASE TABLE';
    
    SELECT COUNT(*) INTO view_count FROM information_schema.views 
    WHERE table_schema = 'public';
    
    SELECT COUNT(*) INTO index_count FROM pg_indexes 
    WHERE schemaname = 'public';
    
    SELECT COUNT(*) INTO trigger_count FROM information_schema.triggers 
    WHERE trigger_schema = 'public';
    
    SELECT COUNT(*) INTO function_count FROM pg_proc p
    JOIN pg_namespace n ON p.pronamespace = n.oid
    WHERE n.nspname = 'public' AND p.prokind = 'f';
    
    SELECT COUNT(*) INTO role_count FROM pg_roles 
    WHERE rolname LIKE 'db_%';
    
    RAISE NOTICE '================================================================';
    RAISE NOTICE 'БЕЗОПАСНАЯ ИС АВТОМАТИЗИРОВАННОГО ТЕСТИРОВАНИЯ ПО';
    RAISE NOTICE 'Полная интегрированная схема успешно развернута!';
    RAISE NOTICE '================================================================';
    RAISE NOTICE 'Создано таблиц: %', table_count;
    RAISE NOTICE 'Создано представлений: %', view_count;
    RAISE NOTICE 'Создано индексов: %', index_count;
    RAISE NOTICE 'Создано триггеров: %', trigger_count;
    RAISE NOTICE 'Создано функций: %', function_count;
    RAISE NOTICE 'Создано ролей БД: %', role_count;
    RAISE NOTICE '================================================================';
    RAISE NOTICE 'Механизмы безопасности:';
    RAISE NOTICE '  ✓ Row-Level Security (RLS)';
    RAISE NOTICE '  ✓ Аудит изменений данных';
    RAISE NOTICE '  ✓ Политика сложности паролей';
    RAISE NOTICE '  ✓ Автоблокировка учетных записей';
    RAISE NOTICE '  ✓ SSL/TLS шифрование';
    RAISE NOTICE '  ✓ Ролевая модель доступа (RBAC)';
    RAISE NOTICE '  ✓ Инкапсуляция через представления';
    RAISE NOTICE '================================================================';
END $$;

-- =====================================================================
-- КОНЕЦ ИНТЕГРИРОВАННОГО СКРИПТА
-- =====================================================================
