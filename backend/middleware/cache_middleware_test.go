package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// userScopedHandler writes a body that depends on the X-User-ID header, standing
// in for a per-user endpoint such as /api/links.
func userScopedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"user":%q}`, r.Header.Get("X-User-ID"))
	})
}

// TestCacheMiddleware_DoesNotLeakAcrossAuthenticatedUsers guards the fix for the
// cross-user cache leak: two authenticated users hitting the same path must each
// receive their own response, never a cached copy of the other user's data.
func TestCacheMiddleware_DoesNotLeakAcrossAuthenticatedUsers(t *testing.T) {
	t.Parallel()

	handler := CacheMiddleware(userScopedHandler())

	req := func(userID string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodGet, "/api/links", nil)
		r.AddCookie(&http.Cookie{Name: "session_token", Value: "token-" + userID})
		r.Header.Set("X-User-ID", userID)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, r)
		return rr
	}

	first := req("alice").Body.String()
	second := req("bob").Body.String()

	if want := `{"user":"alice"}`; first != want {
		t.Fatalf("first response = %q, want %q", first, want)
	}
	if want := `{"user":"bob"}`; second != want {
		t.Fatalf("cross-user cache leak: second response = %q, want %q", second, want)
	}
}

// TestCacheMiddleware_AnonymousResponsesStillCached confirms the fix does not
// disable caching for unauthenticated requests, which only ever see public data.
func TestCacheMiddleware_AnonymousResponsesStillCached(t *testing.T) {
	t.Parallel()

	var calls int
	counting := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"n":%d}`, calls)
	})
	handler := CacheMiddleware(counting)

	get := func() *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodGet, "/api/public-anon-probe", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, r)
		return rr
	}

	first := get()
	second := get()

	if got := second.Header().Get("X-Cache"); got != "HIT" {
		t.Fatalf("second anonymous request X-Cache = %q, want HIT", got)
	}
	if first.Body.String() != second.Body.String() {
		t.Fatalf("anonymous cache miss: %q != %q", first.Body.String(), second.Body.String())
	}
	if calls != 1 {
		t.Fatalf("upstream handler called %d times, want 1 (second served from cache)", calls)
	}
}
