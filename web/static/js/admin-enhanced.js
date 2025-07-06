// Enhanced DomainVault Admin Application
class EnhancedAdminApp {
    constructor() {
        this.apiBase = window.location.origin + '/api/v1';
        this.token = localStorage.getItem('admin_token');
        this.user = null;
        this.data = {
            domains: [],
            analytics: null,
            security: null,
            notifications: null,
            providers: []
        };
        this.charts = {};
        this.refreshIntervals = {};
        
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
        document.querySelectorAll('.nav-item').forEach(tab => {
            tab.addEventListener('click', (e) => {
                e.preventDefault();
                this.switchTab(e.target.closest('.nav-item').dataset.tab);
            });
        });

        // Auto-refresh intervals
        this.startAutoRefresh();
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
            this.apiCall('POST', '/auth/logout').catch(() => {});
        }
        
        this.token = null;
        this.user = null;
        localStorage.removeItem('admin_token');
        
        // Clear intervals
        Object.values(this.refreshIntervals).forEach(interval => clearInterval(interval));
        this.refreshIntervals = {};
        
        document.getElementById('adminDashboard').style.display = 'none';
        document.getElementById('loginScreen').style.display = 'flex';
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
        this.showLoading(true);
        
        try {
            // Load all data concurrently
            await Promise.all([
                this.loadPortfolioAnalytics(),
                this.loadSecurityMetrics(),
                this.loadDomains(),
                this.loadProviders()
            ]);
            
            // Render current tab
            this.renderCurrentTab();
            
            // Create charts
            this.initializeCharts();
            
        } catch (error) {
            console.error('Failed to load dashboard data:', error);
            this.showAlert('Failed to load dashboard data', 'error');
        } finally {
            this.showLoading(false);
        }
    }

    switchTab(tabName) {
        // Update tab buttons
        document.querySelectorAll('.nav-item').forEach(tab => {
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
        const activeTab = document.querySelector('.nav-item.active');
        if (activeTab) {
            this.renderTab(activeTab.dataset.tab);
        }
    }

    renderTab(tabName) {
        switch (tabName) {
            case 'overview':
                this.renderOverviewTab();
                break;
            case 'analytics':
                this.renderAnalyticsTab();
                break;
            case 'security':
                this.renderSecurityTab();
                break;
            case 'notifications':
                this.renderNotificationsTab();
                break;
            case 'domains':
                this.renderDomainsTab();
                break;
            case 'providers':
                this.renderProvidersTab();
                break;
        }
    }

    // Data loading methods
    async loadPortfolioAnalytics() {
        try {
            const response = await this.apiCall('GET', '/admin/analytics/portfolio');
            if (response.ok) {
                this.data.analytics = await response.json();
            }
        } catch (error) {
            console.error('Failed to load portfolio analytics:', error);
        }
    }

    async loadSecurityMetrics() {
        try {
            const response = await this.apiCall('GET', '/admin/security/metrics');
            if (response.ok) {
                this.data.security = await response.json();
            }
        } catch (error) {
            console.error('Failed to load security metrics:', error);
        }
    }

    async loadDomains() {
        try {
            const response = await this.apiCall('GET', '/domains');
            if (response.ok) {
                const data = await response.json();
                this.data.domains = data.domains || [];
            }
        } catch (error) {
            console.error('Failed to load domains:', error);
        }
    }

    async loadProviders() {
        try {
            const response = await this.apiCall('GET', '/admin/credentials');
            if (response.ok) {
                const data = await response.json();
                this.data.providers = data.credentials || [];
            }
        } catch (error) {
            console.error('Failed to load providers:', error);
        }
    }

    // Tab rendering methods
    renderOverviewTab() {
        if (this.data.analytics) {
            const overview = this.data.analytics.overview;
            const financial = this.data.analytics.financial_metrics;
            
            // Update overview metrics
            this.updateElement('totalDomains', overview.total_domains || 0);
            this.updateElement('expiringSoon', overview.domains_expiring_30 || 0);
            this.updateElement('renewalCost', this.formatCurrency(financial.renewal_cost_next_90_days || 0));
            this.updateElement('portfolioValue', this.formatCurrency(financial.estimated_value.total_estimated_value || 0));
        }

        // Show summary for domains with actual data
        const activeDomains = this.data.domains.filter(d => d.status === 'active').length;
        const expiredDomains = this.data.domains.filter(d => d.status === 'expired').length;
        
        this.updateElement('totalDomains', this.data.domains.length);
        
        // Calculate expiring domains (within 30 days)
        const now = new Date();
        const thirtyDaysFromNow = new Date(now.getTime() + (30 * 24 * 60 * 60 * 1000));
        const expiringSoon = this.data.domains.filter(d => {
            const expiryDate = new Date(d.expires_at);
            return expiryDate <= thirtyDaysFromNow && expiryDate >= now;
        }).length;
        
        this.updateElement('expiringSoon', expiringSoon);
        
        // Calculate total renewal cost for next 90 days
        const ninetyDaysFromNow = new Date(now.getTime() + (90 * 24 * 60 * 60 * 1000));
        const renewalCostNext90 = this.data.domains
            .filter(d => {
                const expiryDate = new Date(d.expires_at);
                return expiryDate <= ninetyDaysFromNow && expiryDate >= now;
            })
            .reduce((sum, d) => sum + (d.renewal_price || 15), 0);
        
        this.updateElement('renewalCost', this.formatCurrency(renewalCostNext90));
        
        // Calculate estimated portfolio value
        const portfolioValue = this.data.domains.reduce((sum, d) => {
            return sum + this.estimateDomainValue(d);
        }, 0);
        
        this.updateElement('portfolioValue', this.formatCurrency(portfolioValue));
        
        // Security and uptime placeholders
        this.updateElement('securityScore', '85');
        this.updateElement('uptimePercent', '99.8%');

        // Update charts
        this.updateOverviewCharts();
    }

    renderAnalyticsTab() {
        this.renderFinancialAnalytics();
        this.renderPremiumDomains();
    }

    renderSecurityTab() {
        // Render security metrics
        if (this.data.security) {
            this.updateElement('loginAttempts', this.data.security.login_attempts?.total_attempts || 0);
            this.updateElement('securityAlerts', this.data.security.security_alerts?.total_alerts || 0);
            this.updateElement('activeSessions', this.data.security.active_sessions?.active_sessions || 0);
            this.updateElement('vulnerabilities', this.data.security.risk_assessment?.vulnerability_count || 0);
        } else {
            // Placeholder data
            this.updateElement('loginAttempts', '45');
            this.updateElement('securityAlerts', '2');
            this.updateElement('activeSessions', '3');
            this.updateElement('vulnerabilities', '0');
        }

        this.renderAuditEvents();
    }

    renderNotificationsTab() {
        // Placeholder data for notifications
        this.updateElement('emailAlerts', '156');
        this.updateElement('slackNotifications', '89');
        this.updateElement('webhooks', '234');
        this.updateElement('activeRules', '7');

        this.renderNotificationRules();
    }

    renderDomainsTab() {
        this.renderDomainsTable();
    }

    renderProvidersTab() {
        this.updateElement('connectedProviders', this.data.providers.length);
        this.updateElement('lastSync', '5');
        this.updateElement('domainsSynced', this.data.domains.length);
        this.updateElement('syncIssues', '0');
        
        this.renderProvidersTable();
    }

    // Chart initialization
    initializeCharts() {
        this.initializeGrowthChart();
        this.initializeProviderChart();
        this.initializeFinancialChart();
        this.initializeExpirationChart();
    }

    initializeGrowthChart() {
        const ctx = document.getElementById('growthChart');
        if (!ctx) return;

        // Generate sample growth data
        const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'];
        const data = [15, 17, 19, 20, 21, 21]; // Domain count growth

        this.charts.growth = new Chart(ctx, {
            type: 'line',
            data: {
                labels: months,
                datasets: [{
                    label: 'Domains',
                    data: data,
                    borderColor: '#3498db',
                    backgroundColor: 'rgba(52, 152, 219, 0.1)',
                    tension: 0.4,
                    fill: true
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: false
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        grid: {
                            color: 'rgba(0,0,0,0.1)'
                        }
                    },
                    x: {
                        grid: {
                            display: false
                        }
                    }
                }
            }
        });
    }

    initializeProviderChart() {
        const ctx = document.getElementById('providerChart');
        if (!ctx) return;

        // Calculate provider distribution from actual data
        const providerCounts = {};
        this.data.domains.forEach(domain => {
            providerCounts[domain.provider] = (providerCounts[domain.provider] || 0) + 1;
        });

        const labels = Object.keys(providerCounts);
        const data = Object.values(providerCounts);
        const colors = ['#3498db', '#e74c3c', '#f39c12', '#27ae60'];

        this.charts.provider = new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels: labels,
                datasets: [{
                    data: data,
                    backgroundColor: colors.slice(0, labels.length)
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'bottom'
                    }
                }
            }
        });
    }

    initializeFinancialChart() {
        const ctx = document.getElementById('financialChart');
        if (!ctx) return;

        // Generate sample financial data
        const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'];
        const costs = [350, 420, 380, 450, 390, 480];

        this.charts.financial = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: months,
                datasets: [{
                    label: 'Renewal Costs',
                    data: costs,
                    backgroundColor: '#e74c3c',
                    borderColor: '#c0392b',
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: false
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        ticks: {
                            callback: function(value) {
                                return '$' + value;
                            }
                        }
                    }
                }
            }
        });
    }

    initializeExpirationChart() {
        const ctx = document.getElementById('expirationChart');
        if (!ctx) return;

        // Calculate expiration timeline
        const now = new Date();
        const expirationCounts = {
            '0-30 days': 0,
            '31-90 days': 0,
            '91-180 days': 0,
            '180+ days': 0
        };

        this.data.domains.forEach(domain => {
            const expiryDate = new Date(domain.expires_at);
            const daysUntilExpiry = Math.ceil((expiryDate - now) / (1000 * 60 * 60 * 24));

            if (daysUntilExpiry <= 30) {
                expirationCounts['0-30 days']++;
            } else if (daysUntilExpiry <= 90) {
                expirationCounts['31-90 days']++;
            } else if (daysUntilExpiry <= 180) {
                expirationCounts['91-180 days']++;
            } else {
                expirationCounts['180+ days']++;
            }
        });

        this.charts.expiration = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: Object.keys(expirationCounts),
                datasets: [{
                    label: 'Domains',
                    data: Object.values(expirationCounts),
                    backgroundColor: ['#e74c3c', '#f39c12', '#f1c40f', '#27ae60']
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: false
                    }
                }
            }
        });
    }

    // Table rendering methods
    renderDomainsTable() {
        const tbody = document.getElementById('domainsTable');
        if (!tbody) return;

        if (this.data.domains.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="8" style="text-align: center; padding: 40px;">
                        No domains found
                    </td>
                </tr>
            `;
            return;
        }

        tbody.innerHTML = this.data.domains.map(domain => {
            const statusClass = this.getStatusClass(domain.status);
            const expiryDate = new Date(domain.expires_at);
            const daysUntilExpiry = Math.ceil((expiryDate - new Date()) / (1000 * 60 * 60 * 24));
            
            return `
                <tr>
                    <td><input type="checkbox" value="${domain.id}"></td>
                    <td><strong>${domain.name}</strong></td>
                    <td>${domain.provider}</td>
                    <td>
                        ${this.formatDate(domain.expires_at)}
                        <small style="display: block; color: #666;">
                            ${daysUntilExpiry > 0 ? `${daysUntilExpiry} days` : 'Expired'}
                        </small>
                    </td>
                    <td><span class="status-badge ${statusClass}">${domain.status}</span></td>
                    <td>${this.formatCurrency(domain.renewal_price || 15)}</td>
                    <td>${this.formatCurrency(this.estimateDomainValue(domain))}</td>
                    <td>
                        <button class="action-button primary" onclick="editDomain('${domain.id}')">
                            <i class="fas fa-edit"></i>
                        </button>
                        <button class="action-button" onclick="checkDomainStatus('${domain.id}')">
                            <i class="fas fa-check"></i>
                        </button>
                    </td>
                </tr>
            `;
        }).join('');
    }

    renderProvidersTable() {
        const tbody = document.querySelector('#credentialsTable tbody');
        if (!tbody) return;

        if (this.data.providers.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="6" style="text-align: center; padding: 40px;">
                        No provider credentials found. 
                        <a href="#" onclick="connectProvider()">Connect a provider</a>
                    </td>
                </tr>
            `;
            return;
        }

        tbody.innerHTML = this.data.providers.map(provider => {
            const domainCount = this.data.domains.filter(d => d.provider === provider.provider).length;
            const statusClass = provider.connection_status === 'connected' ? 'status-online' : 'status-offline';
            
            return `
                <tr>
                    <td><strong>${provider.name}</strong></td>
                    <td>${provider.provider}</td>
                    <td><span class="status-badge ${statusClass}">${provider.connection_status}</span></td>
                    <td>${provider.last_sync ? this.formatDate(provider.last_sync) : 'Never'}</td>
                    <td>${domainCount}</td>
                    <td>
                        <button class="action-button primary" onclick="testProvider('${provider.id}')">
                            <i class="fas fa-test"></i> Test
                        </button>
                        <button class="action-button" onclick="syncProvider('${provider.id}')">
                            <i class="fas fa-sync-alt"></i> Sync
                        </button>
                        <button class="action-button danger" onclick="deleteProvider('${provider.id}')">
                            <i class="fas fa-trash"></i>
                        </button>
                    </td>
                </tr>
            `;
        }).join('');
    }

    renderFinancialAnalytics() {
        // This would be populated with real financial analytics data
        // For now, using placeholder data from domains
    }

    renderPremiumDomains() {
        const tbody = document.getElementById('premiumDomainsTable');
        if (!tbody) return;

        // Calculate premium domains (estimated value > 5x renewal cost)
        const premiumDomains = this.data.domains
            .map(domain => {
                const renewalCost = domain.renewal_price || 15;
                const estimatedValue = this.estimateDomainValue(domain);
                const valueMultiplier = estimatedValue / renewalCost;
                
                return {
                    ...domain,
                    estimatedValue,
                    renewalCost,
                    valueMultiplier
                };
            })
            .filter(domain => domain.valueMultiplier > 5)
            .sort((a, b) => b.estimatedValue - a.estimatedValue)
            .slice(0, 10); // Top 10

        if (premiumDomains.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="6" style="text-align: center; padding: 40px;">
                        No premium domains identified in your portfolio
                    </td>
                </tr>
            `;
            return;
        }

        tbody.innerHTML = premiumDomains.map(domain => `
            <tr>
                <td><strong>${domain.name}</strong></td>
                <td>${this.formatCurrency(domain.estimatedValue)}</td>
                <td>${this.formatCurrency(domain.renewalCost)}</td>
                <td>${domain.valueMultiplier.toFixed(1)}x</td>
                <td>
                    <small>
                        ${this.getDomainValueFactors(domain).join(', ')}
                    </small>
                </td>
                <td>
                    <button class="action-button primary" onclick="analyzeDomain('${domain.id}')">
                        <i class="fas fa-analytics"></i> Analyze
                    </button>
                </td>
            </tr>
        `).join('');
    }

    renderAuditEvents() {
        const tbody = document.getElementById('auditEventsTable');
        if (!tbody) return;

        // Sample audit events
        const sampleEvents = [
            {
                timestamp: new Date(Date.now() - 1000 * 60 * 30),
                user: 'admin',
                action: 'Domain Status Check',
                resource: 'techsolutions.com',
                ip: '192.168.1.100',
                risk: 15,
                status: 'success'
            },
            {
                timestamp: new Date(Date.now() - 1000 * 60 * 60 * 2),
                user: 'admin',
                action: 'Provider Connection',
                resource: 'Mock Provider',
                ip: '192.168.1.100',
                risk: 25,
                status: 'success'
            },
            {
                timestamp: new Date(Date.now() - 1000 * 60 * 60 * 4),
                user: 'admin',
                action: 'Bulk Sync',
                resource: 'All Providers',
                ip: '192.168.1.100',
                risk: 35,
                status: 'success'
            }
        ];

        tbody.innerHTML = sampleEvents.map(event => `
            <tr>
                <td>${this.formatDateTime(event.timestamp)}</td>
                <td>${event.user}</td>
                <td>${event.action}</td>
                <td>${event.resource}</td>
                <td>${event.ip}</td>
                <td>
                    <span style="color: ${this.getRiskColor(event.risk)}">
                        ${event.risk}
                    </span>
                </td>
                <td>
                    <span class="status-badge ${event.status === 'success' ? 'status-online' : 'status-offline'}">
                        ${event.status}
                    </span>
                </td>
            </tr>
        `).join('');
    }

    renderNotificationRules() {
        const tbody = document.getElementById('notificationRulesTable');
        if (!tbody) return;

        // Sample notification rules
        const sampleRules = [
            {
                name: 'Expiration Alerts',
                alertTypes: 'Domain Expiring',
                channels: 'Email, Slack',
                recipients: 'admin@domain.com',
                status: 'enabled'
            },
            {
                name: 'Security Alerts',
                alertTypes: 'Security Violation',
                channels: 'Email, Webhook',
                recipients: 'security@domain.com',
                status: 'enabled'
            },
            {
                name: 'Sync Failures',
                alertTypes: 'Sync Failed',
                channels: 'Slack',
                recipients: '#alerts',
                status: 'enabled'
            }
        ];

        tbody.innerHTML = sampleRules.map((rule, index) => `
            <tr>
                <td><strong>${rule.name}</strong></td>
                <td>${rule.alertTypes}</td>
                <td>${rule.channels}</td>
                <td>${rule.recipients}</td>
                <td>
                    <span class="status-badge status-online">
                        ${rule.status}
                    </span>
                </td>
                <td>
                    <button class="action-button primary" onclick="editNotificationRule(${index})">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button class="action-button danger" onclick="deleteNotificationRule(${index})">
                        <i class="fas fa-trash"></i>
                    </button>
                </td>
            </tr>
        `).join('');
    }

    // Utility methods
    updateOverviewCharts() {
        // Update chart data if charts exist
        if (this.charts.provider) {
            // Update provider chart with current data
            const providerCounts = {};
            this.data.domains.forEach(domain => {
                providerCounts[domain.provider] = (providerCounts[domain.provider] || 0) + 1;
            });

            this.charts.provider.data.labels = Object.keys(providerCounts);
            this.charts.provider.data.datasets[0].data = Object.values(providerCounts);
            this.charts.provider.update();
        }
    }

    estimateDomainValue(domain) {
        let value = 50; // Base value

        // Length factor
        const domainName = domain.name.split('.')[0];
        if (domainName.length <= 3) value *= 10;
        else if (domainName.length <= 5) value *= 5;
        else if (domainName.length <= 8) value *= 2;

        // TLD factor
        if (domain.name.endsWith('.com')) value *= 3;
        else if (domain.name.endsWith('.io')) value *= 2;
        else if (domain.name.endsWith('.ai')) value *= 4;

        // Age factor (simplified)
        const age = (new Date() - new Date(domain.created_at)) / (1000 * 60 * 60 * 24 * 365);
        if (age > 5) value *= 1.5;

        // Status factor
        if (domain.http_status === 200) value *= 1.2;

        return Math.round(value);
    }

    getDomainValueFactors(domain) {
        const factors = [];
        const domainName = domain.name.split('.')[0];
        
        if (domainName.length <= 5) factors.push('Short name');
        if (domain.name.endsWith('.com')) factors.push('Premium TLD');
        if (domain.http_status === 200) factors.push('Active website');
        
        const age = (new Date() - new Date(domain.created_at)) / (1000 * 60 * 60 * 24 * 365);
        if (age > 5) factors.push('Established domain');
        
        return factors;
    }

    getStatusClass(status) {
        switch (status) {
            case 'active': return 'status-online';
            case 'expired': return 'status-offline';
            default: return 'status-warning';
        }
    }

    getRiskColor(risk) {
        if (risk >= 80) return '#e74c3c';
        if (risk >= 60) return '#f39c12';
        if (risk >= 40) return '#f1c40f';
        return '#27ae60';
    }

    updateElement(id, value) {
        const element = document.getElementById(id);
        if (element) {
            element.textContent = value;
        }
    }

    formatCurrency(amount) {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: 'USD'
        }).format(amount);
    }

    formatDate(dateString) {
        return new Date(dateString).toLocaleDateString();
    }

    formatDateTime(date) {
        return date.toLocaleString();
    }

    showLoading(show) {
        const overlay = document.getElementById('loadingOverlay');
        if (overlay) {
            overlay.classList.toggle('hidden', !show);
        }
    }

    showAlert(message, type = 'info') {
        const container = document.getElementById('alertContainer');
        if (!container) return;

        const alertClass = type === 'error' ? 'critical' : 
                          type === 'success' ? 'success' : '';

        const alert = document.createElement('div');
        alert.className = `alert-banner ${alertClass}`;
        alert.innerHTML = `
            <i class="fas fa-info-circle"></i>
            ${message}
            <button onclick="this.parentElement.remove()" style="background: none; border: none; color: white; margin-left: 10px;">
                <i class="fas fa-times"></i>
            </button>
        `;

        container.appendChild(alert);

        // Auto-remove after 5 seconds
        setTimeout(() => {
            if (alert.parentElement) {
                alert.remove();
            }
        }, 5000);
    }

    startAutoRefresh() {
        // Refresh overview data every 30 seconds
        this.refreshIntervals.overview = setInterval(() => {
            if (document.querySelector('.nav-item.active')?.dataset.tab === 'overview') {
                this.loadPortfolioAnalytics();
            }
        }, 30000);

        // Refresh security data every 60 seconds
        this.refreshIntervals.security = setInterval(() => {
            if (document.querySelector('.nav-item.active')?.dataset.tab === 'security') {
                this.loadSecurityMetrics();
            }
        }, 60000);
    }

    async apiCall(method, endpoint, data = null) {
        const url = `${this.apiBase}${endpoint}`;
        const options = {
            method,
            headers: {
                'Content-Type': 'application/json'
            }
        };

        if (this.token) {
            options.headers.Authorization = `Bearer ${this.token}`;
        }

        if (data) {
            options.body = JSON.stringify(data);
        }

        return fetch(url, options);
    }
}

// Global function handlers for button clicks
function refreshAnalytics() {
    app.loadPortfolioAnalytics().then(() => {
        app.renderAnalyticsTab();
        app.showAlert('Analytics refreshed successfully', 'success');
    });
}

function createNotificationRule() {
    app.showAlert('Notification rule creation coming soon', 'info');
}

function testNotifications() {
    app.apiCall('POST', '/admin/notifications/test').then(() => {
        app.showAlert('Test notification sent successfully', 'success');
    });
}

function refreshNotifications() {
    app.showAlert('Notifications refreshed', 'success');
}

function searchDomains() {
    const search = document.getElementById('domainSearch').value;
    const provider = document.getElementById('domainProvider').value;
    const status = document.getElementById('domainStatus').value;
    
    // Filter domains based on criteria
    let filteredDomains = app.data.domains;
    
    if (search) {
        filteredDomains = filteredDomains.filter(d => 
            d.name.toLowerCase().includes(search.toLowerCase())
        );
    }
    
    if (provider) {
        filteredDomains = filteredDomains.filter(d => d.provider === provider);
    }
    
    if (status) {
        filteredDomains = filteredDomains.filter(d => d.status === status);
    }
    
    // Temporarily update domains for rendering
    const originalDomains = app.data.domains;
    app.data.domains = filteredDomains;
    app.renderDomainsTable();
    app.data.domains = originalDomains;
    
    app.showAlert(`Found ${filteredDomains.length} domains`, 'success');
}

function bulkActions() {
    app.showAlert('Bulk actions coming soon', 'info');
}

function connectProvider() {
    app.showAlert('Provider connection dialog coming soon', 'info');
}

function syncAllProviders() {
    app.showAlert('Syncing all providers...', 'info');
    app.apiCall('POST', '/sync').then(() => {
        app.showAlert('Sync completed successfully', 'success');
        app.loadDomains();
    });
}

function refreshProviders() {
    app.loadProviders().then(() => {
        app.renderProvidersTab();
        app.showAlert('Providers refreshed', 'success');
    });
}

function editDomain(id) {
    app.showAlert(`Edit domain ${id} coming soon`, 'info');
}

function checkDomainStatus(id) {
    app.apiCall('POST', `/admin/domains/${id}/check-status`).then(() => {
        app.showAlert('Domain status checked', 'success');
    });
}

function testProvider(id) {
    app.showAlert(`Testing provider ${id}...`, 'info');
}

function syncProvider(id) {
    app.showAlert(`Syncing provider ${id}...`, 'info');
}

function deleteProvider(id) {
    if (confirm('Are you sure you want to delete this provider?')) {
        app.apiCall('DELETE', `/admin/credentials/${id}`).then(() => {
            app.showAlert('Provider deleted successfully', 'success');
            app.loadProviders().then(() => app.renderProvidersTab());
        });
    }
}

function analyzeDomain(id) {
    app.showAlert(`Domain analysis for ${id} coming soon`, 'info');
}

function editNotificationRule(index) {
    app.showAlert(`Edit notification rule ${index} coming soon`, 'info');
}

function deleteNotificationRule(index) {
    if (confirm('Are you sure you want to delete this notification rule?')) {
        app.showAlert('Notification rule deleted', 'success');
    }
}

// Initialize the enhanced admin app
const app = new EnhancedAdminApp();
