package repositories

import (
	"context"
	"errors"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/Okabe-Junya/golink-backend/interfaces"
	"github.com/Okabe-Junya/golink-backend/models"
	"google.golang.org/api/iterator"
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
	_, err := r.client.Collection(r.collection).Doc(link.Short).Set(ctx, link)
	return err
}

// GetByShort retrieves a link by its short code
func (r *LinkRepository) GetByShort(ctx context.Context, short string) (*models.Link, error) {
	doc, err := r.client.Collection(r.collection).Doc(short).Get(ctx)
	if err != nil {
		return nil, err
	}

	var link models.Link
	if err := doc.DataTo(&link); err != nil {
		return nil, err
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
			return nil, err
		}

		var link models.Link
		if err := doc.DataTo(&link); err != nil {
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
		return errors.New("link not found")
	}

	// Update the timestamp
	link.UpdatedAt = time.Now()

	// Update the link
	_, err = r.client.Collection(r.collection).Doc(link.Short).Set(ctx, link)
	return err
}

// Delete removes a link by its short code
func (r *LinkRepository) Delete(ctx context.Context, short string) error {
	// Check if the link exists
	_, err := r.GetByShort(ctx, short)
	if err != nil {
		return errors.New("link not found")
	}

	// Delete the link
	_, err = r.client.Collection(r.collection).Doc(short).Delete(ctx)
	return err
}

// IncrementClickCount increments the click count for a link
func (r *LinkRepository) IncrementClickCount(ctx context.Context, short string) error {
	// Get the link
	link, err := r.GetByShort(ctx, short)
	if err != nil {
		return err
	}

	// Increment the click count
	link.ClickCount++
	link.UpdatedAt = time.Now()

	// Update the link
	_, err = r.client.Collection(r.collection).Doc(short).Set(ctx, link)
	return err
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
			return nil, err
		}

		var link models.Link
		if err := doc.DataTo(&link); err != nil {
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
			return nil, err
		}

		var link models.Link
		if err := doc.DataTo(&link); err != nil {
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
		return false, err
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
