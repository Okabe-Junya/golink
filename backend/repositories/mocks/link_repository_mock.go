package mocks

import (
	"context"
	"errors"
	"time"

	"github.com/Okabe-Junya/golink-backend/interfaces"
	"github.com/Okabe-Junya/golink-backend/models"
)

// Ensure MockLinkRepository implements LinkRepositoryInterface
var _ interfaces.LinkRepositoryInterface = (*MockLinkRepository)(nil)

// MockLinkRepository is a mock implementation of the LinkRepository
type MockLinkRepository struct {
	links map[string]*models.Link
}

// NewMockLinkRepository creates a new mock link repository
func NewMockLinkRepository() *MockLinkRepository {
	return &MockLinkRepository{
		links: make(map[string]*models.Link),
	}
}

// Create adds a new link to the mock repository
func (m *MockLinkRepository) Create(ctx context.Context, link *models.Link) error {
	if _, exists := m.links[link.Short]; exists {
		return errors.New("link already exists")
	}
	m.links[link.Short] = link
	return nil
}

// GetByShort retrieves a link by its short code
func (m *MockLinkRepository) GetByShort(ctx context.Context, short string) (*models.Link, error) {
	link, exists := m.links[short]
	if !exists {
		return nil, errors.New("link not found")
	}
	return link, nil
}

// GetAll retrieves all links
func (m *MockLinkRepository) GetAll(ctx context.Context) ([]*models.Link, error) {
	var links []*models.Link
	for _, link := range m.links {
		links = append(links, link)
	}
	return links, nil
}

// Update updates an existing link
func (m *MockLinkRepository) Update(ctx context.Context, link *models.Link) error {
	if _, exists := m.links[link.Short]; !exists {
		return errors.New("link not found")
	}
	link.UpdatedAt = time.Now()
	m.links[link.Short] = link
	return nil
}

// Delete removes a link by its short code
func (m *MockLinkRepository) Delete(ctx context.Context, short string) error {
	if _, exists := m.links[short]; !exists {
		return errors.New("link not found")
	}
	delete(m.links, short)
	return nil
}

// IncrementClickCount increments the click count for a link
func (m *MockLinkRepository) IncrementClickCount(ctx context.Context, short string) error {
	link, exists := m.links[short]
	if !exists {
		return errors.New("link not found")
	}
	link.ClickCount++
	link.UpdatedAt = time.Now()
	return nil
}

// GetByAccessLevel retrieves links by access level
func (m *MockLinkRepository) GetByAccessLevel(ctx context.Context, accessLevel string) ([]*models.Link, error) {
	var links []*models.Link
	for _, link := range m.links {
		if link.AccessLevel == accessLevel {
			links = append(links, link)
		}
	}
	return links, nil
}

// GetByUser retrieves links created by a specific user
func (m *MockLinkRepository) GetByUser(ctx context.Context, userID string) ([]*models.Link, error) {
	var links []*models.Link
	for _, link := range m.links {
		if link.CreatedBy == userID {
			links = append(links, link)
		}
	}
	return links, nil
}

// CheckAccess determines if a user has access to a link
func (m *MockLinkRepository) CheckAccess(ctx context.Context, short string, userID string) (bool, error) {
	link, exists := m.links[short]
	if !exists {
		return false, errors.New("link not found")
	}

	// Public links are accessible to everyone
	if link.AccessLevel == models.AccessLevels.Public {
		return true, nil
	}

	// Private links are only accessible to the creator
	if link.AccessLevel == models.AccessLevels.Private {
		return link.CreatedBy == userID, nil
	}

	// Restricted links are accessible to the creator and allowed users
	if link.AccessLevel == models.AccessLevels.Restricted {
		if link.CreatedBy == userID {
			return true, nil
		}

		// Check if the user is in the allowed users list
		for _, allowedUser := range link.AllowedUsers {
			if allowedUser == userID {
				return true, nil
			}
		}
	}

	return false, nil
}
