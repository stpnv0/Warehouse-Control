'use strict';

/* ═══════════════════════════════════════════════════════════════════════
   State
   ═══════════════════════════════════════════════════════════════════════ */
const state = {
    token: localStorage.getItem('wc_token') || '',
    user: JSON.parse(localStorage.getItem('wc_user') || 'null'),
    items: [],
    selectedItemId: null,
    currentPage: 1,
    auditPage: 1,
    pageSize: 20,
};

/* ═══════════════════════════════════════════════════════════════════════
   Utility Helpers
   ═══════════════════════════════════════════════════════════════════════ */
const $ = (sel) => document.querySelector(sel);
const $$ = (sel) => document.querySelectorAll(sel);

function escHtml(s) {
    const d = document.createElement('div');
    d.textContent = s;
    return d.innerHTML;
}

function showToast(msg, type = 'success') {
    const container = $('#toastContainer');
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.textContent = msg;
    container.appendChild(toast);
    setTimeout(() => {
        toast.style.opacity = '0';
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

function formatDate(iso) {
    if (!iso) return '—';
    const d = new Date(iso);
    return d.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })
        + ' ' + d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });
}

function formatPrice(val) {
    const num = parseFloat(val);
    if (isNaN(num)) return '—';
    return '$' + num.toFixed(2);
}

function actionBadge(action) {
    const cls = { INSERT: 'badge-insert', UPDATE: 'badge-update', DELETE: 'badge-delete' };
    return `<span class="badge ${cls[action] || ''}">${escHtml(action)}</span>`;
}

/* ═══════════════════════════════════════════════════════════════════════
   API Helper
   ═══════════════════════════════════════════════════════════════════════ */
async function api(method, path, body) {
    const opts = { method, headers: {} };
    if (state.token) opts.headers['Authorization'] = 'Bearer ' + state.token;
    if (body !== undefined) {
        opts.headers['Content-Type'] = 'application/json';
        opts.body = JSON.stringify(body);
    }

    const res = await fetch(path, opts);

    if (res.status === 204) return null;
    if (res.status === 401) {
        logout();
        throw new Error('Session expired');
    }

    const data = await res.json().catch(() => null);
    if (!res.ok) {
        throw new Error(data?.error || `Request failed (${res.status})`);
    }
    return data;
}

/* ═══════════════════════════════════════════════════════════════════════
   Auth
   ═══════════════════════════════════════════════════════════════════════ */
async function loadUsers() {
    try {
        const users = await api('GET', '/api/auth/users');
        const sel = $('#userSelect');
        sel.innerHTML = '<option value="">— Select a user —</option>';
        (users || []).forEach(u => {
            const o = document.createElement('option');
            o.value = u.username;
            o.textContent = `${u.username} (${u.role})`;
            sel.appendChild(o);
        });
    } catch (e) {
        showToast('Failed to load users: ' + e.message, 'error');
    }
}

async function login(username, password) {
    try {
        const data = await api('POST', '/api/auth/login', { username, password });
        state.token = data.token;
        state.user = data.user;
        localStorage.setItem('wc_token', data.token);
        localStorage.setItem('wc_user', JSON.stringify(data.user));
        enterApp();
    } catch (e) {
        showToast('Login failed: ' + e.message, 'error');
    }
}

function logout() {
    state.token = '';
    state.user = null;
    state.items = [];
    state.selectedItemId = null;
    localStorage.removeItem('wc_token');
    localStorage.removeItem('wc_user');
    $('#loginScreen').style.display = '';
    $('#appShell').classList.remove('active');
    loadUsers();
}

function enterApp() {
    $('#loginScreen').style.display = 'none';
    $('#appShell').classList.add('active');

    const role = state.user.role;
    $('#headerUsername').textContent = state.user.username;

    const badge = $('#headerRole');
    badge.className = 'badge';
    badge.classList.add(
        role === 'admin' ? 'badge-admin' : role === 'manager' ? 'badge-manager' : 'badge-viewer'
    );
    badge.textContent = role;

    // Permissions
    const canWrite = role === 'admin' || role === 'manager';
    const canAudit = role === 'admin' || role === 'manager';
    $('#addItemBtn').style.display = canWrite ? '' : 'none';
    $('#historyTab').style.display = canAudit ? '' : 'none';

    loadItems();
}

/* ═══════════════════════════════════════════════════════════════════════
   Items CRUD
   ═══════════════════════════════════════════════════════════════════════ */
async function loadItems(page = 1) {
    state.currentPage = page;
    const search = $('#searchInput').value.trim();
    let url = `/api/items?page=${page}&page_size=${state.pageSize}`;
    if (search) url += `&search=${encodeURIComponent(search)}`;

    try {
        const data = await api('GET', url);
        state.items = data.items || [];
        renderItems();
        renderPagination('itemsPagination', data, loadItems);
    } catch (e) {
        showToast('Failed to load items: ' + e.message, 'error');
    }
}

function renderItems() {
    const tbody = $('#itemsBody');
    const role = state.user.role;
    const canEdit = role === 'admin' || role === 'manager';
    const canDelete = role === 'admin';
    const canAudit = role === 'admin' || role === 'manager';

    if (!state.items.length) {
        tbody.innerHTML = '<tr><td colspan="7" class="empty-state">No items found. Add one to get started.</td></tr>';
        return;
    }

    tbody.innerHTML = state.items.map(item => {
        const selected = item.id === state.selectedItemId ? ' selected' : '';
        const clickable = canAudit ? ' clickable' : '';
        return `<tr class="${clickable}${selected}" data-id="${item.id}">
            <td>${escHtml(item.name)}</td>
            <td><code>${escHtml(item.sku)}</code></td>
            <td>${item.quantity}</td>
            <td>${formatPrice(item.price)}</td>
            <td>${escHtml(item.location || '—')}</td>
            <td>${formatDate(item.updated_at)}</td>
            <td><div class="actions-cell">
                ${canEdit ? `<button class="btn btn-outline btn-sm edit-btn" data-id="${item.id}">Edit</button>` : ''}
                ${canDelete ? `<button class="btn btn-danger btn-sm delete-btn" data-id="${item.id}">Delete</button>` : ''}
            </div></td>
        </tr>`;
    }).join('');
}

/* ─── Item table click delegation ─────────────────────────────────── */
$('#itemsBody').addEventListener('click', e => {
    const editBtn = e.target.closest('.edit-btn');
    if (editBtn) {
        e.stopPropagation();
        openEditModal(editBtn.dataset.id);
        return;
    }

    const delBtn = e.target.closest('.delete-btn');
    if (delBtn) {
        e.stopPropagation();
        openDeleteConfirm(delBtn.dataset.id);
        return;
    }

    const row = e.target.closest('tr.clickable');
    if (row) {
        const id = row.dataset.id;
        state.selectedItemId = id;
        renderItems();
        loadItemHistory(id);
    }
});

/* ─── Add / Edit Modal ────────────────────────────────────────────── */
function openAddModal() {
    $('#itemEditId').value = '';
    $('#itemModalTitle').textContent = 'Add Item';
    $('#fieldName').value = '';
    $('#fieldSku').value = '';
    $('#fieldQuantity').value = '';
    $('#fieldPrice').value = '';
    $('#fieldLocation').value = '';
    $('#itemModal').style.display = '';
}

async function openEditModal(id) {
    try {
        const item = await api('GET', `/api/items/${id}`);
        $('#itemEditId').value = item.id;
        $('#itemModalTitle').textContent = 'Edit Item';
        $('#fieldName').value = item.name;
        $('#fieldSku').value = item.sku;
        $('#fieldQuantity').value = item.quantity;
        $('#fieldPrice').value = item.price;
        $('#fieldLocation').value = item.location || '';
        $('#itemModal').style.display = '';
    } catch (e) {
        showToast('Failed to load item: ' + e.message, 'error');
    }
}

function closeItemModal() {
    $('#itemModal').style.display = 'none';
}

async function saveItem() {
    const id = $('#itemEditId').value;
    const payload = {
        name: $('#fieldName').value.trim(),
        sku: $('#fieldSku').value.trim(),
        quantity: parseInt($('#fieldQuantity').value, 10) || 0,
        price: $('#fieldPrice').value,
        location: $('#fieldLocation').value.trim() || null,
    };

    if (!payload.name || !payload.sku) {
        showToast('Name and SKU are required', 'error');
        return;
    }

    try {
        if (id) {
            await api('PUT', `/api/items/${id}`, payload);
            showToast('Item updated', 'success');
        } else {
            await api('POST', '/api/items', payload);
            showToast('Item created', 'success');
        }
        closeItemModal();
        await loadItems(state.currentPage);
        if (state.selectedItemId) loadItemHistory(state.selectedItemId);
    } catch (e) {
        showToast(e.message, 'error');
    }
}

/* ─── Delete Confirm ──────────────────────────────────────────────── */
let pendingDeleteId = null;

function openDeleteConfirm(id) {
    const item = state.items.find(i => i.id === id);
    pendingDeleteId = id;
    $('#confirmMessage').textContent =
        `Are you sure you want to delete "${item ? item.name : 'this item'}"? This action cannot be undone.`;
    $('#confirmModal').style.display = '';
}

function closeConfirmModal() {
    $('#confirmModal').style.display = 'none';
    pendingDeleteId = null;
}

async function confirmDelete() {
    if (!pendingDeleteId) return;
    try {
        await api('DELETE', `/api/items/${pendingDeleteId}`);
        showToast('Item deleted', 'success');
        if (state.selectedItemId === pendingDeleteId) {
            state.selectedItemId = null;
            $('#itemHistoryPanel').style.display = 'none';
        }
        closeConfirmModal();
        await loadItems(state.currentPage);
    } catch (e) {
        showToast(e.message, 'error');
    }
}

/* ═══════════════════════════════════════════════════════════════════════
   Item History (inline panel)
   ═══════════════════════════════════════════════════════════════════════ */
async function loadItemHistory(id) {
    const item = state.items.find(i => i.id === id);
    $('#itemHistoryTitle').textContent = 'History — ' + (item ? item.name : id);
    $('#itemHistoryPanel').style.display = '';

    const tbody = $('#itemHistoryBody');
    tbody.innerHTML = '<tr><td colspan="4" class="empty-state"><div class="spinner spinner-dark"></div></td></tr>';

    try {
        const entries = await api('GET', `/api/items/${id}/audit`) || [];
        if (!entries.length) {
            tbody.innerHTML = '<tr><td colspan="4" class="empty-state">No history records</td></tr>';
            return;
        }
        tbody.innerHTML = entries.map(e => `
            <tr>
                <td>${actionBadge(e.action)}</td>
                <td>${escHtml(e.username || '—')}</td>
                <td>${formatDate(e.changed_at)}</td>
                <td>${renderDiff(e)}</td>
            </tr>
        `).join('');
    } catch (e) {
        tbody.innerHTML = '<tr><td colspan="4" class="empty-state">Failed to load history</td></tr>';
    }

    // Scroll into view
    $('#itemHistoryPanel').scrollIntoView({ behavior: 'smooth', block: 'nearest' });
}

/* ═══════════════════════════════════════════════════════════════════════
   Diff Rendering
   ═══════════════════════════════════════════════════════════════════════ */
function renderDiff(entry) {
    if (entry.action === 'INSERT') {
        return '<div class="diff-block"><span class="diff-added">Created</span></div>';
    }
    if (entry.action === 'DELETE') {
        return '<div class="diff-block"><span class="diff-removed">Deleted</span></div>';
    }

    // UPDATE — use diff object { field: { old, new } }
    const diff = entry.diff;
    if (!diff || typeof diff !== 'object' || Object.keys(diff).length === 0) {
        return '<span class="text-muted">No field changes</span>';
    }

    const lines = Object.entries(diff).map(([field, val]) =>
        `<div class="diff-line">
            <strong>${escHtml(field)}:</strong>
            <span class="diff-modified-old">${escHtml(String(val.old ?? ''))}</span>
            &rarr;
            <span class="diff-modified-new">${escHtml(String(val.new ?? ''))}</span>
        </div>`
    );

    return `<div class="diff-block">${lines.join('')}</div>`;
}

/* ═══════════════════════════════════════════════════════════════════════
   Global Audit History
   ═══════════════════════════════════════════════════════════════════════ */
async function loadAudit(page = 1) {
    state.auditPage = page;

    const action = $('#filterAction').value;
    const dateFrom = $('#filterDateFrom').value;
    const dateTo = $('#filterDateTo').value;

    let url = `/api/audit?page=${page}&page_size=${state.pageSize}`;
    if (action) url += `&action=${action}`;
    if (dateFrom) url += `&date_from=${dateFrom}T00:00:00Z`;
    if (dateTo) url += `&date_to=${dateTo}T23:59:59Z`;

    const tbody = $('#globalHistoryBody');
    tbody.innerHTML = '<tr><td colspan="5" class="empty-state"><div class="spinner spinner-dark"></div></td></tr>';

    try {
        const data = await api('GET', url);
        const entries = data.entries || [];

        if (!entries.length) {
            tbody.innerHTML = '<tr><td colspan="5" class="empty-state">No history records found</td></tr>';
            $('#auditPagination').innerHTML = '';
            return;
        }

        tbody.innerHTML = entries.map(e => `
            <tr>
                <td>${actionBadge(e.action)}</td>
                <td><code>${escHtml((e.item_id || '').substring(0, 8))}…</code></td>
                <td>${escHtml(e.username || '—')}</td>
                <td>${formatDate(e.changed_at)}</td>
                <td>${renderDiff(e)}</td>
            </tr>
        `).join('');

        renderPagination('auditPagination', data, loadAudit);
    } catch (e) {
        tbody.innerHTML = '<tr><td colspan="5" class="empty-state">Failed to load history</td></tr>';
        showToast(e.message, 'error');
    }
}

/* ─── CSV Export ───────────────────────────────────────────────────── */
function exportCsv() {
    const action = $('#filterAction').value;
    const dateFrom = $('#filterDateFrom').value;
    const dateTo = $('#filterDateTo').value;

    const params = [];
    if (action) params.push(`action=${action}`);
    if (dateFrom) params.push(`date_from=${dateFrom}T00:00:00Z`);
    if (dateTo) params.push(`date_to=${dateTo}T23:59:59Z`);

    const url = `/api/audit/export?${params.join('&')}`;

    fetch(url, { headers: { 'Authorization': 'Bearer ' + state.token } })
        .then(res => {
            if (!res.ok) throw new Error('Export failed');
            return res.blob();
        })
        .then(blob => {
            const a = document.createElement('a');
            a.href = URL.createObjectURL(blob);
            a.download = `audit_${new Date().toISOString().slice(0, 10)}.csv`;
            document.body.appendChild(a);
            a.click();
            a.remove();
            URL.revokeObjectURL(a.href);
        })
        .catch(e => showToast('Export failed: ' + e.message, 'error'));
}

/* ═══════════════════════════════════════════════════════════════════════
   Pagination
   ═══════════════════════════════════════════════════════════════════════ */
function renderPagination(containerId, data, loadFn) {
    const container = $(`#${containerId}`);
    if (!data.total_pages || data.total_pages <= 1) {
        container.innerHTML = '';
        return;
    }

    let html = '';

    // Previous
    html += `<button class="btn btn-outline btn-sm" ${data.page <= 1 ? 'disabled' : ''} data-page="${data.page - 1}">←</button>`;

    // Page numbers
    for (let i = 1; i <= data.total_pages; i++) {
        if (data.total_pages > 7) {
            if (i !== 1 && i !== data.total_pages && Math.abs(i - data.page) > 2) {
                if (i === 2 || i === data.total_pages - 1) {
                    html += '<button class="btn btn-outline btn-sm" disabled>…</button>';
                }
                continue;
            }
        }
        const cls = i === data.page ? 'btn btn-primary btn-sm' : 'btn btn-outline btn-sm';
        html += `<button class="${cls}" data-page="${i}">${i}</button>`;
    }

    // Next
    html += `<button class="btn btn-outline btn-sm" ${data.page >= data.total_pages ? 'disabled' : ''} data-page="${data.page + 1}">→</button>`;

    container.innerHTML = html;

    // Bind click events
    container.querySelectorAll('button[data-page]').forEach(btn => {
        if (!btn.disabled) {
            btn.addEventListener('click', () => loadFn(parseInt(btn.dataset.page, 10)));
        }
    });
}

/* ═══════════════════════════════════════════════════════════════════════
   Tab Navigation
   ═══════════════════════════════════════════════════════════════════════ */
$$('.tab-btn').forEach(btn => {
    btn.addEventListener('click', () => {
        $$('.tab-btn').forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        $$('.tab-panel').forEach(p => p.classList.remove('active'));

        const tab = btn.dataset.tab;
        const panelId = 'panel' + tab.charAt(0).toUpperCase() + tab.slice(1);
        $(`#${panelId}`).classList.add('active');

        if (tab === 'history') loadAudit(1);
        if (tab === 'items') loadItems(state.currentPage);
    });
});

/* ═══════════════════════════════════════════════════════════════════════
   Event Bindings
   ═══════════════════════════════════════════════════════════════════════ */

// Login
$('#loginForm').addEventListener('submit', e => {
    e.preventDefault();
    const username = $('#userSelect').value;
    const password = $('#passwordInput').value;
    if (username) login(username, password);
});

// Logout
$('#logoutBtn').addEventListener('click', logout);

// Add Item
$('#addItemBtn').addEventListener('click', openAddModal);

// Item Modal
$('#itemModalClose').addEventListener('click', closeItemModal);
$('#itemModalCancel').addEventListener('click', closeItemModal);
$('#itemModalSave').addEventListener('click', saveItem);
$('#itemModal').addEventListener('click', e => {
    if (e.target === $('#itemModal')) closeItemModal();
});

// Confirm Modal
$('#confirmModalClose').addEventListener('click', closeConfirmModal);
$('#confirmCancel').addEventListener('click', closeConfirmModal);
$('#confirmOk').addEventListener('click', confirmDelete);
$('#confirmModal').addEventListener('click', e => {
    if (e.target === $('#confirmModal')) closeConfirmModal();
});

// Close item history panel
$('#closeItemHistory').addEventListener('click', () => {
    state.selectedItemId = null;
    $('#itemHistoryPanel').style.display = 'none';
    renderItems();
});

// Search
$('#searchBtn').addEventListener('click', () => loadItems(1));
$('#searchInput').addEventListener('keydown', e => {
    if (e.key === 'Enter') loadItems(1);
});

// Audit filters
$('#applyFiltersBtn').addEventListener('click', () => loadAudit(1));
$('#clearFiltersBtn').addEventListener('click', () => {
    $('#filterAction').value = '';
    $('#filterDateFrom').value = '';
    $('#filterDateTo').value = '';
    loadAudit(1);
});
$('#exportCsvBtn').addEventListener('click', exportCsv);

// Keyboard: Escape closes modals
document.addEventListener('keydown', e => {
    if (e.key === 'Escape') {
        if ($('#itemModal').style.display !== 'none') closeItemModal();
        if ($('#confirmModal').style.display !== 'none') closeConfirmModal();
    }
});

/* ═══════════════════════════════════════════════════════════════════════
   Init
   ═══════════════════════════════════════════════════════════════════════ */
(function init() {
    if (state.token && state.user) {
        enterApp();
    } else {
        loadUsers();
    }
})();