package auth

import (
	"context"
	"errors"
	"net/http"
)

// contextKey is a private type for context keys
type contextKey int

const (
	// userContextKey is the key for user information in request contexts
	userContextKey contextKey = iota
)

var (
	// ErrNoUserInContext is returned when no user is found in the context
	ErrNoUserInContext = errors.New("no user found in context")
)

// ContextWithUser adds a user to the given context
func ContextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// UserFromContext extracts a user from the given context
func UserFromContext(ctx context.Context) (*User, error) {
	user, ok := ctx.Value(userContextKey).(*User)
	if !ok || user == nil {
		return nil, ErrNoUserInContext
	}
	return user, nil
}

// GetUserFromRequest extracts a user from the request, either from session or header
func GetUserFromRequest(r *http.Request) (*User, error) {
	// First try to get user from context (set by authentication middleware)
	if user, err := UserFromContext(r.Context()); err == nil {
		return user, nil
	}

	// Try to get user via GetCurrentUser (which handles both session and headers)
	return GetCurrentUser(r)
}
