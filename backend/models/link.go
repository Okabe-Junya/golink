package models

import (
	"time"
)

// Link represents a shortened URL with access control information
type Link struct {
	ID           string    `json:"id" firestore:"id"`                       // Unique identifier (same as Short)
	Short        string    `json:"short" firestore:"short"`                 // Short code for the URL
	URL          string    `json:"url" firestore:"url"`                     // Original URL
	CreatedAt    time.Time `json:"created_at" firestore:"created_at"`       // Creation timestamp
	UpdatedAt    time.Time `json:"updated_at" firestore:"updated_at"`       // Last update timestamp
	CreatedBy    string    `json:"created_by" firestore:"created_by"`       // User who created the link
	AccessLevel  string    `json:"access_level" firestore:"access_level"`   // Public, Private, or Restricted
	AllowedUsers []string  `json:"allowed_users" firestore:"allowed_users"` // List of users allowed to access (for Restricted)
	ClickCount   int       `json:"click_count" firestore:"click_count"`     // Number of times the link has been accessed
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
