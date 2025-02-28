package repositories

import (
	"context"

	"github.com/Okabe-Junya/golink-backend/models"
)

// LinkRepositoryInterface defines the interface for link repository operations
type LinkRepositoryInterface interface {
	Create(ctx context.Context, link *models.Link) error
	GetByShort(ctx context.Context, short string) (*models.Link, error)
	GetAll(ctx context.Context) ([]*models.Link, error)
	Update(ctx context.Context, link *models.Link) error
	Delete(ctx context.Context, short string) error
	IncrementClickCount(ctx context.Context, short string) error
	GetByAccessLevel(ctx context.Context, accessLevel string) ([]*models.Link, error)
	GetByUser(ctx context.Context, userID string) ([]*models.Link, error)
	CheckAccess(ctx context.Context, short string, userID string) (bool, error)
}
