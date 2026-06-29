package notify

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cr4n5/HDU-KillCourse/config"
)

func Test_sendWebhook_RendersTemplateAndHeaders(t *testing.T) {
	t.Parallel()

	var gotMethod string
	var gotContentType string
	var gotCustomHeader string
	var gotBody string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		gotCustomHeader = r.Header.Get("X-Test")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	oldClient := httpClient
	httpClient = srv.Client()
	defer func() { httpClient = oldClient }()

	wh := config.Webhook{
		Url:    srv.URL,
		Method: "POST",
		Headers: map[string]string{
			"X-Test":       "1",
			"Content-Type": "application/json; charset=utf-8",
		},
		BodyTemplate: `{"title":"{{.Title}}","body":"{{.Body}}"}`,
		Enabled:      "1",
	}

	if err := sendWebhook(wh, "标题", "内容"); err != nil {
		t.Fatalf("sendWebhook error: %v", err)
	}
	if gotMethod != "POST" {
		t.Fatalf("method = %q, want POST", gotMethod)
	}
	if gotCustomHeader != "1" {
		t.Fatalf("X-Test header = %q, want 1", gotCustomHeader)
	}
	if gotContentType == "" {
		t.Fatalf("Content-Type should not be empty")
	}
	if !strings.Contains(gotBody, "标题") || !strings.Contains(gotBody, "内容") {
		t.Fatalf("body = %q, want contains rendered title/body", gotBody)
	}
}

func Test_sendBark_PathEscaped(t *testing.T) {
	t.Parallel()

	var gotPath string
	var gotQuery string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	oldClient := httpClient
	httpClient = srv.Client()
	defer func() { httpClient = oldClient }()

	bk := config.Bark{
		Server:  srv.URL,
		Key:     "k/1",
		Sound:   "bell",
		Group:   "g 1",
		URL:     "https://example.com/a?b=1",
		Icon:    "https://example.com/icon.png",
		Enabled: "1",
	}

	if err := sendBark(bk, "t/1", "b/1"); err != nil {
		t.Fatalf("sendBark error: %v", err)
	}
	if !strings.Contains(gotPath, "k%2F1") || !strings.Contains(gotPath, "t%2F1") || !strings.Contains(gotPath, "b%2F1") {
		t.Fatalf("path not escaped as expected: %q", gotPath)
	}
	if !strings.Contains(gotQuery, "sound=bell") || !strings.Contains(gotQuery, "group=g+1") {
		t.Fatalf("query not encoded as expected: %q", gotQuery)
	}
}
