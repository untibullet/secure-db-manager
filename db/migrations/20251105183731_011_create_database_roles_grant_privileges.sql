-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    -- 1. Создание групповых ролей (контейнеры прав)
    -- Роль администратора
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'db_admin') THEN
        CREATE ROLE db_admin WITH NOLOGIN;
    END IF;

    -- Роль Тест-лида
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'db_test_lead') THEN
        CREATE ROLE db_test_lead WITH NOLOGIN;
    END IF;

    -- Роль Инженера по автоматизации
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'db_automation_engineer') THEN
        CREATE ROLE db_automation_engineer WITH NOLOGIN;
    END IF;

    -- Роль Тестировщика
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'db_tester') THEN
        CREATE ROLE db_tester WITH NOLOGIN;
    END IF;

    -- Роль CI/CD системы (технический пользователь)
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'db_cicd_system') THEN
        CREATE ROLE db_cicd_system WITH NOLOGIN;
    END IF;

    -- Роль Гостя
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'db_guest') THEN
        CREATE ROLE db_guest WITH NOLOGIN;
    END IF;
END
$$;

-- 2. Выдача базовых прав на схему (чтобы "видеть" объекты)
GRANT USAGE ON SCHEMA public TO db_admin, db_test_lead, db_automation_engineer, db_tester, db_cicd_system, db_guest;

---------------------------------------------------------------------------
-- 3. Настройка прав доступа (Access Control Lists) согласно Таблице 2.х
---------------------------------------------------------------------------

-- === Роль: Administrator ===
-- Полный доступ к управлению пользователями
GRANT SELECT, INSERT, UPDATE, DELETE ON v_admin_users_and_roles TO db_admin;


-- === Роль: Test_Lead ===
-- Управление тест-планами (CRUD)
GRANT SELECT, INSERT, UPDATE, DELETE ON v_public_test_plans TO db_test_lead;
-- Просмотр всех кейсов
GRANT SELECT ON v_lead_all_cases TO db_test_lead;
-- Просмотр автотестов
GRANT SELECT ON v_engineer_autotests TO db_test_lead;
-- Просмотр окружений
GRANT SELECT ON v_shared_environments TO db_test_lead;
-- Просмотр всех результатов
GRANT SELECT ON v_lead_all_results TO db_test_lead;
-- Управление отчетами (CRUD)
GRANT SELECT, INSERT, DELETE ON v_shared_reports TO db_test_lead;
-- Просмотр статистики (из дополнительных view)
GRANT SELECT ON v_test_case_statistics TO db_test_lead;
GRANT SELECT ON v_active_test_runs TO db_test_lead;
GRANT SELECT ON v_test_execution_summary TO db_test_lead;


-- === Роль: Tester ===
-- Просмотр планов
GRANT SELECT ON v_public_test_plans TO db_tester;
-- Работа со СВОИМИ назначенными кейсами (C, R, U)
-- Примечание: для INSERT/UPDATE view должна быть простой или иметь триггеры
GRANT SELECT, INSERT, UPDATE ON v_tester_assigned_cases TO db_tester;
-- Просмотр окружений
GRANT SELECT ON v_shared_environments TO db_tester;
-- Работа со СВОИМИ результатами (R, C, U)
GRANT SELECT, INSERT, UPDATE ON v_tester_my_results TO db_tester;
-- Просмотр отчетов
GRANT SELECT ON v_shared_reports TO db_tester;


-- === Роль: Automation_Engineer ===
-- Просмотр планов и кейсов
GRANT SELECT ON v_public_test_plans TO db_automation_engineer;
GRANT SELECT ON v_lead_all_cases TO db_automation_engineer;
-- Полное управление автотестами (CRUD)
GRANT SELECT, INSERT, UPDATE, DELETE ON v_engineer_autotests TO db_automation_engineer;
-- Управление окружениями (C, R, U)
GRANT SELECT, INSERT, UPDATE ON v_shared_environments TO db_automation_engineer;
-- Просмотр результатов и отчетов
GRANT SELECT ON v_lead_all_results TO db_automation_engineer;
GRANT SELECT ON v_shared_reports TO db_automation_engineer;


-- === Роль: Guest ===
-- Только чтение публичных данных
GRANT SELECT ON v_public_test_plans TO db_guest;
GRANT SELECT ON v_public_results TO db_guest;
GRANT SELECT ON v_shared_reports TO db_guest;


-- === Роль: CI_CD_System ===
-- Техническая роль, обычно требует доступа к окружениям и запускам
GRANT SELECT ON v_shared_environments TO db_cicd_system;
-- В документе сказано "прямой доступ", но если через View:
GRANT SELECT, INSERT, UPDATE ON v_active_test_runs TO db_cicd_system;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Отзыв прав (каскадно удалит права у наследуемых пользователей)
DROP ROLE IF EXISTS db_guest;
DROP ROLE IF EXISTS db_cicd_system;
DROP ROLE IF EXISTS db_tester;
DROP ROLE IF EXISTS db_automation_engineer;
DROP ROLE IF EXISTS db_test_lead;
DROP ROLE IF EXISTS db_admin;
-- +goose StatementEnd
