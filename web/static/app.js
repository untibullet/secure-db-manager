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

// ... (TABLES, FKS, DATA оставляем без изменений из прошлого ответа) ...
// Для краткости я не дублирую массивы TABLES, FKS и DATA.
// Предполагается, что они объявлены выше.

let activeTable = "test_cases";

function el(id) { return document.getElementById(id); }

// --- Рендеринг ---

function renderTablesList() {
  const root = el("tablesList");
  root.innerHTML = "";
  TABLES.forEach(t => {
    const btn = document.createElement("div");
    btn.className = "nav-item" + (t === activeTable ? " active" : "");
    btn.textContent = t;
    btn.onclick = () => { activeTable = t; refresh(); };
    root.appendChild(btn);
  });
}

function getProcessedData(tableName) {
    let rows = DATA[tableName] || [];
    
    // 1. Фильтрация
    const filterVal = el("filterInput").value.toLowerCase();
    if (filterVal) {
        rows = rows.filter(r => Object.values(r).some(v => String(v).toLowerCase().includes(filterVal)));
    }

    // 2. Сортировка (по первому полю для простоты)
    const sortType = el("sortSelect").value;
    if (rows.length > 0) {
        const key = Object.keys(rows[0])[0]; // Берем ID или первое поле
        rows = [...rows].sort((a, b) => {
            if (a[key] < b[key]) return sortType === 'asc' ? -1 : 1;
            if (a[key] > b[key]) return sortType === 'asc' ? 1 : -1;
            return 0;
        });
    }
    return rows;
}

function renderTable(container, tableName, rows) {
  if (!rows || rows.length === 0) {
    container.innerHTML = `<div class="empty-state">Нет данных</div>`;
    return;
  }
  const cols = Object.keys(rows[0]);
  const thead = `<thead><tr>${cols.map(c => `<th>${c}</th>`).join("")}</tr></thead>`;
  const tbody = `<tbody>${
    rows.map(r => `<tr>${cols.map(c => `<td>${String(r[c] ?? "")}</td>`).join("")}</tr>`).join("")
  }</tbody>`;
  container.innerHTML = `<table class="data-table">${thead}${tbody}</table>`;
}

function renderNeighbors() {
  const root = el("neighbors");
  root.innerHTML = "";

  const ns = new Set();
  FKS.forEach(fk => {
    if (fk.fromTable === activeTable) ns.add(fk.toTable);
    if (fk.toTable === activeTable) ns.add(fk.fromTable);
  });

  if (ns.size === 0) {
    root.innerHTML = `<div class="empty-state">Нет связанных таблиц</div>`;
    return;
  }

  ns.forEach(t => {
    const wrapper = document.createElement("div");
    wrapper.className = "neighbor-card";
    
    // Заголовок соседа
    const header = document.createElement("div");
    header.className = "neighbor-header";
    header.innerHTML = `<span class="neighbor-title">${t}</span>`;
    
    // Мини-тулбар для соседа (как просили в задаче - к каждой таблице)
    const miniToolbar = document.createElement("div");
    miniToolbar.className = "mini-toolbar";
    miniToolbar.innerHTML = `
        <button class="btn btn-xs" onclick="alert('Open modal for ${t}')">Edit</button>
        <select class="input-xs"><option>AZ</option><option>ZA</option></select>
        <input class="input-xs" placeholder="Filter..." style="width:60px">
    `;
    
    header.appendChild(miniToolbar);
    wrapper.appendChild(header);

    const grid = document.createElement("div");
    grid.className = "neighbor-content";
    renderTable(grid, t, DATA[t] || []);
    wrapper.appendChild(grid);
    
    root.appendChild(wrapper);
  });
}

// --- Интерактивность ---

function applyFilter() {
    renderTable(el("activeTable"), activeTable, getProcessedData(activeTable));
}

function applySort() {
    renderTable(el("activeTable"), activeTable, getProcessedData(activeTable));
}

function refresh() {
  renderTablesList();
  el("activeTitle").textContent = `Таблица: ${activeTable}`;
  // Сброс фильтров при смене таблицы
  el("filterInput").value = ""; 
  renderTable(el("activeTable"), activeTable, getProcessedData(activeTable));
  renderNeighbors();
}

// --- Модальное окно ---

function openModal(actionType) {
    const modal = el("modalOverlay");
    const fieldsContainer = el("modalFormFields");
    const title = el("modalTitle");
    
    title.textContent = actionType === 'CREATE' ? `Создание записи в ${activeTable}` : 'Редактирование';
    fieldsContainer.innerHTML = "";

    // Генерируем поля на основе ключей первой записи (или заглушки, если пусто)
    const exampleRow = (DATA[activeTable] && DATA[activeTable][0]) ? DATA[activeTable][0] : { id: '', name: '' };
    
    Object.keys(exampleRow).forEach(key => {
        const div = document.createElement("div");
        div.className = "form-group";
        div.innerHTML = `
            <label>${key}</label>
            <input type="text" class="input-text" name="${key}" placeholder="">
        `;
        fieldsContainer.appendChild(div);
    });

    modal.classList.remove("hidden");
}

function closeModal() {
    el("modalOverlay").classList.add("hidden");
}

function mockAction(action) {
    alert(`MOCK ACTION: ${action} executed for table ${activeTable}`);
    closeModal();
}

// Инициализация
refresh();
