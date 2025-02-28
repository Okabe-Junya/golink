package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"time"

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

// CreateLink handles POST /api/links requests
func (h *LinkHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.Warn("Method not allowed for create link", logger.Fields{"method": r.Method})
		return
	}

	// Parse request body: short code and target URL are expected
	var requestBody struct {
		Short string `json:"short"`
		URL   string `json:"url"`
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

	// Get user ID from auth middleware (from header)
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
		logger.Info("Anonymous user creating link", logger.Fields{"short": requestBody.Short})
	}

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

	// Enforce default access control
	link.AccessLevel = models.AccessLevels.Public
	link.AllowedUsers = []string{}

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
		"short":  link.Short,
		"url":    link.URL,
		"userID": userID,
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

	// Get user ID from context (would be set by auth middleware)
	userID := r.Header.Get("X-User-ID")

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

	// Get user ID from context (would be set by auth middleware)
	userID := r.Header.Get("X-User-ID")

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

	// Get user ID from context (would be set by auth middleware)
	userID := r.Header.Get("X-User-ID")
	// ユーザーIDによるバリデーションを削除

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

	// ユーザーIDによる作成者チェックを削除

	// Only URL can be updated from the front-end
	var requestBody struct {
		URL string `json:"url,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		logger.Error("Failed to decode update request body", err, logger.Fields{"short": short})
		return
	}

	// Update the link
	if requestBody.URL != "" {
		link.URL = requestBody.URL
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
		"short":  short,
		"userID": userID,
		"newURL": link.URL,
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

	// Get user ID from context (would be set by auth middleware)
	userID := r.Header.Get("X-User-ID")

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

	// Delete the link - allow any user to delete
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

	// Get the short code from the URL path
	short := r.URL.Path[1:]
	if short == "" {
		http.Error(w, "Short code is required", http.StatusBadRequest)
		logger.Warn("Short code is missing in redirect request", nil)
		return
	}

	logger.Info("Redirect request received", logger.Fields{"short": short})

	// Get user ID from context (would be set by auth middleware)
	userID := r.Header.Get("X-User-ID")

	// Get the link
	ctx := context.Background()
	link, err := h.repo.GetByShort(ctx, short)
	if err != nil {
		http.Error(w, "Link not found", http.StatusNotFound)
		logger.Error("Link not found for redirect", err, logger.Fields{"short": short})
		return
	}

	// Check access control
	hasAccess := true
	if userID != "" {
		hasAccess, err = h.repo.CheckAccess(ctx, short, userID)
		if err != nil {
			http.Error(w, "Failed to check access", http.StatusInternalServerError)
			logger.Error("Failed to check access for redirect", err, logger.Fields{
				"short":  short,
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
			"short":       short,
			"userID":      userID,
			"accessLevel": link.AccessLevel,
		})
		return
	}

	// Increment the click count in a background goroutine
	go func() {
		// Use a new context for the background operation
		ctx := context.Background()
		if err := h.repo.IncrementClickCount(ctx, short); err != nil {
			logger.Error("Failed to increment click count", err, logger.Fields{"short": short})
		}
	}()

	logger.Info("Redirecting to target URL", logger.Fields{
		"short":     short,
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
