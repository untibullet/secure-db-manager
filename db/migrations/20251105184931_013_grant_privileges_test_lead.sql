-- +goose Up
-- +goose StatementBegin
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
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
REVOKE SELECT ON v_public_test_plans, v_lead_all_cases, v_engineer_autotests, v_shared_environments, v_lead_all_results, v_shared_reports, v_active_test_runs, v_test_case_statistics, v_test_execution_summary FROM db_test_lead;
REVOKE SELECT, INSERT, UPDATE, DELETE ON test_plans, test_cases, test_case_steps, test_runs, test_run_items FROM db_test_lead;
REVOKE SELECT, INSERT, UPDATE ON users, user_roles FROM db_test_lead;
REVOKE SELECT, INSERT, UPDATE ON reports FROM db_test_lead;
REVOKE SELECT ON priorities, statuses, roles, app_versions, env_configs, tool_configs, audit_log, autotests, autotest_versions FROM db_test_lead;
REVOKE USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public FROM db_test_lead;
-- +goose StatementEnd

