-- +goose Up
-- +goose StatementBegin

-- =============================================
-- 1. Вставка Пользователей (3 записи)
-- Пароли указаны в виде заглушек хешей
-- =============================================
INSERT INTO users (username, password_hash, email, full_name, is_active) VALUES
    ('ivan_ivanov', 'hash_admin_123', 'ivan@example.com', 'Иван Иванов', TRUE),
    ('petr_petrov', 'hash_tester_456', 'petr@example.com', 'Петр Петров', TRUE),
    ('guest_user', 'hash_guest_789', 'guest@example.com', 'Гость Системы', TRUE);

-- =============================================
-- 2. Привязка Ролей к Пользователям
-- Используем подзапросы для получения ID
-- =============================================
INSERT INTO user_roles (user_id, role_id, granted_by)
VALUES
    -- Иван (Админ) получает роль ADMIN
    ((SELECT user_id FROM users WHERE username = 'ivan_ivanov'),
     (SELECT role_id FROM roles WHERE code = 'ADMIN'),
     (SELECT user_id FROM users WHERE username = 'ivan_ivanov')),
    
    -- Петр (Тестировщик) получает роль TESTER
    ((SELECT user_id FROM users WHERE username = 'petr_petrov'),
     (SELECT role_id FROM roles WHERE code = 'TESTER'),
     (SELECT user_id FROM users WHERE username = 'ivan_ivanov')),
     
    -- Гость получает роль GUEST
    ((SELECT user_id FROM users WHERE username = 'guest_user'),
     (SELECT role_id FROM roles WHERE code = 'GUEST'),
     (SELECT user_id FROM users WHERE username = 'ivan_ivanov'));

-- =============================================
-- 3. Технические справочники (Конфигурации и Версии)
-- Необходимы для создания Test Runs
-- =============================================
-- Версия приложения
INSERT INTO app_versions (version_string, is_active, created_at) VALUES
    ('1.0.0', FALSE, CURRENT_DATE);

-- Конфигурация окружения (например, QA стенд)
INSERT INTO env_configs (name, description, is_active) VALUES
    ('QA-Env-01', 'Основной стенд тестирования (Linux/Postgres)', TRUE);

-- Конфигурация инструментов (например, Selenium Grid)
INSERT INTO tool_configs (name, description, is_active) VALUES
    ('Selenium-Grid-Chrome', 'Кластер Selenium с браузером Chrome 120', TRUE);

-- =============================================
-- 4. Test Plans (Планы тестирования) - 5 записей
-- =============================================
INSERT INTO test_plans (name, description, priority_id, start_date, end_date, owner_user_id, status) VALUES
    ('Регресс v1.0', 'Полный регрессионный цикл', 
     (SELECT priority_id FROM priorities WHERE code = 'CRITICAL'), 
     CURRENT_DATE, CURRENT_DATE + 7, 
     (SELECT user_id FROM users WHERE username = 'ivan_ivanov'), 'ACTIVE'),
     
    ('Смоук-тест API', 'Быстрая проверка API эндпоинтов', 
     (SELECT priority_id FROM priorities WHERE code = 'HIGH'), 
     CURRENT_DATE, CURRENT_DATE + 1, 
     (SELECT user_id FROM users WHERE username = 'petr_petrov'), 'DRAFT'),
     
    ('UI Тестирование Логина', 'Проверка форм авторизации', 
     (SELECT priority_id FROM priorities WHERE code = 'MEDIUM'), 
     CURRENT_DATE + 2, CURRENT_DATE + 5, 
     (SELECT user_id FROM users WHERE username = 'petr_petrov'), 'COMPLETED'),
     
    ('Нагрузочное тестирование', 'Проверка производительности при 1000 RPS', 
     (SELECT priority_id FROM priorities WHERE code = 'LOW'), 
     CURRENT_DATE + 10, CURRENT_DATE + 12, 
     (SELECT user_id FROM users WHERE username = 'ivan_ivanov'), 'DRAFT'),
     
    ('Проверка безопасности', 'Пентесты формы регистрации', 
     (SELECT priority_id FROM priorities WHERE code = 'CRITICAL'), 
     CURRENT_DATE + 5, CURRENT_DATE + 8, 
     (SELECT user_id FROM users WHERE username = 'ivan_ivanov'), 'ARCHIVED');

-- =============================================
-- 5. Test Cases (Тестовые случаи) - 5 записей
-- =============================================
INSERT INTO test_cases (name, description, priority_id, owner_user_id, is_automated, estimated_duration_minutes) VALUES
    ('TC-001: Успешный вход', 'Вход с валидными данными', 
     (SELECT priority_id FROM priorities WHERE code = 'CRITICAL'), 
     (SELECT user_id FROM users WHERE username = 'petr_petrov'), TRUE, 2),
     
    ('TC-002: Неверный пароль', 'Вход с неверным паролем, ожидается ошибка', 
     (SELECT priority_id FROM priorities WHERE code = 'HIGH'), 
     (SELECT user_id FROM users WHERE username = 'petr_petrov'), TRUE, 1),
     
    ('TC-003: Регистрация нового', 'Создание нового пользователя через UI', 
     (SELECT priority_id FROM priorities WHERE code = 'CRITICAL'), 
     (SELECT user_id FROM users WHERE username = 'petr_petrov'), TRUE, 5),
     
    ('TC-004: Сброс пароля', 'Проверка отправки email для сброса', 
     (SELECT priority_id FROM priorities WHERE code = 'MEDIUM'), 
     (SELECT user_id FROM users WHERE username = 'ivan_ivanov'), FALSE, 10),
     
    ('TC-005: Просмотр профиля', 'Доступ к данным профиля после логина', 
     (SELECT priority_id FROM priorities WHERE code = 'LOW'), 
     (SELECT user_id FROM users WHERE username = 'petr_petrov'), FALSE, 3);

-- =============================================
-- 6. Autotests (Автотесты) - 3 записи
-- Связаны с тест-кейсами, помеченными как is_automated=TRUE
-- =============================================
INSERT INTO autotests (test_case_id, name, description, owner_user_id) VALUES
    ((SELECT test_case_id FROM test_cases WHERE name LIKE 'TC-001%'), 
     'Login_Success_Test.java', 'Автотест успешного входа (Selenium)', 
     (SELECT user_id FROM users WHERE username = 'petr_petrov')),
     
    ((SELECT test_case_id FROM test_cases WHERE name LIKE 'TC-002%'), 
     'Login_Failure_Test.java', 'Автотест ошибки входа', 
     (SELECT user_id FROM users WHERE username = 'petr_petrov')),
     
    ((SELECT test_case_id FROM test_cases WHERE name LIKE 'TC-003%'), 
     'Registration_Flow_Test.py', 'E2E тест регистрации (Playwright)', 
     (SELECT user_id FROM users WHERE username = 'petr_petrov'));

-- =============================================
-- 7. Test Runs (Тестовые прогоны) - 1 запись (техническая необходимость)
-- Создаем прогон, чтобы привязать к нему результаты
-- =============================================
INSERT INTO test_runs (test_plan_id, env_config_id, tool_config_id, version_id, name, start_date, created_by) VALUES
    ((SELECT test_plan_id FROM test_plans WHERE name = 'Регресс v1.0'),
     (SELECT env_config_id FROM env_configs LIMIT 1),
     (SELECT tool_config_id FROM tool_configs LIMIT 1),
     (SELECT version_id FROM app_versions LIMIT 1),
     'Run #1: Nightly Build', 
     CURRENT_TIMESTAMP,
     (SELECT user_id FROM users WHERE username = 'ivan_ivanov'));

-- =============================================
-- 8. Test Run Items (Состав прогона) - 3 записи
-- Добавляем кейсы в созданный прогон
-- =============================================
INSERT INTO test_run_items (test_run_id, test_case_id, execution_order) VALUES
    ((SELECT test_run_id FROM test_runs LIMIT 1), (SELECT test_case_id FROM test_cases WHERE name LIKE 'TC-001%'), 1),
    ((SELECT test_run_id FROM test_runs LIMIT 1), (SELECT test_case_id FROM test_cases WHERE name LIKE 'TC-002%'), 2),
    ((SELECT test_run_id FROM test_runs LIMIT 1), (SELECT test_case_id FROM test_cases WHERE name LIKE 'TC-003%'), 3);

-- =============================================
-- 9. Test Results (Результаты тестов) - 3 записи
-- =============================================
INSERT INTO test_results (run_item_id, status_id, executor_user_id, result_summary, execution_duration_minutes) VALUES
    -- Результат для TC-001: Passed
    ((SELECT run_item_id FROM test_run_items WHERE execution_order = 1), 
     (SELECT status_id FROM statuses WHERE code = 'PASSED'), 
     (SELECT user_id FROM users WHERE username = 'petr_petrov'), 'Успешно авторизован', 2),
     
    -- Результат для TC-002: Passed
    ((SELECT run_item_id FROM test_run_items WHERE execution_order = 2), 
     (SELECT status_id FROM statuses WHERE code = 'PASSED'), 
     (SELECT user_id FROM users WHERE username = 'petr_petrov'), 'Ошибка отобразилась корректно', 1),
     
    -- Результат для TC-003: Failed
    ((SELECT run_item_id FROM test_run_items WHERE execution_order = 3), 
     (SELECT status_id FROM statuses WHERE code = 'FAILED'), 
     (SELECT user_id FROM users WHERE username = 'petr_petrov'), 'Кнопка "Создать" неактивна', 4);

-- =============================================
-- 10. Report Templates (Шаблоны отчетов) - 1 запись
-- =============================================
INSERT INTO report_templates (name, template_content) VALUES
    ('Стандартный отчет о дефектах', '# Отчет о тестировании\n## Статистика\n{{stats_table}}');

-- =============================================
-- 11. Reports (Отчеты) - 2 записи
-- =============================================
INSERT INTO reports (name, template_id, owner_user_id, report_content) VALUES
    ('Еженедельный отчет QA', 
     (SELECT template_id FROM report_templates LIMIT 1), 
     (SELECT user_id FROM users WHERE username = 'ivan_ivanov'), 
     'Содержание отчета за неделю...'),
     
    ('Отчет по инциденту #123', 
     (SELECT template_id FROM report_templates LIMIT 1), 
     (SELECT user_id FROM users WHERE username = 'petr_petrov'), 
     'Описание критического бага в продакшене...');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- 1) Reports -> Report templates (reports ссылаются на templates и users)
DELETE FROM reports
WHERE name IN ('Еженедельный отчет QA', 'Отчет по инциденту #123'); -- удаляем только наши тестовые отчеты

DELETE FROM report_templates
WHERE name = 'Стандартный отчет о дефектах'; -- удаляем только наш шаблон

-- 2) Test results -> Run items -> Runs (цепочка FK)
DELETE FROM test_results
WHERE run_item_id IN (
    SELECT tri.run_item_id
    FROM test_run_items tri
    JOIN test_runs tr ON tr.test_run_id = tri.test_run_id
    WHERE tr.name = 'Run #1: Nightly Build'
);

DELETE FROM test_run_items
WHERE test_run_id IN (
    SELECT test_run_id FROM test_runs WHERE name = 'Run #1: Nightly Build'
);

DELETE FROM test_runs
WHERE name = 'Run #1: Nightly Build';

-- 3) Autotests (ссылаются на test_cases)
DELETE FROM autotests
WHERE name IN ('Login_Success_Test.java', 'Login_Failure_Test.java', 'Registration_Flow_Test.py');

-- 4) Test cases (их нельзя удалить, пока есть autotests/run_items)
DELETE FROM test_cases
WHERE name IN (
    'TC-001: Успешный вход',
    'TC-002: Неверный пароль',
    'TC-003: Регистрация нового',
    'TC-004: Сброс пароля',
    'TC-005: Просмотр профиля'
);

-- 5) Test plans (на них ссылаются test_runs)
DELETE FROM test_plans
WHERE name IN (
    'Регресс v1.0',
    'Смоук-тест API',
    'UI Тестирование Логина',
    'Нагрузочное тестирование',
    'Проверка безопасности'
);

-- 6) Технические справочники, использованные в test_runs (env/tool/app)
DELETE FROM tool_configs
WHERE name = 'Selenium-Grid-Chrome';

DELETE FROM env_configs
WHERE name = 'QA-Env-01';

DELETE FROM app_versions
WHERE version_string = '1.0.0';

-- 7) User roles -> Users (user_roles ссылается на users и roles)
DELETE FROM user_roles
WHERE user_id IN (
    SELECT user_id FROM users WHERE username IN ('ivan_ivanov', 'petr_petrov', 'guest_user')
);

DELETE FROM users
WHERE username IN ('ivan_ivanov', 'petr_petrov', 'guest_user');

-- +goose StatementEnd
