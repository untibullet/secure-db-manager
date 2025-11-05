-- +goose Up
-- +goose StatementBegin
-- User Roles
ALTER TABLE user_roles ADD CONSTRAINT fk_user_roles_to_users 
    FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE;
ALTER TABLE user_roles ADD CONSTRAINT fk_user_roles_to_roles 
    FOREIGN KEY (role_id) REFERENCES roles (role_id) ON DELETE CASCADE;
ALTER TABLE user_roles ADD CONSTRAINT fk_user_roles_granted_by 
    FOREIGN KEY (granted_by) REFERENCES users (user_id) ON DELETE SET NULL;

-- Audit Log
ALTER TABLE audit_log ADD CONSTRAINT fk_audit_log_to_users 
    FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE SET NULL;

-- Env Config Params
ALTER TABLE env_config_params ADD CONSTRAINT fk_env_config_params_to_env_configs 
    FOREIGN KEY (env_config_id) REFERENCES env_configs (env_config_id) ON DELETE CASCADE;

-- Config Errors
ALTER TABLE config_errors ADD CONSTRAINT fk_config_errors_to_env_configs 
    FOREIGN KEY (env_config_id) REFERENCES env_configs (env_config_id) ON DELETE CASCADE;

-- Tool Config Params
ALTER TABLE tool_config_params ADD CONSTRAINT fk_tool_config_params_to_tool_configs 
    FOREIGN KEY (tool_config_id) REFERENCES tool_configs (tool_config_id) ON DELETE CASCADE;

-- Reports
ALTER TABLE reports ADD CONSTRAINT fk_reports_to_report_templates 
    FOREIGN KEY (template_id) REFERENCES report_templates (template_id) ON DELETE RESTRICT;
ALTER TABLE reports ADD CONSTRAINT fk_reports_to_test_runs 
    FOREIGN KEY (test_run_id) REFERENCES test_runs (run_id) ON DELETE CASCADE;
ALTER TABLE reports ADD CONSTRAINT fk_reports_to_users 
    FOREIGN KEY (owner_user_id) REFERENCES users (user_id) ON DELETE SET NULL;

-- Test Plans
ALTER TABLE test_plans ADD CONSTRAINT fk_test_plans_to_app_versions 
    FOREIGN KEY (app_version_id) REFERENCES app_versions (version_id) ON DELETE RESTRICT;
ALTER TABLE test_plans ADD CONSTRAINT fk_test_plans_to_users_owner 
    FOREIGN KEY (owner_user_id) REFERENCES users (user_id) ON DELETE RESTRICT;
ALTER TABLE test_plans ADD CONSTRAINT fk_test_plans_to_statuses 
    FOREIGN KEY (status_id) REFERENCES statuses (status_id) ON DELETE RESTRICT;

-- Test Cases
ALTER TABLE test_cases ADD CONSTRAINT fk_test_cases_to_priorities 
    FOREIGN KEY (priority_id) REFERENCES priorities (priority_id) ON DELETE RESTRICT;
ALTER TABLE test_cases ADD CONSTRAINT fk_test_cases_to_users_author 
    FOREIGN KEY (author_user_id) REFERENCES users (user_id) ON DELETE RESTRICT;

-- Test Case Steps
ALTER TABLE test_case_steps ADD CONSTRAINT fk_test_case_steps_to_test_cases 
    FOREIGN KEY (test_case_id) REFERENCES test_cases (test_case_id) ON DELETE CASCADE;

-- Autotests
ALTER TABLE autotests ADD CONSTRAINT fk_autotests_to_test_cases 
    FOREIGN KEY (test_case_id) REFERENCES test_cases (test_case_id) ON DELETE CASCADE;

-- Autotest Versions
ALTER TABLE autotest_versions ADD CONSTRAINT fk_autotest_versions_to_autotests 
    FOREIGN KEY (autotest_id) REFERENCES autotests (autotest_id) ON DELETE CASCADE;

-- Test Runs
ALTER TABLE test_runs ADD CONSTRAINT fk_test_runs_to_test_plans 
    FOREIGN KEY (test_plan_id) REFERENCES test_plans (test_plan_id) ON DELETE RESTRICT;
ALTER TABLE test_runs ADD CONSTRAINT fk_test_runs_to_users_assignee 
    FOREIGN KEY (assignee_user_id) REFERENCES users (user_id) ON DELETE SET NULL;
ALTER TABLE test_runs ADD CONSTRAINT fk_test_runs_to_statuses 
    FOREIGN KEY (status_id) REFERENCES statuses (status_id) ON DELETE RESTRICT;
ALTER TABLE test_runs ADD CONSTRAINT fk_test_runs_to_env_configs 
    FOREIGN KEY (env_config_id) REFERENCES env_configs (env_config_id) ON DELETE RESTRICT;

-- Test Run Items
ALTER TABLE test_run_items ADD CONSTRAINT fk_test_run_items_to_test_runs 
    FOREIGN KEY (test_run_id) REFERENCES test_runs (run_id) ON DELETE CASCADE;
ALTER TABLE test_run_items ADD CONSTRAINT fk_test_run_items_to_test_cases 
    FOREIGN KEY (test_case_id) REFERENCES test_cases (test_case_id) ON DELETE RESTRICT;
ALTER TABLE test_run_items ADD CONSTRAINT fk_test_run_items_to_users_assignee 
    FOREIGN KEY (assignee_user_id) REFERENCES users (user_id) ON DELETE SET NULL;

-- Test Results
ALTER TABLE test_results ADD CONSTRAINT fk_test_results_to_test_run_items 
    FOREIGN KEY (run_item_id) REFERENCES test_run_items (run_item_id) ON DELETE CASCADE;
ALTER TABLE test_results ADD CONSTRAINT fk_test_results_to_statuses 
    FOREIGN KEY (status_id) REFERENCES statuses (status_id) ON DELETE RESTRICT;
ALTER TABLE test_results ADD CONSTRAINT fk_test_results_to_users_tester 
    FOREIGN KEY (executor_user_id) REFERENCES users (user_id) ON DELETE SET NULL;

-- Test Result Artifacts
ALTER TABLE test_result_artifacts ADD CONSTRAINT fk_test_result_artifacts_to_test_results 
    FOREIGN KEY (test_result_id) REFERENCES test_results (result_id) ON DELETE CASCADE;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
ALTER TABLE test_result_artifacts DROP CONSTRAINT IF EXISTS fk_test_result_artifacts_to_test_results;
ALTER TABLE test_results DROP CONSTRAINT IF EXISTS fk_test_results_to_users_tester;
ALTER TABLE test_results DROP CONSTRAINT IF EXISTS fk_test_results_to_statuses;
ALTER TABLE test_results DROP CONSTRAINT IF EXISTS fk_test_results_to_test_run_items;
ALTER TABLE test_run_items DROP CONSTRAINT IF EXISTS fk_test_run_items_to_users_assignee;
ALTER TABLE test_run_items DROP CONSTRAINT IF EXISTS fk_test_run_items_to_test_cases;
ALTER TABLE test_run_items DROP CONSTRAINT IF EXISTS fk_test_run_items_to_test_runs;
ALTER TABLE test_runs DROP CONSTRAINT IF EXISTS fk_test_runs_to_env_configs;
ALTER TABLE test_runs DROP CONSTRAINT IF EXISTS fk_test_runs_to_statuses;
ALTER TABLE test_runs DROP CONSTRAINT IF EXISTS fk_test_runs_to_users_assignee;
ALTER TABLE test_runs DROP CONSTRAINT IF EXISTS fk_test_runs_to_test_plans;
ALTER TABLE autotest_versions DROP CONSTRAINT IF EXISTS fk_autotest_versions_to_autotests;
ALTER TABLE autotests DROP CONSTRAINT IF EXISTS fk_autotests_to_test_cases;
ALTER TABLE test_case_steps DROP CONSTRAINT IF EXISTS fk_test_case_steps_to_test_cases;
ALTER TABLE test_cases DROP CONSTRAINT IF EXISTS fk_test_cases_to_users_author;
ALTER TABLE test_cases DROP CONSTRAINT IF EXISTS fk_test_cases_to_priorities;
ALTER TABLE test_plans DROP CONSTRAINT IF EXISTS fk_test_plans_to_statuses;
ALTER TABLE test_plans DROP CONSTRAINT IF EXISTS fk_test_plans_to_users_owner;
ALTER TABLE test_plans DROP CONSTRAINT IF EXISTS fk_test_plans_to_app_versions;
ALTER TABLE reports DROP CONSTRAINT IF EXISTS fk_reports_to_users;
ALTER TABLE reports DROP CONSTRAINT IF EXISTS fk_reports_to_test_runs;
ALTER TABLE reports DROP CONSTRAINT IF EXISTS fk_reports_to_report_templates;
ALTER TABLE tool_config_params DROP CONSTRAINT IF EXISTS fk_tool_config_params_to_tool_configs;
ALTER TABLE config_errors DROP CONSTRAINT IF EXISTS fk_config_errors_to_env_configs;
ALTER TABLE env_config_params DROP CONSTRAINT IF EXISTS fk_env_config_params_to_env_configs;
ALTER TABLE audit_log DROP CONSTRAINT IF EXISTS fk_audit_log_to_users;
ALTER TABLE user_roles DROP CONSTRAINT IF EXISTS fk_user_roles_granted_by;
ALTER TABLE user_roles DROP CONSTRAINT IF EXISTS fk_user_roles_to_roles;
ALTER TABLE user_roles DROP CONSTRAINT IF EXISTS fk_user_roles_to_users;
-- +goose StatementEnd
