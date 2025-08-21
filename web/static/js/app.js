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
        const domain = (this.domains || []).find(d => d.id === domainId);
        if (!domain) {
            this.showToast('Domain not found', 'error');
            return;
        }
        this.openDomainModal(domain);
    }

    openDomainModal(domain) {
        // Build modal elements
        const overlay = document.createElement('div');
        overlay.className = 'modal-overlay';
        overlay.id = 'domainDetailsOverlay';

        // Compute derived values
        const daysLeft = this.calculateDaysLeft(domain.expires_at);
        const status = this.getDomainStatus(daysLeft);
        const statusClass = this.getStatusClass(status);
        const daysClass = this.getDaysClass(daysLeft);

        const autoRenewText = domain.auto_renew ? 'Enabled' : 'Disabled';
        const autoRenewClass = domain.auto_renew ? 'status-active' : 'status-inactive';
        const renewalPrice = (domain.renewal_price !== undefined && domain.renewal_price !== null)
            ? `$${Number(domain.renewal_price).toFixed(2)}`
            : '—';

        const httpStatus = domain.http_status ?? '—';
        const statusMessage = domain.status_message || (httpStatus === 200 ? 'OK' : (httpStatus === '—' ? '—' : 'Check site'));
        const lastStatusCheck = domain.last_status_check ? this.formatDate(domain.last_status_check) : '—';

        const tags = Array.isArray(domain.tags) ? domain.tags : [];

        // Modal container
        const container = document.createElement('div');
        container.className = 'modal-container';
        container.innerHTML = `
            <div class="modal-header">
                <h3 class="modal-title">
                    <i class="fas fa-globe"></i>
                    ${domain.name}
                </h3>
                <button class="modal-close" aria-label="Close details" id="domainModalCloseBtn">&times;</button>
            </div>
            <div class="modal-content">
                <div class="card">
                    <div class="card-header">
                        <h4 class="card-title"><i class="fas fa-info-circle"></i> Overview</h4>
                    </div>
                    <div class="card-content">
                        <div class="info-grid">
                            <div class="info-item">
                                <strong>Provider</strong>
                                <span>${this.capitalizeFirst(domain.provider || '—')}</span>
                            </div>
                            <div class="info-item">
                                <strong>Expires</strong>
                                <span>${this.formatDate(domain.expires_at)}</span>
                            </div>
                            <div class="info-item">
                                <strong>Days Left</strong>
                                <span class="days-left ${daysClass}">${daysLeft >= 0 ? daysLeft : 'Expired'}</span>
                            </div>
                            <div class="info-item">
                                <strong>Status</strong>
                                <span class="status-badge ${statusClass}">${status}</span>
                            </div>
                            <div class="info-item">
                                <strong>Auto-Renew</strong>
                                <span class="status-badge ${autoRenewClass}">${autoRenewText}</span>
                            </div>
                            <div class="info-item">
                                <strong>Renewal Price</strong>
                                <span>${renewalPrice}</span>
                            </div>
                        </div>
                        ${tags.length ? `
                        <div style="margin-top:16px;">
                            <strong style="display:block; margin-bottom:8px; color:#374151; text-transform:uppercase; font-size:12px; letter-spacing:.05em;">Tags</strong>
                            <div style="display:flex; gap:8px; flex-wrap:wrap;">
                                ${tags.map(t => `<span style="background:#f1f5f9; color:#374151; padding:4px 8px; border-radius:9999px; font-size:12px; border:1px solid #e5e7eb;">${t}</span>`).join('')}
                            </div>
                        </div>` : ''}
                    </div>
                </div>

                <div class="card">
                    <div class="card-header">
                        <h4 class="card-title"><i class="fas fa-heartbeat"></i> Website Status</h4>
                    </div>
                    <div class="card-content">
                        <div class="info-grid">
                            <div class="info-item">
                                <strong>HTTP Status</strong>
                                <span class="website-status ${httpStatus === 200 ? 'status-200' : (httpStatus === '—' ? '' : 'status-error')}">${httpStatus}</span>
                            </div>
                            <div class="info-item">
                                <strong>Last Check</strong>
                                <span>${lastStatusCheck}</span>
                            </div>
                            <div class="info-item" style="grid-column: 1 / -1;">
                                <strong>Message</strong>
                                <span>${statusMessage}</span>
                            </div>
                        </div>
                    </div>
                </div>

                ${domain.project_id || domain.category_id ? `
                <div class="card">
                    <div class="card-header">
                        <h4 class="card-title"><i class="fas fa-folder"></i> Organization</h4>
                    </div>
                    <div class="card-content">
                        <div class="info-grid">
                            ${domain.category_id ? `
                            <div class="info-item">
                                <strong>Category</strong>
                                <span>${domain.category_id}</span>
                            </div>` : ''}
                            ${domain.project_id ? `
                            <div class="info-item">
                                <strong>Project</strong>
                                <span>${domain.project_id}</span>
                            </div>` : ''}
                        </div>
                    </div>
                </div>` : ''}
            </div>
            <div class="modal-footer">
                <a href="http://${domain.name}" target="_blank" rel="noopener" class="btn btn-secondary">
                    <i class="fas fa-external-link-alt"></i> Open Site
                </a>
                <button class="btn btn-primary" id="domainModalCloseBtnFooter">
                    <i class="fas fa-times"></i> Close
                </button>
            </div>
        `;

        overlay.appendChild(container);
        document.body.appendChild(overlay);

        // Close handlers
        const close = () => this.closeDomainModal();
        overlay.addEventListener('click', (e) => {
            if (e.target === overlay) close();
        });
        container.querySelector('#domainModalCloseBtn').addEventListener('click', close);
        container.querySelector('#domainModalCloseBtnFooter').addEventListener('click', close);
    }

    closeDomainModal() {
        const overlay = document.getElementById('domainDetailsOverlay');
        if (overlay) {
            document.body.removeChild(overlay);
        }
    }
}

// Initialize the application when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.app = new DomainVaultApp();
});
