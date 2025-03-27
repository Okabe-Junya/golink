package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/Okabe-Junya/golink-backend/logger"
	"github.com/Okabe-Junya/golink-backend/models"
	"github.com/Okabe-Junya/golink-backend/pkg/config"
	"google.golang.org/api/option"
)

func main() {
	// Command-line flags
	var (
		createStatsCollection bool
		migrateExpiredLinks   bool
		dryRun                bool
	)

	flag.BoolVar(&createStatsCollection, "create-stats", false, "Create link_stats collection")
	flag.BoolVar(&migrateExpiredLinks, "migrate-expired", false, "Migrate expired links")
	flag.BoolVar(&dryRun, "dry-run", false, "Run in dry-run mode (no changes)")
	flag.Parse()

	// Load config
	cfg := config.New()

	// Initialize Firebase
	client, err := initFirebase(cfg.Firebase)
	if err != nil {
		logger.Fatal("Failed to initialize Firebase", err, nil)
	}
	defer client.Close()

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Run migrations
	if createStatsCollection {
		if err := createLinkStatsCollection(ctx, client, dryRun); err != nil {
			logger.Fatal("Failed to create link_stats collection", err, nil)
		}
	}

	if migrateExpiredLinks {
		if err := updateExpiredLinks(ctx, client, dryRun); err != nil {
			logger.Fatal("Failed to migrate expired links", err, nil)
		}
	}

	logger.Info("Migration completed successfully", nil)
}

// initFirebase initializes Firebase and returns a Firestore client
func initFirebase(cfg config.FirebaseConfig) (*firestore.Client, error) {
	ctx := context.Background()
	var opt option.ClientOption

	switch {
	case cfg.CredentialsJSON != "":
		opt = option.WithCredentialsJSON([]byte(cfg.CredentialsJSON))
	case cfg.CredentialsFile != "":
		opt = option.WithCredentialsFile(cfg.CredentialsFile)
	default:
		return nil, fmt.Errorf("no Firebase credentials provided")
	}

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firestore: %w", err)
	}

	return client, nil
}

// createLinkStatsCollection creates the link_stats collection
func createLinkStatsCollection(ctx context.Context, client *firestore.Client, dryRun bool) error {
	logger.Info("Creating link_stats collection", logger.Fields{
		"dry_run": dryRun,
	})

	// Get all links
	linksIter := client.Collection("links").Documents(ctx)
	batch := client.Batch()
	count := 0

	for {
		doc, err := linksIter.Next()
		if err != nil {
			break
		}

		var link models.Link
		if err := doc.DataTo(&link); err != nil {
			logger.Error("Failed to parse link", err, logger.Fields{
				"document_id": doc.Ref.ID,
			})
			continue
		}

		// Check if stats already exist
		statsRef := client.Collection("link_stats").Doc(link.Short)
		_, err = statsRef.Get(ctx)
		if err == nil {
			// Stats exist, skip
			logger.Info("Stats already exist", logger.Fields{
				"short": link.Short,
			})
			continue
		}

		// Create new stats
		stats := models.NewLinkStats(link.Short)
		stats.TotalClicks = link.ClickCount

		if !dryRun {
			batch.Set(statsRef, stats)
			count++

			// Execute batch when it reaches 500 operations (Firestore limit)
			if count%500 == 0 {
				if _, err := batch.Commit(ctx); err != nil {
					return fmt.Errorf("failed to commit batch: %w", err)
				}
				batch = client.Batch()
				logger.Info("Batch committed", logger.Fields{
					"count": count,
				})
			}
		} else {
			logger.Info("Would create stats", logger.Fields{
				"short":        link.Short,
				"total_clicks": link.ClickCount,
			})
			count++
		}
	}

	// Commit any remaining operations
	if !dryRun && count%500 != 0 {
		if _, err := batch.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit final batch: %w", err)
		}
	}

	logger.Info("Link stats migration completed", logger.Fields{
		"count":   count,
		"dry_run": dryRun,
	})

	return nil
}

// updateExpiredLinks updates the is_expired field for links that are past their expiry date
func updateExpiredLinks(ctx context.Context, client *firestore.Client, dryRun bool) error {
	logger.Info("Updating expired links", logger.Fields{
		"dry_run": dryRun,
	})

	now := time.Now()
	query := client.Collection("links").Where("expires_at", "<", now).Where("is_expired", "==", false)
	linksIter := query.Documents(ctx)
	count := 0

	batch := client.Batch()

	for {
		doc, err := linksIter.Next()
		if err != nil {
			break
		}

		var link models.Link
		if err := doc.DataTo(&link); err != nil {
			logger.Error("Failed to parse link", err, logger.Fields{
				"document_id": doc.Ref.ID,
			})
			continue
		}

		logger.Info("Found expired link", logger.Fields{
			"short":      link.Short,
			"expires_at": link.ExpiresAt,
		})

		if !dryRun {
			// Update the link
			link.IsExpired = true
			link.UpdatedAt = now
			batch.Set(doc.Ref, link)
			count++

			// Execute batch when it reaches 500 operations (Firestore limit)
			if count%500 == 0 {
				if _, err := batch.Commit(ctx); err != nil {
					return fmt.Errorf("failed to commit batch: %w", err)
				}
				batch = client.Batch()
				logger.Info("Batch committed", logger.Fields{
					"count": count,
				})
			}
		} else {
			logger.Info("Would update link", logger.Fields{
				"short":      link.Short,
				"expires_at": link.ExpiresAt.Format(time.RFC3339),
			})
			count++
		}
	}

	// Commit any remaining operations
	if !dryRun && count%500 != 0 {
		if _, err := batch.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit final batch: %w", err)
		}
	}

	logger.Info("Expired links migration completed", logger.Fields{
		"count":   count,
		"dry_run": dryRun,
	})

	return nil
}
