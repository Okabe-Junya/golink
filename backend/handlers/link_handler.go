package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Okabe-Junya/golink-backend/auth"
	"github.com/Okabe-Junya/golink-backend/interfaces"
	"github.com/Okabe-Junya/golink-backend/logger"
	"github.com/Okabe-Junya/golink-backend/models"
)

// LinkHandler handles HTTP requests for link operations
type LinkHandler struct {
	repo interfaces.LinkRepositoryInterface
}

// NewLinkHandler creates a new LinkHandler
func NewLinkHandler(repo interfaces.LinkRepositoryInterface) *LinkHandler {
	return &LinkHandler{
		repo: repo,
	}
}

// getUserFromContext extracts the user from request context
func getUserFromContext(r *http.Request) (string, string) {
	// Try to get authenticated user from context
	if user, ok := r.Context().Value("user").(*auth.User); ok && user != nil {
		return user.ID, user.Email
	}

	// Fall back to header for backward compatibility
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}
	return userID, ""
}

// CreateLink handles POST /api/links requests
func (h *LinkHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.Warn("Method not allowed for create link", logger.Fields{"method": r.Method})
		return
	}

	// Parse request body: short code and target URL are expected
	var requestBody struct {
		Short        string   `json:"short"`
		URL          string   `json:"url"`
		AccessLevel  string   `json:"access_level,omitempty"`
		ExpiresAt    string   `json:"expires_at,omitempty"`
		AllowedUsers []string `json:"allowed_users,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		logger.Error("Failed to decode request body", err, nil)
		return
	}

	// Validate required field
	if requestBody.Short == "" {
		http.Error(w, "Short code is required", http.StatusBadRequest)
		logger.Warn("Missing short code in request", nil)
		return
	}

	// If URL is not provided, set a default
	targetURL := requestBody.URL
	if targetURL == "" {
		targetURL = "https://example.com"
		logger.Info("No URL provided, using default", logger.Fields{
			"short":      requestBody.Short,
			"defaultURL": targetURL,
		})
	}

	// Validate short code format (alphanumeric and hyphen only)
	validShortCode := regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString(requestBody.Short)
	if !validShortCode {
		http.Error(w, "Short code must contain only letters, numbers, and hyphens", http.StatusBadRequest)
		logger.Warn("Invalid short code format", logger.Fields{"short": requestBody.Short})
		return
	}

	// Get user ID from context
	userID, userEmail := getUserFromContext(r)
	logger.Info("User creating link", logger.Fields{
		"userID": userID,
		"email":  userEmail,
		"short":  requestBody.Short,
	})

	ctx := context.Background()

	// Check if short code already exists
	existingLink, err := h.repo.GetByShort(ctx, requestBody.Short)
	if err == nil && existingLink != nil {
		http.Error(w, "Short code already exists", http.StatusConflict)
		logger.Warn("Attempted to create link with existing short code", logger.Fields{
			"short":  requestBody.Short,
			"userID": userID,
		})
		return
	}

	// Create a new link with the target URL
	link := models.NewLink(requestBody.Short, targetURL, userID)

	// Set access level if provided, otherwise use default
	if requestBody.AccessLevel != "" &&
		(requestBody.AccessLevel == models.AccessLevels.Public ||
			requestBody.AccessLevel == models.AccessLevels.Private ||
			requestBody.AccessLevel == models.AccessLevels.Restricted) {
		link.AccessLevel = requestBody.AccessLevel
	} else {
		link.AccessLevel = models.AccessLevels.Public
	}

	// Set allowed users if provided and access level is restricted
	if link.AccessLevel == models.AccessLevels.Restricted && len(requestBody.AllowedUsers) > 0 {
		link.AllowedUsers = requestBody.AllowedUsers
	} else {
		link.AllowedUsers = []string{}
	}

	// Set expiry time if provided
	if requestBody.ExpiresAt != "" {
		expiryTime, err := time.Parse(time.RFC3339, requestBody.ExpiresAt)
		if err != nil {
			http.Error(w, "Invalid expiry date format. Use RFC3339 format (e.g. 2025-12-31T23:59:59Z)", http.StatusBadRequest)
			logger.Error("Failed to parse expiry date", err, logger.Fields{
				"expiryDate": requestBody.ExpiresAt,
				"shortCode":  requestBody.Short,
			})
			return
		}

		// Ensure expiry time is in the future
		if expiryTime.Before(time.Now()) {
			http.Error(w, "Expiry date must be in the future", http.StatusBadRequest)
			logger.Warn("Attempted to set expiry date in the past", logger.Fields{
				"expiryDate": expiryTime.String(),
				"shortCode":  requestBody.Short,
			})
			return
		}

		link.SetExpiry(expiryTime)
		logger.Info("Link expiry set", logger.Fields{
			"shortCode":  requestBody.Short,
			"expiryDate": expiryTime.String(),
		})
	}

	// Save the link
	if err := h.repo.Create(ctx, link); err != nil {
		http.Error(w, "Failed to create link", http.StatusInternalServerError)
		logger.Error("Failed to create link in database", err, logger.Fields{
			"short":  requestBody.Short,
			"userID": userID,
		})
		return
	}

	logger.Info("Link created successfully", logger.Fields{
		"short":       link.Short,
		"url":         link.URL,
		"userID":      userID,
		"accessLevel": link.AccessLevel,
	})

	// Return the created link
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(link); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// GetLinks handles GET /api/links requests
func (h *LinkHandler) GetLinks(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.Warn("Method not allowed for get links", logger.Fields{"method": r.Method})
		return
	}

	// Get user ID from context
	userID, _ := getUserFromContext(r)

	// Get query parameters
	accessLevel := r.URL.Query().Get("access_level")
	createdBy := r.URL.Query().Get("created_by")
	logger.Info("Getting links with filters", logger.Fields{
		"userID":      userID,
		"accessLevel": accessLevel,
		"createdBy":   createdBy,
	})

	ctx := context.Background()
	var links []*models.Link
	var err error

	// Filter by access level if provided
	switch accessLevel {
	case "":
		if createdBy != "" {
			// Filter by creator if provided
			links, err = h.repo.GetByUser(ctx, createdBy)
		} else {
			// Get all links
			links, err = h.repo.GetAll(ctx)
		}
	default:
		links, err = h.repo.GetByAccessLevel(ctx, accessLevel)
	}

	if err != nil {
		http.Error(w, "Failed to get links", http.StatusInternalServerError)
		logger.Error("Failed to retrieve links", err, logger.Fields{
			"userID":      userID,
			"accessLevel": accessLevel,
			"createdBy":   createdBy,
		})
		return
	}

	// Filter links based on access control if user ID is provided
	if userID != "" {
		var filteredLinks []*models.Link
		for _, link := range links {
			// Check if the link is expired and update the flag if needed
			if link.IsLinkExpired() && !link.IsExpired {
				link.IsExpired = true
				if err := h.repo.Update(ctx, link); err != nil {
					logger.Error("Failed to update link expired status", err, logger.Fields{
						"short": link.Short,
					})
				}
			}

			// Public links are accessible to everyone
			if link.AccessLevel == models.AccessLevels.Public {
				filteredLinks = append(filteredLinks, link)
				continue
			}

			// Private links are only accessible to the creator
			if link.AccessLevel == models.AccessLevels.Private && link.CreatedBy == userID {
				filteredLinks = append(filteredLinks, link)
				continue
			}

			// Restricted links are accessible to the creator and allowed users
			if link.AccessLevel == models.AccessLevels.Restricted {
				if link.CreatedBy == userID {
					filteredLinks = append(filteredLinks, link)
					continue
				}

				// Check if the user is in the allowed users list
				for _, allowedUser := range link.AllowedUsers {
					if allowedUser == userID {
						filteredLinks = append(filteredLinks, link)
						break
					}
				}
			}
		}
		links = filteredLinks
	}

	logger.Info("Retrieved links", logger.Fields{
		"count":  len(links),
		"userID": userID,
	})

	// Return the links
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(links); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// GetLink handles GET /api/links/{short} requests
func (h *LinkHandler) GetLink(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.Warn("Method not allowed for get link", logger.Fields{"method": r.Method})
		return
	}

	// Get the short code from the URL path
	short := r.URL.Path[len("/api/links/"):]
	if short == "" {
		http.Error(w, "Short code is required", http.StatusBadRequest)
		logger.Warn("Short code is missing in get link request", nil)
		return
	}

	// Get user ID from context
	userID, _ := getUserFromContext(r)

	logger.Info("Getting link details", logger.Fields{
		"short":  short,
		"userID": userID,
	})

	// Get the link
	ctx := context.Background()
	link, err := h.repo.GetByShort(ctx, short)
	if err != nil {
		http.Error(w, "Link not found", http.StatusNotFound)
		logger.Error("Failed to find link", err, logger.Fields{"short": short})
		return
	}

	// Check access control if user ID is provided
	if userID != "" {
		hasAccess, err := h.repo.CheckAccess(ctx, short, userID)
		if err != nil {
			http.Error(w, "Failed to check access", http.StatusInternalServerError)
			logger.Error("Failed to check access for get link", err, logger.Fields{
				"short":  short,
				"userID": userID,
			})
			return
		}
		if !hasAccess {
			http.Error(w, "Access denied", http.StatusForbidden)
			logger.Warn("Access denied for get link", logger.Fields{
				"short":       short,
				"userID":      userID,
				"accessLevel": link.AccessLevel,
			})
			return
		}
	}

	logger.Info("Link details retrieved successfully", logger.Fields{
		"short":  short,
		"userID": userID,
	})

	// Return the link
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(link); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// UpdateLink handles PUT /api/links/{short} requests
func (h *LinkHandler) UpdateLink(w http.ResponseWriter, r *http.Request) {
	// Only allow PUT method
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.Warn("Method not allowed for update link", logger.Fields{"method": r.Method})
		return
	}

	// Get the short code from the URL path
	short := r.URL.Path[len("/api/links/"):]
	if short == "" {
		http.Error(w, "Short code is required", http.StatusBadRequest)
		logger.Warn("Short code is missing in update request", nil)
		return
	}

	// Get user ID from context
	userID, _ := getUserFromContext(r)
	logger.Info("Update link request received", logger.Fields{
		"short":  short,
		"userID": userID,
	})

	// Get the existing link
	ctx := context.Background()
	link, err := h.repo.GetByShort(ctx, short)
	if err != nil {
		http.Error(w, "Link not found", http.StatusNotFound)
		logger.Error("Link not found for update", err, logger.Fields{"short": short})
		return
	}

	// Only the creator can update this link
	// If userId is anonymous, this allows also update by anonymous users
	if userID != "anonymous" && link.CreatedBy != userID && auth.IsAuthEnabled() {
		http.Error(w, "Only the creator can update this link", http.StatusForbidden)
		logger.Warn("Unauthorized update attempt", logger.Fields{
			"short":       short,
			"requestUser": userID,
			"creatorUser": link.CreatedBy,
		})
		return
	}

	var requestBody struct {
		URL          string   `json:"url,omitempty"`
		AccessLevel  string   `json:"access_level,omitempty"`
		ExpiresAt    string   `json:"expires_at,omitempty"`
		AllowedUsers []string `json:"allowed_users,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		logger.Error("Failed to decode update request body", err, logger.Fields{"short": short})
		return
	}

	// Update the link fields
	if requestBody.URL != "" {
		link.URL = requestBody.URL
	}

	// Update access level if provided
	if requestBody.AccessLevel != "" &&
		(requestBody.AccessLevel == models.AccessLevels.Public ||
			requestBody.AccessLevel == models.AccessLevels.Private ||
			requestBody.AccessLevel == models.AccessLevels.Restricted) {
		link.AccessLevel = requestBody.AccessLevel
	}

	// Update allowed users if provided and access level is restricted
	var updateErr error
	if link.AccessLevel == models.AccessLevels.Restricted && requestBody.AllowedUsers != nil {
		link.AllowedUsers = requestBody.AllowedUsers
		if updateErr = h.repo.Update(ctx, link); updateErr != nil {
			logger.Error("Failed to update link allowed users", updateErr, logger.Fields{"short": short})
		}
	}

	// Update expiry time if provided
	if requestBody.ExpiresAt != "" {
		expiryTime, err := time.Parse(time.RFC3339, requestBody.ExpiresAt)
		if err != nil {
			http.Error(w, "Invalid expiry date format. Use RFC3339 format (e.g. 2025-12-31T23:59:59Z)", http.StatusBadRequest)
			logger.Error("Failed to parse expiry date in update", err, logger.Fields{
				"expiryDate": requestBody.ExpiresAt,
				"shortCode":  short,
			})
			return
		}

		// Ensure expiry time is in the future
		if expiryTime.Before(time.Now()) {
			http.Error(w, "Expiry date must be in the future", http.StatusBadRequest)
			logger.Warn("Attempted to set expiry date in the past during update", logger.Fields{
				"expiryDate": expiryTime.String(),
				"shortCode":  short,
			})
			return
		}

		link.SetExpiry(expiryTime)
		logger.Info("Link expiry updated", logger.Fields{
			"shortCode":  short,
			"expiryDate": expiryTime.String(),
		})
	} else if requestBody.ExpiresAt == "" && !link.ExpiresAt.IsZero() {
		// If expiresAt is explicitly set to empty string, remove the expiration
		link.ExpiresAt = time.Time{}
		link.IsExpired = false
		logger.Info("Link expiry removed", logger.Fields{
			"shortCode": short,
		})
	}

	link.UpdatedAt = time.Now()

	// Save the updated link
	if err := h.repo.Update(ctx, link); err != nil {
		http.Error(w, "Failed to update link", http.StatusInternalServerError)
		logger.Error("Failed to update link in database", err, logger.Fields{
			"short":  short,
			"userID": userID,
		})
		return
	}

	logger.Info("Link updated successfully", logger.Fields{
		"short":       short,
		"userID":      userID,
		"newURL":      link.URL,
		"accessLevel": link.AccessLevel,
	})

	// Return the updated link
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(link); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// DeleteLink handles DELETE /api/links/{short} requests
func (h *LinkHandler) DeleteLink(w http.ResponseWriter, r *http.Request) {
	// Only allow DELETE method
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.Warn("Method not allowed for delete link", logger.Fields{"method": r.Method})
		return
	}

	// Get the short code from the URL path
	short := r.URL.Path[len("/api/links/"):]
	if short == "" {
		http.Error(w, "Short code is required", http.StatusBadRequest)
		logger.Warn("Short code is missing in delete request", nil)
		return
	}

	// Get user ID from context
	userID, _ := getUserFromContext(r)

	// Get the existing link
	ctx := context.Background()
	link, err := h.repo.GetByShort(ctx, short)
	if err != nil {
		http.Error(w, "Link not found", http.StatusNotFound)
		logger.Error("Failed to find link for deletion", err, logger.Fields{
			"short":  short,
			"userID": userID,
		})
		return
	}

	// 認証が無効または匿名ユーザーの場合は権限チェックをスキップ
	// それ以外の場合は作成者のみが削除可能
	if userID != "anonymous" && link.CreatedBy != userID && auth.IsAuthEnabled() {
		http.Error(w, "Only the creator can delete this link", http.StatusForbidden)
		logger.Warn("Unauthorized delete attempt", logger.Fields{
			"short":       short,
			"requestUser": userID,
			"creatorUser": link.CreatedBy,
		})
		return
	}

	// Delete the link
	if err := h.repo.Delete(ctx, short); err != nil {
		http.Error(w, "Failed to delete link", http.StatusInternalServerError)
		logger.Error("Failed to delete link", err, logger.Fields{
			"short":  short,
			"userID": userID,
		})
		return
	}

	logger.Info("Link successfully deleted", logger.Fields{
		"short":           short,
		"userID":          userID,
		"originalCreator": link.CreatedBy,
	})

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// RedirectLink handles GET /{short} requests
func (h *LinkHandler) RedirectLink(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.Warn("Method not allowed for redirect", logger.Fields{"method": r.Method})
		return
	}

	// Skip static file requests and special paths
	path := r.URL.Path[1:] // Remove leading slash
	if path == "" || path == "index.html" || path == "favicon.ico" ||
		strings.HasPrefix(path, "static/") || strings.HasPrefix(path, "assets/") {
		http.NotFound(w, r)
		return
	}

	logger.Info("Redirect request received", logger.Fields{"short": path})

	// Get user ID from context
	userID, _ := getUserFromContext(r)

	// Get the link
	ctx := context.Background()
	link, err := h.repo.GetByShort(ctx, path)
	if err != nil {
		http.Error(w, "Link not found", http.StatusNotFound)
		logger.Error("Link not found for redirect", err, logger.Fields{"short": path})
		return
	}

	// Check if the link is expired
	if link.IsLinkExpired() {
		// Mark the link as expired in the database if not already marked
		if !link.IsExpired {
			link.IsExpired = true
			err := h.repo.Update(ctx, link)
			if err != nil {
				logger.Error("Failed to mark link as expired", err, logger.Fields{"short": path})
			}
		}

		http.Error(w, "This link has expired", http.StatusGone)
		logger.Info("Expired link access attempt", logger.Fields{
			"short":     path,
			"userID":    userID,
			"expiresAt": link.ExpiresAt.Format(time.RFC3339),
		})
		return
	}

	// Check access control
	hasAccess := true
	if userID != "" {
		hasAccess, err = h.repo.CheckAccess(ctx, path, userID)
		if err != nil {
			http.Error(w, "Failed to check access", http.StatusInternalServerError)
			logger.Error("Failed to check access for redirect", err, logger.Fields{
				"short":  path,
				"userID": userID,
			})
			return
		}
	} else if link.AccessLevel != models.AccessLevels.Public {
		// If no user ID is provided and the link is not public, deny access
		hasAccess = false
	}

	if !hasAccess {
		http.Error(w, "Access denied", http.StatusForbidden)
		logger.Warn("Access denied for redirect", logger.Fields{
			"short":       path,
			"userID":      userID,
			"accessLevel": link.AccessLevel,
		})
		return
	}

	// Increment the click count in a background goroutine
	go func() {
		// Use a new context for the background operation
		ctx := context.Background()
		if err := h.repo.IncrementClickCount(ctx, path); err != nil {
			logger.Error("Failed to increment click count", err, logger.Fields{"short": path})
		}
	}()

	logger.Info("Redirecting to target URL", logger.Fields{
		"short":     path,
		"targetURL": link.URL,
		"userID":    userID,
	})

	// Redirect to the original URL
	http.Redirect(w, r, link.URL, http.StatusFound)
}

// HealthCheck handles GET /health requests
func (h *LinkHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if we can connect to the database
	ctx := context.Background()
	_, err := h.repo.GetAll(ctx)

	response := map[string]string{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if err != nil {
		response["status"] = "unhealthy"
		response["error"] = "Database connection failed"
		w.WriteHeader(http.StatusServiceUnavailable)
		logger.Error("Health check failed", err, nil)
	} else {
		logger.Info("Health check passed", nil)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode health check response", http.StatusInternalServerError)
	}
}

// DeleteExpiredLinks handles DELETE /api/links/expired requests
func (h *LinkHandler) DeleteExpiredLinks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userID, _ := getUserFromContext(r)

	ctx := context.Background()
	links, err := h.repo.GetAll(ctx)
	if err != nil {
		http.Error(w, "Failed to get links", http.StatusInternalServerError)
		logger.Error("Failed to get links for bulk deletion", err, nil)
		return
	}

	var deletedCount int
	for _, link := range links {
		// Only delete links that the user has permission to delete
		if link.CreatedBy != userID && userID != "anonymous" {
			continue
		}

		// Check if the link is expired
		if link.IsLinkExpired() {
			if err := h.repo.Delete(ctx, link.Short); err != nil {
				logger.Error("Failed to delete expired link", err, logger.Fields{
					"short": link.Short,
				})
				continue
			}
			deletedCount++
			logger.Info("Deleted expired link", logger.Fields{
				"short":  link.Short,
				"userID": userID,
			})
		}
	}

	// Return success response with count
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"deleted_count": deletedCount,
		"message":       "Expired links deleted successfully",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
