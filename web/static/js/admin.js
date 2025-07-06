// DomainVault Admin Application
class AdminApp {
    constructor() {
        this.apiBase = window.location.origin + '/api/v1';
        this.token = localStorage.getItem('admin_token');
        this.user = null;
        this.domains = [];
        this.categories = [];
        this.projects = [];
        this.credentials = [];
        
        this.init();
    }

    async init() {
        // Check if already logged in
        if (this.token) {
            if (await this.validateToken()) {
                this.showDashboard();
                return;
            } else {
                this.logout();
            }
        }
        
        this.setupEventListeners();
    }

    setupEventListeners() {
        // Login form
        document.getElementById('loginForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.login();
        });

        // Logout button
        document.getElementById('logoutBtn').addEventListener('click', () => {
            this.logout();
        });

        // Tab switching
        document.querySelectorAll('.admin-tab').forEach(tab => {
            tab.addEventListener('click', (e) => {
                this.switchTab(e.target.dataset.tab);
            });
        });
    }

    async login() {
        const username = document.getElementById('username').value;
        const password = document.getElementById('password').value;
        const loginBtn = document.getElementById('loginBtn');
        const loginText = document.getElementById('loginText');
        const loginSpinner = document.getElementById('loginSpinner');
        const errorMessage = document.getElementById('errorMessage');

        // Show loading state
        loginBtn.disabled = true;
        loginText.style.display = 'none';
        loginSpinner.style.display = 'inline';
        errorMessage.style.display = 'none';

        try {
            const response = await fetch(`${this.apiBase}/auth/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ username, password })
            });

            const data = await response.json();

            if (response.ok) {
                this.token = data.token;
                this.user = data.user;
                localStorage.setItem('admin_token', this.token);
                this.showDashboard();
            } else {
                this.showError(data.error || 'Login failed');
            }
        } catch (error) {
            this.showError('Connection error: ' + error.message);
        } finally {
            // Reset button state
            loginBtn.disabled = false;
            loginText.style.display = 'inline';
            loginSpinner.style.display = 'none';
        }
    }

    async validateToken() {
        try {
            const response = await this.apiCall('GET', '/admin/domains');
            return response.ok;
        } catch (error) {
            return false;
        }
    }

    logout() {
        if (this.token) {
            // Call logout endpoint
            this.apiCall('POST', '/auth/logout').catch(() => {
                // Ignore errors on logout
            });
        }
        
        this.token = null;
        this.user = null;
        localStorage.removeItem('admin_token');
        
        document.getElementById('adminDashboard').style.display = 'none';
        document.getElementById('loginScreen').style.display = 'flex';
        
        // Clear form
        document.getElementById('loginForm').reset();
    }

    showError(message) {
        const errorMessage = document.getElementById('errorMessage');
        errorMessage.textContent = message;
        errorMessage.style.display = 'block';
    }

    async showDashboard() {
        document.getElementById('loginScreen').style.display = 'none';
        document.getElementById('adminDashboard').style.display = 'block';
        
        if (this.user) {
            document.getElementById('userEmail').textContent = this.user.email;
        }
        
        // Load initial data
        await this.loadDashboardData();
    }

    async loadDashboardData() {
        try {
            // Load all data concurrently
            await Promise.all([
                this.loadDomains(),
                this.loadCategories(),
                this.loadProjects(),
                this.loadCredentials()
            ]);
            
            // Render current tab
            this.renderCurrentTab();
        } catch (error) {
            console.error('Failed to load dashboard data:', error);
        }
    }

    switchTab(tabName) {
        // Update tab buttons
        document.querySelectorAll('.admin-tab').forEach(tab => {
            tab.classList.remove('active');
        });
        document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');

        // Update tab content
        document.querySelectorAll('.tab-content').forEach(content => {
            content.classList.remove('active');
        });
        document.getElementById(`${tabName}Tab`).classList.add('active');

        // Render tab content
        this.renderTab(tabName);
    }

    renderCurrentTab() {
        const activeTab = document.querySelector('.admin-tab.active');
        if (activeTab) {
            this.renderTab(activeTab.dataset.tab);
        }
    }

    renderTab(tabName) {
        switch (tabName) {
            case 'domains':
                this.renderDomainsTab();
                break;
            case 'dns':
                this.renderDNSTab();
                break;
            case 'categories':
                this.renderCategoriesTab();
                break;
            case 'projects':
                this.renderProjectsTab();
                break;
            case 'credentials':
                this.renderCredentialsTab();
                break;
            case 'bulk':
                this.renderBulkTab();
                break;
        }
    }

    async apiCall(method, endpoint, data = null) {
        const options = {
            method,
            headers: {
                'Content-Type': 'application/json'
            }
        };

        if (this.token) {
            options.headers['Authorization'] = `Bearer ${this.token}`;
        }

        if (data) {
            options.body = JSON.stringify(data);
        }

        const response = await fetch(`${this.apiBase}${endpoint}`, options);
        
        if (response.status === 401) {
            this.logout();
            throw new Error('Authentication required');
        }
        
        return response;
    }

    async loadDomains() {
        try {
            const response = await this.apiCall('GET', '/domains');
            if (response.ok) {
                const data = await response.json();
                this.domains = data.domains || [];
            }
        } catch (error) {
            console.error('Failed to load domains:', error);
        }
    }

    async loadCategories() {
        try {
            const response = await this.apiCall('GET', '/admin/categories');
            if (response.ok) {
                const data = await response.json();
                this.categories = data.categories || [];
            }
        } catch (error) {
            console.error('Failed to load categories:', error);
        }
    }

    async loadProjects() {
        try {
            const response = await this.apiCall('GET', '/admin/projects');
            if (response.ok) {
                const data = await response.json();
                this.projects = data.projects || [];
            }
        } catch (error) {
            console.error('Failed to load projects:', error);
        }
    }

    async loadCredentials() {
        try {
            const response = await this.apiCall('GET', '/admin/credentials');
            if (response.ok) {
                const data = await response.json();
                this.credentials = data.credentials || [];
            }
        } catch (error) {
            console.error('Failed to load credentials:', error);
        }
    }

    renderDomainsTab() {
        const container = document.getElementById('domainsTable');
        
        if (this.domains.length === 0) {
            container.innerHTML = '<p>No domains found.</p>';
            return;
        }

        let html = `
            <table style="width: 100%; border-collapse: collapse;">
                <thead>
                    <tr style="background: #f9fafb; border-bottom: 1px solid #e5e7eb;">
                        <th style="padding: 1rem; text-align: left;">Domain</th>
                        <th style="padding: 1rem; text-align: left;">Provider</th>
                        <th style="padding: 1rem; text-align: left;">Expires</th>
                        <th style="padding: 1rem; text-align: left;">Status</th>
                        <th style="padding: 1rem; text-align: left;">Auto-Renew</th>
                        <th style="padding: 1rem; text-align: left;">Actions</th>
                    </tr>
                </thead>
                <tbody>
        `;

        this.domains.forEach(domain => {
            const daysLeft = this.calculateDaysLeft(domain.expires_at);
            const statusClass = this.getStatusClass(daysLeft);
            
            html += `
                <tr style="border-bottom: 1px solid #e5e7eb;">
                    <td style="padding: 1rem;">${domain.name}</td>
                    <td style="padding: 1rem;">${domain.provider}</td>
                    <td style="padding: 1rem;">${this.formatDate(domain.expires_at)}</td>
                    <td style="padding: 1rem;">
                        <span class="status-indicator ${statusClass}">
                            ${daysLeft >= 0 ? `${daysLeft} days` : 'Expired'}
                        </span>
                    </td>
                    <td style="padding: 1rem;">
                        <span class="status-indicator ${domain.auto_renew ? 'status-active' : 'status-inactive'}">
                            ${domain.auto_renew ? 'Enabled' : 'Disabled'}
                        </span>
                    </td>
                    <td style="padding: 1rem;">
                        <button class="action-button btn-primary" onclick="adminApp.editDomain('${domain.id}')">
                            <i class="fas fa-edit"></i> Edit
                        </button>
                        <button class="action-button btn-secondary" onclick="adminApp.viewDNS('${domain.id}')">
                            <i class="fas fa-globe"></i> DNS
                        </button>
                    </td>
                </tr>
            `;
        });

        html += '</tbody></table>';
        container.innerHTML = html;
    }

    renderDNSTab() {
        // Populate domain selector
        const select = document.getElementById('dnsSelect');
        select.innerHTML = '<option value="">Choose a domain...</option>';
        
        this.domains.forEach(domain => {
            const option = document.createElement('option');
            option.value = domain.id;
            option.textContent = domain.name;
            select.appendChild(option);
        });
    }

    renderCategoriesTab() {
        const container = document.getElementById('categoriesTable');
        
        if (this.categories.length === 0) {
            container.innerHTML = '<p>No categories found.</p>';
            return;
        }

        let html = `
            <table style="width: 100%; border-collapse: collapse;">
                <thead>
                    <tr style="background: #f9fafb; border-bottom: 1px solid #e5e7eb;">
                        <th style="padding: 1rem; text-align: left;">Name</th>
                        <th style="padding: 1rem; text-align: left;">Description</th>
                        <th style="padding: 1rem; text-align: left;">Color</th>
                        <th style="padding: 1rem; text-align: left;">Actions</th>
                    </tr>
                </thead>
                <tbody>
        `;

        this.categories.forEach(category => {
            html += `
                <tr style="border-bottom: 1px solid #e5e7eb;">
                    <td style="padding: 1rem;">${category.name}</td>
                    <td style="padding: 1rem;">${category.description || ''}</td>
                    <td style="padding: 1rem;">
                        <span style="display: inline-block; width: 20px; height: 20px; background: ${category.color}; border-radius: 4px;"></span>
                        ${category.color}
                    </td>
                    <td style="padding: 1rem;">
                        <button class="action-button btn-primary" onclick="adminApp.editCategory('${category.id}')">
                            <i class="fas fa-edit"></i> Edit
                        </button>
                        <button class="action-button btn-danger" onclick="adminApp.deleteCategory('${category.id}')">
                            <i class="fas fa-trash"></i> Delete
                        </button>
                    </td>
                </tr>
            `;
        });

        html += '</tbody></table>';
        container.innerHTML = html;
    }

    renderProjectsTab() {
        const container = document.getElementById('projectsTable');
        
        if (this.projects.length === 0) {
            container.innerHTML = '<p>No projects found.</p>';
            return;
        }

        let html = `
            <table style="width: 100%; border-collapse: collapse;">
                <thead>
                    <tr style="background: #f9fafb; border-bottom: 1px solid #e5e7eb;">
                        <th style="padding: 1rem; text-align: left;">Name</th>
                        <th style="padding: 1rem; text-align: left;">Description</th>
                        <th style="padding: 1rem; text-align: left;">Color</th>
                        <th style="padding: 1rem; text-align: left;">Actions</th>
                    </tr>
                </thead>
                <tbody>
        `;

        this.projects.forEach(project => {
            html += `
                <tr style="border-bottom: 1px solid #e5e7eb;">
                    <td style="padding: 1rem;">${project.name}</td>
                    <td style="padding: 1rem;">${project.description || ''}</td>
                    <td style="padding: 1rem;">
                        <span style="display: inline-block; width: 20px; height: 20px; background: ${project.color}; border-radius: 4px;"></span>
                        ${project.color}
                    </td>
                    <td style="padding: 1rem;">
                        <button class="action-button btn-primary" onclick="adminApp.editProject('${project.id}')">
                            <i class="fas fa-edit"></i> Edit
                        </button>
                        <button class="action-button btn-danger" onclick="adminApp.deleteProject('${project.id}')">
                            <i class="fas fa-trash"></i> Delete
                        </button>
                    </td>
                </tr>
            `;
        });

        html += '</tbody></table>';
        container.innerHTML = html;
    }

    renderCredentialsTab() {
        const container = document.getElementById('credentialsTable');
        
        if (this.credentials.length === 0) {
            container.innerHTML = '<p>No credentials found.</p>';
            return;
        }

        let html = `
            <table style="width: 100%; border-collapse: collapse;">
                <thead>
                    <tr style="background: #f9fafb; border-bottom: 1px solid #e5e7eb;">
                        <th style="padding: 1rem; text-align: left;">Name</th>
                        <th style="padding: 1rem; text-align: left;">Provider</th>
                        <th style="padding: 1rem; text-align: left;">Status</th>
                        <th style="padding: 1rem; text-align: left;">Last Sync</th>
                        <th style="padding: 1rem; text-align: left;">Actions</th>
                    </tr>
                </thead>
                <tbody>
        `;

        this.credentials.forEach(cred => {
            html += `
                <tr style="border-bottom: 1px solid #e5e7eb;">
                    <td style="padding: 1rem;">${cred.name}</td>
                    <td style="padding: 1rem;">${cred.provider}</td>
                    <td style="padding: 1rem;">
                        <span class="status-indicator ${cred.enabled ? 'status-active' : 'status-inactive'}">
                            ${cred.enabled ? 'Enabled' : 'Disabled'}
                        </span>
                    </td>
                    <td style="padding: 1rem;">${cred.last_sync ? this.formatDate(cred.last_sync) : 'Never'}</td>
                    <td style="padding: 1rem;">
                        <button class="action-button btn-primary" onclick="adminApp.editCredentials('${cred.id}')">
                            <i class="fas fa-edit"></i> Edit
                        </button>
                        <button class="action-button btn-secondary" onclick="adminApp.testCredentials('${cred.id}')">
                            <i class="fas fa-check"></i> Test
                        </button>
                        <button class="action-button btn-danger" onclick="adminApp.deleteCredentials('${cred.id}')">
                            <i class="fas fa-trash"></i> Delete
                        </button>
                    </td>
                </tr>
            `;
        });

        html += '</tbody></table>';
        container.innerHTML = html;
    }

    renderBulkTab() {
        // Populate decommission domains
        const container = document.getElementById('decommissionDomains');
        
        let html = '';
        this.domains.forEach(domain => {
            html += `
                <div style="margin: 0.5rem 0;">
                    <label>
                        <input type="checkbox" name="decommissionDomain" value="${domain.id}">
                        ${domain.name} (${domain.provider})
                    </label>
                </div>
            `;
        });
        
        container.innerHTML = html;
    }

    // Action methods
    async triggerSync() {
        try {
            const response = await this.apiCall('POST', '/admin/sync/manual', {
                force_refresh: true
            });
            
            if (response.ok) {
                alert('Sync initiated successfully');
                await this.loadDomains();
                this.renderDomainsTab();
            } else {
                const error = await response.json();
                alert('Sync failed: ' + error.error);
            }
        } catch (error) {
            alert('Sync failed: ' + error.message);
        }
    }

    async loadDNSRecords() {
        const domainId = document.getElementById('dnsSelect').value;
        if (!domainId) {
            alert('Please select a domain');
            return;
        }

        try {
            const response = await this.apiCall('GET', `/admin/domains/${domainId}/dns`);
            
            if (response.ok) {
                const data = await response.json();
                this.renderDNSRecords(data.records);
            } else {
                const error = await response.json();
                alert('Failed to load DNS records: ' + error.error);
            }
        } catch (error) {
            alert('Failed to load DNS records: ' + error.message);
        }
    }

    renderDNSRecords(records) {
        const container = document.getElementById('dnsRecords');
        
        if (records.length === 0) {
            container.innerHTML = '<p>No DNS records found.</p>';
            return;
        }

        let html = `
            <table style="width: 100%; border-collapse: collapse;">
                <thead>
                    <tr style="background: #f9fafb; border-bottom: 1px solid #e5e7eb;">
                        <th style="padding: 1rem; text-align: left;">Type</th>
                        <th style="padding: 1rem; text-align: left;">Name</th>
                        <th style="padding: 1rem; text-align: left;">Value</th>
                        <th style="padding: 1rem; text-align: left;">TTL</th>
                        <th style="padding: 1rem; text-align: left;">Actions</th>
                    </tr>
                </thead>
                <tbody>
        `;

        records.forEach(record => {
            html += `
                <tr style="border-bottom: 1px solid #e5e7eb;">
                    <td style="padding: 1rem;">${record.type}</td>
                    <td style="padding: 1rem;">${record.name}</td>
                    <td style="padding: 1rem;">${record.value}</td>
                    <td style="padding: 1rem;">${record.ttl}</td>
                    <td style="padding: 1rem;">
                        <button class="action-button btn-primary" onclick="adminApp.editDNSRecord('${record.id}')">
                            <i class="fas fa-edit"></i> Edit
                        </button>
                        <button class="action-button btn-danger" onclick="adminApp.deleteDNSRecord('${record.id}')">
                            <i class="fas fa-trash"></i> Delete
                        </button>
                    </td>
                </tr>
            `;
        });

        html += '</tbody></table>';
        container.innerHTML = html;
    }

    async bulkPurchase() {
        const domainsText = document.getElementById('purchaseDomains').value;
        const provider = document.getElementById('purchaseProvider').value;
        
        const domains = domainsText.split('\n').filter(d => d.trim()).map(d => d.trim());
        
        if (domains.length === 0) {
            alert('Please enter domains to purchase');
            return;
        }

        try {
            const response = await this.apiCall('POST', '/admin/domains/bulk-purchase', {
                domains,
                provider,
                credentials_id: 'default', // This would come from a selector
                years: 1,
                auto_renew: true
            });
            
            if (response.ok) {
                const data = await response.json();
                alert(`Bulk purchase initiated for ${domains.length} domains`);
                document.getElementById('purchaseDomains').value = '';
            } else {
                const error = await response.json();
                alert('Bulk purchase failed: ' + error.error);
            }
        } catch (error) {
            alert('Bulk purchase failed: ' + error.message);
        }
    }

    async bulkDecommission() {
        const checkboxes = document.querySelectorAll('input[name="decommissionDomain"]:checked');
        const domainIds = Array.from(checkboxes).map(cb => cb.value);
        
        if (domainIds.length === 0) {
            alert('Please select domains to decommission');
            return;
        }

        const stopAutoRenew = document.getElementById('stopAutoRenew').checked;
        const transferOut = document.getElementById('transferOut').checked;

        if (!confirm(`Decommission ${domainIds.length} domains?`)) {
            return;
        }

        try {
            const response = await this.apiCall('POST', '/admin/domains/bulk-decommission', {
                domain_ids: domainIds,
                stop_auto_renew: stopAutoRenew,
                transfer_out: transferOut,
                delete_dns: false
            });
            
            if (response.ok) {
                const data = await response.json();
                alert(`Decommissioned ${data.processed} out of ${data.total} domains`);
                await this.loadDomains();
                this.renderBulkTab();
            } else {
                const error = await response.json();
                alert('Bulk decommission failed: ' + error.error);
            }
        } catch (error) {
            alert('Bulk decommission failed: ' + error.message);
        }
    }

    // Utility methods
    calculateDaysLeft(expiresAt) {
        const now = new Date();
        const expires = new Date(expiresAt);
        const diffTime = expires - now;
        return Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    }

    getStatusClass(daysLeft) {
        if (daysLeft < 0) return 'status-inactive';
        if (daysLeft <= 30) return 'status-inactive';
        return 'status-active';
    }

    formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric'
        });
    }

    // Placeholder methods for other actions
    editDomain(id) {
        alert('Edit domain: ' + id);
    }

    viewDNS(id) {
        this.switchTab('dns');
        document.getElementById('dnsSelect').value = id;
        this.loadDNSRecords();
    }

    showCreateCategory() {
        alert('Create category dialog would open here');
    }

    editCategory(id) {
        alert('Edit category: ' + id);
    }

    deleteCategory(id) {
        if (confirm('Delete this category?')) {
            alert('Delete category: ' + id);
        }
    }

    showCreateProject() {
        alert('Create project dialog would open here');
    }

    editProject(id) {
        alert('Edit project: ' + id);
    }

    deleteProject(id) {
        if (confirm('Delete this project?')) {
            alert('Delete project: ' + id);
        }
    }

    showCreateCredentials() {
        alert('Create credentials dialog would open here');
    }

    editCredentials(id) {
        alert('Edit credentials: ' + id);
    }

    deleteCredentials(id) {
        if (confirm('Delete these credentials?')) {
            alert('Delete credentials: ' + id);
        }
    }

    testCredentials(id) {
        alert('Test credentials: ' + id);
    }

    editDNSRecord(id) {
        alert('Edit DNS record: ' + id);
    }

    deleteDNSRecord(id) {
        if (confirm('Delete this DNS record?')) {
            alert('Delete DNS record: ' + id);
        }
    }
}

// Initialize the admin app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.adminApp = new AdminApp();
});
