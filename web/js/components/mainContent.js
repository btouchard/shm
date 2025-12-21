// SPDX-License-Identifier: AGPL-3.0-or-later

/**
 * Main Content component - Handles the main scrollable area
 */
export default () => ({
    /**
     * Get the dashboard store
     */
    get store() {
        return this.$store.dashboard;
    },

    /**
     * Get displayed groups
     */
    get displayedGroups() {
        return this.store.getDisplayedGroups();
    },

    /**
     * Check if loading
     */
    get isLoading() {
        return this.store.loading;
    },

    /**
     * Check if there are no groups to display
     */
    get isEmpty() {
        return Object.keys(this.displayedGroups).length === 0 && !this.isLoading;
    },

    /**
     * Handle scroll for lazy loading
     */
    handleScroll(event) {
        if (this.store.selectedApp === null) return;

        const el = event.target;
        if (el.scrollTop + el.clientHeight >= el.scrollHeight - 200) {
            this.store.loadMoreInstances(this.store.selectedApp);
        }
    }
});
