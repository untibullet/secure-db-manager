-- +goose Up
-- +goose StatementBegin
-- Триггеры для автоматического обновления updated_at
CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
WHEN (OLD IS DISTINCT FROM NEW)
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_testcases_updated_at
BEFORE UPDATE ON test_cases
FOR EACH ROW
WHEN (OLD IS DISTINCT FROM NEW)
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_testplans_updated_at
BEFORE UPDATE ON test_plans
FOR EACH ROW
WHEN (OLD IS DISTINCT FROM NEW)
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_envconfigs_updated_at
BEFORE UPDATE ON env_configs
FOR EACH ROW
WHEN (OLD IS DISTINCT FROM NEW)
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_toolconfigs_updated_at
BEFORE UPDATE ON tool_configs
FOR EACH ROW
WHEN (OLD IS DISTINCT FROM NEW)
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_autotests_updated_at
BEFORE UPDATE ON autotests
FOR EACH ROW
WHEN (OLD IS DISTINCT FROM NEW)
EXECUTE FUNCTION update_updated_at_column();

-- Триггеры аудита для критических таблиц
-- users
CREATE TRIGGER trg_users_audit_ins
AFTER INSERT ON users
FOR EACH ROW
EXECUTE FUNCTION audit_users();

CREATE TRIGGER trg_users_audit_upd
AFTER UPDATE ON users
FOR EACH ROW
WHEN (OLD IS DISTINCT FROM NEW)
EXECUTE FUNCTION audit_users();

CREATE TRIGGER trg_users_audit_del
AFTER DELETE ON users
FOR EACH ROW
EXECUTE FUNCTION audit_users();

-- user_roles
CREATE TRIGGER trg_user_roles_audit_ins
AFTER INSERT ON user_roles
FOR EACH ROW
EXECUTE FUNCTION audit_user_roles();

CREATE TRIGGER trg_user_roles_audit_upd
AFTER UPDATE ON user_roles
FOR EACH ROW
WHEN (OLD IS DISTINCT FROM NEW)
EXECUTE FUNCTION audit_user_roles();

CREATE TRIGGER trg_user_roles_audit_del
AFTER DELETE ON user_roles
FOR EACH ROW
EXECUTE FUNCTION audit_user_roles();


-- test_cases
CREATE TRIGGER trg_test_cases_audit_ins
AFTER INSERT ON test_cases
FOR EACH ROW
EXECUTE FUNCTION audit_test_cases();

CREATE TRIGGER trg_test_cases_audit_upd
AFTER UPDATE ON test_cases
FOR EACH ROW
WHEN (OLD IS DISTINCT FROM NEW)
EXECUTE FUNCTION audit_test_cases();

CREATE TRIGGER trg_test_cases_audit_del
AFTER DELETE ON test_cases
FOR EACH ROW
EXECUTE FUNCTION audit_test_cases();


-- test_plans
CREATE TRIGGER trg_test_plans_audit_ins
AFTER INSERT ON test_plans
FOR EACH ROW
EXECUTE FUNCTION audit_test_plans();

CREATE TRIGGER trg_test_plans_audit_upd
AFTER UPDATE ON test_plans
FOR EACH ROW
WHEN (OLD IS DISTINCT FROM NEW)
EXECUTE FUNCTION audit_test_plans();

CREATE TRIGGER trg_test_plans_audit_del
AFTER DELETE ON test_plans
FOR EACH ROW
EXECUTE FUNCTION audit_test_plans();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_users_updated_at         ON users;
DROP TRIGGER IF EXISTS trg_testcases_updated_at     ON test_cases;
DROP TRIGGER IF EXISTS trg_testplans_updated_at     ON test_plans;
DROP TRIGGER IF EXISTS trg_envconfigs_updated_at    ON env_configs;
DROP TRIGGER IF EXISTS trg_toolconfigs_updated_at   ON tool_configs;
DROP TRIGGER IF EXISTS trg_autotests_updated_at     ON autotests;

DROP TRIGGER IF EXISTS trg_users_audit_ins        ON users;
DROP TRIGGER IF EXISTS trg_users_audit_upd        ON users;
DROP TRIGGER IF EXISTS trg_users_audit_del        ON users;

DROP TRIGGER IF EXISTS trg_user_roles_audit_ins   ON user_roles;
DROP TRIGGER IF EXISTS trg_user_roles_audit_upd   ON user_roles;
DROP TRIGGER IF EXISTS trg_user_roles_audit_del   ON user_roles;

DROP TRIGGER IF EXISTS trg_test_cases_audit_ins   ON test_cases;
DROP TRIGGER IF EXISTS trg_test_cases_audit_upd   ON test_cases;
DROP TRIGGER IF EXISTS trg_test_cases_audit_del   ON test_cases;

DROP TRIGGER IF EXISTS trg_test_plans_audit_ins   ON test_plans;
DROP TRIGGER IF EXISTS trg_test_plans_audit_upd   ON test_plans;
DROP TRIGGER IF EXISTS trg_test_plans_audit_del   ON test_plans;
-- +goose StatementEnd
