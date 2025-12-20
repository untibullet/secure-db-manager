// ---- 1) Мок-список таблиц (минимальный, под вашу задачу)
const TABLES = [
  "users", "priorities", "statuses", "roles", "user_roles",
  "test_plans", "test_cases", "autotests",
  "test_runs", "test_run_items", "test_results",
  "report_templates", "reports",
];

// ---- 2) FK-граф (минимально нужный для “соседей”)
const FKS = [
  { fromTable: "test_plans", fromKey: "owner_user_id", toTable: "users", toKey: "user_id" },
  { fromTable: "test_plans", fromKey: "priority_id",  toTable: "priorities", toKey: "priority_id" },

  { fromTable: "test_cases", fromKey: "owner_user_id", toTable: "users", toKey: "user_id" },
  { fromTable: "test_cases", fromKey: "priority_id",  toTable: "priorities", toKey: "priority_id" },

  { fromTable: "autotests", fromKey: "test_case_id", toTable: "test_cases", toKey: "test_case_id" },

  { fromTable: "reports", fromKey: "template_id", toTable: "report_templates", toKey: "template_id" },
  { fromTable: "reports", fromKey: "owner_user_id", toTable: "users", toKey: "user_id" },

  { fromTable: "test_run_items", fromKey: "test_run_id", toTable: "test_runs", toKey: "test_run_id" },
  { fromTable: "test_run_items", fromKey: "test_case_id", toTable: "test_cases", toKey: "test_case_id" },

  { fromTable: "test_results", fromKey: "run_item_id", toTable: "test_run_items", toKey: "run_item_id" },
  { fromTable: "test_results", fromKey: "status_id", toTable: "statuses", toKey: "status_id" },
  { fromTable: "test_results", fromKey: "executor_user_id", toTable: "users", toKey: "user_id" },
];

// ---- 3) Мок-данные (очень маленькие)
const DATA = {
  users: [
    { user_id: 1, username: "ivan_ivanov", full_name: "Иван Иванов" },
    { user_id: 2, username: "petr_petrov", full_name: "Петр Петров" },
    { user_id: 3, username: "guest_user", full_name: "Гость Системы" },
  ],
  priorities: [
    { priority_id: 1, code: "CRITICAL", name: "Критический" },
    { priority_id: 2, code: "HIGH", name: "Высокий" },
  ],
  statuses: [
    { status_id: 1, code: "PASSED", name: "Пройден" },
    { status_id: 2, code: "FAILED", name: "Провален" },
  ],
  test_plans: [
    { test_plan_id: 10, name: "Регресс v1.0", priority_id: 1, owner_user_id: 1, status: "ACTIVE" },
    { test_plan_id: 11, name: "Смоук-тест API", priority_id: 2, owner_user_id: 2, status: "DRAFT" },
  ],
  test_cases: [
    { test_case_id: 100, name: "TC-001: Успешный вход", priority_id: 1, owner_user_id: 2, is_automated: true },
    { test_case_id: 101, name: "TC-002: Неверный пароль", priority_id: 2, owner_user_id: 2, is_automated: true },
  ],
  autotests: [
    { autotest_id: 1000, test_case_id: 100, name: "Login_Success_Test.java", owner_user_id: 2 },
    { autotest_id: 1001, test_case_id: 101, name: "Login_Failure_Test.java", owner_user_id: 2 },
  ],
  test_runs: [
    { test_run_id: 500, name: "Run #1: Nightly Build", test_plan_id: 10 },
  ],
  test_run_items: [
    { run_item_id: 700, test_run_id: 500, test_case_id: 100, execution_order: 1 },
    { run_item_id: 701, test_run_id: 500, test_case_id: 101, execution_order: 2 },
  ],
  test_results: [
    { test_result_id: 900, run_item_id: 700, status_id: 1, executor_user_id: 2, result_summary: "OK" },
    { test_result_id: 901, run_item_id: 701, status_id: 2, executor_user_id: 2, result_summary: "Fail" },
  ],
  report_templates: [
    { template_id: 77, name: "Стандартный отчет о дефектах" },
  ],
  reports: [
    { report_id: 88, name: "Еженедельный отчет QA", template_id: 77, owner_user_id: 1 },
  ],
};

let activeTable = "test_cases";

function el(id) { return document.getElementById(id); }

function renderTablesList() {
  const root = el("tablesList");
  root.innerHTML = "";
  TABLES.forEach(t => {
    const btn = document.createElement("button");
    btn.className = "tableBtn" + (t === activeTable ? " active" : "");
    btn.textContent = t;
    btn.onclick = () => { activeTable = t; refresh(); };
    root.appendChild(btn);
  });
}

function renderTable(container, tableName, rows) {
  if (!rows || rows.length === 0) {
    container.innerHTML = `<div class="empty">Нет данных</div>`;
    return;
  }
  const cols = Object.keys(rows[0]);
  const thead = `<thead><tr>${cols.map(c => `<th>${c}</th>`).join("")}</tr></thead>`;
  const tbody = `<tbody>${
    rows.map(r => `<tr>${cols.map(c => `<td>${String(r[c] ?? "")}</td>`).join("")}</tr>`).join("")
  }</tbody>`;
  container.innerHTML = `<table class="tbl">${thead}${tbody}</table>`;
}

function neighborsFor(tableName) {
  // соседи = таблицы, которые ссылаются на активную ИЛИ на которые ссылается активная
  const out = new Set();
  FKS.forEach(fk => {
    if (fk.fromTable === tableName) out.add(fk.toTable);
    if (fk.toTable === tableName) out.add(fk.fromTable);
  });
  return [...out];
}

function renderNeighbors() {
  const root = el("neighbors");
  root.innerHTML = "";

  const ns = neighborsFor(activeTable);
  if (ns.length === 0) {
    root.innerHTML = `<div class="empty">Нет связей</div>`;
    return;
  }

  ns.forEach(t => {
    const card = document.createElement("div");
    card.className = "neighborCard";
    const title = document.createElement("div");
    title.className = "neighborTitle";
    title.textContent = t;

    const grid = document.createElement("div");
    grid.className = "neighborGrid";
    renderTable(grid, t, DATA[t]);

    card.appendChild(title);
    card.appendChild(grid);
    root.appendChild(card);
  });
}

function refresh() {
  renderTablesList();
  el("activeTitle").textContent = `Таблица: ${activeTable}`;
  renderTable(el("activeTable"), activeTable, DATA[activeTable]);
  renderNeighbors();
}

refresh();
