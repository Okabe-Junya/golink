package mocks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Okabe-Junya/golink-backend/models"
	"github.com/Okabe-Junya/golink-backend/pkg/errors"
)

// MockLinkRepository is a mock implementation of the LinkRepository interface
type MockLinkRepository struct {
	// links is a map of short URLs to link models
	links map[string]*models.Link
	// mutex to protect concurrent access to the links map
	mutex sync.RWMutex
}

// NewMockLinkRepository creates a new instance of MockLinkRepository
func NewMockLinkRepository() *MockLinkRepository {
	return &MockLinkRepository{
		links: make(map[string]*models.Link),
	}
}

// Create adds a new link to the mock repository
func (r *MockLinkRepository) Create(ctx context.Context, link *models.Link) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if the link already exists
	if _, exists := r.links[link.Short]; exists {
		return errors.NewAlreadyExists(fmt.Sprintf("Link '%s' already exists", link.Short))
	}

	// Set the timestamps
	now := time.Now()
	link.CreatedAt = now
	link.UpdatedAt = now

	// Store a copy of the link
	linkCopy := *link
	r.links[link.Short] = &linkCopy

	return nil
}

// GetByShort retrieves a link by its short code
func (r *MockLinkRepository) GetByShort(ctx context.Context, short string) (*models.Link, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	link, exists := r.links[short]
	if !exists {
		return nil, errors.NewNotFound(fmt.Sprintf("Link '%s' not found", short))
	}

	// Make a copy of the link to prevent mutation
	linkCopy := *link

	// Update expiry status if needed
	if !linkCopy.ExpiresAt.IsZero() && time.Now().After(linkCopy.ExpiresAt) && !linkCopy.IsExpired {
		linkCopy.IsExpired = true

		// Use a goroutine to avoid blocking while holding the lock
		go func(shortURL string) {
			link, err := r.GetByShort(context.Background(), shortURL)
			if err == nil {
				link.IsExpired = true
				if err := r.Update(context.Background(), link); err != nil {
					// We're in a goroutine, so just log the error or handle it appropriately
					// For simplicity in tests, we'll ignore it
				}
			}
		}(short)
	}

	return &linkCopy, nil
}

// GetAll retrieves all links
func (r *MockLinkRepository) GetAll(ctx context.Context) ([]*models.Link, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	links := make([]*models.Link, 0, len(r.links))
	for _, link := range r.links {
		linkCopy := *link
		links = append(links, &linkCopy)
	}

	return links, nil
}

// Update updates an existing link
func (r *MockLinkRepository) Update(ctx context.Context, link *models.Link) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if the link exists
	if _, exists := r.links[link.Short]; !exists {
		return errors.NewNotFound(fmt.Sprintf("Link '%s' not found", link.Short))
	}

	// Update the timestamp
	link.UpdatedAt = time.Now()

	// Store a copy of the link
	linkCopy := *link
	r.links[link.Short] = &linkCopy

	return nil
}

// Delete removes a link by its short code
func (r *MockLinkRepository) Delete(ctx context.Context, short string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if the link exists
	if _, exists := r.links[short]; !exists {
		return errors.NewNotFound(fmt.Sprintf("Link '%s' not found", short))
	}

	// Delete the link
	delete(r.links, short)

	return nil
}

// IncrementClickCount increments the click count for a link
func (r *MockLinkRepository) IncrementClickCount(ctx context.Context, short string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if the link exists
	link, exists := r.links[short]
	if !exists {
		return errors.NewNotFound(fmt.Sprintf("Link '%s' not found", short))
	}

	// Increment the click count
	link.ClickCount++
	link.UpdatedAt = time.Now()

	return nil
}

// GetExpiredLinks retrieves all expired links
func (r *MockLinkRepository) GetExpiredLinks(ctx context.Context) ([]*models.Link, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	now := time.Now()
	var expiredLinks []*models.Link

	for _, link := range r.links {
		if !link.ExpiresAt.IsZero() && link.ExpiresAt.Before(now) && !link.IsExpired {
			linkCopy := *link
			expiredLinks = append(expiredLinks, &linkCopy)
		}
	}

	return expiredLinks, nil
}

// GetLinksByExpiryStatus retrieves links by their expiry status
func (r *MockLinkRepository) GetLinksByExpiryStatus(ctx context.Context, isExpired bool) ([]*models.Link, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var filteredLinks []*models.Link

	for _, link := range r.links {
		if link.IsExpired == isExpired {
			linkCopy := *link
			filteredLinks = append(filteredLinks, &linkCopy)
		}
	}

	return filteredLinks, nil
}

// GetLinkStats retrieves statistics for a link
func (r *MockLinkRepository) GetLinkStats(ctx context.Context, short string) (*models.LinkStats, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Check if the link exists
	link, exists := r.links[short]
	if !exists {
		return nil, errors.NewNotFound(fmt.Sprintf("Link '%s' not found", short))
	}

	// Create stats from link
	stats := models.NewLinkStats(short)
	stats.TotalClicks = link.ClickCount

	return stats, nil
}
