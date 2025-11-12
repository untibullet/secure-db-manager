-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'db_admin') THEN
        CREATE ROLE db_admin WITH LOGIN PASSWORD NULL VALID UNTIL '2026-12-31' CONNECTION LIMIT 5;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'db_test_lead') THEN
        CREATE ROLE db_test_lead WITH LOGIN PASSWORD NULL VALID UNTIL '2026-12-31' CONNECTION LIMIT 10;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'db_automation_engineer') THEN
        CREATE ROLE db_automation_engineer WITH LOGIN PASSWORD NULL VALID UNTIL '2026-12-31' CONNECTION LIMIT 15;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'db_tester') THEN
        CREATE ROLE db_tester WITH LOGIN PASSWORD NULL VALID UNTIL '2026-12-31' CONNECTION LIMIT 20;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'db_cicd_system') THEN
        CREATE ROLE db_cicd_system WITH LOGIN PASSWORD NULL VALID UNTIL '2026-12-31' CONNECTION LIMIT 10;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'db_developer') THEN
        CREATE ROLE db_developer WITH LOGIN PASSWORD NULL VALID UNTIL '2026-12-31' CONNECTION LIMIT 15;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'db_guest') THEN
        CREATE ROLE db_guest WITH LOGIN PASSWORD NULL VALID UNTIL '2026-12-31' CONNECTION LIMIT 30;
    END IF;
END
$$;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP ROLE IF EXISTS db_admin;
DROP ROLE IF EXISTS db_test_lead;
DROP ROLE IF EXISTS db_automation_engineer;
DROP ROLE IF EXISTS db_tester;
DROP ROLE IF EXISTS db_cicd_system;
DROP ROLE IF EXISTS db_developer;
DROP ROLE IF EXISTS db_guest;
-- +goose StatementEnd
