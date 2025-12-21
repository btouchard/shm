// SPDX-License-Identifier: AGPL-3.0-or-later

/**
 * API utilities for the SHM dashboard
 */

const API_BASE = '/api/v1/admin';

/**
 * Fetch dashboard statistics
 * @returns {Promise<{total_instances: number, active_instances: number, per_app_counts?: Object}>}
 */
export async function fetchStats() {
    const response = await fetch(`${API_BASE}/stats`);
    if (!response.ok) throw new Error('Failed to fetch stats');
    return response.json();
}

/**
 * Fetch applications list
 * @returns {Promise<Array<{name: string, slug: string, github_url?: string, stars?: number}>>}
 */
export async function fetchApplications() {
    const response = await fetch(`${API_BASE}/applications`);
    if (!response.ok) throw new Error('Failed to fetch applications');
    return response.json();
}

/**
 * Fetch instances with pagination and filtering
 * @param {Object} options - Query options
 * @param {number} [options.offset=0] - Pagination offset
 * @param {number} [options.limit=50] - Page size
 * @param {string} [options.app] - Filter by app name
 * @param {string} [options.query] - Search query
 * @returns {Promise<Array>}
 */
export async function fetchInstances({ offset = 0, limit = 50, app = null, query = null } = {}) {
    const params = new URLSearchParams();
    params.set('offset', offset);
    params.set('limit', limit);

    if (app) params.set('app', app);
    if (query?.trim()) params.set('q', query.trim());

    const response = await fetch(`${API_BASE}/instances?${params.toString()}`);
    if (!response.ok) throw new Error('Failed to fetch instances');
    return response.json();
}

/**
 * Fetch metrics time series for an application
 * @param {string} appName - Application name
 * @param {string} [period='24h'] - Time period (24h, 7d, 30d, 3m, 1y, all)
 * @returns {Promise<{timestamps: string[], metrics: Object}>}
 */
export async function fetchMetrics(appName, period = '24h') {
    const response = await fetch(
        `${API_BASE}/metrics/${encodeURIComponent(appName)}?period=${period}`
    );
    if (!response.ok) throw new Error('Failed to fetch metrics');
    return response.json();
}

/**
 * Update an application's metadata
 * @param {string} slug - Application slug
 * @param {Object} data - Update data
 * @param {string} [data.github_url] - GitHub repository URL
 * @param {string} [data.logo_url] - Custom logo URL
 * @returns {Promise<{status: string, message: string}>}
 */
export async function updateApplication(slug, data) {
    const response = await fetch(`${API_BASE}/applications/${encodeURIComponent(slug)}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    });
    if (!response.ok) {
        const error = await response.text();
        throw new Error(error || 'Failed to update application');
    }
    return response.json();
}

/**
 * Refresh GitHub stars for an application
 * @param {string} slug - Application slug
 * @returns {Promise<{status: string, message: string}>}
 */
export async function refreshApplicationStars(slug) {
    const response = await fetch(`${API_BASE}/applications/${encodeURIComponent(slug)}/refresh-stars`, {
        method: 'POST'
    });
    if (!response.ok) {
        const error = await response.text();
        throw new Error(error || 'Failed to refresh stars');
    }
    return response.json();
}
