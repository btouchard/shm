// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"net/http"
	"strings"

	"github.com/btouchard/shm/internal/services/badge"
)

// BadgeInstances generates a badge showing the number of active instances.
// GET /badge/{app-slug}/instances
func (h *Handlers) BadgeInstances(w http.ResponseWriter, r *http.Request) {
	// Extract app slug from path
	appSlug := extractSlugFromPath(r.URL.Path, "/badge/", "/instances")
	if appSlug == "" {
		renderErrorBadge(w, "invalid slug")
		return
	}

	// Get active instances count
	count, err := h.dashboard.GetActiveInstancesCount(r.Context(), appSlug)
	if err != nil {
		h.logger.Warn("failed to get instances count", "slug", appSlug, "error", err)
		renderErrorBadge(w, "error")
		return
	}

	// Determine color based on count
	color := badge.GetInstancesColor(count)

	// Allow custom color via query param
	if customColor := r.URL.Query().Get("color"); customColor != "" {
		color = "#" + strings.TrimPrefix(customColor, "#")
	}

	// Allow custom label
	label := r.URL.Query().Get("label")
	if label == "" {
		label = "instances"
	}

	// Generate badge
	b := badge.NewBadge(label, badge.FormatNumber(float64(count)), color)
	renderSVGBadge(w, b.ToSVG())
}

// BadgeVersion generates a badge showing the most used version.
// GET /badge/{app-slug}/version
func (h *Handlers) BadgeVersion(w http.ResponseWriter, r *http.Request) {
	// Extract app slug from path
	appSlug := extractSlugFromPath(r.URL.Path, "/badge/", "/version")
	if appSlug == "" {
		renderErrorBadge(w, "invalid slug")
		return
	}

	// Get most used version
	version, err := h.dashboard.GetMostUsedVersion(r.Context(), appSlug)
	if err != nil {
		h.logger.Warn("failed to get version", "slug", appSlug, "error", err)
		renderErrorBadge(w, "error")
		return
	}

	if version == "" {
		version = "no data"
	}

	// Allow custom color via query param
	color := badge.ColorPurple
	if customColor := r.URL.Query().Get("color"); customColor != "" {
		color = "#" + strings.TrimPrefix(customColor, "#")
	}

	// Allow custom label
	label := r.URL.Query().Get("label")
	if label == "" {
		label = "version"
	}

	// Generate badge
	b := badge.NewBadge(label, version, color)
	renderSVGBadge(w, b.ToSVG())
}

// BadgeMetric generates a badge showing an aggregated metric value.
// GET /badge/{app-slug}/metric/{metric-name}
func (h *Handlers) BadgeMetric(w http.ResponseWriter, r *http.Request) {
	// Extract app slug and metric name from path
	// Path format: /badge/{slug}/metric/{name}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/badge/"), "/")
	if len(parts) < 3 || parts[1] != "metric" {
		renderErrorBadge(w, "invalid path")
		return
	}

	appSlug := parts[0]
	metricName := parts[2]

	if appSlug == "" || metricName == "" {
		renderErrorBadge(w, "invalid params")
		return
	}

	// Get aggregated metric
	value, err := h.dashboard.GetAggregatedMetric(r.Context(), appSlug, metricName)
	if err != nil {
		h.logger.Warn("failed to get metric", "slug", appSlug, "metric", metricName, "error", err)
		renderErrorBadge(w, "error")
		return
	}

	// Determine color based on value
	color := badge.GetMetricColor(value)

	// Allow custom color via query param
	if customColor := r.URL.Query().Get("color"); customColor != "" {
		color = "#" + strings.TrimPrefix(customColor, "#")
	}

	// Allow custom label
	label := r.URL.Query().Get("label")
	if label == "" {
		label = metricName
	}

	// Generate badge
	b := badge.NewBadge(label, badge.FormatNumber(value), color)
	renderSVGBadge(w, b.ToSVG())
}

// BadgeCombined generates a combined badge showing metric value and instance count.
// GET /badge/{app-slug}/combined?metric=users_count
func (h *Handlers) BadgeCombined(w http.ResponseWriter, r *http.Request) {
	// Extract app slug from path
	appSlug := extractSlugFromPath(r.URL.Path, "/badge/", "/combined")
	if appSlug == "" {
		renderErrorBadge(w, "invalid slug")
		return
	}

	// Get metric name from query param (default: users_count)
	metricName := r.URL.Query().Get("metric")
	if metricName == "" {
		metricName = "users_count"
	}

	// Get combined stats
	metricValue, instanceCount, err := h.dashboard.GetCombinedStats(r.Context(), appSlug, metricName)
	if err != nil {
		h.logger.Warn("failed to get combined stats", "slug", appSlug, "metric", metricName, "error", err)
		renderErrorBadge(w, "error")
		return
	}

	// Allow custom color
	color := badge.ColorIndigo
	if customColor := r.URL.Query().Get("color"); customColor != "" {
		color = "#" + strings.TrimPrefix(customColor, "#")
	}

	// Allow custom label
	label := r.URL.Query().Get("label")
	if label == "" {
		label = "adoption"
	}

	// Format compact: "1.2k / 42"
	value := badge.FormatNumber(metricValue) + " / " + badge.FormatNumber(float64(instanceCount))

	// Generate badge
	b := badge.NewBadge(label, value, color)
	renderSVGBadge(w, b.ToSVG())
}

// Helper functions

// extractSlugFromPath extracts the app slug from a badge path.
// Example: /badge/my-app/instances -> "my-app"
func extractSlugFromPath(path, prefix, suffix string) string {
	s := strings.TrimPrefix(path, prefix)
	s = strings.TrimSuffix(s, suffix)
	return s
}

// renderSVGBadge writes an SVG badge to the response with proper headers.
func renderSVGBadge(w http.ResponseWriter, svg string) {
	w.Header().Set("Content-Type", "image/svg+xml;charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300") // 5 minutes cache
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(svg))
}

// renderErrorBadge renders an error badge.
func renderErrorBadge(w http.ResponseWriter, message string) {
	b := badge.NewBadge("error", message, badge.ColorRed)
	renderSVGBadge(w, b.ToSVG())
}
