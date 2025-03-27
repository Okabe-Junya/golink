package repositories

import (
	"context"

	"github.com/Okabe-Junya/golink-backend/models"
)

// LinkRepositoryInterface defines the interface for link repository operations
// All methods return either the expected result or an error
// Errors should be created using the errors package from pkg/errors
type LinkRepositoryInterface interface {
	// Create creates a new link
	Create(ctx context.Context, link *models.Link) error

	// GetByShort retrieves a link by its short code
	GetByShort(ctx context.Context, short string) (*models.Link, error)

	// GetAll retrieves all links
	GetAll(ctx context.Context) ([]*models.Link, error)

	// Update updates an existing link
	Update(ctx context.Context, link *models.Link) error

	// Delete removes a link by its short code
	Delete(ctx context.Context, short string) error

	// IncrementClickCount increments the click count for a link
	IncrementClickCount(ctx context.Context, short string) error

	// GetByAccessLevel retrieves links by access level
	GetByAccessLevel(ctx context.Context, accessLevel string) ([]*models.Link, error)

	// GetByUser retrieves links created by a specific user
	GetByUser(ctx context.Context, userID string) ([]*models.Link, error)

	// CheckAccess determines if a user has access to a link
	CheckAccess(ctx context.Context, short string, userID string) (bool, error)

	// GetExpiredLinks retrieves all expired links
	GetExpiredLinks(ctx context.Context) ([]*models.Link, error)

	// GetLinksByExpiryStatus retrieves links by their expiry status
	GetLinksByExpiryStatus(ctx context.Context, isExpired bool) ([]*models.Link, error)

	// GetLinkStats retrieves statistics for a link
	GetLinkStats(ctx context.Context, short string) (*models.LinkStats, error)
}
