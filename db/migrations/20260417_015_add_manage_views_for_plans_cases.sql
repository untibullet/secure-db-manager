-- +goose Up
-- +goose StatementBegin

-- Плоское представление test_plans для записи (auto-updatable, AD-4)
CREATE OR REPLACE VIEW v_test_plans_manage AS
SELECT
    test_plan_id,
    name,
    description,
    priority_id,
    start_date,
    end_date,
    acceptance_criteria,
    owner_user_id,
    version_no,
    status,
    created_at,
    updated_at
FROM test_plans;

GRANT SELECT, INSERT, UPDATE, DELETE ON v_test_plans_manage TO db_test_lead;


-- Плоское представление test_cases для записи (auto-updatable, AD-4)
CREATE OR REPLACE VIEW v_test_cases_manage AS
SELECT
    test_case_id,
    priority_id,
    name,
    description,
    owner_user_id,
    is_automated,
    is_active,
    estimated_duration_minutes,
    created_at,
    updated_at
FROM test_cases;

GRANT SELECT, INSERT, UPDATE, DELETE ON v_test_cases_manage TO db_test_lead;
GRANT SELECT, INSERT, UPDATE ON v_test_cases_manage TO db_automation_engineer;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS v_test_plans_manage;
DROP VIEW IF EXISTS v_test_cases_manage;
-- +goose StatementEnd
