// SPDX-License-Identifier: AGPL-3.0-or-later

/**
 * Icon helper utilities for Phosphor Icons
 */

/**
 * Get the appropriate icon class for a resource type
 * @param {string} key - Resource key
 * @returns {string} Phosphor icon class
 */
export function getResourceIcon(key) {
    if (key.includes('mem')) return 'ph-memory';
    if (key.includes('cpu')) return 'ph-cpu';
    if (key.includes('uptime')) return 'ph-clock';
    if (key.includes('goroutines')) return 'ph-arrows-split';
    return 'ph-activity';
}

/**
 * Get the appropriate icon class for a tag value
 * @param {string} val - Tag value
 * @returns {string} Phosphor icon class
 */
export function getIconForTag(val) {
    const v = String(val).toLowerCase();
    if (v.includes('linux')) return 'ph-linux-logo';
    if (v.includes('windows')) return 'ph-windows-logo';
    if (v.includes('darwin') || v.includes('apple')) return 'ph-apple-logo';
    if (v.includes('docker')) return 'ph-container';
    return 'ph-tag';
}

/**
 * Get the appropriate icon class for an OS name
 * @param {string} os - Operating system name
 * @returns {string} Phosphor icon class
 */
export function getOSIcon(os) {
    const v = String(os).toLowerCase();
    if (v.includes('linux')) return 'ph-linux-logo';
    if (v.includes('windows')) return 'ph-windows-logo';
    if (v.includes('darwin') || v.includes('macos')) return 'ph-apple-logo';
    if (v.includes('bsd')) return 'ph-ghost';
    return 'ph-desktop';
}
