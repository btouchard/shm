// SPDX-License-Identifier: AGPL-3.0-or-later

/**
 * Navbar component - Top navigation bar with instance search
 */
export default () => ({
    searchQuery: '',

    init() {
        // Watch for search query changes
        this.$watch('searchQuery', (value) => {
            this.$store.dashboard.instanceSearchQuery = value;
            this.$store.dashboard.handleInstanceSearch();
        });
    },

    /**
     * Get the dashboard store
     */
    get store() {
        return this.$store.dashboard;
    },

    /**
     * Get the current title
     */
    get title() {
        return this.store.selectedApp || 'All Applications';
    },

    /**
     * Get instance count text
     */
    get countText() {
        const count = this.store.selectedApp
            ? this.store.getAppTotalCount(this.store.selectedApp)
            : this.store.stats.total_instances;
        return `(${count} instances)`;
    },

    /**
     * Check if currently searching
     */
    get isSearching() {
        return this.store.searchingInstances;
    }
});
