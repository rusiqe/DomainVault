// State management
const state = {
    isAuthenticated: false,
    authToken: null,
    currentSection: 'dashboard',
    dashboardData: null,
    analyticsData: null,
    securityData: null,
    notificationsData: null,
    domainsData: null,
    providersData: null
};

// Initialize the application
document.addEventListener('DOMContentLoaded', () => {
    initializeApp();
});

function initializeApp() {
    // Check if user is authenticated
    const token = localStorage.getItem('authToken');
    if (token) {
        state.authToken = token;
        state.isAuthenticated = true;
        hideLoginOverlay();
        loadDashboard();
    } else {
        showLoginOverlay();
    }

    // Set up event listeners
    setupEventListeners();
}

function setupEventListeners() {
    // Login form submission
    const loginForm = document.getElementById('loginForm');
    if (loginForm) {
        loginForm.addEventListener('submit', handleLogin);
    }

    // Menu item clicks
    document.querySelectorAll('.menu-item').forEach(item => {
        item.addEventListener('click', (e) => {
            e.preventDefault();
            const section = item.getAttribute('data-section');
            if (section) {
                switchSection(section);
            }
        });
    });

    // Logout functionality
    const logoutBtn = document.querySelector('.logout-btn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', logout);
    }
}

// Authentication functions
async function handleLogin(e) {
    e.preventDefault();
    
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    try {
        const response = await fetch('/api/v1/auth/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ username, password })
        });

        if (response.ok) {
            const data = await response.json();
            state.authToken = data.token;
            state.isAuthenticated = true;
            localStorage.setItem('authToken', data.token);
            
            hideLoginOverlay();
            loadDashboard();
        } else {
            showError('Invalid credentials. Please try again.');
        }
    } catch (error) {
        showError('Login failed. Please check your connection.');
        console.error('Login error:', error);
    }
}

function logout() {
    state.isAuthenticated = false;
    state.authToken = null;
    localStorage.removeItem('authToken');
    showLoginOverlay();
}

function showLoginOverlay() {
    document.getElementById('loginOverlay').style.display = 'flex';
}

function hideLoginOverlay() {
    document.getElementById('loginOverlay').style.display = 'none';
}

function showError(message) {
    alert(message);
}

// Navigation functions
function switchSection(sectionName) {
    // Update active menu item
    document.querySelectorAll('.menu-item').forEach(item => {
        item.classList.remove('active');
    });
    document.querySelector(`[data-section="${sectionName}"]`).classList.add('active');

    // Hide all sections
    document.querySelectorAll('.content-section').forEach(section => {
        section.classList.remove('active');
    });

    // Show selected section
    document.getElementById(sectionName).classList.add('active');

    // Update state
    state.currentSection = sectionName;

    // Load section data
    loadSectionData(sectionName);
}

async function loadSectionData(section) {
    switch (section) {
        case 'dashboard':
            await loadDashboard();
            break;
        case 'analytics':
            await loadAnalytics();
            break;
        case 'domains':
            await loadDomains();
            break;
        case 'providers':
            await loadProviders();
            break;
        case 'security':
            await loadSecurity();
            break;
        case 'notifications':
            await loadNotifications();
            break;
        case 'audit':
            await loadAuditLog();
            break;
        case 'dns':
            await loadDNSManagement();
            break;
        default:
            console.log(`Section ${section} not implemented yet`);
    }
}

// API helper function
async function apiCall(endpoint, options = {}) {
    const defaultOptions = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${state.authToken}`
        }
    };

    const finalOptions = { ...defaultOptions, ...options };
    
    try {
        const response = await fetch(endpoint, finalOptions);
        
        if (response.status === 401) {
            logout();
            return null;
        }
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        return await response.json();
    } catch (error) {
        console.error('API call failed:', error);
        return null;
    }
}

// Dashboard functions
async function loadDashboard() {
    try {
        // Load portfolio analytics
        const portfolioData = await apiCall('/api/v1/admin/analytics/portfolio');
        if (portfolioData) {
            updateDashboardMetrics(portfolioData);
        }

        // Load recent domains
        const domainsData = await apiCall('/api/v1/domains');
        if (domainsData) {
            updateRecentDomainsTable(domainsData.domains || []);
        }
    } catch (error) {
        console.error('Error loading dashboard:', error);
    }
}

function updateDashboardMetrics(data) {
    const overview = data.overview || {};
    const financial = data.financial_metrics || {};

    // Update metric cards
    document.getElementById('totalDomains').textContent = overview.total_domains || '-';
    document.getElementById('activeDomains').textContent = overview.active_domains || '-';
    document.getElementById('expiringDomains').textContent = overview.domains_expiring_30 || '-';
    document.getElementById('totalValue').textContent = financial.total_estimated_value ? 
        `$${financial.total_estimated_value.toLocaleString()}` : '$-';
}

function updateRecentDomainsTable(domains) {
    const tbody = document.querySelector('#recentDomainsTable tbody');
    if (!tbody) return;

    if (domains.length === 0) {
        tbody.innerHTML = '<tr><td colspan="5" class="loading">No domains found</td></tr>';
        return;
    }

    tbody.innerHTML = domains.slice(0, 10).map(domain => `
        <tr>
            <td>${domain.name}</td>
            <td><span class="status-badge ${getStatusClass(domain.status)}">${domain.status}</span></td>
            <td>${formatDate(domain.expires_at)}</td>
            <td>${domain.provider || '-'}</td>
            <td>
                <button class="btn btn-secondary" onclick="viewDomain('${domain.id}')">View</button>
                <button class="btn btn-primary" onclick="editDomain('${domain.id}')">Edit</button>
            </td>
        </tr>
    `).join('');
}

// Analytics functions
async function loadAnalytics() {
    try {
        const portfolioData = await apiCall('/api/v1/admin/analytics/portfolio');
        const financialData = await apiCall('/api/v1/admin/analytics/financial');
        
        if (portfolioData) {
            // Create charts here
            createPortfolioChart(portfolioData);
        }
        
        if (financialData) {
            createFinancialChart(financialData);
        }
    } catch (error) {
        console.error('Error loading analytics:', error);
    }
}

function createPortfolioChart(data) {
    const ctx = document.getElementById('portfolioChart');
    if (!ctx) return;

    // Basic chart implementation - you can enhance this with Chart.js
    const chartContainer = ctx.getContext('2d');
    chartContainer.fillStyle = '#1877f2';
    chartContainer.fillRect(0, 0, ctx.width, ctx.height);
    chartContainer.fillStyle = 'white';
    chartContainer.font = '16px sans-serif';
    chartContainer.textAlign = 'center';
    chartContainer.fillText('Portfolio Growth Chart', ctx.width/2, ctx.height/2);
}

function createFinancialChart(data) {
    const ctx = document.getElementById('financialChart');
    if (!ctx) return;

    const chartContainer = ctx.getContext('2d');
    chartContainer.fillStyle = '#42b883';
    chartContainer.fillRect(0, 0, ctx.width, ctx.height);
    chartContainer.fillStyle = 'white';
    chartContainer.font = '16px sans-serif';
    chartContainer.textAlign = 'center';
    chartContainer.fillText('Financial Overview Chart', ctx.width/2, ctx.height/2);
}

// Domains functions
async function loadDomains() {
    try {
        const data = await apiCall('/api/v1/domains');
        if (data && data.domains) {
            updateAllDomainsTable(data.domains);
        }
    } catch (error) {
        console.error('Error loading domains:', error);
    }
}

function updateAllDomainsTable(domains) {
    const tbody = document.querySelector('#allDomainsTable tbody');
    if (!tbody) return;

    if (domains.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" class="loading">No domains found</td></tr>';
        return;
    }

    tbody.innerHTML = domains.map(domain => `
        <tr>
            <td>${domain.name}</td>
            <td><span class="status-badge ${getStatusClass(domain.status)}">${domain.status}</span></td>
            <td>${formatDate(domain.expires_at)}</td>
            <td>${domain.provider || '-'}</td>
            <td>${domain.category_name || '-'}</td>
            <td>
                <button class="btn btn-secondary" onclick="viewDomain('${domain.id}')">View</button>
                <button class="btn btn-primary" onclick="editDomain('${domain.id}')">Edit</button>
                <button class="btn btn-danger" onclick="deleteDomain('${domain.id}')">Delete</button>
            </td>
        </tr>
    `).join('');
}

// Providers functions
async function loadProviders() {
    try {
        // Load supported providers
        const supportedData = await apiCall('/api/v1/admin/providers/supported');
        if (supportedData) {
            updateSupportedProvidersList(supportedData.providers || []);
            populateProviderSelect(supportedData.providers || []);
        }

        // Load connected providers
        const connectedData = await apiCall('/api/v1/admin/providers/connected');
        if (connectedData) {
            updateConnectedProvidersList(connectedData.providers || []);
        }

        // Load auto-sync status
        const autoSyncData = await apiCall('/api/v1/admin/providers/auto-sync/status');
        if (autoSyncData) {
            updateAutoSyncStatus(autoSyncData);
        }
    } catch (error) {
        console.error('Error loading providers:', error);
        showNotification('Failed to load provider data', 'error');
    }
}

function updateSupportedProvidersList(providers) {
    const container = document.getElementById('supportedProvidersContainer');
    if (!container) return;

    if (providers.length === 0) {
        container.innerHTML = '<div class="loading">No supported providers found</div>';
        return;
    }

    container.innerHTML = providers.map(provider => `
        <div class="supported-provider">
            <div class="supported-provider-info">
                <div class="supported-provider-logo">
                    <i class="${getProviderIcon(provider.name)}"></i>
                </div>
                <div>
                    <h3>${provider.display_name || provider.name}</h3>
                    <p>${provider.description || 'Domain provider'}</p>
                    <small>Required credentials: ${provider.required_credentials ? provider.required_credentials.join(', ') : 'API Key'}</small>
                </div>
            </div>
            <button class="btn btn-primary" onclick="showAddProviderModal('${provider.name}')">
                <i class="fas fa-plus"></i>
                Connect
            </button>
        </div>
    `).join('');
}

function updateConnectedProvidersList(providers) {
    const container = document.getElementById('connectedProvidersContainer');
    if (!container) return;

    if (providers.length === 0) {
        container.innerHTML = '<div class="loading">No connected providers found</div>';
        return;
    }

    container.innerHTML = providers.map(provider => `
        <div class="provider-card">
            <div class="provider-header">
                <div class="provider-info">
                    <div class="provider-icon">
                        <i class="${getProviderIcon(provider.provider)}"></i>
                    </div>
                    <div class="provider-meta">
                        <h3>${provider.name || provider.provider}</h3>
                        <p>${provider.account_name}</p>
                    </div>
                </div>
            </div>
            <div class="provider-status">
                <div class="status-indicator ${getConnectionStatus(provider.connection_status)}">
                    <i class="fas fa-circle"></i>
                    ${provider.connection_status || 'Unknown'}
                </div>
            </div>
            <div class="provider-stats">
                <div class="stat-item">
                    <div class="stat-value">${provider.domain_count || '0'}</div>
                    <div class="stat-label">Domains</div>
                </div>
                <div class="stat-item">
                    <div class="stat-value">${provider.auto_sync ? 'On' : 'Off'}</div>
                    <div class="stat-label">Auto-Sync</div>
                </div>
            </div>
            <div class="provider-actions">
                <button class="btn btn-secondary" onclick="showProviderDetails('${provider.id}')">
                    <i class="fas fa-info-circle"></i>
                    Details
                </button>
                <button class="btn btn-success" onclick="syncProviderById('${provider.id}')">
                    <i class="fas fa-sync"></i>
                    Sync
                </button>
            </div>
            ${provider.last_sync ? `<div style="margin-top: 12px; font-size: 12px; color: #65676b;">Last sync: ${formatDate(provider.last_sync)}</div>` : ''}
        </div>
    `).join('');
}

function updateAutoSyncStatus(status) {
    const statusElement = document.getElementById('autoSyncStatus');
    const toggleButton = document.getElementById('autoSyncToggle');
    
    if (!statusElement || !toggleButton) return;

    const isRunning = status.running || false;
    const activeProviders = status.active_providers || 0;
    
    statusElement.innerHTML = `
        <div class="status-indicator ${isRunning ? 'online' : 'offline'}">
            <i class="fas fa-circle"></i>
            Auto-Sync ${isRunning ? 'Active' : 'Disabled'}
        </div>
        <div class="sync-details">
            <p>${isRunning ? `Monitoring ${activeProviders} provider(s) for automatic synchronization.` : 'Enable auto-sync to automatically synchronize domains from all connected providers based on their configured intervals.'}</p>
        </div>
    `;
    
    toggleButton.innerHTML = `
        <i class="fas fa-${isRunning ? 'stop' : 'play'}"></i>
        ${isRunning ? 'Stop Auto-Sync' : 'Start Auto-Sync'}
    `;
    toggleButton.className = `btn ${isRunning ? 'btn-danger' : 'btn-success'}`;
}

function populateProviderSelect(providers) {
    const select = document.getElementById('providerSelect');
    if (!select) return;

    select.innerHTML = '<option value="">Select a provider...</option>';
    providers.forEach(provider => {
        const option = document.createElement('option');
        option.value = provider.name;
        option.textContent = provider.display_name || provider.name;
        option.dataset.credentials = JSON.stringify(provider.required_credentials || ['api_key']);
        select.appendChild(option);
    });
}

function getProviderIcon(provider) {
    const icons = {
        'namecheap': 'fas fa-shopping-cart',
        'godaddy': 'fas fa-globe',
        'cloudflare': 'fas fa-cloud',
        'route53': 'fab fa-aws',
        'digitalocean': 'fab fa-digital-ocean',
        'default': 'fas fa-server'
    };
    return icons[provider?.toLowerCase()] || icons.default;
}

function getConnectionStatus(status) {
    const statusMap = {
        'connected': 'online',
        'disconnected': 'offline',
        'error': 'offline',
        'syncing': 'syncing'
    };
    return statusMap[status?.toLowerCase()] || 'offline';
}

// Security functions
async function loadSecurity() {
    try {
        const data = await apiCall('/api/v1/admin/security/metrics');
        // Load security metrics and update the interface
        console.log('Security data:', data);
    } catch (error) {
        console.error('Error loading security:', error);
    }
}

// Notifications functions
async function loadNotifications() {
    try {
        const data = await apiCall('/api/v1/admin/notifications/rules');
        // Load notifications and update the interface
        console.log('Notifications data:', data);
    } catch (error) {
        console.error('Error loading notifications:', error);
    }
}

// Audit Log functions
async function loadAuditLog() {
    try {
        const data = await apiCall('/api/v1/admin/security/audit');
        // Load audit log and update the interface
        console.log('Audit log data:', data);
    } catch (error) {
        console.error('Error loading audit log:', error);
    }
}

// Utility functions
function getStatusClass(status) {
    switch (status?.toLowerCase()) {
        case 'active':
            return 'status-active';
        case 'expired':
            return 'status-expired';
        case 'expiring':
            return 'status-expiring';
        default:
            return 'status-active';
    }
}

function formatDate(dateString) {
    if (!dateString) return '-';
    return new Date(dateString).toLocaleDateString();
}

// Action functions (placeholders)
function viewDomain(id) {
    console.log('View domain:', id);
}

function editDomain(id) {
    console.log('Edit domain:', id);
}

function deleteDomain(id) {
    if (confirm('Are you sure you want to delete this domain?')) {
        console.log('Delete domain:', id);
    }
}

// Provider Modal Functions
function showAddProviderModal(providerName = '') {
    const modal = document.getElementById('addProviderModal');
    if (!modal) return;
    
    // Reset form
    document.getElementById('addProviderForm').reset();
    document.getElementById('credentialsContainer').innerHTML = '';
    document.getElementById('autoSyncSettings').style.display = 'none';
    
    // Pre-select provider if specified
    if (providerName) {
        document.getElementById('providerSelect').value = providerName;
        onProviderSelectChange();
    }
    
    modal.style.display = 'flex';
}

function hideAddProviderModal() {
    const modal = document.getElementById('addProviderModal');
    if (modal) {
        modal.style.display = 'none';
    }
}

function showProviderDetails(providerId) {
    const modal = document.getElementById('providerDetailsModal');
    if (!modal) return;
    
    // Store current provider ID
    window.currentProviderId = providerId;
    
    // Load provider details
    loadProviderDetails(providerId);
    
    modal.style.display = 'flex';
}

function hideProviderDetailsModal() {
    const modal = document.getElementById('providerDetailsModal');
    if (modal) {
        modal.style.display = 'none';
    }
}

async function loadProviderDetails(providerId) {
    try {
        const provider = await apiCall(`/api/v1/admin/providers/connected/${providerId}`);
        if (provider) {
            updateProviderDetailsModal(provider);
        }
    } catch (error) {
        console.error('Error loading provider details:', error);
        showNotification('Failed to load provider details', 'error');
    }
}

function updateProviderDetailsModal(provider) {
    const content = document.getElementById('providerDetailsContent');
    if (!content) return;
    
    content.innerHTML = `
        <div class="provider-details">
            <div class="detail-section">
                <h3>Basic Information</h3>
                <div class="detail-row">
                    <strong>Provider:</strong> ${provider.provider}
                </div>
                <div class="detail-row">
                    <strong>Display Name:</strong> ${provider.name}
                </div>
                <div class="detail-row">
                    <strong>Account:</strong> ${provider.account_name}
                </div>
                <div class="detail-row">
                    <strong>Status:</strong> 
                    <span class="status-badge ${getConnectionStatus(provider.connection_status)}">
                        ${provider.connection_status}
                    </span>
                </div>
            </div>
            
            <div class="detail-section">
                <h3>Sync Configuration</h3>
                <div class="detail-row">
                    <strong>Auto-Sync:</strong> ${provider.auto_sync ? 'Enabled' : 'Disabled'}
                </div>
                ${provider.auto_sync ? `
                    <div class="detail-row">
                        <strong>Sync Interval:</strong> ${provider.sync_interval_hours} hours
                    </div>
                ` : ''}
                <div class="detail-row">
                    <strong>Last Sync:</strong> ${provider.last_sync ? formatDate(provider.last_sync) : 'Never'}
                </div>
                <div class="detail-row">
                    <strong>Domain Count:</strong> ${provider.domain_count || 0}
                </div>
            </div>
            
            <div class="detail-section">
                <h3>System Information</h3>
                <div class="detail-row">
                    <strong>Connected:</strong> ${formatDate(provider.created_at)}
                </div>
                <div class="detail-row">
                    <strong>Last Updated:</strong> ${formatDate(provider.updated_at)}
                </div>
                ${provider.sync_status ? `
                    <div class="detail-row">
                        <strong>Sync Status:</strong> ${provider.sync_status}
                    </div>
                ` : ''}
            </div>
        </div>
    `;
}

// Provider Operations
function onProviderSelectChange() {
    const select = document.getElementById('providerSelect');
    const container = document.getElementById('credentialsContainer');
    
    if (!select || !container) return;
    
    const selectedOption = select.options[select.selectedIndex];
    if (!selectedOption || !selectedOption.value) {
        container.innerHTML = '';
        return;
    }
    
    // Get required credentials for the selected provider
    const requiredCredentials = JSON.parse(selectedOption.dataset.credentials || '["api_key"]');
    
    // Generate credential input fields
    container.innerHTML = requiredCredentials.map(credential => {
        const fieldName = credential.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
        const inputType = credential.toLowerCase().includes('secret') || credential.toLowerCase().includes('key') ? 'password' : 'text';
        
        return `
            <div class="credentials-field">
                <label class="form-label">${fieldName}</label>
                <input type="${inputType}" id="credential_${credential}" class="form-control" 
                       placeholder="Enter ${fieldName.toLowerCase()}" required>
                <div class="credentials-help">
                    ${getCredentialHelp(credential)}
                </div>
            </div>
        `;
    }).join('');
}

function getCredentialHelp(credential) {
    const helpTexts = {
        'api_key': 'Your API key from the provider\'s developer console',
        'api_secret': 'Your API secret key (keep this secure)',
        'username': 'Your account username',
        'token': 'Access token from your provider account',
        'access_key': 'Access key for API authentication',
        'secret_key': 'Secret key for API authentication'
    };
    return helpTexts[credential] || 'Required credential for authentication';
}

async function testProviderConnection() {
    const formData = getProviderFormData();
    if (!formData) return;
    
    const testBtn = document.getElementById('testBtn');
    const originalText = testBtn.innerHTML;
    
    try {
        testBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Testing...';
        testBtn.disabled = true;
        
        const response = await apiCall('/api/v1/admin/providers/test', {
            method: 'POST',
            body: JSON.stringify({
                provider: formData.provider,
                credentials: formData.credentials
            })
        });
        
        if (response && response.success) {
            showNotification('Connection test successful!', 'success');
            testBtn.innerHTML = '<i class="fas fa-check"></i> Test Passed';
            testBtn.className = 'btn btn-success';
        } else {
            showNotification(response?.message || 'Connection test failed', 'error');
            testBtn.innerHTML = '<i class="fas fa-times"></i> Test Failed';
            testBtn.className = 'btn btn-danger';
        }
    } catch (error) {
        console.error('Connection test error:', error);
        showNotification('Connection test failed', 'error');
        testBtn.innerHTML = '<i class="fas fa-times"></i> Test Failed';
        testBtn.className = 'btn btn-danger';
    } finally {
        testBtn.disabled = false;
        // Reset button after 3 seconds
        setTimeout(() => {
            testBtn.innerHTML = originalText;
            testBtn.className = 'btn btn-primary';
        }, 3000);
    }
}

async function connectProvider() {
    const formData = getProviderFormData();
    if (!formData) return;
    
    const connectBtn = document.getElementById('connectBtn');
    const originalText = connectBtn.innerHTML;
    
    try {
        connectBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Connecting...';
        connectBtn.disabled = true;
        
        const response = await apiCall('/api/v1/admin/providers/connect', {
            method: 'POST',
            body: JSON.stringify(formData)
        });
        
        if (response && response.success) {
            showNotification('Provider connected successfully!', 'success');
            hideAddProviderModal();
            refreshProviders();
        } else {
            showNotification(response?.error || 'Failed to connect provider', 'error');
        }
    } catch (error) {
        console.error('Connect provider error:', error);
        showNotification('Failed to connect provider', 'error');
    } finally {
        connectBtn.innerHTML = originalText;
        connectBtn.disabled = false;
    }
}

function getProviderFormData() {
    const provider = document.getElementById('providerSelect').value;
    const name = document.getElementById('providerName').value;
    const accountName = document.getElementById('accountName').value;
    const testConnection = document.getElementById('testConnection').checked;
    const autoSync = document.getElementById('enableAutoSync').checked;
    const syncIntervalHours = parseInt(document.getElementById('syncIntervalHours').value);
    const initialSync = document.getElementById('initialSync').checked;
    
    if (!provider || !name || !accountName) {
        showNotification('Please fill in all required fields', 'error');
        return null;
    }
    
    // Collect credentials
    const credentials = {};
    const credentialInputs = document.querySelectorAll('#credentialsContainer input[id^="credential_"]');
    
    for (const input of credentialInputs) {
        const credentialName = input.id.replace('credential_', '');
        if (!input.value.trim()) {
            showNotification(`Please enter ${credentialName.replace(/_/g, ' ')}`, 'error');
            return null;
        }
        credentials[credentialName] = input.value.trim();
    }
    
    return {
        provider,
        name,
        account_name: accountName,
        credentials,
        test_connection: testConnection,
        auto_sync: autoSync,
        sync_interval_hours: autoSync ? syncIntervalHours : null,
        initial_sync: initialSync
    };
}

// Provider Operations
async function refreshProviders() {
    await loadProviders();
}

async function syncProviderById(providerId) {
    try {
        const response = await apiCall(`/api/v1/admin/providers/${providerId}/sync`, {
            method: 'POST'
        });
        
        if (response) {
            showNotification('Sync initiated successfully', 'success');
            setTimeout(refreshProviders, 2000); // Refresh after 2 seconds
        }
    } catch (error) {
        console.error('Sync provider error:', error);
        showNotification('Failed to sync provider', 'error');
    }
}

async function syncAllProviders() {
    try {
        const response = await apiCall('/api/v1/admin/providers/sync-all', {
            method: 'POST'
        });
        
        if (response) {
            showNotification('Sync all providers initiated', 'success');
            setTimeout(refreshProviders, 2000);
        }
    } catch (error) {
        console.error('Sync all providers error:', error);
        showNotification('Failed to sync providers', 'error');
    }
}

async function toggleAutoSync() {
    try {
        const statusData = await apiCall('/api/v1/admin/providers/auto-sync/status');
        const isRunning = statusData?.running || false;
        
        const endpoint = isRunning ? '/api/v1/admin/providers/auto-sync/stop' : '/api/v1/admin/providers/auto-sync/start';
        const response = await apiCall(endpoint, { method: 'POST' });
        
        if (response) {
            showNotification(response.message, 'success');
            setTimeout(() => loadProviders(), 1000);
        }
    } catch (error) {
        console.error('Toggle auto-sync error:', error);
        showNotification('Failed to toggle auto-sync', 'error');
    }
}

async function removeProvider() {
    if (!window.currentProviderId) return;
    
    if (!confirm('Are you sure you want to remove this provider? This action cannot be undone.')) {
        return;
    }
    
    try {
        const response = await apiCall(`/api/v1/admin/providers/connected/${window.currentProviderId}`, {
            method: 'DELETE'
        });
        
        if (response) {
            showNotification('Provider removed successfully', 'success');
            hideProviderDetailsModal();
            refreshProviders();
        }
    } catch (error) {
        console.error('Remove provider error:', error);
        showNotification('Failed to remove provider', 'error');
    }
}

// Event listeners for modal interactions
document.addEventListener('DOMContentLoaded', () => {
    // Auto-sync settings toggle
    const autoSyncCheckbox = document.getElementById('enableAutoSync');
    const autoSyncSettings = document.getElementById('autoSyncSettings');
    
    if (autoSyncCheckbox && autoSyncSettings) {
        autoSyncCheckbox.addEventListener('change', (e) => {
            autoSyncSettings.style.display = e.target.checked ? 'block' : 'none';
        });
    }
    
    // Provider select change
    const providerSelect = document.getElementById('providerSelect');
    if (providerSelect) {
        providerSelect.addEventListener('change', onProviderSelectChange);
    }
    
    // Modal close on outside click
    document.addEventListener('click', (e) => {
        if (e.target.classList.contains('modal-overlay')) {
            if (e.target.id === 'addProviderModal') {
                hideAddProviderModal();
            } else if (e.target.id === 'providerDetailsModal') {
                hideProviderDetailsModal();
            }
        }
    });
});

// Notification function
function showNotification(message, type = 'info') {
    // For now, use alert - you can replace this with a proper notification system
    alert(message);
}

// ============================================================================
// DNS MANAGEMENT FUNCTIONS
// ============================================================================

// State for DNS management
let currentDomainId = null;
let currentDNSRecords = [];
let currentEditingRecord = null;

// Load DNS management section
async function loadDNSManagement() {
    try {
        // Load domains for selection and bulk operations
        await loadDomainsForDNS();
        
        // Load DNS templates
        await loadDNSTemplates();
    } catch (error) {
        console.error('Error loading DNS management:', error);
        showNotification('Failed to load DNS management', 'error');
    }
}

// Load domains for DNS management
async function loadDomainsForDNS() {
    try {
        const data = await apiCall('/api/v1/domains');
        if (data && data.domains) {
            populateDomainsSelect(data.domains);
            // Store domains for bulk operations
            window.availableDomains = data.domains;
        }
    } catch (error) {
        console.error('Error loading domains for DNS:', error);
    }
}

function populateDomainsSelect(domains) {
    const select = document.getElementById('dnsDomainsSelect');
    if (!select) return;

    select.innerHTML = '<option value="">Choose a domain to manage DNS records...</option>';
    domains.forEach(domain => {
        const option = document.createElement('option');
        option.value = domain.id;
        option.textContent = domain.name;
        option.dataset.provider = domain.provider;
        option.dataset.status = domain.status;
        option.dataset.expires = domain.expires_at;
        select.appendChild(option);
    });
}

// Load DNS records for selected domain
async function loadDNSRecords() {
    const select = document.getElementById('dnsDomainsSelect');
    if (!select || !select.value) {
        hideDNSRecordsCard();
        return;
    }

    currentDomainId = select.value;
    const selectedOption = select.options[select.selectedIndex];
    
    // Show domain info
    showDomainInfo({
        provider: selectedOption.dataset.provider,
        status: selectedOption.dataset.status,
        expires: selectedOption.dataset.expires,
        nameServers: 'ns1.example.com, ns2.example.com' // This would come from API
    });

    try {
        const records = await apiCall(`/api/v1/admin/domains/${currentDomainId}/dns`);
        if (records) {
            currentDNSRecords = records.records || [];
            updateDNSRecordsTable(currentDNSRecords);
            updateDNSAnalytics(currentDNSRecords);
            showDNSRecordsCard();
        }
    } catch (error) {
        console.error('Error loading DNS records:', error);
        showNotification('Failed to load DNS records', 'error');
    }
}

function showDomainInfo(info) {
    document.getElementById('domainProvider').textContent = info.provider;
    document.getElementById('domainStatus').textContent = info.status;
    document.getElementById('domainStatus').className = `status-badge status-${info.status.toLowerCase()}`;
    document.getElementById('domainExpiry').textContent = formatDate(info.expires);
    document.getElementById('domainNameServers').textContent = info.nameServers;
    document.getElementById('selectedDomainInfo').style.display = 'block';
}

function showDNSRecordsCard() {
    document.getElementById('dnsRecordsCard').style.display = 'block';
    document.getElementById('dnsAnalyticsCard').style.display = 'block';
}

function hideDNSRecordsCard() {
    document.getElementById('dnsRecordsCard').style.display = 'none';
    document.getElementById('dnsAnalyticsCard').style.display = 'none';
    document.getElementById('selectedDomainInfo').style.display = 'none';
}

function updateDNSRecordsTable(records) {
    const tbody = document.querySelector('#dnsRecordsTable tbody');
    if (!tbody) return;

    if (records.length === 0) {
        tbody.innerHTML = '<tr><td colspan="7" class="loading">No DNS records found for this domain</td></tr>';
        return;
    }

    tbody.innerHTML = records.map(record => `
        <tr>
            <td><span class="record-type-badge record-type-${record.type}">${record.type}</span></td>
            <td><strong>${record.name}</strong></td>
            <td><span class="record-value" title="${record.value}">${record.value}</span></td>
            <td><span class="ttl-badge">${record.ttl}</span></td>
            <td>${record.priority ? `<span class="record-priority">${record.priority}</span>` : '-'}</td>
            <td>${formatDate(record.updated_at)}</td>
            <td>
                <div class="dns-record-actions">
                    <button class="btn btn-secondary" onclick="editDNSRecord('${record.id}')" title="Edit">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button class="btn btn-danger" onclick="deleteDNSRecord('${record.id}')" title="Delete">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </td>
        </tr>
    `).join('');
}

function updateDNSAnalytics(records) {
    // Calculate statistics
    const totalRecords = records.length;
    const recordTypes = [...new Set(records.map(r => r.type))].length;
    const avgTTL = records.length > 0 ? Math.round(records.reduce((sum, r) => sum + r.ttl, 0) / records.length / 60) : 0;
    const lastModified = records.length > 0 ? Math.max(...records.map(r => new Date(r.updated_at).getTime())) : null;

    // Update display
    document.getElementById('totalRecords').textContent = totalRecords;
    document.getElementById('recordTypes').textContent = recordTypes;
    document.getElementById('avgTTL').textContent = avgTTL;
    document.getElementById('lastModified').textContent = lastModified ? formatDate(new Date(lastModified)) : '-';

    // Update chart (basic implementation)
    updateRecordTypesChart(records);
}

function updateRecordTypesChart(records) {
    // This is a basic implementation - you could use Chart.js for better charts
    const canvas = document.getElementById('recordTypesChart');
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    
    // Count record types
    const typeCounts = {};
    records.forEach(record => {
        typeCounts[record.type] = (typeCounts[record.type] || 0) + 1;
    });

    // Draw simple bar chart
    const types = Object.keys(typeCounts);
    const maxCount = Math.max(...Object.values(typeCounts));
    const barWidth = canvas.width / types.length;
    
    ctx.fillStyle = '#1877f2';
    ctx.font = '12px sans-serif';
    
    types.forEach((type, index) => {
        const height = (typeCounts[type] / maxCount) * (canvas.height - 40);
        const x = index * barWidth;
        const y = canvas.height - height - 20;
        
        ctx.fillRect(x + 10, y, barWidth - 20, height);
        ctx.fillStyle = '#1c1e21';
        ctx.fillText(type, x + barWidth/2 - 10, canvas.height - 5);
        ctx.fillText(typeCounts[type], x + barWidth/2 - 5, y - 5);
        ctx.fillStyle = '#1877f2';
    });
}

// DNS Record Modal Functions
function showAddRecordModal() {
    if (!currentDomainId) {
        showNotification('Please select a domain first', 'error');
        return;
    }
    
    document.getElementById('addDNSRecordForm').reset();
    document.getElementById('priorityWeightSection').style.display = 'none';
    document.getElementById('recordPreview').style.display = 'none';
    document.getElementById('addDNSRecordModal').style.display = 'flex';
}

function hideAddRecordModal() {
    document.getElementById('addDNSRecordModal').style.display = 'none';
}

function updateRecordForm() {
    const type = document.getElementById('recordType').value;
    const prioritySection = document.getElementById('priorityWeightSection');
    const weightGroup = document.getElementById('weightGroup');
    const portGroup = document.getElementById('portGroup');
    const helpText = document.getElementById('recordValueHelp');
    
    // Hide all additional fields first
    prioritySection.style.display = 'none';
    weightGroup.style.display = 'none';
    portGroup.style.display = 'none';
    
    // Update help text and show relevant fields
    switch (type) {
        case 'A':
            helpText.textContent = 'Enter IPv4 address (e.g., 192.168.1.1)';
            break;
        case 'AAAA':
            helpText.textContent = 'Enter IPv6 address (e.g., 2001:db8::1)';
            break;
        case 'CNAME':
            helpText.textContent = 'Enter target domain (e.g., example.com)';
            break;
        case 'MX':
            helpText.textContent = 'Enter mail server hostname (e.g., mail.example.com)';
            prioritySection.style.display = 'block';
            break;
        case 'TXT':
            helpText.textContent = 'Enter text content (e.g., "v=spf1 include:_spf.example.com ~all")';
            break;
        case 'NS':
            helpText.textContent = 'Enter nameserver hostname (e.g., ns1.example.com)';
            break;
        case 'SRV':
            helpText.textContent = 'Enter target hostname (e.g., target.example.com)';
            prioritySection.style.display = 'block';
            weightGroup.style.display = 'block';
            portGroup.style.display = 'block';
            break;
        default:
            helpText.textContent = 'Enter the record value';
    }
}

function previewDNSRecord() {
    const formData = getDNSRecordFormData();
    if (!formData) return;
    
    const preview = document.getElementById('recordPreview');
    const content = document.getElementById('recordPreviewContent');
    
    let previewText = `${formData.name} ${formData.ttl} IN ${formData.type}`;
    
    if (formData.type === 'MX') {
        previewText += ` ${formData.priority} ${formData.value}`;
    } else if (formData.type === 'SRV') {
        previewText += ` ${formData.priority} ${formData.weight} ${formData.port} ${formData.value}`;
    } else {
        previewText += ` ${formData.value}`;
    }
    
    content.textContent = previewText;
    preview.style.display = 'block';
}

async function saveDNSRecord() {
    const formData = getDNSRecordFormData();
    if (!formData) return;
    
    const saveBtn = document.querySelector('#addDNSRecordModal .btn-success');
    const originalText = saveBtn.innerHTML;
    
    try {
        saveBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Saving...';
        saveBtn.disabled = true;
        
        const response = await apiCall(`/api/v1/admin/domains/${currentDomainId}/dns`, {
            method: 'POST',
            body: JSON.stringify(formData)
        });
        
        if (response) {
            showNotification('DNS record created successfully', 'success');
            hideAddRecordModal();
            await loadDNSRecords(); // Refresh the records
        }
    } catch (error) {
        console.error('Save DNS record error:', error);
        showNotification('Failed to save DNS record', 'error');
    } finally {
        saveBtn.innerHTML = originalText;
        saveBtn.disabled = false;
    }
}

function getDNSRecordFormData() {
    const type = document.getElementById('recordType').value;
    const name = document.getElementById('recordName').value;
    const value = document.getElementById('recordValue').value;
    const ttlSelect = document.getElementById('recordTTL').value;
    const customTTL = document.getElementById('customTTL').value;
    
    if (!type || !name || !value) {
        showNotification('Please fill in all required fields', 'error');
        return null;
    }
    
    const ttl = ttlSelect === 'custom' ? parseInt(customTTL) : parseInt(ttlSelect);
    
    const record = {
        domain_id: currentDomainId,
        type,
        name,
        value,
        ttl
    };
    
    // Add priority, weight, port for specific record types
    if (type === 'MX' || type === 'SRV') {
        const priority = document.getElementById('recordPriority').value;
        if (!priority) {
            showNotification('Priority is required for ' + type + ' records', 'error');
            return null;
        }
        record.priority = parseInt(priority);
    }
    
    if (type === 'SRV') {
        const weight = document.getElementById('recordWeight').value;
        const port = document.getElementById('recordPort').value;
        if (!weight || !port) {
            showNotification('Weight and Port are required for SRV records', 'error');
            return null;
        }
        record.weight = parseInt(weight);
        record.port = parseInt(port);
    }
    
    return record;
}

// DNS Templates
async function loadDNSTemplates() {
    try {
        const templates = await apiCall('/api/v1/admin/dns/templates');
        if (templates) {
            populateDNSTemplates(templates);
        }
    } catch (error) {
        console.error('Error loading DNS templates:', error);
    }
}

function populateDNSTemplates(templates) {
    const grid = document.getElementById('dnsTemplatesGrid');
    if (!grid) return;
    
    const templateDescriptions = {
        'basic_website': 'Basic website setup with A and CNAME records',
        'email_hosting': 'Email hosting configuration with MX and SPF records',
        'cdn_setup': 'CDN configuration with multiple CNAME records'
    };
    
    grid.innerHTML = Object.entries(templates).map(([name, records]) => `
        <div class="template-card" onclick="applyDNSTemplate('${name}')">
            <div class="template-title">${name.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}</div>
            <div class="template-description">${templateDescriptions[name] || 'DNS template'}</div>
            <div class="template-records">${records.length} records</div>
        </div>
    `).join('');
}

function showDNSTemplatesModal() {
    document.getElementById('dnsTemplatesModal').style.display = 'flex';
}

function hideDNSTemplatesModal() {
    document.getElementById('dnsTemplatesModal').style.display = 'none';
}

async function applyDNSTemplate(templateName) {
    if (!currentDomainId) {
        showNotification('Please select a domain first', 'error');
        return;
    }
    
    if (!confirm(`Apply ${templateName} template? This will add multiple DNS records.`)) {
        return;
    }
    
    try {
        const response = await apiCall(`/api/v1/admin/domains/${currentDomainId}/dns/template`, {
            method: 'POST',
            body: JSON.stringify({ template: templateName })
        });
        
        if (response) {
            showNotification('DNS template applied successfully', 'success');
            hideDNSTemplatesModal();
            await loadDNSRecords();
        }
    } catch (error) {
        console.error('Apply template error:', error);
        showNotification('Failed to apply DNS template', 'error');
    }
}

// DNS Record Management
async function editDNSRecord(recordId) {
    const record = currentDNSRecords.find(r => r.id === recordId);
    if (!record) return;
    
    currentEditingRecord = record;
    
    // Populate edit form
    document.getElementById('editRecordType').value = record.type;
    document.getElementById('editRecordName').value = record.name;
    document.getElementById('editRecordValue').value = record.value;
    document.getElementById('editRecordTTL').value = record.ttl;
    
    if (record.priority) {
        document.getElementById('editRecordPriority').value = record.priority;
    }
    if (record.weight) {
        document.getElementById('editRecordWeight').value = record.weight;
    }
    if (record.port) {
        document.getElementById('editRecordPort').value = record.port;
    }
    
    updateEditRecordForm();
    document.getElementById('editDNSRecordModal').style.display = 'flex';
}

function hideEditRecordModal() {
    document.getElementById('editDNSRecordModal').style.display = 'none';
    currentEditingRecord = null;
}

function updateEditRecordForm() {
    const type = document.getElementById('editRecordType').value;
    const prioritySection = document.getElementById('editPriorityWeightSection');
    const weightGroup = document.getElementById('editWeightGroup');
    const portGroup = document.getElementById('editPortGroup');
    
    prioritySection.style.display = 'none';
    weightGroup.style.display = 'none';
    portGroup.style.display = 'none';
    
    if (type === 'MX' || type === 'SRV') {
        prioritySection.style.display = 'block';
    }
    if (type === 'SRV') {
        weightGroup.style.display = 'block';
        portGroup.style.display = 'block';
    }
}

async function updateDNSRecord() {
    if (!currentEditingRecord) return;
    
    const formData = {
        type: document.getElementById('editRecordType').value,
        name: document.getElementById('editRecordName').value,
        value: document.getElementById('editRecordValue').value,
        ttl: parseInt(document.getElementById('editRecordTTL').value)
    };
    
    if (formData.type === 'MX' || formData.type === 'SRV') {
        formData.priority = parseInt(document.getElementById('editRecordPriority').value);
    }
    if (formData.type === 'SRV') {
        formData.weight = parseInt(document.getElementById('editRecordWeight').value);
        formData.port = parseInt(document.getElementById('editRecordPort').value);
    }
    
    try {
        const response = await apiCall(`/api/v1/admin/dns/${currentEditingRecord.id}`, {
            method: 'PUT',
            body: JSON.stringify(formData)
        });
        
        if (response) {
            showNotification('DNS record updated successfully', 'success');
            hideEditRecordModal();
            await loadDNSRecords();
        }
    } catch (error) {
        console.error('Update DNS record error:', error);
        showNotification('Failed to update DNS record', 'error');
    }
}

async function deleteDNSRecord(recordId) {
    if (!confirm('Are you sure you want to delete this DNS record?')) {
        return;
    }
    
    try {
        const response = await apiCall(`/api/v1/admin/dns/${recordId}`, {
            method: 'DELETE'
        });
        
        if (response) {
            showNotification('DNS record deleted successfully', 'success');
            await loadDNSRecords();
        }
    } catch (error) {
        console.error('Delete DNS record error:', error);
        showNotification('Failed to delete DNS record', 'error');
    }
}

// DNS Filtering
function filterDNSRecords() {
    const typeFilter = document.getElementById('recordTypeFilter').value;
    const nameFilter = document.getElementById('nameFilter').value.toLowerCase();
    const ttlFilter = document.getElementById('ttlFilter').value;
    
    let filteredRecords = currentDNSRecords;
    
    if (typeFilter) {
        filteredRecords = filteredRecords.filter(record => record.type === typeFilter);
    }
    
    if (nameFilter) {
        filteredRecords = filteredRecords.filter(record => 
            record.name.toLowerCase().includes(nameFilter) || 
            record.value.toLowerCase().includes(nameFilter)
        );
    }
    
    if (ttlFilter) {
        filteredRecords = filteredRecords.filter(record => record.ttl == ttlFilter);
    }
    
    updateDNSRecordsTable(filteredRecords);
}

// DNS Import/Export
function showImportDNSModal() {
    if (!currentDomainId) {
        showNotification('Please select a domain first', 'error');
        return;
    }
    document.getElementById('importDNSModal').style.display = 'flex';
}

function hideImportDNSModal() {
    document.getElementById('importDNSModal').style.display = 'none';
}

async function importDNSRecords() {
    const format = document.getElementById('importFormat').value;
    const data = document.getElementById('importData').value;
    const replace = document.getElementById('replaceRecords').checked;
    
    if (!data.trim()) {
        showNotification('Please enter DNS records data', 'error');
        return;
    }
    
    try {
        const response = await apiCall(`/api/v1/admin/domains/${currentDomainId}/dns/import`, {
            method: 'POST',
            body: JSON.stringify({
                format,
                data,
                replace
            })
        });
        
        if (response) {
            showNotification('DNS records imported successfully', 'success');
            hideImportDNSModal();
            await loadDNSRecords();
        }
    } catch (error) {
        console.error('Import DNS records error:', error);
        showNotification('Failed to import DNS records', 'error');
    }
}

async function exportDNSRecords() {
    if (!currentDomainId) {
        showNotification('Please select a domain first', 'error');
        return;
    }
    
    try {
        const response = await apiCall(`/api/v1/admin/domains/${currentDomainId}/dns/export?format=bind`);
        if (response && response.data) {
            // Create and download file
            const blob = new Blob([response.data], { type: 'text/plain' });
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `${document.getElementById('dnsDomainsSelect').selectedOptions[0].text}.zone`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);
            
            showNotification('DNS records exported successfully', 'success');
        }
    } catch (error) {
        console.error('Export DNS records error:', error);
        showNotification('Failed to export DNS records', 'error');
    }
}

function refreshDNSRecords() {
    if (currentDomainId) {
        loadDNSRecords();
    }
}

// Event listeners for DNS management
document.addEventListener('DOMContentLoaded', () => {
    // TTL selector change
    const ttlSelect = document.getElementById('recordTTL');
    const customTTLGroup = document.getElementById('customTTLGroup');
    
    if (ttlSelect && customTTLGroup) {
        ttlSelect.addEventListener('change', (e) => {
            customTTLGroup.style.display = e.target.value === 'custom' ? 'block' : 'none';
        });
    }
});

// Global functions for button clicks
window.viewDomain = viewDomain;
window.editDomain = editDomain;
window.deleteDomain = deleteDomain;
window.logout = logout;
window.showAddProviderModal = showAddProviderModal;
window.hideAddProviderModal = hideAddProviderModal;
window.showProviderDetails = showProviderDetails;
window.hideProviderDetailsModal = hideProviderDetailsModal;
window.testProviderConnection = testProviderConnection;
window.connectProvider = connectProvider;
window.refreshProviders = refreshProviders;
window.syncProviderById = syncProviderById;
window.syncAllProviders = syncAllProviders;
window.toggleAutoSync = toggleAutoSync;
window.removeProvider = removeProvider;
window.editProvider = () => console.log('Edit provider - to be implemented');
window.syncProvider = () => syncProviderById(window.currentProviderId);

// DNS Management global functions
window.loadDNSRecords = loadDNSRecords;
window.showAddRecordModal = showAddRecordModal;
window.hideAddRecordModal = hideAddRecordModal;
window.updateRecordForm = updateRecordForm;
window.previewDNSRecord = previewDNSRecord;
window.saveDNSRecord = saveDNSRecord;
window.showDNSTemplatesModal = showDNSTemplatesModal;
window.hideDNSTemplatesModal = hideDNSTemplatesModal;
window.applyDNSTemplate = applyDNSTemplate;
window.editDNSRecord = editDNSRecord;
window.hideEditRecordModal = hideEditRecordModal;
window.updateEditRecordForm = updateEditRecordForm;
window.updateDNSRecord = updateDNSRecord;
window.deleteDNSRecord = deleteDNSRecord;
window.filterDNSRecords = filterDNSRecords;
window.showImportDNSModal = showImportDNSModal;
window.hideImportDNSModal = hideImportDNSModal;
window.importDNSRecords = importDNSRecords;
window.exportDNSRecords = exportDNSRecords;
window.refreshDNSRecords = refreshDNSRecords;

// Bulk DNS Management Functions
let bulkOperationType = null;
let bulkOperationData = null;
let csvData = null;

// Tab switching for bulk DNS management
function switchBulkTab(tabName) {
    // Update tab buttons
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');
    
    // Update tab content
    document.querySelectorAll('.bulk-tab-content').forEach(content => {
        content.classList.remove('active');
    });
    document.getElementById(tabName).classList.add('active');
}

// Load domains for bulk operations
async function loadDomainsForBulkOperations() {
    try {
        const domains = await apiCall('/api/v1/admin/domains');
        if (domains && domains.domains) {
            // Store domains for reference
            window.availableDomains = domains.domains;
        }
    } catch (error) {
        console.error('Failed to load domains for bulk operations:', error);
    }
}

// Domain selection helpers
function selectAllDomains(textareaId) {
    const textarea = document.getElementById(textareaId);
    if (textarea && window.availableDomains) {
        // Load all available domains from the API
        const allDomainNames = window.availableDomains.map(domain => domain.name).join('\n');
        textarea.value = allDomainNames;
        showNotification(`Loaded ${window.availableDomains.length} domains`, 'success');
    } else {
        // If domains not loaded, try to load them
        loadDomainsForBulkOperations().then(() => {
            if (window.availableDomains) {
                const allDomainNames = window.availableDomains.map(domain => domain.name).join('\n');
                textarea.value = allDomainNames;
                showNotification(`Loaded ${window.availableDomains.length} domains`, 'success');
            }
        });
    }
}

function clearDomainSelection(textareaId) {
    const textarea = document.getElementById(textareaId);
    if (textarea) {
        textarea.value = '';
        showNotification('Domain list cleared', 'info');
    }
}

// Nameserver presets
function applyNameserverPreset() {
    const preset = document.getElementById('nsPresets').value;
    const ns1 = document.getElementById('ns1');
    const ns2 = document.getElementById('ns2');
    const ns3 = document.getElementById('ns3');
    const ns4 = document.getElementById('ns4');
    
    switch (preset) {
        case 'cloudflare':
            ns1.value = 'ns1.cloudflare.com';
            ns2.value = 'ns2.cloudflare.com';
            ns3.value = '';
            ns4.value = '';
            break;
        case 'godaddy':
            ns1.value = 'ns1.godaddy.com';
            ns2.value = 'ns2.godaddy.com';
            ns3.value = '';
            ns4.value = '';
            break;
        case 'namecheap':
            ns1.value = 'dns1.namecheap.com';
            ns2.value = 'dns2.namecheap.com';
            ns3.value = '';
            ns4.value = '';
            break;
        case 'aws':
            ns1.value = 'ns-1.awsdns-00.com';
            ns2.value = 'ns-2.awsdns-00.net';
            ns3.value = 'ns-3.awsdns-00.org';
            ns4.value = 'ns-4.awsdns-00.co.uk';
            break;
        case 'custom':
            ns1.value = '';
            ns2.value = '';
            ns3.value = '';
            ns4.value = '';
            break;
    }
}

// Preview bulk IP changes
function previewBulkIpChanges() {
    const domainsTextarea = document.getElementById('bulkIpDomains');
    const domainNames = parseDomainList(domainsTextarea.value);
    const newIp = document.getElementById('bulkNewIp').value;
    const recordName = document.getElementById('bulkRecordName').value || '@';
    const ttl = document.getElementById('bulkTtl').value;
    
    if (domainNames.length === 0) {
        showNotification('Please enter at least one domain name', 'error');
        return;
    }
    
    if (!newIp || !isValidIP(newIp)) {
        showNotification('Please enter a valid IP address', 'error');
        return;
    }
    
    const previewContainer = document.getElementById('bulkIpPreview');
    const changes = domainNames.map(domainName => {
        return {
            domainName: domainName.trim(),
            recordName: recordName,
            newValue: newIp,
            ttl: ttl,
            type: 'A'
        };
    });
    
    bulkOperationData = {
        type: 'ip',
        changes: changes
    };
    
    const previewHtml = changes.map(change => `
        <div class="preview-item">
            <div class="preview-domain">${change.domainName}</div>
            <div class="preview-change new">Add/Update A record: ${change.recordName} → ${change.newValue} (TTL: ${change.ttl}s)</div>
        </div>
    `).join('');
    
    previewContainer.innerHTML = previewHtml;
    document.getElementById('bulkIpApplyBtn').disabled = false;
}

// Preview bulk nameserver changes
function previewBulkNsChanges() {
    const domainsTextarea = document.getElementById('bulkNsDomains');
    const domainNames = parseDomainList(domainsTextarea.value);
    const ns1 = document.getElementById('ns1').value;
    const ns2 = document.getElementById('ns2').value;
    const ns3 = document.getElementById('ns3').value;
    const ns4 = document.getElementById('ns4').value;
    
    if (domainNames.length === 0) {
        showNotification('Please enter at least one domain name', 'error');
        return;
    }
    
    if (!ns1 || !ns2) {
        showNotification('Please enter at least 2 nameservers', 'error');
        return;
    }
    
    const nameservers = [ns1, ns2];
    if (ns3) nameservers.push(ns3);
    if (ns4) nameservers.push(ns4);
    
    const previewContainer = document.getElementById('bulkNsPreview');
    const changes = domainNames.map(domainName => {
        return {
            domainName: domainName.trim(),
            nameservers: nameservers
        };
    });
    
    bulkOperationData = {
        type: 'nameservers',
        changes: changes
    };
    
    const previewHtml = changes.map(change => `
        <div class="preview-item">
            <div class="preview-domain">${change.domainName}</div>
            <div class="preview-change update">Update nameservers: ${change.nameservers.join(', ')}</div>
        </div>
    `).join('');
    
    previewContainer.innerHTML = previewHtml;
    document.getElementById('bulkNsApplyBtn').disabled = false;
}

// CSV Upload handling
function handleCsvUpload(event) {
    const file = event.target.files[0];
    if (!file) return;
    
    if (!file.name.endsWith('.csv')) {
        showNotification('Please select a CSV file', 'error');
        return;
    }
    
    const reader = new FileReader();
    reader.onload = function(e) {
        const csv = e.target.result;
        parseCsvData(csv);
    };
    reader.readAsText(file);
}

function parseCsvData(csvText) {
    const lines = csvText.split('\n').filter(line => line.trim());
    if (lines.length < 2) {
        showNotification('CSV file must contain at least a header and one data row', 'error');
        return;
    }
    
    const headers = lines[0].split(',').map(h => h.trim());
    const expectedHeaders = ['domain', 'record_type', 'name', 'value', 'ttl', 'nameserver1', 'nameserver2'];
    
    if (!expectedHeaders.every(h => headers.includes(h))) {
        showNotification('CSV headers do not match expected format', 'error');
        return;
    }
    
    const data = [];
    for (let i = 1; i < lines.length; i++) {
        const values = lines[i].split(',').map(v => v.trim());
        const row = {};
        headers.forEach((header, index) => {
            row[header] = values[index] || '';
        });
        data.push(row);
    }
    
    csvData = data;
    displayCsvPreview(data);
    document.getElementById('csvPreviewSection').style.display = 'block';
    document.getElementById('csvActions').style.display = 'flex';
}

function displayCsvPreview(data) {
    const tbody = document.querySelector('#csvPreviewTable tbody');
    const rows = data.slice(0, 10).map(row => { // Show first 10 rows
        const currentSettings = 'Loading...';
        const proposedChanges = [];
        
        if (row.record_type && row.name && row.value) {
            proposedChanges.push(`${row.record_type} record: ${row.name} → ${row.value}`);
        }
        if (row.nameserver1 && row.nameserver2) {
            proposedChanges.push(`Nameservers: ${row.nameserver1}, ${row.nameserver2}`);
        }
        
        return `
            <tr>
                <td>${row.domain}</td>
                <td>${currentSettings}</td>
                <td>${proposedChanges.join('<br>')}</td>
                <td><span class="status-badge status-warning">Pending</span></td>
            </tr>
        `;
    }).join('');
    
    tbody.innerHTML = rows;
    if (data.length > 10) {
        tbody.innerHTML += `<tr><td colspan="4">... and ${data.length - 10} more rows</td></tr>`;
    }
}

function validateCsvData() {
    if (!csvData) {
        showNotification('No CSV data to validate', 'error');
        return;
    }
    
    const errors = [];
    csvData.forEach((row, index) => {
        if (!row.domain) {
            errors.push(`Row ${index + 2}: Missing domain`);
        }
        if (row.record_type && !['A', 'AAAA', 'CNAME', 'MX', 'TXT', 'NS', 'SRV'].includes(row.record_type)) {
            errors.push(`Row ${index + 2}: Invalid record type`);
        }
        if (row.record_type === 'A' && row.value && !isValidIP(row.value)) {
            errors.push(`Row ${index + 2}: Invalid IP address`);
        }
    });
    
    if (errors.length > 0) {
        showNotification(`Validation errors:\n${errors.join('\n')}`, 'error');
    } else {
        showNotification('CSV data validation passed', 'success');
        bulkOperationData = {
            type: 'csv',
            changes: csvData
        };
        document.getElementById('csvApplyBtn').disabled = false;
    }
}

function downloadCsvTemplate() {
    const template = `domain,record_type,name,value,ttl,nameserver1,nameserver2
example.com,A,@,192.168.1.100,3600,ns1.cloudflare.com,ns2.cloudflare.com
test.com,A,www,192.168.1.101,3600,ns1.godaddy.com,ns2.godaddy.com
mysite.org,CNAME,blog,myblog.wordpress.com,1800,,`;
    
    const blob = new Blob([template], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'bulk_dns_template.csv';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
}

// Bulk confirmation modal
function showBulkConfirmationModal(operationType) {
    if (!bulkOperationData) {
        showNotification('No bulk operation data available', 'error');
        return;
    }
    
    bulkOperationType = operationType;
    const modal = document.getElementById('bulkConfirmationModal');
    const summaryDiv = document.getElementById('bulkConfirmationSummary');
    
    let summaryHtml = '';
    switch (operationType) {
        case 'ip':
            summaryHtml = `
                <h4>Bulk IP Assignment Summary</h4>
                <p><strong>${bulkOperationData.changes.length}</strong> domains will be updated with IP address <strong>${bulkOperationData.changes[0].newValue}</strong></p>
                <ul>${bulkOperationData.changes.map(c => `<li>${c.domainName}</li>`).join('')}</ul>
            `;
            break;
        case 'nameservers':
            summaryHtml = `
                <h4>Bulk Nameserver Changes Summary</h4>
                <p><strong>${bulkOperationData.changes.length}</strong> domains will have their nameservers updated</p>
                <p><strong>New nameservers:</strong> ${bulkOperationData.changes[0].nameservers.join(', ')}</p>
                <ul>${bulkOperationData.changes.map(c => `<li>${c.domainName}</li>`).join('')}</ul>
            `;
            break;
        case 'csv':
            summaryHtml = `
                <h4>CSV Bulk Changes Summary</h4>
                <p><strong>${bulkOperationData.changes.length}</strong> domains will be processed from CSV data</p>
                <p>Changes include DNS records and nameserver updates as specified in the uploaded file.</p>
            `;
            break;
    }
    
    summaryDiv.innerHTML = summaryHtml;
    modal.style.display = 'flex';
    
    // Setup confirmation checkbox listener
    const confirmCheckbox = document.getElementById('confirmBulkChanges');
    const passwordInput = document.getElementById('bulkConfirmPassword');
    const executeBtn = document.getElementById('executeBulkBtn');
    
    const checkEnableButton = () => {
        executeBtn.disabled = !(confirmCheckbox.checked && passwordInput.value.length > 0);
    };
    
    confirmCheckbox.addEventListener('change', checkEnableButton);
    passwordInput.addEventListener('input', checkEnableButton);
}

function hideBulkConfirmationModal() {
    document.getElementById('bulkConfirmationModal').style.display = 'none';
    document.getElementById('bulkConfirmPassword').value = '';
    document.getElementById('confirmBulkChanges').checked = false;
    document.getElementById('executeBulkBtn').disabled = true;
}

// Execute bulk changes
async function executeBulkChanges() {
    const password = document.getElementById('bulkConfirmPassword').value;
    
    if (!password) {
        showNotification('Password is required', 'error');
        return;
    }
    
    try {
        let endpoint, payload;
        
        switch (bulkOperationType) {
            case 'ip':
                endpoint = '/api/v1/admin/dns/bulk/ip';
                payload = {
                    password: password,
                    operations: bulkOperationData.changes.map(change => ({
                        domain_name: change.domainName,
                        record_name: change.recordName,
                        ip_address: change.newValue,
                        ttl: parseInt(change.ttl)
                    }))
                };
                break;
            case 'nameservers':
                endpoint = '/api/v1/admin/dns/bulk/nameservers';
                payload = {
                    password: password,
                    operations: bulkOperationData.changes.map(change => ({
                        domain_name: change.domainName,
                        nameservers: change.nameservers
                    }))
                };
                break;
            case 'csv':
                endpoint = '/api/v1/admin/dns/bulk/csv';
                payload = {
                    password: password,
                    csv_data: bulkOperationData.changes
                };
                break;
        }
        
        const response = await apiCall(endpoint, {
            method: 'POST',
            body: JSON.stringify(payload)
        });
        
        if (response) {
            showNotification(`Bulk ${bulkOperationType} operation completed successfully`, 'success');
            hideBulkConfirmationModal();
            
            // Reset forms and preview
            bulkOperationData = null;
            bulkOperationType = null;
            resetBulkForms();
        }
    } catch (error) {
        console.error('Bulk operation failed:', error);
        showNotification('Bulk operation failed', 'error');
    }
}

function resetBulkForms() {
    // Reset IP form
    document.getElementById('bulkIpDomains').value = '';
    document.getElementById('bulkNewIp').value = '';
    document.getElementById('bulkRecordName').value = '@';
    document.getElementById('bulkTtl').value = '3600';
    document.getElementById('bulkIpPreview').innerHTML = '<p class="preview-empty">Enter domains and IP to preview changes</p>';
    document.getElementById('bulkIpApplyBtn').disabled = true;
    
    // Reset nameserver form
    document.getElementById('bulkNsDomains').value = '';
    document.getElementById('ns1').value = '';
    document.getElementById('ns2').value = '';
    document.getElementById('ns3').value = '';
    document.getElementById('ns4').value = '';
    document.getElementById('nsPresets').value = '';
    document.getElementById('bulkNsPreview').innerHTML = '<p class="preview-empty">Enter domains and nameservers to preview changes</p>';
    document.getElementById('bulkNsApplyBtn').disabled = true;
    
    // Reset CSV form
    document.getElementById('csvFileInput').value = '';
    document.getElementById('csvPreviewSection').style.display = 'none';
    document.getElementById('csvActions').style.display = 'none';
    document.getElementById('csvApplyBtn').disabled = true;
    csvData = null;
}

// Utility functions
function isValidIP(ip) {
    const ipRegex = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
    return ipRegex.test(ip);
}

function parseDomainList(input) {
    if (!input || typeof input !== 'string') {
        return [];
    }
    
    // Split by line breaks, commas, or semicolons and filter out empty entries
    return input
        .split(/[\n,;]+/)
        .map(domain => domain.trim())
        .filter(domain => domain.length > 0)
        .filter(domain => {
            // Basic domain validation
            const domainRegex = /^[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]?\.[a-zA-Z]{2,}$/;
            return domainRegex.test(domain);
        });
}

function isDomainValid(domain) {
    const domainRegex = /^[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]?\.[a-zA-Z]{2,}$/;
    return domainRegex.test(domain);
}

// Global function exports for bulk DNS management
window.switchBulkTab = switchBulkTab;
window.selectAllDomains = selectAllDomains;
window.clearDomainSelection = clearDomainSelection;
window.applyNameserverPreset = applyNameserverPreset;
window.previewBulkIpChanges = previewBulkIpChanges;
window.previewBulkNsChanges = previewBulkNsChanges;
window.handleCsvUpload = handleCsvUpload;
window.validateCsvData = validateCsvData;
window.downloadCsvTemplate = downloadCsvTemplate;
window.showBulkConfirmationModal = showBulkConfirmationModal;
window.hideBulkConfirmationModal = hideBulkConfirmationModal;
window.executeBulkChanges = executeBulkChanges;
