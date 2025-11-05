-- +goose Up
-- +goose StatementBegin
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
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS test_result_artifacts;

DROP TABLE IF EXISTS test_results;

DROP TABLE IF EXISTS test_run_items;

DROP TABLE IF EXISTS test_runs;
-- +goose StatementEnd
