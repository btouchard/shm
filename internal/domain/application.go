// SPDX-License-Identifier: AGPL-3.0-or-later

package domain

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// ApplicationID is a validated application identifier (UUID format).
type ApplicationID string

// NewApplicationID creates and validates an ApplicationID.
func NewApplicationID(id string) (ApplicationID, error) {
	if id == "" {
		return "", ErrInvalidApplicationID
	}
	if !uuidRegex.MatchString(id) {
		return "", fmt.Errorf("%w: invalid UUID format", ErrInvalidApplicationID)
	}
	return ApplicationID(id), nil
}

// String returns the string representation.
func (id ApplicationID) String() string {
	return string(id)
}

// AppSlug is a validated, URL-safe application slug.
type AppSlug string

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// NewAppSlug creates and validates an AppSlug.
func NewAppSlug(slug string) (AppSlug, error) {
	if slug == "" {
		return "", fmt.Errorf("%w: slug cannot be empty", ErrInvalidAppSlug)
	}

	if len(slug) > 100 {
		return "", fmt.Errorf("%w: slug too long (max 100 chars)", ErrInvalidAppSlug)
	}

	if !slugRegex.MatchString(slug) {
		return "", fmt.Errorf("%w: must be lowercase alphanumeric with hyphens only", ErrInvalidAppSlug)
	}

	return AppSlug(slug), nil
}

// String returns the string representation.
func (s AppSlug) String() string {
	return string(s)
}

// Slugify converts a string to a valid slug format.
func Slugify(s string) AppSlug {
	// Convert to lowercase
	result := strings.ToLower(s)

	// Replace accented characters
	replacements := map[rune]string{
		'à': "a", 'á': "a", 'â': "a", 'ã': "a", 'ä': "a", 'å': "a",
		'è': "e", 'é': "e", 'ê': "e", 'ë': "e",
		'ì': "i", 'í': "i", 'î': "i", 'ï': "i",
		'ò': "o", 'ó': "o", 'ô': "o", 'õ': "o", 'ö': "o",
		'ù': "u", 'ú': "u", 'û': "u", 'ü': "u",
		'ý': "y", 'ÿ': "y",
		'ñ': "n", 'ç': "c",
	}

	var builder strings.Builder
	for _, r := range result {
		if replacement, ok := replacements[r]; ok {
			builder.WriteString(replacement)
		} else if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '_' || r == '.' {
			// Replace any separator with a hyphen
			builder.WriteRune('-')
		} else {
			// Replace other special characters with hyphen if they're between alphanumeric chars
			// Otherwise skip them
			if builder.Len() > 0 {
				lastChar := result[len(result)-1]
				if (lastChar >= 'a' && lastChar <= 'z') || (lastChar >= '0' && lastChar <= '9') {
					builder.WriteRune('-')
				}
			}
		}
	}

	// Clean up hyphens
	result = strings.Trim(builder.String(), "-")
	result = regexp.MustCompile(`-+`).ReplaceAllString(result, "-")

	if result == "" {
		result = "app"
	}

	// Already validated format, safe to cast
	return AppSlug(result)
}

// GitHubURL is a validated GitHub repository URL.
type GitHubURL string

var githubURLRegex = regexp.MustCompile(`^https://github\.com/[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+/?$`)

// NewGitHubURL creates and validates a GitHubURL.
func NewGitHubURL(urlStr string) (GitHubURL, error) {
	if urlStr == "" {
		return "", nil // Empty is allowed (nullable field)
	}

	// Parse URL
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidGitHubURL, err)
	}

	// Must be HTTPS
	if parsed.Scheme != "https" && parsed.Scheme != "" {
		return "", fmt.Errorf("%w: must use HTTPS", ErrInvalidGitHubURL)
	}

	// Must be github.com
	if parsed.Host != "github.com" && parsed.Host != "" {
		return "", fmt.Errorf("%w: must be github.com domain", ErrInvalidGitHubURL)
	}

	// Normalize URL
	normalized := fmt.Sprintf("https://github.com%s", parsed.Path)
	normalized = strings.TrimSuffix(normalized, "/")

	// Validate format
	if !githubURLRegex.MatchString(normalized) {
		return "", fmt.Errorf("%w: must be https://github.com/owner/repo", ErrInvalidGitHubURL)
	}

	return GitHubURL(normalized), nil
}

// String returns the string representation.
func (u GitHubURL) String() string {
	return string(u)
}

// OwnerAndRepo extracts the owner and repository name from the URL.
func (u GitHubURL) OwnerAndRepo() (owner, repo string, err error) {
	if u == "" {
		return "", "", fmt.Errorf("%w: empty URL", ErrInvalidGitHubURL)
	}

	parts := strings.Split(strings.TrimPrefix(string(u), "https://github.com/"), "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("%w: invalid format", ErrInvalidGitHubURL)
	}

	return parts[0], parts[1], nil
}

// Application represents an application that can have multiple instances.
type Application struct {
	ID        ApplicationID
	Slug      AppSlug
	Name      string
	GitHubURL GitHubURL
	Stars     int
	StarsUpdatedAt *time.Time
	LogoURL   string // Optional custom logo
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewApplication creates a new Application with validation.
func NewApplication(slug, name string) (*Application, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidApplication)
	}

	appSlug, err := NewAppSlug(slug)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &Application{
		Slug:      appSlug,
		Name:      name,
		Stars:     0,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// SetGitHubURL updates the GitHub repository URL.
func (a *Application) SetGitHubURL(urlStr string) error {
	githubURL, err := NewGitHubURL(urlStr)
	if err != nil {
		return err
	}
	a.GitHubURL = githubURL
	a.UpdatedAt = time.Now().UTC()
	return nil
}

// SetLogoURL updates the custom logo URL.
func (a *Application) SetLogoURL(logoURL string) {
	a.LogoURL = logoURL
	a.UpdatedAt = time.Now().UTC()
}

// UpdateStars updates the GitHub stars count and timestamp.
func (a *Application) UpdateStars(stars int) {
	if stars < 0 {
		stars = 0
	}
	a.Stars = stars
	now := time.Now().UTC()
	a.StarsUpdatedAt = &now
	a.UpdatedAt = now
}

// NeedsStarsRefresh returns true if stars data is stale (older than 1 hour).
func (a *Application) NeedsStarsRefresh() bool {
	if a.GitHubURL == "" {
		return false // No GitHub URL, nothing to refresh
	}
	if a.StarsUpdatedAt == nil {
		return true // Never fetched
	}
	return time.Since(*a.StarsUpdatedAt) > time.Hour
}
