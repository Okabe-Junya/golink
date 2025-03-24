package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/Okabe-Junya/golink-backend/interfaces"
	"github.com/Okabe-Junya/golink-backend/logger"
	"github.com/Okabe-Junya/golink-backend/middleware"
	"github.com/Okabe-Junya/golink-backend/models"
)

// AnalyticsHandler provides analytics endpoints for link usage
type AnalyticsHandler struct {
	repo interfaces.LinkRepositoryInterface
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(repo interfaces.LinkRepositoryInterface) *AnalyticsHandler {
	return &AnalyticsHandler{
		repo: repo,
	}
}

// GetLinkStats handles GET /api/analytics/links/{short} requests
func (h *AnalyticsHandler) GetLinkStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.RespondWithError(w, http.StatusMethodNotAllowed, middleware.ErrBadRequest, "Method not allowed")
		return
	}

	// Get the short code from the URL path
	short := r.URL.Path[len("/api/analytics/links/"):]
	if short == "" {
		middleware.RespondWithError(w, http.StatusBadRequest, middleware.ErrBadRequest, "Short code is required")
		return
	}

	// Get user ID from context
	userID, _ := getUserFromContext(r)

	// Get the link
	ctx := context.Background()
	link, err := h.repo.GetByShort(ctx, short)
	if err != nil {
		middleware.RespondWithError(w, http.StatusNotFound, middleware.ErrNotFound, "Link not found")
		return
	}

	// Check if user has permission to view stats
	// Only creator can view stats if link is not public
	if link.AccessLevel != models.AccessLevels.Public && link.CreatedBy != userID {
		hasAccess, err := h.repo.CheckAccess(ctx, short, userID)
		if err != nil || !hasAccess {
			middleware.RespondWithError(w, http.StatusForbidden, middleware.ErrForbidden, "Access denied")
			return
		}
	}

	// Prepare stats response
	stats := map[string]interface{}{
		"link_id":      link.ID,
		"short":        link.Short,
		"url":          link.URL,
		"click_count":  link.ClickCount,
		"created_at":   link.CreatedAt,
		"age_days":     time.Since(link.CreatedAt).Hours() / 24,
		"access_level": link.AccessLevel,
		"is_expired":   link.IsExpired,
	}

	// Add expiry information if set
	if !link.ExpiresAt.IsZero() {
		stats["expires_at"] = link.ExpiresAt
		if link.IsLinkExpired() {
			stats["is_expired"] = true
		}
	}

	// If the link has been used, calculate average clicks per day
	if !link.CreatedAt.IsZero() {
		var daysSinceCreation float64
		if link.IsExpired {
			// For expired links, calculate average based on the time until expiration
			daysSinceCreation = link.ExpiresAt.Sub(link.CreatedAt).Hours() / 24
		} else {
			daysSinceCreation = time.Since(link.CreatedAt).Hours() / 24
		}
		if daysSinceCreation > 0 {
			stats["avg_clicks_per_day"] = float64(link.ClickCount) / daysSinceCreation
		}
	}

	logger.Info("Analytics retrieved for link", logger.Fields{
		"short":       short,
		"userID":      userID,
		"click_count": link.ClickCount,
	})

	// Return the stats
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		middleware.RespondWithError(w, http.StatusInternalServerError, middleware.ErrInternalServerError, "Failed to encode response")
	}
}

// GetTopLinks handles GET /api/analytics/top requests
func (h *AnalyticsHandler) GetTopLinks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.RespondWithError(w, http.StatusMethodNotAllowed, middleware.ErrBadRequest, "Method not allowed")
		return
	}

	// Get user ID from context
	userID, _ := getUserFromContext(r)

	// Optional: limit parameter (default: 10)
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
			limit = n
		}
	}

	// Get all links
	ctx := context.Background()
	links, err := h.repo.GetAll(ctx)
	if err != nil {
		middleware.RespondWithError(w, http.StatusInternalServerError, middleware.ErrInternalServerError, "Failed to retrieve links")
		return
	}

	// Filter links based on access control
	var accessibleLinks []*models.Link
	for _, link := range links {
		if link.AccessLevel == models.AccessLevels.Public {
			accessibleLinks = append(accessibleLinks, link)
			continue
		}

		if link.CreatedBy == userID {
			accessibleLinks = append(accessibleLinks, link)
			continue
		}

		if link.AccessLevel == models.AccessLevels.Restricted {
			for _, allowedUser := range link.AllowedUsers {
				if allowedUser == userID {
					accessibleLinks = append(accessibleLinks, link)
					break
				}
			}
		}
	}

	// Sort links by click count (descending)
	sort.Slice(accessibleLinks, func(i, j int) bool {
		return accessibleLinks[i].ClickCount > accessibleLinks[j].ClickCount
	})

	// Limit the number of results
	if len(accessibleLinks) > limit {
		accessibleLinks = accessibleLinks[:limit]
	}

	logger.Info("Top links retrieved", logger.Fields{
		"userID": userID,
		"count":  len(accessibleLinks),
		"limit":  limit,
	})

	// Return the top links
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(accessibleLinks); err != nil {
		middleware.RespondWithError(w, http.StatusInternalServerError, middleware.ErrInternalServerError, "Failed to encode response")
	}
}
