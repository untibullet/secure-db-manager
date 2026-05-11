// ---- Auth ----
function getToken() { return localStorage.getItem('jwt'); }
function clearAuth() { localStorage.removeItem('jwt'); }
function redirectLogin() { clearAuth(); window.location.href = '/login'; }

// ---- API ----
async function apiFetch(path, opts) {
    opts = opts || {};
    var headers = { 'Content-Type': 'application/json' };
    var token = getToken();
    if (token) headers.Authorization = 'Bearer ' + token;
    var res = await fetch(path, Object.assign({}, opts, { headers: headers }));
    if (res.status === 401) { redirectLogin(); throw new Error('401'); }
    if (res.status === 204) return null;
    var body = await res.json();
    if (!res.ok) {
        var e = new Error(body.message || 'HTTP ' + res.status);
        e.status = res.status;
        throw e;
    }
    return body;
}

// ---- State ----
var currentUser = null;
var activeSection = null;
var pageOffset = 0;
var pageLimit = 20;
var activeFilters = {};
var cachedRoles = [];
var expandedSet = {};

function el(id) { return document.getElementById(id); }

// ---- Formatters ----
function fmtDate(v) {
    if (!v) return '';
    var d = (v && typeof v === 'object' && v.Time) ? v.Time : v;
    if (!d) return '';
    try { return new Date(d).toLocaleDateString('ru-RU'); } catch(e) { return String(d); }
}

function fmtDT(v) {
    if (!v) return '';
    var d = (v && typeof v === 'object' && v.Time) ? v.Time : v;
    if (!d) return '';
    try { return new Date(d).toLocaleString('ru-RU'); } catch(e) { return String(d); }
}

function nullVal(v) {
    if (v == null) return '';
    if (typeof v !== 'object') return String(v);
    if (!v.Valid) return '';
    if (v.String  != null) return v.String;
    if (v.Int64   != null) return String(v.Int64);
    if (v.Float64 != null) return String(v.Float64);
    if (v.Time    != null) return v.Time;
    return '';
}

function boolStr(v) { return v ? 'Да' : 'Нет'; }

var STATUS_COLORS = {
    DRAFT: '#6b7280', ACTIVE: '#16a34a', CLOSED: '#9ca3af',
    PASSED: '#16a34a', FAILED: '#dc2626', IN_PROGRESS: '#d97706',
    BLOCKED: '#7c3aed', PLANNED: '#2563eb', COMPLETED: '#16a34a',
    ABORTED: '#dc2626', SKIPPED: '#9ca3af'
};
function statusBadge(v) {
    var c = STATUS_COLORS[v] || '#6b7280';
    return '<span style="background:' + c + ';color:#fff;padding:2px 8px;border-radius:4px;font-size:0.75rem">' + esc(v || '') + '</span>';
}

function esc(s) {
    return String(s)
        .replace(/&/g,'&amp;').replace(/</g,'&lt;')
        .replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}

// ---- Form helpers ----
function fv(id)  { var e = el(id); return e ? e.value.trim() : ''; }
function fvb(id) { var e = el(id); return e ? e.checked : false; }
function fvi(id) { return parseInt(fv(id)) || 0; }

function fg(label, inputHtml) {
    return '<div class="form-group"><label>' + esc(label) + '</label>' + inputHtml + '</div>';
}
function inp(id, val, placeholder, type) {
    return '<input class="input-text" id="' + esc(id) + '" type="' + esc(type || 'text') + '" value="' + esc(val || '') + '" placeholder="' + esc(placeholder || '') + '">';
}
function inpNum(id, val) {
    return '<input class="input-text" id="' + esc(id) + '" type="number" value="' + (val || '') + '">';
}
function ta(id, val, rows) {
    return '<textarea class="input-text" id="' + esc(id) + '" rows="' + (rows || 3) + '">' + esc(val || '') + '</textarea>';
}
function sel(id, opts, val) {
    var h = '<select class="input-select" id="' + esc(id) + '">';
    opts.forEach(function(o) {
        var v = (o && o.value !== undefined) ? String(o.value) : String(o);
        var l = (o && o.label  !== undefined) ? o.label : o;
        h += '<option value="' + esc(v) + '"' + (v === String(val || '') ? ' selected' : '') + '>' + esc(l) + '</option>';
    });
    return h + '</select>';
}
function chk(id, checked, label) {
    return '<label style="display:flex;align-items:center;gap:8px"><input type="checkbox" id="' + esc(id) + '"' + (checked ? ' checked' : '') + '> ' + esc(label) + '</label>';
}
function dateInp(id, val) {
    var d = '';
    if (val) { try { d = new Date(val).toISOString().substring(0, 10); } catch(e){} }
    return '<input class="input-text" id="' + esc(id) + '" type="date" value="' + esc(d) + '">';
}
function rolesSelect(id, val) {
    var opts = [{ value: '', label: '— выберите роль —' }].concat(
        cachedRoles.map(function(r) { return { value: r.role_id, label: r.name + ' (' + r.code + ')' }; })
    );
    return sel(id, opts, val || '');
}

// ---- Modal ----
var _onSave = null;

function openModal(title, bodyHtml, onSave) {
    el('modalTitle').textContent = title;
    el('modalBody').innerHTML = bodyHtml;
    el('modalError').textContent = '';
    el('modalSave').disabled = false;
    _onSave = onSave;
    el('modal').classList.remove('hidden');
}

function closeModal() {
    el('modal').classList.add('hidden');
    _onSave = null;
}

async function doModalSave() {
    if (!_onSave) return;
    el('modalSave').disabled = true;
    el('modalError').textContent = '';
    try {
        await _onSave();
        closeModal();
        reload();
    } catch(e) {
        el('modalError').textContent = e.message || 'Ошибка';
        el('modalSave').disabled = false;
    }
}

function reload() {
    if (activeSection) loadSection(activeSection, pageOffset, activeFilters);
}

// ---- Status ----
function showStatus(msg, isErr) {
    var s = el('statusMsg');
    s.textContent = msg;
    s.style.display = msg ? '' : 'none';
    s.style.color = isErr ? 'var(--danger-color)' : 'var(--text-secondary)';
}

// ---- Sub-table helpers ----
function renderSubContent(container, rows, parentRow, subCfg) {
    var h = '<div style="padding:10px 16px">';
    h += '<div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:8px">';
    h += '<span style="font-weight:600;font-size:0.85rem;color:var(--text-secondary)">' + esc(subCfg.label) + '</span>';
    if (subCfg.createLabel) h += '<button class="btn btn-xs btn-primary sub-create">+ ' + esc(subCfg.createLabel) + '</button>';
    h += '</div>';

    if (!rows || !rows.length) {
        h += '<div style="font-style:italic;color:var(--text-secondary);font-size:0.85rem;padding:4px 0">Нет данных</div>';
    } else {
        h += '<table class="data-table" style="font-size:0.85rem"><thead><tr>';
        subCfg.cols.forEach(function(c) { h += '<th>' + esc(c.label) + '</th>'; });
        if (subCfg.deleteUrl) h += '<th style="width:1px"></th>';
        h += '</tr></thead><tbody>';
        rows.forEach(function(row) {
            h += '<tr>';
            subCfg.cols.forEach(function(c) {
                var v = row[c.key];
                if (c.fmt) v = c.fmt(v, row);
                else v = esc(nullVal(v) || (v != null ? String(v) : ''));
                h += '<td>' + v + '</td>';
            });
            if (subCfg.deleteUrl) {
                h += '<td><button class="btn btn-xs btn-danger sub-del" data-del="' + esc(String(subCfg.rowId(row))) + '">✕</button></td>';
            }
            h += '</tr>';
        });
        h += '</tbody></table>';
    }
    h += '</div>';
    container.innerHTML = h;

    var createBtn = container.querySelector('.sub-create');
    if (createBtn && subCfg.openCreate) {
        createBtn.addEventListener('click', function() {
            subCfg.openCreate(parentRow, function() { loadSubContent(container, parentRow, subCfg); });
        });
    }
    if (subCfg.deleteUrl) {
        container.querySelectorAll('.sub-del').forEach(function(btn) {
            btn.addEventListener('click', function() {
                if (!confirm('Удалить?')) return;
                var id = btn.getAttribute('data-del');
                apiFetch(subCfg.deleteUrl(id, parentRow), { method: 'DELETE' })
                    .then(function() { loadSubContent(container, parentRow, subCfg); })
                    .catch(function(e) { alert('Ошибка: ' + e.message); });
            });
        });
    }
}

async function loadSubContent(container, parentRow, subCfg) {
    container.innerHTML = '<div style="padding:10px 16px;color:var(--text-secondary);font-style:italic">Загрузка...</div>';
    try {
        var data = await apiFetch(subCfg.url(parentRow));
        var rows = Array.isArray(data) ? data : (data && data.data ? data.data : []);
        renderSubContent(container, rows, parentRow, subCfg);
    } catch(e) {
        container.innerHTML = '<div style="padding:10px 16px;color:var(--danger-color)">Ошибка: ' + esc(e.message) + '</div>';
    }
}

// ---- Table rendering ----
function renderTable(rows, cols, actions, subCfg) {
    expandedSet = {};
    var container = el('activeTable');
    if (!rows || !rows.length) {
        container.innerHTML = '<div class="empty-state">Нет данных</div>';
        return;
    }
    var hasActions = actions && actions.length;
    var colSpan = cols.length + (hasActions ? 1 : 0) + (subCfg ? 1 : 0);

    var h = '<table class="data-table"><thead><tr>';
    if (subCfg) h += '<th style="width:28px"></th>';
    cols.forEach(function(c) { h += '<th>' + esc(c.label) + '</th>'; });
    if (hasActions) h += '<th></th>';
    h += '</tr></thead><tbody>';

    rows.forEach(function(row, i) {
        h += '<tr class="data-row" id="dr-' + i + '">';
        if (subCfg) {
            h += '<td class="expand-toggle" id="et-' + i + '" style="text-align:center;cursor:pointer;user-select:none">▶</td>';
        }
        cols.forEach(function(c) {
            var v = row[c.key];
            if (c.fmt) v = c.fmt(v, row);
            else {
                // handle sql.Null* types
                var raw = (v && typeof v === 'object' && 'Valid' in v) ? nullVal(v) : v;
                v = esc(raw != null ? String(raw) : '');
            }
            h += '<td>' + v + '</td>';
        });
        if (hasActions) {
            h += '<td class="row-actions">';
            actions.forEach(function(a) {
                h += '<button class="btn btn-xs action-btn" data-action="' + esc(a.id) + '" data-i="' + i + '">' + esc(a.label) + '</button>';
            });
            h += '</td>';
        }
        h += '</tr>';
        if (subCfg) {
            h += '<tr class="sub-row" id="sr-' + i + '" style="display:none"><td colspan="' + colSpan + '" class="sub-cell"><div class="sub-content" id="sc-' + i + '"></div></td></tr>';
        }
    });
    h += '</tbody></table>';
    container.innerHTML = h;

    // wire expand toggles
    if (subCfg) {
        rows.forEach(function(row, i) {
            var td = el('et-' + i);
            if (!td) return;
            td.addEventListener('click', function() {
                var sr = el('sr-' + i);
                var sc = el('sc-' + i);
                if (!sr || !sc) return;
                if (expandedSet[i]) {
                    sr.style.display = 'none';
                    td.textContent = '▶';
                    el('dr-' + i).classList.remove('row-expanded');
                    delete expandedSet[i];
                } else {
                    sr.style.display = '';
                    td.textContent = '▼';
                    el('dr-' + i).classList.add('row-expanded');
                    expandedSet[i] = true;
                    loadSubContent(sc, row, subCfg);
                }
            });
        });
    }

    // wire action buttons
    if (hasActions) {
        container.querySelectorAll('.action-btn').forEach(function(btn) {
            btn.addEventListener('click', function() {
                var actionId = btn.getAttribute('data-action');
                var i = parseInt(btn.getAttribute('data-i'));
                var a = actions.find(function(x) { return x.id === actionId; });
                if (a) a.handler(rows[i]);
            });
        });
    }
}

// ---- Pagination ----
function renderPagination(meta, section, filters) {
    var p = el('pagination');
    if (!meta) { p.style.display = 'none'; return; }
    var from = meta.offset + 1;
    var to = Math.min(meta.offset + meta.limit, meta.total);
    if (meta.total === 0) { p.style.display = 'none'; return; }
    p.style.display = '';
    el('pageInfo').textContent = 'Показано ' + from + '–' + to + ' из ' + meta.total;
    el('btnPrev').disabled = meta.offset <= 0;
    el('btnNext').disabled = to >= meta.total;
    el('btnPrev').onclick = function() {
        pageOffset = Math.max(0, meta.offset - pageLimit);
        loadSection(section, pageOffset, filters);
    };
    el('btnNext').onclick = function() {
        pageOffset = meta.offset + pageLimit;
        loadSection(section, pageOffset, filters);
    };
}

// ========================
// SECTION DEFINITIONS
// ========================

// Steps sub-table (for test cases)
var stepsSubCfg = {
    label: 'Шаги', createLabel: 'Добавить шаг',
    url: function(row) { return '/api/test-cases/' + row.test_case_id + '/steps'; },
    cols: [
        { key: 'step_order',      label: '#' },
        { key: 'action_text',     label: 'Действие' },
        { key: 'expected_result', label: 'Ожидаемый результат' }
    ],
    rowId: function(row) { return row.step_id; },
    deleteUrl: function(id, p) { return '/api/test-cases/' + p.test_case_id + '/steps/' + id; },
    openCreate: function(parentRow, refresh) {
        openModal('Добавить шаг',
            fg('Порядок', inpNum('f_order', '')) +
            fg('Действие', ta('f_action', '')) +
            fg('Ожидаемый результат', ta('f_expected', '')),
            async function() {
                await apiFetch('/api/test-cases/' + parentRow.test_case_id + '/steps', {
                    method: 'POST',
                    body: JSON.stringify({ order: fvi('f_order'), action: fv('f_action'), expected: fv('f_expected') })
                });
                refresh();
            }
        );
    }
};

// Run items sub-table
var runItemsSubCfg = {
    label: 'Тест-кейсы прогона', createLabel: 'Добавить кейс',
    url: function(row) { return '/api/runs/' + row.test_run_id + '/items'; },
    cols: [
        { key: 'run_item_id',     label: 'ID' },
        { key: 'test_case_id',    label: 'Case ID' },
        { key: 'execution_order', label: 'Порядок' }
    ],
    rowId: function(row) { return row.run_item_id; },
    deleteUrl: function(id, p) { return '/api/runs/' + p.test_run_id + '/items/' + id; },
    openCreate: function(parentRow, refresh) {
        openModal('Добавить тест-кейс',
            fg('Test Case ID', inpNum('f_case_id', '')) +
            fg('Порядок', inpNum('f_order', '')),
            async function() {
                await apiFetch('/api/runs/' + parentRow.test_run_id + '/items', {
                    method: 'POST',
                    body: JSON.stringify({ test_case_id: fvi('f_case_id'), order: fvi('f_order') })
                });
                refresh();
            }
        );
    }
};

// Artifacts sub-table
var artifactsSubCfg = {
    label: 'Артефакты', createLabel: 'Добавить артефакт',
    url: function(row) { return '/api/results/' + row.test_result_id + '/artifacts'; },
    cols: [
        { key: 'artifact_id', label: 'ID' },
        { key: 'kind',        label: 'Тип' },
        { key: 'file_path',   label: 'Путь' },
        { key: 'created_at',  label: 'Добавлен', fmt: fmtDate }
    ],
    rowId: function(row) { return row.artifact_id; },
    deleteUrl: null,
    openCreate: function(parentRow, refresh) {
        openModal('Добавить артефакт',
            fg('Тип', sel('f_kind', ['LOG', 'SCREENSHOT', 'REPORT', 'OTHER'], '')) +
            fg('Путь к файлу', inp('f_path', '', '/path/to/file')),
            async function() {
                await apiFetch('/api/results/' + parentRow.test_result_id + '/artifacts', {
                    method: 'POST',
                    body: JSON.stringify({ kind: fv('f_kind'), path: fv('f_path') })
                });
                refresh();
            }
        );
    }
};

// Autotest versions sub-table
var versionsSubCfg = {
    label: 'Версии', createLabel: 'Добавить версию',
    url: function(row) { return '/api/autotests/' + row.autotest_id + '/versions'; },
    cols: [
        { key: 'version_id',     label: 'ID' },
        { key: 'version_string', label: 'Версия' },
        { key: 'commit_hash',    label: 'Commit', fmt: nullVal },
        { key: 'created_at',     label: 'Дата', fmt: fmtDate }
    ],
    rowId: function(row) { return row.version_id; },
    deleteUrl: null,
    openCreate: function(parentRow, refresh) {
        openModal('Добавить версию',
            fg('Версия (напр. 1.2.3)', inp('f_ver', '', '1.2.3')) +
            fg('Commit Hash', inp('f_commit', '')),
            async function() {
                await apiFetch('/api/autotests/' + parentRow.autotest_id + '/versions', {
                    method: 'POST',
                    body: JSON.stringify({ version_string: fv('f_ver'), commit_hash: fv('f_commit') })
                });
                refresh();
            }
        );
    }
};

// ---- Per-section form functions ----
function testPlanForm(row) {
    return fg('Название *', inp('f_name', row && row.test_plan_name)) +
           fg('Описание', ta('f_desc', row && nullVal(row.description))) +
           fg('Priority ID', inpNum('f_prio', '')) +
           fg('Owner User ID', inpNum('f_owner', '')) +
           fg('Статус', sel('f_status', ['DRAFT', 'ACTIVE', 'CLOSED'], row && row.status || 'DRAFT')) +
           fg('Дата начала', dateInp('f_start', row && row.start_date)) +
           fg('Дата окончания', dateInp('f_end', row && row.end_date)) +
           fg('Критерии приёмки', ta('f_acceptance', ''));
}
function testPlanDTO() {
    return { name: fv('f_name'), description: fv('f_desc'), priority_id: fvi('f_prio'),
             owner_id: fvi('f_owner'), status: fv('f_status'), acceptance_criteria: fv('f_acceptance'),
             start_date: fv('f_start') ? new Date(fv('f_start')).toISOString() : null,
             end_date:   fv('f_end')   ? new Date(fv('f_end')).toISOString()   : null };
}

function testCaseForm(row) {
    return fg('Название *', inp('f_name', row && row.test_case_name)) +
           fg('Описание', ta('f_desc', row && nullVal(row.description))) +
           fg('Priority ID', inpNum('f_prio', '')) +
           fg('Owner User ID', inpNum('f_owner', '')) +
           fg('', chk('f_auto', row && row.is_automated, 'Автоматизированный'));
}
function testCaseDTO() {
    return { name: fv('f_name'), description: fv('f_desc'),
             priority_id: fvi('f_prio'), owner_id: fvi('f_owner'), is_automated: fvb('f_auto') };
}

function runForm() {
    return fg('Название *', inp('f_name', '')) +
           fg('Test Plan ID', inpNum('f_plan', '')) +
           fg('Version ID', inpNum('f_ver', '')) +
           fg('Environment ID', inpNum('f_env', '')) +
           fg('Tool Config ID', inpNum('f_tool', ''));
}
function runDTO() {
    return { name: fv('f_name'), plan_id: fvi('f_plan'), version_id: fvi('f_ver'),
             env_id: fvi('f_env'), tool_config_id: fvi('f_tool') };
}

function resultForm(row) {
    return fg('Run Item ID', inpNum('f_item', '')) +
           fg('Status ID', inpNum('f_status', '')) +
           fg('Краткое описание', ta('f_summary', row && row.summary));
}
function resultDTO() {
    return { run_item_id: fvi('f_item'), status_id: fvi('f_status'), summary: fv('f_summary') };
}

function autotestForm(row) {
    return fg('Название *', inp('f_name', row && row.autotest_name)) +
           fg('Описание', ta('f_desc', row && nullVal(row.description))) +
           fg('Test Case ID', inpNum('f_case', '')) +
           fg('Owner User ID', inpNum('f_owner', '')) +
           fg('', chk('f_active', row ? row.is_active : true, 'Активен'));
}
function autotestDTO() {
    return { name: fv('f_name'), description: fv('f_desc'), test_case_id: fvi('f_case'),
             owner_id: fvi('f_owner'), is_active: fvb('f_active') };
}

function userCreateForm() {
    return fg('Логин *', inp('f_user', '')) +
           fg('ФИО', inp('f_fn', '')) +
           fg('Email', inp('f_email', '', 'user@example.com', 'email')) +
           fg('Пароль *', inp('f_pass', '', '', 'password')) +
           fg('Начальная роль', rolesSelect('f_role', ''));
}

// ---- Sections config ----
var SECTIONS = [
    {
        id: 'test-plans', label: 'Test Plans', url: '/api/test-plans', paged: true,
        cols: [
            { key: 'test_plan_id',   label: 'ID' },
            { key: 'test_plan_name', label: 'Название' },
            { key: 'status',         label: 'Статус', fmt: statusBadge },
            { key: 'priority',       label: 'Приоритет' },
            { key: 'start_date',     label: 'Начало', fmt: fmtDate },
            { key: 'end_date',       label: 'Конец', fmt: fmtDate },
            { key: 'owner',          label: 'Владелец', fmt: nullVal }
        ],
        filters: [
            { name: 'status', type: 'select', label: 'Статус',
              opts: [{ value: '', label: 'Все' }, 'DRAFT', 'ACTIVE', 'CLOSED'] }
        ],
        idKey: 'test_plan_id',
        createUrl: '/api/test-plans',
        editUrl:   function(r) { return '/api/test-plans/' + r.test_plan_id; },
        deleteUrl: function(r) { return '/api/test-plans/' + r.test_plan_id; },
        formHtml: testPlanForm, buildDTO: testPlanDTO
    },
    {
        id: 'test-cases', label: 'Test Cases', url: '/api/test-cases', paged: true,
        cols: [
            { key: 'test_case_id',   label: 'ID' },
            { key: 'test_case_name', label: 'Название' },
            { key: 'priority',       label: 'Приоритет' },
            { key: 'is_automated',   label: 'Авто', fmt: boolStr },
            { key: 'is_active',      label: 'Активен', fmt: boolStr },
            { key: 'owner',          label: 'Владелец', fmt: nullVal }
        ],
        filters: [
            { name: 'is_active',    type: 'select', label: 'Активен',
              opts: [{ value: '', label: 'Все' }, { value: 'true', label: 'Да' }, { value: 'false', label: 'Нет' }] },
            { name: 'is_automated', type: 'select', label: 'Авто',
              opts: [{ value: '', label: 'Все' }, { value: 'true', label: 'Да' }, { value: 'false', label: 'Нет' }] }
        ],
        idKey: 'test_case_id',
        createUrl: '/api/test-cases',
        editUrl:   function(r) { return '/api/test-cases/' + r.test_case_id; },
        deleteUrl: function(r) { return '/api/test-cases/' + r.test_case_id; },
        formHtml: testCaseForm, buildDTO: testCaseDTO,
        sub: stepsSubCfg
    },
    {
        id: 'runs', label: 'Runs', url: '/api/runs', paged: true,
        cols: [
            { key: 'test_run_id',    label: 'ID' },
            { key: 'run_name',       label: 'Название' },
            { key: 'plan_name',      label: 'План' },
            { key: 'status',         label: 'Статус', fmt: statusBadge },
            { key: 'passed_tests',   label: 'Пройдено' },
            { key: 'failed_tests',   label: 'Провалено' },
            { key: 'blocked_tests',  label: 'Блок.' }
        ],
        filters: [
            { name: 'status', type: 'select', label: 'Статус',
              opts: [{ value: '', label: 'Все' }, 'PLANNED', 'IN_PROGRESS', 'COMPLETED', 'ABORTED'] }
        ],
        idKey: 'test_run_id',
        createUrl: '/api/runs',
        editUrl: null,
        deleteUrl: function(r) { return '/api/runs/' + r.test_run_id; },
        formHtml: runForm, buildDTO: runDTO,
        extraActions: [{
            id: 'upd-status', label: 'Статус',
            handler: function(row) {
                openModal('Изменить статус прогона #' + row.test_run_id,
                    fg('Новый статус', sel('f_status', ['PLANNED', 'IN_PROGRESS', 'COMPLETED', 'ABORTED'], row.status)),
                    async function() {
                        await apiFetch('/api/runs/' + row.test_run_id + '/status', {
                            method: 'PATCH', body: JSON.stringify({ status: fv('f_status') })
                        });
                    }
                );
            }
        }],
        sub: runItemsSubCfg
    },
    {
        id: 'results', label: 'Results', url: '/api/results', paged: true,
        cols: [
            { key: 'test_result_id',  label: 'ID' },
            { key: 'test_case_name',  label: 'Тест-кейс' },
            { key: 'status',          label: 'Статус', fmt: statusBadge },
            { key: 'execution_date',  label: 'Дата', fmt: fmtDate },
            { key: 'test_run_name',   label: 'Прогон' },
            { key: 'executor',        label: 'Исполнитель' }
        ],
        filters: [
            { name: 'status', type: 'select', label: 'Статус',
              opts: [{ value: '', label: 'Все' }, 'PASSED', 'FAILED', 'IN_PROGRESS', 'BLOCKED'] }
        ],
        idKey: 'test_result_id',
        createUrl: '/api/results',
        editUrl:   function(r) { return '/api/results/' + r.test_result_id; },
        deleteUrl: function(r) { return '/api/results/' + r.test_result_id; },
        formHtml: resultForm, buildDTO: resultDTO,
        sub: artifactsSubCfg
    },
    {
        id: 'autotests', label: 'Autotests', url: '/api/autotests', paged: true,
        cols: [
            { key: 'autotest_id',    label: 'ID' },
            { key: 'autotest_name',  label: 'Название' },
            { key: 'version_string', label: 'Версия' },
            { key: 'is_active',      label: 'Активен', fmt: boolStr },
            { key: 'commit_hash',    label: 'Commit', fmt: nullVal },
            { key: 'author',         label: 'Автор', fmt: nullVal }
        ],
        filters: [
            { name: 'is_active', type: 'select', label: 'Активен',
              opts: [{ value: '', label: 'Все' }, { value: 'true', label: 'Да' }, { value: 'false', label: 'Нет' }] }
        ],
        idKey: 'autotest_id',
        createUrl: '/api/autotests',
        editUrl:   function(r) { return '/api/autotests/' + r.autotest_id; },
        deleteUrl: function(r) { return '/api/autotests/' + r.autotest_id; },
        formHtml: autotestForm, buildDTO: autotestDTO,
        sub: versionsSubCfg
    },
    {
        id: 'environments', label: 'Environments', url: '/api/environments', paged: false,
        cols: [
            { key: 'env_config_id',    label: 'ID' },
            { key: 'environment_name', label: 'Название' },
            { key: 'description',      label: 'Описание', fmt: nullVal },
            { key: 'is_active',        label: 'Активна', fmt: boolStr },
            { key: 'params_count',     label: 'Параметры' }
        ],
        filters: [], idKey: 'env_config_id',
        createUrl: null, editUrl: null, deleteUrl: null
    },
    {
        id: 'reports', label: 'Reports', url: '/api/reports', paged: true,
        cols: [
            { key: 'report_id',      label: 'ID' },
            { key: 'report_name',    label: 'Название' },
            { key: 'template',       label: 'Шаблон' },
            { key: 'creation_date',  label: 'Дата', fmt: fmtDate },
            { key: 'author',         label: 'Автор', fmt: nullVal }
        ],
        filters: [], idKey: 'report_id',
        createUrl: null, editUrl: null, deleteUrl: null
    },
    {
        id: 'admin-users', label: 'Пользователи', url: '/api/admin/users', paged: true, adminOnly: true,
        cols: [
            { key: 'user_id',   label: 'ID' },
            { key: 'username',  label: 'Логин' },
            { key: 'full_name', label: 'ФИО', fmt: nullVal },
            { key: 'email',     label: 'Email', fmt: nullVal },
            { key: 'is_active', label: 'Активен', fmt: boolStr },
            { key: 'roles',     label: 'Роли', fmt: nullVal }
        ],
        filters: [
            { name: 'is_active', type: 'select', label: 'Активен',
              opts: [{ value: '', label: 'Все' }, { value: 'true', label: 'Да' }, { value: 'false', label: 'Нет' }] }
        ],
        idKey: 'user_id',
        createUrl: '/api/admin/users',
        editUrl: null, deleteUrl: null,
        formHtml: userCreateForm,
        buildDTO: function() {
            return { username: fv('f_user'), full_name: fv('f_fn'), email: fv('f_email'),
                     password: fv('f_pass'), role_id: fvi('f_role') };
        },
        extraActions: [
            {
                id: 'toggle', label: 'Вкл/Выкл',
                handler: function(row) {
                    if (!confirm((row.is_active ? 'Деактивировать' : 'Активировать') + ' ' + row.username + '?')) return;
                    apiFetch('/api/admin/users/' + row.user_id + '/active', {
                        method: 'PATCH', body: JSON.stringify({ is_active: !row.is_active })
                    }).then(reload).catch(function(e) { alert('Ошибка: ' + e.message); });
                }
            },
            {
                id: 'role', label: 'Роль +',
                handler: function(row) {
                    openModal('Назначить роль: ' + row.username,
                        fg('Роль', rolesSelect('f_role', '')),
                        async function() {
                            await apiFetch('/api/admin/users/' + row.user_id + '/roles', {
                                method: 'POST', body: JSON.stringify({ role_id: fvi('f_role') })
                            });
                        }
                    );
                }
            }
        ]
    },
    {
        id: 'admin-audit', label: 'Журнал аудита', url: '/api/admin/audit-log', paged: true, adminOnly: true,
        cols: [
            { key: 'audit_id',   label: 'ID' },
            { key: 'table_name', label: 'Таблица' },
            { key: 'operation',  label: 'Операция', fmt: statusBadge },
            { key: 'record_id',  label: 'Record ID' },
            { key: 'user_id',    label: 'User ID', fmt: nullVal },
            { key: 'changed_at', label: 'Время', fmt: fmtDT }
        ],
        filters: [
            { name: 'table_name', type: 'text',   label: 'Таблица',   placeholder: 'test_plans...' },
            { name: 'operation',  type: 'select',  label: 'Операция',
              opts: [{ value: '', label: 'Все' }, 'INSERT', 'UPDATE', 'DELETE'] }
        ],
        idKey: 'audit_id', createUrl: null, editUrl: null, deleteUrl: null
    }
];

// ---- Sidebar ----
function renderSidebar() {
    var root = el('tablesList');
    root.innerHTML = '';
    SECTIONS.forEach(function(s) {
        if (s.adminOnly && (!currentUser || currentUser.db_role !== 'ADMIN')) return;
        var div = document.createElement('div');
        div.className = 'nav-item' + (activeSection && activeSection.id === s.id ? ' active' : '');
        div.textContent = s.label;
        div.onclick = function() { navigateTo(s); };
        root.appendChild(div);
    });
}

// ---- Toolbar ----
function renderToolbar(section) {
    el('toolbar').style.display = '';
    var filtersEl = el('toolbarFilters');
    filtersEl.innerHTML = '';

    (section.filters || []).forEach(function(f) {
        var wrap = document.createElement('div');
        wrap.style.cssText = 'display:flex;align-items:center;gap:6px';

        var lbl = document.createElement('span');
        lbl.textContent = f.label;
        lbl.style.cssText = 'font-size:0.85rem;color:var(--text-secondary)';
        wrap.appendChild(lbl);

        var ctrl;
        if (f.type === 'select') {
            ctrl = document.createElement('select');
            ctrl.className = 'input-select input-xs';
            (f.opts || []).forEach(function(o) {
                var opt = document.createElement('option');
                var v = (o && o.value !== undefined) ? String(o.value) : String(o);
                opt.value = v;
                opt.textContent = (o && o.label !== undefined) ? o.label : String(o);
                if (activeFilters[f.name] !== undefined && String(activeFilters[f.name]) === v) opt.selected = true;
                ctrl.appendChild(opt);
            });
        } else {
            ctrl = document.createElement('input');
            ctrl.className = 'input-text input-xs';
            ctrl.placeholder = f.placeholder || '';
            ctrl.value = activeFilters[f.name] || '';
        }

        ctrl.addEventListener('change', function() {
            activeFilters[f.name] = ctrl.value;
            pageOffset = 0;
            loadSection(section, 0, activeFilters);
        });
        wrap.appendChild(ctrl);
        filtersEl.appendChild(wrap);
    });

    var btnCreate = el('btnCreate');
    if (section.createUrl && section.formHtml) {
        btnCreate.style.display = '';
        btnCreate.onclick = function() {
            openModal('Создать: ' + section.label, section.formHtml(null), async function() {
                var dto = section.buildDTO(null);
                await apiFetch(section.createUrl, { method: 'POST', body: JSON.stringify(dto) });
            });
        };
    } else {
        btnCreate.style.display = 'none';
    }
}

// ---- Navigate / Load ----
function navigateTo(section) {
    activeSection = section;
    pageOffset = 0;
    activeFilters = {};
    expandedSet = {};
    renderSidebar();
    el('activeTitle').textContent = section.label;
    renderToolbar(section);
    loadSection(section, 0, {});
}

async function loadSection(section, offset, filters) {
    showStatus('Загрузка...', false);
    el('activeTable').innerHTML = '';
    el('pagination').style.display = 'none';

    var url = section.url;
    var params = [];
    if (section.paged) {
        params.push('limit=' + pageLimit);
        params.push('offset=' + offset);
    }
    Object.keys(filters).forEach(function(k) {
        var v = filters[k];
        if (v !== '' && v !== null && v !== undefined) {
            params.push(encodeURIComponent(k) + '=' + encodeURIComponent(v));
        }
    });
    if (params.length) url += '?' + params.join('&');

    try {
        var resp = await apiFetch(url);
        showStatus('', false);

        var rows, meta;
        if (section.paged && resp && resp.data !== undefined) {
            rows = resp.data || [];
            meta = resp.meta;
        } else {
            rows = Array.isArray(resp) ? resp : [];
            meta = null;
        }

        // Build actions
        var actions = [];
        if (section.editUrl && section.formHtml) {
            actions.push({ id: 'edit', label: 'Ред.', handler: function(row) {
                openModal('Редактировать #' + row[section.idKey], section.formHtml(row), async function() {
                    await apiFetch(section.editUrl(row), { method: 'PUT', body: JSON.stringify(section.buildDTO(row)) });
                });
            }});
        }
        if (section.deleteUrl) {
            actions.push({ id: 'del', label: 'Удал.', handler: function(row) {
                if (!confirm('Удалить #' + row[section.idKey] + '?')) return;
                apiFetch(section.deleteUrl(row), { method: 'DELETE' })
                    .then(reload)
                    .catch(function(e) { showStatus('Ошибка: ' + e.message, true); });
            }});
        }
        if (section.extraActions) {
            section.extraActions.forEach(function(a) { actions.push(a); });
        }

        renderTable(rows, section.cols, actions, section.sub || null);
        renderPagination(meta, section, filters);

    } catch(e) {
        if (e.message === '401') return;
        if (e.status === 403) {
            showStatus('Доступ запрещён (403)', true);
        } else {
            showStatus('Ошибка: ' + e.message, true);
        }
    }
}

// ---- Logout ----
async function logout() {
    try { await apiFetch('/api/auth/logout', { method: 'POST' }); } catch(e) {}
    redirectLogin();
}

// ---- Init ----
async function init() {
    if (!getToken()) { redirectLogin(); return; }

    try { currentUser = await apiFetch('/api/auth/me'); } catch(e) {
        if (e.message !== '401') showStatus('Ошибка соединения: ' + e.message, true);
        return;
    }

    el('currentUser').textContent = currentUser.username + ' (' + currentUser.db_role + ')';
    el('logoutLink').addEventListener('click', function(e) { e.preventDefault(); logout(); });

    if (currentUser.db_role === 'ADMIN') {
        try { cachedRoles = await apiFetch('/api/admin/roles') || []; } catch(e) { cachedRoles = []; }
    }

    renderSidebar();

    var first = SECTIONS.find(function(s) {
        return !s.adminOnly || currentUser.db_role === 'ADMIN';
    });
    if (first) navigateTo(first);
}

init();
