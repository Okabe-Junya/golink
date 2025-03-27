package repositories

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/Okabe-Junya/golink-backend/interfaces"
	"github.com/Okabe-Junya/golink-backend/models"
	"github.com/Okabe-Junya/golink-backend/pkg/errors"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LinkRepository handles database operations for links
type LinkRepository struct {
	client     *firestore.Client
	collection string
}

// Ensure LinkRepository implements LinkRepositoryInterface
var _ interfaces.LinkRepositoryInterface = (*LinkRepository)(nil)

// NewLinkRepository creates a new LinkRepository
func NewLinkRepository(client *firestore.Client) *LinkRepository {
	return &LinkRepository{
		client:     client,
		collection: "links",
	}
}

// Create adds a new link to the database
func (r *LinkRepository) Create(ctx context.Context, link *models.Link) error {
	// Check if the link already exists
	existingLink, err := r.GetByShort(ctx, link.Short)
	if err == nil && existingLink != nil {
		return errors.NewAlreadyExists(fmt.Sprintf("Link '%s' already exists", link.Short))
	}

	// Only proceed if the error is a "not found" error
	if err != nil {
		if !errors.Is(err, errors.ErrNotFound) {
			return errors.Wrap(err, "Error checking if link exists")
		}
	}

	// Set the timestamps
	now := time.Now()
	link.CreatedAt = now
	link.UpdatedAt = now

	// Create the link
	_, err = r.client.Collection(r.collection).Doc(link.Short).Set(ctx, link)
	if err != nil {
		return errors.NewInternalError(fmt.Errorf("Error creating link: %w", err))
	}

	return nil
}

// GetByShort retrieves a link by its short code
func (r *LinkRepository) GetByShort(ctx context.Context, short string) (*models.Link, error) {
	doc, err := r.client.Collection(r.collection).Doc(short).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, errors.NewNotFound(fmt.Sprintf("Link '%s' not found", short))
		}
		return nil, errors.NewInternalError(fmt.Errorf("Error retrieving link: %w", err))
	}

	var link models.Link
	if err := doc.DataTo(&link); err != nil {
		return nil, errors.NewInternalError(fmt.Errorf("Error converting link data: %w", err))
	}

	// Update expiry status if needed
	if !link.ExpiresAt.IsZero() && time.Now().After(link.ExpiresAt) && !link.IsExpired {
		link.IsExpired = true
		// Update the link in the background, don't block the response
		go func() {
			// Create a new context for the background operation
			bgCtx := context.Background()
			if err := r.Update(bgCtx, &link); err != nil {
				// We're in a goroutine, so we can't return the error
				// Ideally, we would log this error
			}
		}()
	}

	return &link, nil
}

// GetAll retrieves all links
func (r *LinkRepository) GetAll(ctx context.Context) ([]*models.Link, error) {
	iter := r.client.Collection(r.collection).Documents(ctx)
	var links []*models.Link

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.NewInternalError(fmt.Errorf("Error retrieving links: %w", err))
		}

		var link models.Link
		if err := doc.DataTo(&link); err != nil {
			// Log error but continue with next document
			continue
		}
		links = append(links, &link)
	}

	return links, nil
}

// Update updates an existing link
func (r *LinkRepository) Update(ctx context.Context, link *models.Link) error {
	// Check if the link exists
	_, err := r.GetByShort(ctx, link.Short)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return errors.NewNotFound(fmt.Sprintf("Link '%s' not found", link.Short))
		}
		return errors.Wrap(err, "Error checking if link exists")
	}

	// Update the timestamp
	link.UpdatedAt = time.Now()

	// Update the link
	_, err = r.client.Collection(r.collection).Doc(link.Short).Set(ctx, link)
	if err != nil {
		return errors.NewInternalError(fmt.Errorf("Error updating link: %w", err))
	}

	return nil
}

// Delete removes a link by its short code
func (r *LinkRepository) Delete(ctx context.Context, short string) error {
	// Check if the link exists
	_, err := r.GetByShort(ctx, short)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return errors.NewNotFound(fmt.Sprintf("Link '%s' not found", short))
		}
		return errors.Wrap(err, "Error checking if link exists")
	}

	// Delete the link
	_, err = r.client.Collection(r.collection).Doc(short).Delete(ctx)
	if err != nil {
		return errors.NewInternalError(fmt.Errorf("Error deleting link: %w", err))
	}

	return nil
}

// IncrementClickCount increments the click count for a link
func (r *LinkRepository) IncrementClickCount(ctx context.Context, short string) error {
	// Get the link
	link, err := r.GetByShort(ctx, short)
	if err != nil {
		return err // Already wrapped by GetByShort
	}

	// Increment the click count
	link.ClickCount++
	link.UpdatedAt = time.Now()

	// Update the link
	_, err = r.client.Collection(r.collection).Doc(short).Set(ctx, link)
	if err != nil {
		return errors.NewInternalError(fmt.Errorf("Error updating click count: %w", err))
	}

	return nil
}

// GetByAccessLevel retrieves links by access level
func (r *LinkRepository) GetByAccessLevel(ctx context.Context, accessLevel string) ([]*models.Link, error) {
	query := r.client.Collection(r.collection).Where("access_level", "==", accessLevel)
	iter := query.Documents(ctx)
	var links []*models.Link

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.NewInternalError(fmt.Errorf("Error retrieving links by access level: %w", err))
		}

		var link models.Link
		if err := doc.DataTo(&link); err != nil {
			// Log error but continue with next document
			continue
		}
		links = append(links, &link)
	}

	return links, nil
}

// GetByUser retrieves links created by a specific user
func (r *LinkRepository) GetByUser(ctx context.Context, userID string) ([]*models.Link, error) {
	query := r.client.Collection(r.collection).Where("created_by", "==", userID)
	iter := query.Documents(ctx)
	var links []*models.Link

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.NewInternalError(fmt.Errorf("Error retrieving links by user: %w", err))
		}

		var link models.Link
		if err := doc.DataTo(&link); err != nil {
			// Log error but continue with next document
			continue
		}
		links = append(links, &link)
	}

	return links, nil
}

// CheckAccess determines if a user has access to a link
func (r *LinkRepository) CheckAccess(ctx context.Context, short string, userID string) (bool, error) {
	link, err := r.GetByShort(ctx, short)
	if err != nil {
		return false, err // Already wrapped by GetByShort
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

// GetExpiredLinks retrieves all expired links
func (r *LinkRepository) GetExpiredLinks(ctx context.Context) ([]*models.Link, error) {
	now := time.Now()
	query := r.client.Collection(r.collection).Where("expires_at", "<", now).Where("is_expired", "==", false)
	iter := query.Documents(ctx)
	var links []*models.Link

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.NewInternalError(fmt.Errorf("Error retrieving expired links: %w", err))
		}

		var link models.Link
		if err := doc.DataTo(&link); err != nil {
			// Log error but continue with next document
			continue
		}
		links = append(links, &link)
	}

	return links, nil
}

// GetLinksByExpiryStatus retrieves links by their expiry status
func (r *LinkRepository) GetLinksByExpiryStatus(ctx context.Context, isExpired bool) ([]*models.Link, error) {
	query := r.client.Collection(r.collection).Where("is_expired", "==", isExpired)
	iter := query.Documents(ctx)
	var links []*models.Link

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.NewInternalError(fmt.Errorf("Error retrieving links by expiry status: %w", err))
		}

		var link models.Link
		if err := doc.DataTo(&link); err != nil {
			// Log error but continue with next document
			continue
		}
		links = append(links, &link)
	}

	return links, nil
}

// GetLinkStats retrieves statistics for a link
func (r *LinkRepository) GetLinkStats(ctx context.Context, short string) (*models.LinkStats, error) {
	// Check if the link exists
	link, err := r.GetByShort(ctx, short)
	if err != nil {
		return nil, err // Already wrapped by GetByShort
	}

	// Check if stats document exists
	statsDoc, err := r.client.Collection("link_stats").Doc(short).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			// Create new stats if not found
			stats := models.NewLinkStats(short)
			_, err = r.client.Collection("link_stats").Doc(short).Set(ctx, stats)
			if err != nil {
				return nil, errors.NewInternalError(fmt.Errorf("Error creating link stats: %w", err))
			}
			return stats, nil
		}
		return nil, errors.NewInternalError(fmt.Errorf("Error retrieving link stats: %w", err))
	}

	// Parse stats document
	var stats models.LinkStats
	if err := statsDoc.DataTo(&stats); err != nil {
		return nil, errors.NewInternalError(fmt.Errorf("Error converting link stats data: %w", err))
	}

	// Update total clicks from link
	stats.TotalClicks = link.ClickCount

	return &stats, nil
}
