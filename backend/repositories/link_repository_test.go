package repositories_test

import (
	"context"
	"testing"

	"github.com/Okabe-Junya/golink-backend/models"
	"github.com/Okabe-Junya/golink-backend/repositories/mocks"
	"github.com/stretchr/testify/assert"
)

// createTestLink creates a test link
func createTestLink(short, url, userID string) *models.Link {
	return models.NewLink(short, url, userID)
}

func TestMockRepositoryImplementation(t *testing.T) {
	// This test verifies that our mock repository works as expected
	mockRepo := mocks.NewMockLinkRepository()
	ctx := context.Background()

	// Create a test link
	link := createTestLink("test", "https://example.com", "user1")

	// Test Create
	err := mockRepo.Create(ctx, link)
	assert.NoError(t, err)

	// Test GetByShort
	retrievedLink, err := mockRepo.GetByShort(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, link.Short, retrievedLink.Short)
	assert.Equal(t, link.URL, retrievedLink.URL)

	// Test GetAll
	links, err := mockRepo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, links, 1)

	// Test Update
	link.URL = "https://updated.com"
	err = mockRepo.Update(ctx, link)
	assert.NoError(t, err)

	// Verify update
	retrievedLink, err = mockRepo.GetByShort(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, "https://updated.com", retrievedLink.URL)

	// Test IncrementClickCount
	err = mockRepo.IncrementClickCount(ctx, "test")
	assert.NoError(t, err)

	// Verify click count
	retrievedLink, err = mockRepo.GetByShort(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, 1, retrievedLink.ClickCount)

	// Test GetByAccessLevel
	links, err = mockRepo.GetByAccessLevel(ctx, models.AccessLevels.Public)
	assert.NoError(t, err)
	assert.Len(t, links, 1)

	// Test GetByUser
	links, err = mockRepo.GetByUser(ctx, "user1")
	assert.NoError(t, err)
	assert.Len(t, links, 1)

	// Test CheckAccess
	hasAccess, err := mockRepo.CheckAccess(ctx, "test", "user1")
	assert.NoError(t, err)
	assert.True(t, hasAccess)

	// Test Delete
	err = mockRepo.Delete(ctx, "test")
	assert.NoError(t, err)

	// Verify deletion
	_, err = mockRepo.GetByShort(ctx, "test")
	assert.Error(t, err)
}

func TestAccessControl(t *testing.T) {
	mockRepo := mocks.NewMockLinkRepository()
	ctx := context.Background()

	// Create a public link
	publicLink := createTestLink("public", "https://public.com", "user1")
	publicLink.AccessLevel = models.AccessLevels.Public
	mockRepo.Create(ctx, publicLink)

	// Create a private link
	privateLink := createTestLink("private", "https://private.com", "user1")
	privateLink.AccessLevel = models.AccessLevels.Private
	mockRepo.Create(ctx, privateLink)

	// Create a restricted link
	restrictedLink := createTestLink("restricted", "https://restricted.com", "user1")
	restrictedLink.AccessLevel = models.AccessLevels.Restricted
	restrictedLink.AllowedUsers = []string{"user2"}
	mockRepo.Create(ctx, restrictedLink)

	// Test public link access
	hasAccess, err := mockRepo.CheckAccess(ctx, "public", "user2")
	assert.NoError(t, err)
	assert.True(t, hasAccess, "Anyone should have access to public links")

	// Test private link access - owner
	hasAccess, err = mockRepo.CheckAccess(ctx, "private", "user1")
	assert.NoError(t, err)
	assert.True(t, hasAccess, "Owner should have access to private links")

	// Test private link access - non-owner
	hasAccess, err = mockRepo.CheckAccess(ctx, "private", "user2")
	assert.NoError(t, err)
	assert.False(t, hasAccess, "Non-owner should not have access to private links")

	// Test restricted link access - owner
	hasAccess, err = mockRepo.CheckAccess(ctx, "restricted", "user1")
	assert.NoError(t, err)
	assert.True(t, hasAccess, "Owner should have access to restricted links")

	// Test restricted link access - allowed user
	hasAccess, err = mockRepo.CheckAccess(ctx, "restricted", "user2")
	assert.NoError(t, err)
	assert.True(t, hasAccess, "Allowed user should have access to restricted links")

	// Test restricted link access - non-allowed user
	hasAccess, err = mockRepo.CheckAccess(ctx, "restricted", "user3")
	assert.NoError(t, err)
	assert.False(t, hasAccess, "Non-allowed user should not have access to restricted links")
}
