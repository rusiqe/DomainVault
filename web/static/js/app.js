// DomainVault Frontend Application
class DomainVaultApp {
    constructor() {
        this.apiBase = window.location.origin + '/api/v1';
        this.currentSection = 'dashboard';
        this.domains = [];
        this.summary = null;
        this.syncStatus = null;
        
        this.init();
    }

    async init() {
        this.setupEventListeners();
        this.showLoading();
        
        try {
            await this.loadInitialData();
            this.hideLoading();
            this.showToast('Application loaded successfully', 'success');
        } catch (error) {
            this.hideLoading();
            this.showToast('Failed to load initial data: ' + error.message, 'error');
            console.error('Init error:', error);
        }
    }

    setupEventListeners() {
        // Navigation
        document.querySelectorAll('.nav-link').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const section = link.dataset.section;
                this.showSection(section);
            });
        });

        // Sync buttons
        document.getElementById('syncBtn').addEventListener('click', () => this.triggerSync());
        document.getElementById('manualSyncBtn').addEventListener('click', () => this.triggerSync());

        // Search and filters
        document.getElementById('searchInput').addEventListener('input', (e) => {
            this.filterDomains();
        });
        
        document.getElementById('providerFilter').addEventListener('change', (e) => {
            this.filterDomains();
        });

        // Expiring filter tabs
        document.querySelectorAll('.filter-tab').forEach(tab => {
            tab.addEventListener('click', (e) => {
                document.querySelectorAll('.filter-tab').forEach(t => t.classList.remove('active'));
                tab.classList.add('active');
                const days = tab.dataset.days;
                this.loadExpiringDomains(days);
            });
        });

        // Auto-refresh data every 30 seconds
        setInterval(() => {
            this.refreshData();
        }, 30000);

        // Check status every 10 seconds
        setInterval(() => {
            this.checkStatus();
        }, 10000);
    }

    async loadInitialData() {
        await Promise.all([
            this.loadDomains(),
            this.loadSummary(),
            this.loadSyncStatus()
        ]);

        this.renderDashboard();
        this.renderDomains();
        this.renderExpiringDomains();
        this.renderSyncStatus();
    }

    async refreshData() {
        try {
            await this.loadSummary();
            this.updateDashboardStats();
        } catch (error) {
            console.error('Refresh error:', error);
        }
    }

    async checkStatus() {
        try {
            const response = await fetch(`${this.apiBase}/health`);
            const statusIndicator = document.getElementById('statusIndicator');
            const statusDot = statusIndicator.querySelector('.status-dot');
            const statusText = statusIndicator.querySelector('.status-text');
            
            if (response.ok) {
                statusDot.className = 'status-dot status-online';
                statusText.textContent = 'Online';
            } else {
                statusDot.className = 'status-dot status-offline';
                statusText.textContent = 'Offline';
            }
        } catch (error) {
            const statusIndicator = document.getElementById('statusIndicator');
            const statusDot = statusIndicator.querySelector('.status-dot');
            const statusText = statusIndicator.querySelector('.status-text');
            statusDot.className = 'status-dot status-offline';
            statusText.textContent = 'Offline';
        }
    }

    async loadDomains() {
        const response = await fetch(`${this.apiBase}/domains`);
        if (!response.ok) throw new Error('Failed to load domains');
        const data = await response.json();
        this.domains = data.domains || [];
    }

    async loadSummary() {
        const response = await fetch(`${this.apiBase}/domains/summary`);
        if (!response.ok) throw new Error('Failed to load summary');
        this.summary = await response.json();
    }

    async loadSyncStatus() {
        const response = await fetch(`${this.apiBase}/sync/status`);
        if (!response.ok) throw new Error('Failed to load sync status');
        this.syncStatus = await response.json();
    }

    async loadExpiringDomains(days = 30) {
        try {
            const response = await fetch(`${this.apiBase}/domains/expiring?days=${days}`);
            if (!response.ok) throw new Error('Failed to load expiring domains');
            const data = await response.json();
            this.renderExpiringDomainsList(data.domains || []);
        } catch (error) {
            this.showToast('Failed to load expiring domains: ' + error.message, 'error');
        }
    }

    async triggerSync() {
        try {
            this.showSyncProgress();
            const response = await fetch(`${this.apiBase}/sync`, { method: 'POST' });
            
            if (!response.ok) throw new Error('Failed to trigger sync');
            
            this.showToast('Sync started successfully', 'success');
            
            // Simulate sync progress
            this.animateSyncProgress();
            
            // Refresh data after sync
            setTimeout(async () => {
                await this.loadInitialData();
                this.hideSyncProgress();
                this.showToast('Sync completed successfully', 'success');
            }, 3000);
            
        } catch (error) {
            this.hideSyncProgress();
            this.showToast('Failed to start sync: ' + error.message, 'error');
        }
    }

    showSection(section) {
        // Update navigation
        document.querySelectorAll('.nav-link').forEach(link => {
            link.classList.remove('active');
        });
        document.querySelector(`[data-section="${section}"]`).classList.add('active');

        // Update content
        document.querySelectorAll('.content-section').forEach(section => {
            section.classList.remove('active');
        });
        document.getElementById(`${section}-section`).classList.add('active');

        this.currentSection = section;

        // Load section-specific data
        if (section === 'expiring') {
            this.loadExpiringDomains();
        }
    }

    renderDashboard() {
        this.updateDashboardStats();
        this.renderProviderDistribution();
        this.renderExpirationTimeline();
    }

    updateDashboardStats() {
        if (!this.summary) return;

        document.getElementById('totalDomains').textContent = this.summary.total || 0;
        document.getElementById('expiringDomains').textContent = this.summary.expiring_in?.['30_days'] || 0;
        document.getElementById('activeProviders').textContent = Object.keys(this.summary.by_provider || {}).length;
        
        if (this.summary.last_sync) {
            const lastSync = new Date(this.summary.last_sync);
            document.getElementById('lastSync').textContent = this.formatTimeAgo(lastSync);
        }
    }

    renderProviderDistribution() {
        const container = document.getElementById('providerList');
        if (!this.summary?.by_provider) {
            container.innerHTML = '<p class="text-gray-500">No provider data available</p>';
            return;
        }

        container.innerHTML = '';
        
        Object.entries(this.summary.by_provider).forEach(([provider, count]) => {
            const item = document.createElement('div');
            item.className = 'provider-item';
            item.innerHTML = `
                <div class="provider-info">
                    <div class="provider-icon">${provider.charAt(0).toUpperCase()}</div>
                    <div class="provider-details">
                        <h4>${this.capitalizeFirst(provider)}</h4>
                        <p>${count} domain${count !== 1 ? 's' : ''}</p>
                    </div>
                </div>
                <div class="provider-count">${count}</div>
            `;
            container.appendChild(item);
        });
    }

    renderExpirationTimeline() {
        const container = document.getElementById('expirationTimeline');
        if (!this.summary?.expiring_in) {
            container.innerHTML = '<p class="text-gray-500">No expiration data available</p>';
            return;
        }

        container.innerHTML = '';

        const timeframes = [
            { key: '30_days', label: '30 Days', color: '#dc2626' },
            { key: '90_days', label: '90 Days', color: '#d97706' },
            { key: '365_days', label: '1 Year', color: '#059669' }
        ];

        timeframes.forEach(timeframe => {
            const count = this.summary.expiring_in[timeframe.key] || 0;
            const item = document.createElement('div');
            item.className = 'provider-item';
            item.innerHTML = `
                <div class="provider-info">
                    <div class="provider-icon" style="background: ${timeframe.color}">
                        <i class="fas fa-calendar-alt"></i>
                    </div>
                    <div class="provider-details">
                        <h4>Within ${timeframe.label}</h4>
                        <p>${count} domain${count !== 1 ? 's' : ''} expiring</p>
                    </div>
                </div>
                <div class="provider-count">${count}</div>
            `;
            container.appendChild(item);
        });
    }

    renderDomains() {
        const tbody = document.getElementById('domainsTableBody');
        const resultsCount = document.getElementById('resultsCount');
        
        if (!this.domains || this.domains.length === 0) {
            tbody.innerHTML = '<tr><td colspan="6" class="text-center text-gray-500 py-8">No domains found</td></tr>';
            resultsCount.textContent = '0 domains';
            return;
        }

        tbody.innerHTML = '';
        
        this.domains.forEach(domain => {
            const daysLeft = this.calculateDaysLeft(domain.expires_at);
            const status = this.getDomainStatus(daysLeft);
            const statusClass = this.getStatusClass(status);
            const daysClass = this.getDaysClass(daysLeft);

            const row = document.createElement('tr');
            row.innerHTML = `
                <td>
                    <div class="domain-name">${domain.name}</div>
                </td>
                <td>
                    <span class="provider-badge">${this.capitalizeFirst(domain.provider)}</span>
                </td>
                <td>${this.formatDate(domain.expires_at)}</td>
                <td>
                    <span class="status-badge ${statusClass}">${status}</span>
                </td>
                <td>
                    <span class="days-left ${daysClass}">${daysLeft >= 0 ? daysLeft : 'Expired'}</span>
                </td>
                <td>
                    <button class="btn btn-secondary btn-sm" onclick="app.viewDomain('${domain.id}')">
                        <i class="fas fa-eye"></i>
                    </button>
                </td>
            `;
            tbody.appendChild(row);
        });

        resultsCount.textContent = `${this.domains.length} domain${this.domains.length !== 1 ? 's' : ''}`;
        this.populateProviderFilter();
    }

    renderExpiringDomains() {
        this.loadExpiringDomains(30);
    }

    renderExpiringDomainsList(domains) {
        const container = document.getElementById('expiringDomainsList');
        
        if (!domains || domains.length === 0) {
            container.innerHTML = '<div class="text-center text-gray-500 py-8">No expiring domains found</div>';
            return;
        }

        container.innerHTML = '';
        
        domains.forEach(domain => {
            const daysLeft = this.calculateDaysLeft(domain.expires_at);
            const severity = daysLeft <= 7 ? 'critical' : 'warning';

            const card = document.createElement('div');
            card.className = `expiring-domain-card ${severity}`;
            card.innerHTML = `
                <div>
                    <h4 class="domain-name">${domain.name}</h4>
                    <p class="text-gray-600">Provider: ${this.capitalizeFirst(domain.provider)}</p>
                    <p class="text-gray-600">Expires: ${this.formatDate(domain.expires_at)}</p>
                </div>
                <div class="text-right">
                    <div class="days-left ${this.getDaysClass(daysLeft)} text-2xl font-bold">
                        ${daysLeft >= 0 ? daysLeft : 'Expired'}
                    </div>
                    <div class="text-sm text-gray-500">days left</div>
                </div>
            `;
            container.appendChild(card);
        });
    }

    renderSyncStatus() {
        const container = document.getElementById('providersStatus');
        
        if (!this.syncStatus?.providers) {
            container.innerHTML = '<div class="text-center text-gray-500 py-8">No provider status available</div>';
            return;
        }

        container.innerHTML = '';
        
        Object.values(this.syncStatus.providers).forEach(provider => {
            const card = document.createElement('div');
            card.className = 'provider-status-card';
            card.innerHTML = `
                <div class="flex items-center justify-between mb-4">
                    <h4 class="text-lg font-semibold">${this.capitalizeFirst(provider.name)}</h4>
                    <span class="status-badge ${provider.enabled ? 'status-active' : 'status-offline'}">
                        ${provider.enabled ? 'Active' : 'Inactive'}
                    </span>
                </div>
                <div class="text-sm text-gray-600">
                    <p>Last Sync: ${provider.last_sync || 'Never'}</p>
                    ${provider.error ? `<p class="text-red-600">Error: ${provider.error}</p>` : ''}
                </div>
            `;
            container.appendChild(card);
        });
    }

    filterDomains() {
        const searchTerm = document.getElementById('searchInput').value.toLowerCase();
        const providerFilter = document.getElementById('providerFilter').value;
        
        let filteredDomains = this.domains;
        
        if (searchTerm) {
            filteredDomains = filteredDomains.filter(domain => 
                domain.name.toLowerCase().includes(searchTerm)
            );
        }
        
        if (providerFilter) {
            filteredDomains = filteredDomains.filter(domain => 
                domain.provider === providerFilter
            );
        }
        
        // Temporarily store the original domains and replace with filtered
        const originalDomains = this.domains;
        this.domains = filteredDomains;
        this.renderDomains();
        this.domains = originalDomains;
    }

    populateProviderFilter() {
        const select = document.getElementById('providerFilter');
        const currentValue = select.value;
        
        // Get unique providers
        const providers = [...new Set(this.domains.map(domain => domain.provider))];
        
        // Clear and repopulate
        select.innerHTML = '<option value="">All Providers</option>';
        providers.forEach(provider => {
            const option = document.createElement('option');
            option.value = provider;
            option.textContent = this.capitalizeFirst(provider);
            select.appendChild(option);
        });
        
        // Restore selection
        select.value = currentValue;
    }

    showSyncProgress() {
        const progressFill = document.getElementById('syncProgress');
        const progressText = document.getElementById('syncProgressText');
        
        progressFill.style.width = '0%';
        progressText.textContent = 'Initializing sync...';
    }

    animateSyncProgress() {
        const progressFill = document.getElementById('syncProgress');
        const progressText = document.getElementById('syncProgressText');
        
        let progress = 0;
        const interval = setInterval(() => {
            progress += Math.random() * 30;
            if (progress > 90) progress = 90;
            
            progressFill.style.width = progress + '%';
            
            if (progress < 30) {
                progressText.textContent = 'Connecting to providers...';
            } else if (progress < 60) {
                progressText.textContent = 'Fetching domain data...';
            } else if (progress < 90) {
                progressText.textContent = 'Updating database...';
            } else {
                progressText.textContent = 'Finalizing sync...';
                clearInterval(interval);
            }
        }, 500);
    }

    hideSyncProgress() {
        const progressFill = document.getElementById('syncProgress');
        const progressText = document.getElementById('syncProgressText');
        
        progressFill.style.width = '100%';
        progressText.textContent = 'Sync completed successfully!';
        
        setTimeout(() => {
            progressFill.style.width = '0%';
            progressText.textContent = 'Ready to sync';
        }, 2000);
    }

    showLoading() {
        document.getElementById('loadingOverlay').classList.add('show');
    }

    hideLoading() {
        document.getElementById('loadingOverlay').classList.remove('show');
    }

    showToast(message, type = 'info') {
        const container = document.getElementById('toastContainer');
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        
        const icon = {
            success: 'fa-check-circle',
            error: 'fa-exclamation-circle',
            warning: 'fa-exclamation-triangle',
            info: 'fa-info-circle'
        }[type] || 'fa-info-circle';
        
        toast.innerHTML = `
            <i class="fas ${icon}"></i>
            <span>${message}</span>
        `;
        
        container.appendChild(toast);
        
        // Animate in
        setTimeout(() => toast.classList.add('show'), 100);
        
        // Auto remove
        setTimeout(() => {
            toast.classList.remove('show');
            setTimeout(() => container.removeChild(toast), 300);
        }, 5000);
    }

    // Utility functions
    calculateDaysLeft(expiresAt) {
        const now = new Date();
        const expires = new Date(expiresAt);
        const diffTime = expires - now;
        return Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    }

    getDomainStatus(daysLeft) {
        if (daysLeft < 0) return 'Expired';
        if (daysLeft <= 7) return 'Critical';
        if (daysLeft <= 30) return 'Expiring';
        return 'Active';
    }

    getStatusClass(status) {
        const classes = {
            'Active': 'status-active',
            'Expiring': 'status-expiring',
            'Critical': 'status-expired',
            'Expired': 'status-expired'
        };
        return classes[status] || 'status-active';
    }

    getDaysClass(daysLeft) {
        if (daysLeft < 0 || daysLeft <= 7) return 'critical';
        if (daysLeft <= 30) return 'warning';
        return 'safe';
    }

    formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric'
        });
    }

    formatTimeAgo(date) {
        const now = new Date();
        const diffInSeconds = Math.floor((now - date) / 1000);
        
        if (diffInSeconds < 60) return 'Just now';
        if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m ago`;
        if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}h ago`;
        return `${Math.floor(diffInSeconds / 86400)}d ago`;
    }

    capitalizeFirst(str) {
        return str.charAt(0).toUpperCase() + str.slice(1);
    }

    viewDomain(domainId) {
        // Placeholder for domain detail view
        this.showToast(`View domain: ${domainId}`, 'info');
    }
}

// Initialize the application when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.app = new DomainVaultApp();
});
