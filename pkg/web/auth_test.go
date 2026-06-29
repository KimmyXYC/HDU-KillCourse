package web

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cr4n5/HDU-KillCourse/config"
)

func basicHeader(user, pass string) string {
	token := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
	return "Basic " + token
}

func TestBasicAuthWrap_DisabledWhenCredentialsEmpty(t *testing.T) {
	called := false
	next := func(w http.ResponseWriter, r *http.Request) { called = true }

	webCfg := config.WebConfig{Auth: config.WebAuthConfig{Enabled: "1", Username: "", Password: ""}}
	h := basicAuthWrap(webCfg, next)

	r := httptest.NewRequest(http.MethodGet, "http://example/", nil)
	w := httptest.NewRecorder()
	h(w, r)

	if !called {
		t.Fatalf("expected handler to be called when credentials are empty")
	}
}

func TestBasicAuthWrap_UnauthorizedWithoutHeader(t *testing.T) {
	called := false
	next := func(w http.ResponseWriter, r *http.Request) { called = true }

	webCfg := config.WebConfig{Auth: config.WebAuthConfig{Enabled: "1", Username: "admin", Password: "password"}}
	h := basicAuthWrap(webCfg, next)

	r := httptest.NewRequest(http.MethodGet, "http://example/", nil)
	w := httptest.NewRecorder()
	h(w, r)

	if called {
		t.Fatalf("expected handler not to be called")
	}
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
	if got := w.Header().Get("WWW-Authenticate"); got == "" {
		t.Fatalf("expected WWW-Authenticate header to be set")
	}
}

func TestBasicAuthWrap_UnauthorizedWithWrongCredentials(t *testing.T) {
	called := false
	next := func(w http.ResponseWriter, r *http.Request) { called = true }

	webCfg := config.WebConfig{Auth: config.WebAuthConfig{Enabled: "1", Username: "admin", Password: "password"}}
	h := basicAuthWrap(webCfg, next)

	r := httptest.NewRequest(http.MethodGet, "http://example/", nil)
	r.Header.Set("Authorization", basicHeader("admin", "wrong"))
	w := httptest.NewRecorder()
	h(w, r)

	if called {
		t.Fatalf("expected handler not to be called")
	}
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestBasicAuthWrap_AllowsWithCorrectCredentials(t *testing.T) {
	called := false
	next := func(w http.ResponseWriter, r *http.Request) { called = true }

	webCfg := config.WebConfig{Auth: config.WebAuthConfig{Enabled: "1", Username: "admin", Password: "password"}}
	h := basicAuthWrap(webCfg, next)

	r := httptest.NewRequest(http.MethodGet, "http://example/", nil)
	r.Header.Set("Authorization", basicHeader("admin", "password"))
	w := httptest.NewRecorder()
	h(w, r)

	if !called {
		t.Fatalf("expected handler to be called")
	}
}
