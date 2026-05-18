package security

import (
	"net/url"
	"testing"

	"go-crud-db-p2/config"
)

func TestGoogleOAuthAuthCodeURLPromptsAccountSelection(t *testing.T) {
	googleOAuth := NewGoogleOAuth(config.GoogleConfig{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "http://localhost:8080/api/v1/auth/google/callback",
	})

	loginURL, err := googleOAuth.AuthCodeURL("test-state")
	if err != nil {
		t.Fatalf("AuthCodeURL() error = %v", err)
	}

	parsedURL, err := url.Parse(loginURL)
	if err != nil {
		t.Fatalf("Parse login URL error = %v", err)
	}

	if prompt := parsedURL.Query().Get("prompt"); prompt != "select_account" {
		t.Fatalf("prompt = %q, want %q", prompt, "select_account")
	}
}
