package models_test

import (
	"testing"
	"time"

	"github.com/Okabe-Junya/golink-backend/models"
	"github.com/stretchr/testify/assert"
)

func TestNewLink(t *testing.T) {
	// Test creating a new link
	short := "test-link"
	url := "https://example.com"
	userID := "user123"

	link := models.NewLink(short, url, userID)

	// Verify all fields are set correctly
	assert.Equal(t, short, link.ID)
	assert.Equal(t, short, link.Short)
	assert.Equal(t, url, link.URL)
	assert.Equal(t, userID, link.CreatedBy)
	assert.Equal(t, models.AccessLevels.Public, link.AccessLevel)
	assert.Empty(t, link.AllowedUsers)
	assert.Equal(t, 0, link.ClickCount)

	// Verify timestamps are set
	assert.WithinDuration(t, time.Now(), link.CreatedAt, 2*time.Second)
	assert.WithinDuration(t, time.Now(), link.UpdatedAt, 2*time.Second)
}

func TestAccessLevels(t *testing.T) {
	// Test that access levels are defined correctly
	assert.Equal(t, "Public", models.AccessLevels.Public)
	assert.Equal(t, "Private", models.AccessLevels.Private)
	assert.Equal(t, "Restricted", models.AccessLevels.Restricted)
}

func TestLinkFields(t *testing.T) {
	// Test that all fields are properly defined with the correct tags
	link := &models.Link{
		ID:           "test-id",
		Short:        "test-short",
		URL:          "https://example.com",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		CreatedBy:    "user123",
		AccessLevel:  models.AccessLevels.Private,
		AllowedUsers: []string{"user456", "user789"},
		ClickCount:   42,
	}

	// Verify field values
	assert.Equal(t, "test-id", link.ID)
	assert.Equal(t, "test-short", link.Short)
	assert.Equal(t, "https://example.com", link.URL)
	assert.Equal(t, "user123", link.CreatedBy)
	assert.Equal(t, models.AccessLevels.Private, link.AccessLevel)
	assert.Equal(t, []string{"user456", "user789"}, link.AllowedUsers)
	assert.Equal(t, 42, link.ClickCount)
}
