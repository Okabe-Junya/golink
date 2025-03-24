package models

import (
	"time"
)

// Link represents a shortened URL with access control information
type Link struct {
	CreatedAt    time.Time `json:"created_at" firestore:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" firestore:"updated_at"`
	ExpiresAt    time.Time `json:"expires_at,omitempty" firestore:"expires_at,omitempty"`
	ID           string    `json:"id" firestore:"id"`
	Short        string    `json:"short" firestore:"short"`
	URL          string    `json:"url" firestore:"url"`
	CreatedBy    string    `json:"created_by" firestore:"created_by"`
	AccessLevel  string    `json:"access_level" firestore:"access_level"`
	AllowedUsers []string  `json:"allowed_users" firestore:"allowed_users"`
	ClickCount   int       `json:"click_count" firestore:"click_count"`
	IsExpired    bool      `json:"is_expired" firestore:"is_expired"`
}

// NewLink creates a new Link with default values
func NewLink(short, url, createdBy string) *Link {
	now := time.Now()
	return &Link{
		ID:           short,
		Short:        short,
		URL:          url,
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    createdBy,
		AccessLevel:  "Public", // Default to public access
		AllowedUsers: []string{},
		ClickCount:   0,
		IsExpired:    false, // Default to not expired
	}
}

// SetExpiry sets the expiration time for a link
func (l *Link) SetExpiry(expires time.Time) {
	l.ExpiresAt = expires
	l.UpdatedAt = time.Now()
}

// IsLinkExpired checks if a link is expired
func (l *Link) IsLinkExpired() bool {
	// If ExpiresAt is zero, the link never expires
	if l.ExpiresAt.IsZero() {
		return false
	}
	// Check if current time is past the expiration time
	return time.Now().After(l.ExpiresAt)
}

// IsExpiringOrExpired checks if a link is expired or will expire soon
func (l *Link) IsExpiringOrExpired() (bool, string) {
	if l.ExpiresAt.IsZero() {
		return false, ""
	}

	if l.IsExpired {
		return true, "expired"
	}

	now := time.Now()
	daysUntilExpiry := l.ExpiresAt.Sub(now).Hours() / 24

	switch {
	case daysUntilExpiry < 0:
		return true, "expired"
	case daysUntilExpiry < 1:
		return true, "expiring_today"
	case daysUntilExpiry < 7:
		return true, "expiring_soon"
	default:
		return false, ""
	}
}

// AccessLevels defines the possible access levels for a link
var AccessLevels = struct {
	Public     string
	Private    string
	Restricted string
}{
	Public:     "Public",     // Anyone can access
	Private:    "Private",    // Only the creator can access
	Restricted: "Restricted", // Only specific users can access
}
