// API Client for xpanel
class APIClient {
  constructor(baseURL = null) {
    // Dynamically detect base URL from current location if not provided
    if (!baseURL) {
      baseURL = `${window.location.protocol}//${window.location.host}`;
    }
    this.baseURL = baseURL;
    this.token = localStorage.getItem('access_token');
    this.refreshToken = localStorage.getItem('refresh_token');
  }

  // Authentication
  async login(email, password) {
    const response = await this.request('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password })
    }, false);

    console.log('Login response:', response);

    if (response.success) {
      // Backend returns tokens in nested object: response.data.tokens
      console.log('Tokens object:', response.data.tokens);
      this.token = response.data.tokens.access_token;
      this.refreshToken = response.data.tokens.refresh_token;
      localStorage.setItem('access_token', this.token);
      localStorage.setItem('refresh_token', this.refreshToken);
      localStorage.setItem('user', JSON.stringify(response.data.user));
      console.log('Stored access_token:', this.token);
    }

    return response;
  }

  async register(email, password) {
    return this.request('/api/v1/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password })
    }, false);
  }

  async logout() {
    // Try to call logout endpoint, but don't fail if it errors
    try {
      if (this.token) {
        await this.request('/api/v1/auth/logout', { method: 'POST' });
      }
    } catch (error) {
      // Ignore errors during logout
      console.log('Logout API call failed, clearing local tokens anyway');
    } finally {
      this.clearTokens();
      window.location.href = '/web/login.html';
    }
  }

  async refreshAccessToken() {
    const response = await fetch(`${this.baseURL}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ refresh_token: this.refreshToken })
    });

    const data = await response.json();

    if (data.success && data.data && data.data.tokens) {
      // Backend returns tokens in nested object
      this.token = data.data.tokens.access_token;
      localStorage.setItem('access_token', this.token);
      return true;
    }

    return false;
  }

  // Generic request method
  async request(endpoint, options = {}, requiresAuth = true) {
    const headers = {
      'Content-Type': 'application/json',
      ...options.headers
    };

    if (requiresAuth && this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    try {
      const response = await fetch(`${this.baseURL}${endpoint}`, {
        ...options,
        headers
      });

      // Handle 401 unauthorized - but only retry once
      if (response.status === 401 && requiresAuth && !options._isRetry) {
        const refreshed = await this.refreshAccessToken();
        if (refreshed) {
          // Retry the original request with new token
          headers['Authorization'] = `Bearer ${this.token}`;
          const retryResponse = await fetch(`${this.baseURL}${endpoint}`, {
            ...options,
            headers,
            _isRetry: true  // Prevent infinite loop
          });
          return retryResponse.json();
        } else {
          // Refresh failed, logout
          this.clearTokens();
          window.location.href = '/web/login.html';
          throw new Error('Session expired');
        }
      }

      return response.json();
    } catch (error) {
      console.error('API Request failed:', error);
      throw error;
    }
  }

  // Clear tokens helper
  clearTokens() {
    this.token = null;
    this.refreshToken = null;
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user');
  }

  // User APIs
  async getProfile() {
    return this.request('/api/v1/user/profile');
  }

  async getSubscription() {
    return this.request('/api/v1/user/subscription');
  }

  async getDevices() {
    return this.request('/api/v1/user/devices');
  }

  async getVPNConfig() {
    return this.request('/api/v1/user/config');
  }

  // Admin - Users
  async adminGetUsers(page = 1, limit = 20) {
    return this.request(`/api/v1/admin/users?page=${page}&page_size=${limit}`);
  }

  async adminGetUser(id) {
    return this.request(`/api/v1/admin/users/${id}`);
  }

  async adminUpdateUser(id, data) {
    return this.request(`/api/v1/admin/users/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data)
    });
  }

  async adminCreateUser(data) {
    return this.request('/api/v1/admin/users', {
      method: 'POST',
      body: JSON.stringify(data)
    });
  }

  async adminSuspendUser(id) {
    return this.request(`/api/v1/admin/users/${id}/suspend`, {
      method: 'POST'
    });
  }

  async adminActivateUser(id) {
    return this.request(`/api/v1/admin/users/${id}/activate`, {
      method: 'POST'
    });
  }

  async getUserLinks(id) {
    return this.request(`/api/v1/admin/users/${id}/links`);
  }

  async adminUpdateUserSubscription(userId, data) {
    return this.request(`/api/v1/admin/users/${userId}/subscription`, {
      method: 'PUT',
      body: JSON.stringify(data)
    });
  }

  async adminDeleteUser(id) {
    return this.request(`/api/v1/admin/users/${id}`, {
      method: 'DELETE'
    });
  }

  // Admin - Nodes
  async adminGetNodes() {
    return this.request('/api/v1/admin/nodes');
  }

  async adminGetNode(id) {
    return this.request(`/api/v1/admin/nodes/${id}`);
  }

  async adminCreateNode(data) {
    return this.request('/api/v1/admin/nodes', {
      method: 'POST',
      body: JSON.stringify(data)
    });
  }

  async adminUpdateNode(id, data) {
    return this.request(`/api/v1/admin/nodes/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data)
    });
  }

  async adminDeleteNode(id) {
    return this.request(`/api/v1/admin/nodes/${id}`, {
      method: 'DELETE'
    });
  }

  // Admin - Subscriptions
  async adminGetSubscriptions(page = 1, limit = 20) {
    return this.request(`/api/v1/admin/subscriptions?page=${page}&limit=${limit}`);
  }

  async adminExtendSubscription(id, days) {
    return this.request(`/api/v1/admin/subscriptions/${id}/extend`, {
      method: 'POST',
      body: JSON.stringify({ days })
    });
  }

  async adminResetDataUsage(id) {
    return this.request(`/api/v1/admin/subscriptions/${id}/reset-data`, {
      method: 'POST'
    });
  }

  // Admin - Statistics
  async adminGetStats() {
    return this.request('/api/v1/admin/stats/overview');
  }

  // Nodes (public)
  async getNodes() {
    return this.request('/api/v1/nodes');
  }

  // System Config (Admin)
  async getSystemConfig() {
    return this.request('/api/v1/admin/system/config');
  }

  async updateSystemConfig(key, value) {
    return this.request(`/api/v1/admin/system/config/${key}`, {
      method: 'PUT',
      body: JSON.stringify({ value })
    });
  }

  async reloadSystemConfig() {
    return this.request('/api/v1/admin/system/config/reload', {
      method: 'POST'
    });
  }

  async revealConfigValue(key) {
    return this.request(`/api/v1/admin/system/config/${key}/reveal`);
  }

  // Plan Management (Admin)
  async getPlans() {
    return this.request('/api/v1/admin/plans');
  }

  async getPlan(id) {
    return this.request(`/api/v1/admin/plans/${id}`);
  }

  async createPlan(data) {
    return this.request('/api/v1/admin/plans', {
      method: 'POST',
      body: JSON.stringify(data)
    });
  }

  async updatePlan(id, data) {
    return this.request(`/api/v1/admin/plans/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data)
    });
  }

  async deletePlan(id) {
    return this.request(`/api/v1/admin/plans/${id}`, {
      method: 'DELETE'
    });
  }

  async assignNodesToPlan(planId, nodeIds) {
    return this.request(`/api/v1/admin/plans/${planId}/nodes`, {
      method: 'PUT',
      body: JSON.stringify({ node_ids: nodeIds })
    });
  }

  async getPlanNodes(planId) {
    return this.request(`/api/v1/admin/plans/${planId}/nodes`);
  }

  async getPlanUsers(planId) {
    return this.request(`/api/v1/admin/plans/${planId}/users`);
  }

  async assignPlanToUser(userId, planId, autoRenew = false) {
    return this.request(`/api/v1/admin/users/${userId}/plan`, {
      method: 'PUT',
      body: JSON.stringify({ plan_id: planId, auto_renew: autoRenew })
    });
  }

  // Administrator Management (Admin)
  async getAdministrators() {
    return this.request('/api/v1/admin/administrators');
  }

  async createAdministrator(data) {
    return this.request('/api/v1/admin/administrators', {
      method: 'POST',
      body: JSON.stringify(data)
    });
  }

  async changeAdminPassword(adminId, data) {
    return this.request(`/api/v1/admin/administrators/${adminId}/password`, {
      method: 'PUT',
      body: JSON.stringify(data)
    });
  }

  async deleteAdministrator(adminId) {
    return this.request(`/api/v1/admin/administrators/${adminId}`, {
      method: 'DELETE'
    });
  }
}

// Export instance
const api = new APIClient();
