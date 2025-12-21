// SPDX-License-Identifier: AGPL-3.0-or-later

import { fetchStats, fetchApplications, fetchInstances } from '../utils/api.js';

/**
 * Main dashboard store - manages global application state
 * Access via Alpine.$store('dashboard') or this.$store.dashboard in components
 */
export default {
    // Loading states
    loading: false,
    loadingMore: false,
    searchingInstances: false,

    // Core data
    stats: { total_instances: 0, active_instances: 0, per_app_counts: {} },
    applications: [],
    rawInstances: [],
    groupedInstances: {},
    refreshKey: 0,

    // Selection state
    selectedApp: null,
    selectedInstance: null,
    currentResourceKeys: [],
    editingApp: null,

    // Search state
    searchQuery: '',
    instanceSearchQuery: '',

    // Pagination
    apiOffset: 0,
    hasMoreFromApi: true,
    loadedPages: {},

    // Constants
    API_PAGE_SIZE: 50,
    INSTANCES_PER_PAGE: 25,
    INSTANCES_PREVIEW_COUNT: 5,

    // Debounce timer
    _searchDebounce: null,

    /**
     * Initialize the store - fetch initial data
     */
    async init() {
        await this.fetchInitialData();
    },

    /**
     * Select an application (or null for all apps view)
     */
    selectApp(appName) {
        this.selectedApp = appName;
        this.fetchInstancesOnly();
    },

    /**
     * Open the instance detail drawer
     */
    openDrawer(instance, resourceKeys) {
        this.selectedInstance = instance;
        this.currentResourceKeys = resourceKeys || [];
    },

    /**
     * Close the instance detail drawer
     */
    closeDrawer() {
        this.selectedInstance = null;
        this.currentResourceKeys = [];
    },

    /**
     * Open the application edit modal
     */
    openEditModal(appSlug) {
        const app = this.applications.find(a => a.slug === appSlug);
        if (app) {
            this.editingApp = app;
        }
    },

    /**
     * Close the application edit modal
     */
    closeEditModal() {
        this.editingApp = null;
    },

    /**
     * Handle instance search with debounce
     */
    handleInstanceSearch() {
        clearTimeout(this._searchDebounce);
        this.searchingInstances = true;
        this._searchDebounce = setTimeout(() => {
            this.fetchInstancesOnly().then(() => {
                this.searchingInstances = false;
            });
        }, 400);
    },

    /**
     * Fetch all initial data (stats, apps, instances)
     */
    async fetchInitialData() {
        this.loading = true;
        this.apiOffset = 0;
        this.hasMoreFromApi = true;
        this.rawInstances = [];

        try {
            const [stats, applications, instances] = await Promise.all([
                fetchStats(),
                fetchApplications(),
                fetchInstances({
                    offset: 0,
                    limit: this.API_PAGE_SIZE,
                    app: this.selectedApp,
                    query: this.instanceSearchQuery
                })
            ]);

            this.stats = stats;
            this.applications = applications;
            this.rawInstances = instances;
            this.apiOffset = instances.length;
            this.hasMoreFromApi = instances.length >= this.API_PAGE_SIZE;
            this.refreshKey++;

            this.processData();
        } catch (e) {
            console.error('Failed to fetch initial data:', e);
        } finally {
            this.loading = false;
        }
    },

    /**
     * Fetch only instances (for search/filter operations)
     */
    async fetchInstancesOnly() {
        this.loading = true;
        this.apiOffset = 0;
        this.hasMoreFromApi = true;
        this.rawInstances = [];

        try {
            const instances = await fetchInstances({
                offset: 0,
                limit: this.API_PAGE_SIZE,
                app: this.selectedApp,
                query: this.instanceSearchQuery
            });

            this.rawInstances = instances;
            this.apiOffset = instances.length;
            this.hasMoreFromApi = instances.length >= this.API_PAGE_SIZE;

            this.processData();
        } catch (e) {
            console.error('Failed to fetch instances:', e);
        } finally {
            this.loading = false;
        }
    },

    /**
     * Fetch more instances from API (pagination)
     */
    async fetchMoreFromApi() {
        if (this.loadingMore || !this.hasMoreFromApi) return;

        this.loadingMore = true;
        try {
            const instances = await fetchInstances({
                offset: this.apiOffset,
                limit: this.API_PAGE_SIZE,
                app: this.selectedApp,
                query: this.instanceSearchQuery
            });

            if (instances.length < this.API_PAGE_SIZE) {
                this.hasMoreFromApi = false;
            }

            if (instances.length > 0) {
                this.apiOffset += instances.length;
                // Use push with spread for better performance on large arrays
                this.rawInstances.push(...instances);
                this.processData();
            }
        } catch (e) {
            console.error('Failed to fetch more instances:', e);
        } finally {
            this.loadingMore = false;
        }
    },

    /**
     * Process raw instances into grouped structure
     */
    processData() {
        const groups = {};

        // Build lookup maps from applications for GitHub info and logo
        const appInfoBySlug = {};
        const appInfoByName = {};
        for (const app of this.applications) {
            const info = {
                slug: app.slug,
                githubUrl: app.github_url || null,
                githubStars: app.stars || 0,
                logoUrl: app.logo_url || null
            };
            appInfoBySlug[app.slug] = info;
            appInfoByName[app.name] = info;
        }

        for (const inst of this.rawInstances) {
            const appName = inst.app_name || 'Unknown App';

            if (!groups[appName]) {
                const appInfo = appInfoBySlug[inst.app_slug] || appInfoByName[appName] || {};
                groups[appName] = {
                    instances: [],
                    businessKeys: new Set(),
                    resourceKeys: new Set(),
                    stringKeys: new Set(),
                    sums: {},
                    osCounts: {},
                    lastSeenGlobal: null,
                    githubUrl: appInfo.githubUrl || null,
                    githubStars: appInfo.githubStars || 0,
                    logoUrl: appInfo.logoUrl || null,
                    appSlug: inst.app_slug || appInfo.slug || ''
                };
            }

            const group = groups[appName];
            const lastSeen = new Date(inst.last_seen_at);

            if (!group.lastSeenGlobal || lastSeen > group.lastSeenGlobal) {
                group.lastSeenGlobal = lastSeen;
            }

            group.instances.push(inst);

            if (inst.metrics) {
                if (inst.metrics.sys_os) {
                    const os = inst.metrics.sys_os;
                    group.osCounts[os] = (group.osCounts[os] || 0) + 1;
                }

                for (const [key, val] of Object.entries(inst.metrics)) {
                    if (typeof val === 'number') {
                        const isResource = key.startsWith('sys_') || key.startsWith('app_') ||
                            ['cpu', 'mem', 'uptime', 'goroutines'].some(k => key.includes(k));

                        if (isResource) {
                            group.resourceKeys.add(key);
                        } else {
                            group.businessKeys.add(key);
                            group.sums[key] = (group.sums[key] || 0) + val;
                        }
                    } else if (typeof val === 'string' && val.length < 25) {
                        group.stringKeys.add(key);
                    }
                }
            }
        }

        // Convert Sets to sorted arrays
        for (const group of Object.values(groups)) {
            group.businessKeys = [...group.businessKeys].sort();
            group.resourceKeys = [...group.resourceKeys].sort();
            group.stringKeys = [...group.stringKeys].sort();
        }

        this.groupedInstances = groups;
    },

    /**
     * Get filtered applications based on search query
     */
    get filteredApplications() {
        if (!this.searchQuery.trim()) {
            return this.applications;
        }
        const query = this.searchQuery.toLowerCase();
        return this.applications.filter(app =>
            app.name.toLowerCase().includes(query)
        );
    },

    /**
     * Get total count for an app (from stats or loaded data)
     */
    getAppTotalCount(appName) {
        if (this.stats.per_app_counts?.[appName] !== undefined) {
            return this.stats.per_app_counts[appName];
        }
        const group = this.groupedInstances[appName];
        return group ? group.instances.length : 0;
    },

    /**
     * Get displayed groups with pagination applied
     */
    getDisplayedGroups() {
        const isSingleAppView = this.selectedApp !== null;
        let groups = this.groupedInstances;

        // Filter by selected app
        if (isSingleAppView && this.groupedInstances[this.selectedApp]) {
            groups = { [this.selectedApp]: this.groupedInstances[this.selectedApp] };
        }

        // Apply sidebar search filter
        if (this.searchQuery.trim()) {
            const query = this.searchQuery.toLowerCase();
            groups = Object.fromEntries(
                Object.entries(groups).filter(([name]) =>
                    name.toLowerCase().includes(query)
                )
            );
        }

        const result = {};

        for (const [appName, group] of Object.entries(groups)) {
            if (!group) continue;
            const instances = group.instances;

            if (instances.length) {
                if (isSingleAppView) {
                    const currentPage = this.loadedPages[appName] || 0;
                    const displayCount = (currentPage + 1) * this.INSTANCES_PER_PAGE;
                    result[appName] = {
                        ...group,
                        displayedInstances: instances.slice(0, displayCount),
                        hasMore: displayCount < instances.length || this.hasMoreFromApi
                    };
                } else {
                    result[appName] = {
                        ...group,
                        displayedInstances: instances.slice(0, this.INSTANCES_PREVIEW_COUNT),
                        hasMore: false
                    };
                }
            }
        }

        return result;
    },

    /**
     * Load more instances for a specific app (lazy loading)
     */
    async loadMoreInstances(appName) {
        if (this.selectedApp === null) return;

        const group = this.groupedInstances[appName];
        if (!group) return;

        const currentPage = this.loadedPages[appName] || 0;
        const nextPage = currentPage + 1;
        const endIdx = (nextPage + 1) * this.INSTANCES_PER_PAGE;

        if (endIdx < group.instances.length) {
            this.loadedPages[appName] = nextPage;
        } else if (this.hasMoreFromApi) {
            await this.fetchMoreFromApi();
            this.loadedPages[appName] = nextPage;
        }
    }
};
