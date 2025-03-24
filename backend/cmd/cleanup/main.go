package main

import (
	"context"
	"flag"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/Okabe-Junya/golink-backend/logger"
	"github.com/Okabe-Junya/golink-backend/repositories"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "Perform a dry run without actually deleting any links")
	olderThan := flag.Int("older-than", 30, "Delete expired links older than this many days")
	flag.Parse()

	logger.Info("Starting cleanup job", logger.Fields{
		"dryRun":    *dryRun,
		"olderThan": *olderThan,
	})

	// Initialize Firestore client
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, os.Getenv("PROJECT_ID"))
	if err != nil {
		logger.Error("Failed to initialize Firestore client", err, nil)
		return
	}
	defer client.Close()

	// Initialize repository
	repo := repositories.NewLinkRepository(client)

	// Get all links
	links, err := repo.GetAll(ctx)
	if err != nil {
		logger.Error("Failed to get links", err, nil)
		return
	}

	cutoffDate := time.Now().AddDate(0, 0, -*olderThan)
	var processedCount, expiredCount int

	for _, link := range links {
		processedCount++

		// Skip if not expired
		if !link.IsLinkExpired() {
			continue
		}

		// Skip if not old enough
		if link.ExpiresAt.After(cutoffDate) {
			continue
		}

		expiredCount++

		if *dryRun {
			logger.Info("Would delete expired link", logger.Fields{
				"short":     link.Short,
				"url":       link.URL,
				"expiredAt": link.ExpiresAt,
			})
			continue
		}

		// Delete the link
		if err := repo.Delete(ctx, link.Short); err != nil {
			logger.Error("Failed to delete expired link", err, logger.Fields{
				"short": link.Short,
			})
			continue
		}

		logger.Info("Deleted expired link", logger.Fields{
			"short":     link.Short,
			"url":       link.URL,
			"expiredAt": link.ExpiresAt,
		})
	}

	logger.Info("Cleanup job completed", logger.Fields{
		"processed": processedCount,
		"expired":   expiredCount,
		"dryRun":    *dryRun,
	})
}
