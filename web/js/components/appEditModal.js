// SPDX-License-Identifier: AGPL-3.0-or-later

import { updateApplication, refreshApplicationStars } from '../utils/api.js';

/**
 * Application edit modal component
 */
export default () => ({
    // Form state
    form: {
        github_url: '',
        logo_url: ''
    },

    // UI state
    saving: false,
    refreshingStars: false,
    error: null,
    success: null,

    /**
     * Initialize - watch for editingApp changes
     */
    init() {
        this.$watch('$store.dashboard.editingApp', (app) => {
            if (app) {
                this.form.github_url = app.github_url || '';
                this.form.logo_url = app.logo_url || '';
                this.error = null;
                this.success = null;
            }
        });
    },

    /**
     * Check if modal is open
     */
    get isOpen() {
        return this.$store.dashboard.editingApp !== null;
    },

    /**
     * Get the application being edited
     */
    get app() {
        return this.$store.dashboard.editingApp;
    },

    /**
     * Close the modal
     */
    close() {
        this.$store.dashboard.closeEditModal();
        this.error = null;
        this.success = null;
    },

    /**
     * Save changes
     */
    async save() {
        const app = this.$store.dashboard.editingApp;
        if (!app) return;

        this.saving = true;
        this.error = null;
        this.success = null;

        const newGithubUrl = this.form.github_url.trim();
        const oldGithubUrl = app.github_url || '';
        const githubUrlChanged = newGithubUrl !== oldGithubUrl && newGithubUrl !== '';

        try {
            await updateApplication(app.slug, {
                github_url: newGithubUrl,
                logo_url: this.form.logo_url.trim()
            });

            // If GitHub URL changed, refresh stars
            if (githubUrlChanged) {
                this.success = 'Updating stars...';
                try {
                    await refreshApplicationStars(app.slug);
                    this.success = 'Application updated & stars synced!';
                } catch (e) {
                    console.warn('[SHM] Failed to refresh stars:', e);
                    this.success = 'Application updated (stars sync failed)';
                }
            } else {
                this.success = 'Application updated successfully';
            }

            // Refresh data after a short delay to show success message
            setTimeout(async () => {
                await this.$store.dashboard.fetchInitialData();
                this.close();
            }, 1000);
        } catch (e) {
            this.error = e.message || 'Failed to save changes';
        } finally {
            this.saving = false;
        }
    },

    /**
     * Refresh GitHub stars
     */
    async refreshStars() {
        const app = this.$store.dashboard.editingApp;
        if (!app || !this.form.github_url) return;

        this.refreshingStars = true;
        this.error = null;

        try {
            await refreshApplicationStars(app.slug);
            this.success = 'Stars refreshed successfully';

            // Refresh data
            await this.$store.dashboard.fetchInitialData();
        } catch (e) {
            this.error = e.message || 'Failed to refresh stars';
        } finally {
            this.refreshingStars = false;
        }
    },

    /**
     * Check if GitHub URL is valid
     */
    get isValidGitHubUrl() {
        if (!this.form.github_url) return true;
        // Support repos with dots, underscores and longer org/repo names
        return /^https:\/\/github\.com\/[\w.-]+\/[\w.-]+\/?$/.test(this.form.github_url);
    },

    /**
     * Check if Logo URL is valid
     * Only allows HTTPS URLs with common image extensions or known CDN domains
     */
    get isValidLogoUrl() {
        if (!this.form.logo_url) return true;
        const url = this.form.logo_url.trim();

        // Must be HTTPS
        if (!url.startsWith('https://')) return false;

        // Block dangerous protocols
        if (url.startsWith('javascript:') || url.startsWith('data:')) return false;

        // Allow common image extensions or known safe domains
        const allowedExtensions = /\.(png|jpg|jpeg|gif|svg|webp|ico)(\?.*)?$/i;
        const allowedDomains = [
            'github.com',
            'githubusercontent.com',
            'raw.githubusercontent.com',
            'avatars.githubusercontent.com',
            'cdn.jsdelivr.net',
            'unpkg.com',
            'cloudflare.com',
            'imgur.com',
            'i.imgur.com'
        ];

        try {
            const parsedUrl = new URL(url);
            const isAllowedDomain = allowedDomains.some(d => parsedUrl.hostname.endsWith(d));
            const hasAllowedExtension = allowedExtensions.test(parsedUrl.pathname);
            return isAllowedDomain || hasAllowedExtension;
        } catch (e) {
            return false;
        }
    },

    /**
     * Check if form is valid
     */
    get isFormValid() {
        return this.isValidGitHubUrl && this.isValidLogoUrl;
    },

    /**
     * Check if form has changes
     */
    get hasChanges() {
        const app = this.$store.dashboard.editingApp;
        if (!app) return false;
        return (
            this.form.github_url !== (app.github_url || '') ||
            this.form.logo_url !== (app.logo_url || '')
        );
    }
});
