package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/Okabe-Junya/golink-backend/models"
	"github.com/Okabe-Junya/golink-backend/repositories/mocks"
	"github.com/stretchr/testify/assert"
)

// createTestLink creates a test link
func createTestLink(short, url, userID string) *models.Link {
	return &models.Link{
		ID:        short,
		Short:     short,
		URL:       url,
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestMockRepositoryImplementation(t *testing.T) {
	mockRepo := mocks.NewMockLinkRepository()
	ctx := context.Background()

	// Test Create
	link := createTestLink("test", "https://example.com", "user1")
	err := mockRepo.Create(ctx, link)
	assert.NoError(t, err)

	// Test Create with duplicate short code
	duplicateLink := createTestLink("test", "https://another.com", "user2")
	err = mockRepo.Create(ctx, duplicateLink)
	assert.Error(t, err, "Should not allow duplicate short codes")

	// Test GetByShort
	retrievedLink, err := mockRepo.GetByShort(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, link.Short, retrievedLink.Short)
	assert.Equal(t, link.URL, retrievedLink.URL)

	// Test GetByShort with non-existent link
	_, err = mockRepo.GetByShort(ctx, "nonexistent")
	assert.Error(t, err, "Should return error for non-existent link")

	// Test GetAll
	links, err := mockRepo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, links, 1)

	// Test Update
	link.URL = "https://updated.com"
	err = mockRepo.Update(ctx, link)
	assert.NoError(t, err)

	// Test Update non-existent link
	nonExistentLink := createTestLink("nonexistent", "https://test.com", "user1")
	err = mockRepo.Update(ctx, nonExistentLink)
	assert.Error(t, err, "Should return error when updating non-existent link")

	// Test IncrementClickCount
	err = mockRepo.IncrementClickCount(ctx, "test")
	assert.NoError(t, err)

	// Test IncrementClickCount for non-existent link
	err = mockRepo.IncrementClickCount(ctx, "nonexistent")
	assert.Error(t, err, "Should return error when incrementing click count for non-existent link")

	// Verify click count
	retrievedLink, err = mockRepo.GetByShort(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, 1, retrievedLink.ClickCount)

	// Test Delete
	err = mockRepo.Delete(ctx, "test")
	assert.NoError(t, err)

	// Test Delete non-existent link
	err = mockRepo.Delete(ctx, "nonexistent")
	assert.Error(t, err, "Should return error when deleting non-existent link")

	// Verify deletion
	_, err = mockRepo.GetByShort(ctx, "test")
	assert.Error(t, err)
}

func TestAccessControl(t *testing.T) {
	mockRepo := mocks.NewMockLinkRepository()
	ctx := context.Background()

	// Create test links with different access levels
	publicLink := createTestLink("public", "https://public.com", "user1")
	publicLink.AccessLevel = models.AccessLevels.Public

	privateLink := createTestLink("private", "https://private.com", "user1")
	privateLink.AccessLevel = models.AccessLevels.Private

	restrictedLink := createTestLink("restricted", "https://restricted.com", "user1")
	restrictedLink.AccessLevel = models.AccessLevels.Restricted
	restrictedLink.AllowedUsers = []string{"user1", "user2"}

	// Add links to repository
	assert.NoError(t, mockRepo.Create(ctx, publicLink))
	assert.NoError(t, mockRepo.Create(ctx, privateLink))
	assert.NoError(t, mockRepo.Create(ctx, restrictedLink))

	tests := []struct {
		name          string
		linkID        string
		userID        string
		expectAccess  bool
		expectedCount int
	}{
		{
			name:         "Public Link - Anonymous User",
			linkID:       "public",
			userID:       "",
			expectAccess: true,
		},
		{
			name:         "Private Link - Owner",
			linkID:       "private",
			userID:       "user1",
			expectAccess: true,
		},
		{
			name:         "Private Link - Other User",
			linkID:       "private",
			userID:       "user2",
			expectAccess: false,
		},
		{
			name:         "Restricted Link - Allowed User",
			linkID:       "restricted",
			userID:       "user2",
			expectAccess: true,
		},
		{
			name:         "Restricted Link - Not Allowed User",
			linkID:       "restricted",
			userID:       "user3",
			expectAccess: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hasAccess, err := mockRepo.CheckAccess(ctx, tc.linkID, tc.userID)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectAccess, hasAccess)
		})
	}

	// Test GetByAccessLevel
	publicLinks, err := mockRepo.GetByAccessLevel(ctx, models.AccessLevels.Public)
	assert.NoError(t, err)
	assert.NotEmpty(t, publicLinks)
	assert.Equal(t, "public", publicLinks[0].Short)

	privateLinks, err := mockRepo.GetByAccessLevel(ctx, models.AccessLevels.Private)
	assert.NoError(t, err)
	assert.NotEmpty(t, privateLinks)
	assert.Equal(t, "private", privateLinks[0].Short)

	// Test GetByUser
	user1Links, err := mockRepo.GetByUser(ctx, "user1")
	assert.NoError(t, err)
	assert.Len(t, user1Links, 3)

	user2Links, err := mockRepo.GetByUser(ctx, "user2")
	assert.NoError(t, err)
	assert.Empty(t, user2Links)
}

func TestBoundaryValues(t *testing.T) {
	mockRepo := mocks.NewMockLinkRepository()
	ctx := context.Background()

	tests := []struct {
		name      string
		link      *models.Link
		expectErr bool
	}{
		{
			name:      "Nil Link",
			link:      nil,
			expectErr: true,
		},
		{
			name: "Empty Short Code",
			link: &models.Link{
				Short:     "",
				URL:       "https://example.com",
				CreatedBy: "user1",
			},
			expectErr: true,
		},
		{
			name: "Empty URL",
			link: &models.Link{
				Short:     "test",
				URL:       "",
				CreatedBy: "user1",
			},
			expectErr: true,
		},
		{
			name: "Empty User ID",
			link: &models.Link{
				Short:     "test",
				URL:       "https://example.com",
				CreatedBy: "",
			},
			expectErr: true,
		},
		{
			name: "Invalid Access Level",
			link: &models.Link{
				Short:       "test",
				URL:         "https://example.com",
				CreatedBy:   "user1",
				AccessLevel: "invalid",
			},
			expectErr: true,
		},
		{
			name: "Invalid URL Format",
			link: &models.Link{
				Short:     "test",
				URL:       "not-a-url",
				CreatedBy: "user1",
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := mockRepo.Create(ctx, tc.link)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
