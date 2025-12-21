// SPDX-License-Identifier: AGPL-3.0-or-later

import { formatNumber, formatKey, formatResourceKey } from '../utils/formatters.js';
import { getResourceIcon } from '../utils/icons.js';

/**
 * Drawer component - Instance detail panel
 */
export default () => ({
    // Expose utilities to template
    formatNumber,
    formatKey,
    formatResourceKey,
    getResourceIcon,

    /**
     * Get the dashboard store
     */
    get store() {
        return this.$store.dashboard;
    },

    /**
     * Get selected instance
     */
    get instance() {
        return this.store.selectedInstance;
    },

    /**
     * Get resource keys
     */
    get resourceKeys() {
        return this.store.currentResourceKeys;
    },

    /**
     * Check if drawer is open
     */
    get isOpen() {
        return this.instance !== null;
    },

    /**
     * Close the drawer
     */
    close() {
        this.store.closeDrawer();
    }
});
