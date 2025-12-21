// SPDX-License-Identifier: AGPL-3.0-or-later

/**
 * SHM Dashboard - Alpine.js Application Bootstrap
 */

console.log('[SHM] Loading modules...');

// Import Alpine as ES Module (this ensures proper load order)
// Note: SRI for ES modules must be handled via importmap or build tools
import Alpine from 'https://cdn.jsdelivr.net/npm/alpinejs@3.14.3/dist/module.esm.js';

// Import stores
import dashboardStore from './stores/dashboard.js';
import chartsStore from './stores/charts.js';

// Import components
import sidebar from './components/sidebar.js';
import navbar from './components/navbar.js';
import appSection from './components/appSection.js';
import mainContent from './components/mainContent.js';
import drawer from './components/drawer.js';
import appEditModal from './components/appEditModal.js';

console.log('[SHM] All modules loaded, registering with Alpine...');

// Make Alpine available globally
window.Alpine = Alpine;

// Register stores BEFORE starting Alpine
Alpine.store('dashboard', dashboardStore);
Alpine.store('charts', chartsStore);

// Register components BEFORE starting Alpine
Alpine.data('sidebar', sidebar);
Alpine.data('navbar', navbar);
Alpine.data('appSection', appSection);
Alpine.data('mainContent', mainContent);
Alpine.data('drawer', drawer);
Alpine.data('appEditModal', appEditModal);

console.log('[SHM] Starting Alpine...');

// Start Alpine
Alpine.start();

console.log('[SHM] Alpine started, initializing dashboard...');

// Initialize dashboard store
Alpine.store('dashboard').init();

console.log('[SHM] Dashboard initialized');

// Cleanup on page unload
window.addEventListener('beforeunload', () => {
    Alpine.store('charts')?.destroy();
});
