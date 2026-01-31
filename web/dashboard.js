// Dashboard JavaScript - Consolidated Plans and Settings
requireAuth();
requireAdmin();

const user = getCurrentUser();
// Populate topbar user info
document.getElementById('userEmail').textContent = user.email;
// Populate sidebar user info
document.getElementById('userEmailSidebar').textContent = user.email;
document.getElementById('userRoleSidebar').innerHTML = '<span class="badge badge-success" style="font-size: 0.65rem;">Admin</span>';

// Build simplified navigation
const navItems = [
    { label: 'Dashboard', page: 'dashboard.html', icon: '📊' },
    { label: 'Users', page: 'users.html', icon: '👥' },
    { label: 'Plans', page: 'plans.html', icon: '📋' },
    { label: 'Nodes', page: 'nodes.html', icon: '🌐' },
    { label: 'Subscriptions', page: 'subscriptions.html', icon: '💎' },
    { label: 'Settings', page: 'settings.html', icon: '⚙️' },
];

const nav = document.getElementById('sidebarNav');
navItems.forEach(item => {
    const navItem = document.createElement('div');
    navItem.className = 'nav-item' + (window.location.pathname.includes(item.page) ? ' active' : '');
    navItem.innerHTML = `${item.icon} ${item.label}`;
    navItem.onclick = () => window.location.href = item.page;
    nav.appendChild(navItem);
});

// Tab Switching
function switchTab(tabId) {
    // Update tab buttons
    document.querySelectorAll('.dashboard-tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('.tab-panel').forEach(p => p.classList.remove('active'));

    // Find and activate the correct tab button
    const tabs = document.querySelectorAll('.dashboard-tab');
    const tabIndex = ['overview', 'plans', 'settings'].indexOf(tabId);
    if (tabIndex >= 0 && tabs[tabIndex]) {
        tabs[tabIndex].classList.add('active');
    }

    // Activate the correct panel
    const panel = document.getElementById(tabId);
    if (panel) {
        panel.classList.add('active');
    }

    // Load tab content
    if (tabId === 'overview') loadOverview();
    else if (tabId === 'plans') loadPlans();
    else if (tabId === 'settings') {
        loadSystemConfig();
        loadAdministrators();
    }

    // Update URL hash
    window.location.hash = tabId;
}

// ============ OVERVIEW TAB ============
async function loadOverview() {
    const content = document.getElementById('overviewContent');
    showLoading(content);

    try {
        // Fetch both stats and activity in parallel
        const [statsResp, activityResp] = await Promise.all([
            api.adminGetStats(),
            api.request('/api/v1/admin/activity/stats', { method: 'GET' }).catch(() => ({ success: false }))
        ]);

        if (statsResp.success) {
            const data = statsResp.data;
            const onlineUsers = activityResp.success ? activityResp.data.online : 0;
            const totalActiveUsers = activityResp.success ? activityResp.data.total : (data.active_users || 0);

            content.innerHTML = `
                <div class="stats-grid">
                    <div class="stat-card">
                        <div class="stat-label">Total Users</div>
                        <div class="stat-value">${data.total_users || 0}</div>
                        <div class="stat-meta" style="color: var(--success);">Active: ${data.active_users || 0}</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-label">Users Online</div>
                        <div class="stat-value" style="color: var(--success);">${onlineUsers}</div>
                        <div class="stat-meta">Connected right now</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-label">Total Nodes</div>
                        <div class="stat-value">${data.total_nodes || 0}</div>
                        <div class="stat-meta" style="color: var(--success);">Online: ${data.online_nodes || 0}</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-label">Subscription Plans</div>
                        <div class="stat-value">${data.total_subscriptions || 0}</div>
                        <div class="stat-meta">Active plans</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-label">Total Traffic</div>
                        <div class="stat-value">${formatBytes(data.total_traffic || 0)}</div>
                        <div class="stat-meta">All time</div>
                    </div>
                </div>
                <div class="card">
                    <h3 class="card-title">Quick Actions</h3>
                    <div style="display: flex; gap: var(--space-md); flex-wrap: wrap;">
                        <button class="btn btn-primary" onclick="window.location.href='users.html'">Manage Users</button>
                        <button class="btn btn-primary" onclick="window.location.href='nodes.html'">Manage Nodes</button>
                        <button class="btn btn-secondary" onclick="window.location.href='subscriptions.html'">Subscriptions</button>
                        <button class="btn btn-secondary" onclick="window.location.href='settings.html'">System Settings</button>
                    </div>
                </div>
            `;
        }
    } catch (error) {
        content.innerHTML = '<div class="card"><p style="color: var(--error);">Failed to load dashboard</p></div>';
    }
}


// ============ PLANS TAB ============
let allPlans = [];
let allNodes = [];

async function loadPlans() {
    const tbody = document.getElementById('plansTableBody');
    showLoading(tbody);

    try {
        const response = await api.getPlans();
        if (response.success) {
            allPlans = response.data.plans || [];
            renderPlans(allPlans);
        } else {
            tbody.innerHTML = '<tr><td colspan="10" style="text-align: center; color: var(--error);">Failed to load plans</td></tr>';
        }
    } catch (error) {
        tbody.innerHTML = '<tr><td colspan="10" style="text-align: center; color: var(--error);">Failed to load plans</td></tr>';
    }
}

function renderPlans(plans) {
    const tbody = document.getElementById('plansTableBody');
    if (!plans || plans.length === 0) {
        tbody.innerHTML = '<tr><td colspan="10" style="text-align: center; color: var(--text-muted);">No plans found</td></tr>';
        return;
    }

    tbody.innerHTML = plans.map(plan => {
        const hasUsersNoNodes = (plan.user_count || 0) > 0 && (plan.node_count || 0) === 0;
        const nodesWarning = hasUsersNoNodes ?
            '<span title="Users on this plan cannot connect - no nodes assigned!" style="cursor: help; color: var(--warning);">⚠️</span>' : '';

        return `
            <tr>
                <td>${plan.id}</td>
                <td><strong>${plan.display_name}</strong><br><small style="color: var(--text-muted);">${plan.name}</small></td>
                <td><span class="badge badge-info">${plan.duration}</span><br><small>${plan.duration_days} days</small></td>
                <td><strong>$${plan.price.toFixed(2)}</strong></td>
                <td>${plan.data_limit_gb === 0 ? '<span class="badge badge-success">Unlimited</span>' : formatBytes(plan.data_limit_bytes)}</td>
                <td>${plan.max_devices}</td>
                <td>
                    <span class="badge ${plan.node_count > 0 ? 'badge-success' : 'badge-secondary'}">${plan.node_count || 0} nodes</span>
                    ${nodesWarning}
                </td>
                <td><span class="badge ${plan.user_count > 0 ? 'badge-info' : 'badge-secondary'}">${plan.user_count || 0} users</span></td>
                <td>${getStatusBadge(plan.status)}</td>
                <td>
                    <button class="btn btn-sm btn-secondary" onclick="editPlan(${plan.id})">Edit</button>
                    <button class="btn btn-sm btn-info" onclick="showNodeAssignment(${plan.id})">Nodes</button>
                    <button class="btn btn-sm btn-danger" onclick="deletePlan(${plan.id})">Delete</button>
                </td>
            </tr>
        `;
    }).join('');
}

function getStatusBadge(status) {
    const badges = {
        'active': '<span class="badge badge-success">Active</span>',
        'archived': '<span class="badge badge-secondary">Archived</span>'
    };
    return badges[status] || `<span class="badge">${status}</span>`;
}

function showAddPlanModal() {
    document.getElementById('planModalTitle').textContent = 'Create Plan';
    document.getElementById('planForm').reset();
    document.getElementById('planId').value = '';
    document.getElementById('planName').disabled = false;
    document.getElementById('planModal').classList.remove('hidden');
}

async function editPlan(id) {
    try {
        const response = await api.getPlan(id);
        if (response.success) {
            const plan = response.data.plan;
            document.getElementById('planModalTitle').textContent = 'Edit Plan';
            document.getElementById('planId').value = plan.id;
            document.getElementById('planName').value = plan.name;
            document.getElementById('planName').disabled = true;
            document.getElementById('planDisplayName').value = plan.display_name;
            document.getElementById('planDuration').value = plan.duration;
            document.getElementById('planPrice').value = plan.price;
            document.getElementById('planDataLimit').value = plan.data_limit_gb;
            document.getElementById('planMaxDevices').value = plan.max_devices;
            document.getElementById('planDescription').value = plan.description || '';
            document.getElementById('planFeatures').value = plan.features ? plan.features.join(', ') : '';
            document.getElementById('planModal').classList.remove('hidden');
        }
    } catch (error) {
        showToast('Failed to load plan details', 'error');
    }
}

async function savePlan() {
    const planId = document.getElementById('planId').value;
    const data = {
        name: document.getElementById('planName').value,
        display_name: document.getElementById('planDisplayName').value,
        duration: document.getElementById('planDuration').value,
        price: parseFloat(document.getElementById('planPrice').value),
        data_limit_gb: parseInt(document.getElementById('planDataLimit').value),
        max_devices: parseInt(document.getElementById('planMaxDevices').value),
        description: document.getElementById('planDescription').value,
        features: document.getElementById('planFeatures').value
    };

    try {
        let response;
        if (planId) {
            response = await api.updatePlan(planId, data);
        } else {
            response = await api.createPlan(data);
        }

        if (response.success) {
            showToast(planId ? 'Plan updated successfully' : 'Plan created successfully', 'success');
            closePlanModal();
            loadPlans();
        } else {
            showToast(response.message || 'Failed to save plan', 'error');
        }
    } catch (error) {
        showToast('Operation failed: ' + error.message, 'error');
    }
}

async function deletePlan(id) {
    const plan = allPlans.find(p => p.id === id);
    confirm(`Are you sure you want to delete "${plan.display_name}"? This action cannot be undone.`, async () => {
        try {
            const response = await api.deletePlan(id);
            if (response.success) {
                showToast('Plan archived successfully', 'success');
                loadPlans();
            } else {
                showToast(response.message || 'Failed to delete plan', 'error');
            }
        } catch (error) {
            showToast('Failed to delete plan', 'error');
        }
    });
}

async function showNodeAssignment(planId) {
    document.getElementById('nodePlanId').value = planId;

    try {
        const nodesResponse = await api.getNodes();
        if (nodesResponse.success) {
            allNodes = nodesResponse.data.nodes || [];
        }

        const planNodesResponse = await api.getPlanNodes(planId);
        const assignedNodeIds = planNodesResponse.success ?
            (planNodesResponse.data.nodes || []).map(n => n.id) : [];

        const nodesList = document.getElementById('nodesList');
        if (allNodes.length === 0) {
            nodesList.innerHTML = '<p style="text-align: center; color: var(--text-muted);">No nodes available</p>';
        } else {
            nodesList.innerHTML = allNodes.map(node => `
                <div class="form-group" style="margin-bottom: var(--space-sm);">
                    <label style="display: flex; align-items: center; cursor: pointer;">
                        <input type="checkbox" 
                               class="node-checkbox" 
                               value="${node.id}" 
                               ${assignedNodeIds.includes(node.id) ? 'checked' : ''}
                               style="margin-right: var(--space-sm);">
                        <div>
                            <strong>${node.name}</strong>
                            <small style="color: var(--text-muted); display: block;">
                                ${node.country || 'Unknown'} - ${node.status}
                            </small>
                        </div>
                    </label>
                </div>
            `).join('');
        }

        document.getElementById('nodeModal').classList.remove('hidden');
    } catch (error) {
        showToast('Failed to load nodes', 'error');
    }
}

async function saveNodeAssignment() {
    const planId = document.getElementById('nodePlanId').value;
    const checkboxes = document.querySelectorAll('.node-checkbox:checked');
    const nodeIds = Array.from(checkboxes).map(cb => parseInt(cb.value));

    try {
        const response = await api.assignNodesToPlan(planId, nodeIds);
        if (response.success) {
            showToast('Nodes assigned successfully', 'success');
            closeNodeModal();
            loadPlans();
        } else {
            showToast('Failed to assign nodes', 'error');
        }
    } catch (error) {
        showToast('Failed to assign nodes', 'error');
    }
}

function closePlanModal() {
    document.getElementById('planModal').classList.add('hidden');
}

function closeNodeModal() {
    document.getElementById('nodeModal').classList.add('hidden');
}

// ============ SETTINGS TAB ============
async function loadSystemConfig() {
    try {
        const response = await api.getSystemConfig();
        if (response.success) {
            renderConfigs(response.data.configs || []);
        }
    } catch (error) {
        showToast('Failed to load configuration', 'error');
    }
}

function renderConfigs(configs) {
    const list = document.getElementById('configList');
    if (!configs || configs.length === 0) {
        list.innerHTML = '<p style="color: var(--text-muted);">No configuration found</p>';
        return;
    }

    list.innerHTML = configs.map(cfg => `
        <div class="config-item">
            <div style="flex: 1;">
                <div style="font-weight: 600; margin-bottom: var(--space-xs);">${cfg.key}</div>
                <div style="font-size: 0.875rem; color: var(--text-secondary); font-family: var(--font-mono);">${cfg.value}</div>
                ${cfg.description ? `<div style="font-size: 0.75rem; color: var(--text-muted); margin-top: var(--space-xs);">${cfg.description}</div>` : ''}
                ${cfg.updated_at ? `<div style="font-size: 0.75rem; color: var(--text-muted); margin-top: var(--space-xs);">Last updated: ${formatDate(cfg.updated_at)}</div>` : ''}
            </div>
            <button class="btn btn-sm btn-primary" onclick='editConfig(${JSON.stringify(cfg)})'>Edit</button>
        </div>
    `).join('');
}

function editConfig(cfg) {
    document.getElementById('editKey').value = cfg.key;
    document.getElementById('editLabel').textContent = cfg.key.replace(/_/g, ' ').toUpperCase();
    document.getElementById('editValue').value = cfg.encrypted ? '' : cfg.value;
    document.getElementById('editValue').placeholder = cfg.encrypted ? 'Enter new value...' : '';
    document.getElementById('editDescription').textContent = cfg.description || '';
    document.getElementById('editModal').classList.remove('hidden');
}

function closeEditModal() {
    document.getElementById('editModal').classList.add('hidden');
}

async function saveConfig() {
    const key = document.getElementById('editKey').value;
    const value = document.getElementById('editValue').value;

    if (!value) {
        showToast('Value cannot be empty', 'error');
        return;
    }

    try {
        const response = await api.updateSystemConfig(key, value);
        if (response.success) {
            showToast('Configuration updated successfully', 'success');
            closeEditModal();
            loadSystemConfig();
        } else {
            showToast(response.message || 'Failed to update', 'error');
        }
    } catch (error) {
        showToast('Operation failed', 'error');
    }
}

async function reloadConfig() {
    try {
        const response = await api.reloadSystemConfig();
        if (response.success) {
            showToast('Configuration cache reloaded', 'success');
        }
    } catch (error) {
        showToast('Failed to reload cache', 'error');
    }
}

// Administrator Management Functions

async function loadAdministrators() {
    const tbody = document.getElementById('administratorsTableBody');
    tbody.innerHTML = '<tr><td colspan="4" style="text-align: center;">Loading...</td></tr>';

    try {
        const response = await api.getAdministrators();
        if (response.success && response.data.administrators) {
            const admins = response.data.administrators;
            if (admins.length === 0) {
                tbody.innerHTML = '<tr><td colspan="4" style="text-align: center; color: var(--text-muted);">No administrators found</td></tr>';
                return;
            }

            tbody.innerHTML = admins.map(admin => `
                <tr>
                    <td>${admin.id}</td>
                    <td>${admin.email}</td>
                    <td>${new Date(admin.created_at).toLocaleDateString()}</td>
                    <td>
                        <button class="btn btn-sm btn-secondary" onclick="showChangePasswordModal(${admin.id})">Change Password</button>
                        ${admin.id !== user.id ? `<button class="btn btn-sm btn-danger" onclick="deleteAdmin(${admin.id}, '${admin.email}')">Delete</button>` : ''}
                    </td>
                </tr>
            `).join('');
        } else {
            tbody.innerHTML = '<tr><td colspan="4" style="text-align: center; color: var(--error);">Failed to load administrators</td></tr>';
        }
    } catch (error) {
        tbody.innerHTML = '<tr><td colspan="4" style="text-align: center; color: var(--error);">Error loading administrators</td></tr>';
    }
}

function showAddAdminModal() {
    document.getElementById('addAdminForm').reset();
    document.getElementById('addAdminModal').classList.remove('hidden');
}

function closeAddAdminModal() {
    document.getElementById('addAdminModal').classList.add('hidden');
}

async function createAdmin() {
    const email = document.getElementById('adminEmail').value;
    const password = document.getElementById('adminPassword').value;

    try {
        const response = await api.createAdministrator({ email, password });
        if (response.success) {
            showToast('Administrator created successfully', 'success');
            closeAddAdminModal();
            loadAdministrators();
        } else {
            showToast(response.message || 'Failed to create administrator', 'error');
        }
    } catch (error) {
        showToast('Failed to create administrator', 'error');
    }
}

function showChangePasswordModal(adminId) {
    document.getElementById('changePasswordForm').reset();
    document.getElementById('changePasswordAdminId').value = adminId;
    document.getElementById('changePasswordModal').classList.remove('hidden');
}

function closeChangePasswordModal() {
    document.getElementById('changePasswordModal').classList.add('hidden');
}

async function savePasswordChange() {
    const adminId = document.getElementById('changePasswordAdminId').value;
    const currentPassword = document.getElementById('currentPassword').value;
    const newPassword = document.getElementById('newPassword').value;

    try {
        const response = await api.changeAdminPassword(adminId, {
            current_password: currentPassword,
            new_password: newPassword
        });
        if (response.success) {
            showToast('Password changed successfully', 'success');
            closeChangePasswordModal();
        } else {
            showToast(response.message || 'Failed to change password', 'error');
        }
    } catch (error) {
        showToast('Invalid current password or update failed', 'error');
    }
}

async function deleteAdmin(adminId, email) {
    confirm(`Are you sure you want to delete administrator "${email}"? This action cannot be undone.`, async () => {
        try {
            const response = await api.deleteAdministrator(adminId);
            if (response.success) {
                showToast('Administrator deleted successfully', 'success');
                loadAdministrators();
            } else {
                showToast(response.message || 'Failed to delete administrator', 'error');
            }
        } catch (error) {
            showToast('Failed to delete administrator', 'error');
        }
    });
}

// Initialize - check URL hash or load overview by default
const hash = window.location.hash.substring(1); // Remove the #
if (hash && ['overview', 'plans', 'settings'].includes(hash)) {
    switchTab(hash);
} else {
    loadOverview();
}
