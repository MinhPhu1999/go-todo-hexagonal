package security

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go-crud-db-p2/config"
	domain "go-crud-db-p2/internal/core/domain/platform"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	googleUserInfoURL = "https://openidconnect.googleapis.com/v1/userinfo"
)

type GoogleOAuth struct {
	config *oauth2.Config
	client *http.Client
}

func NewGoogleOAuth(cfg config.Config) *GoogleOAuth {
	return &GoogleOAuth{
		config: &oauth2.Config{
			ClientID:     strings.TrimSpace(cfg.Google.ClientID),
			ClientSecret: strings.TrimSpace(cfg.Google.ClientSecret),
			RedirectURL:  strings.TrimSpace(cfg.Google.RedirectURL),
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
		client: http.DefaultClient,
	}
}

func (googleOAuth *GoogleOAuth) AuthCodeURL(state string) (string, error) {
	if !googleOAuth.ready() {
		return "", domain.ErrGoogleAuthUnavailable
	}
	return googleOAuth.config.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "select_account"),
	), nil
}

func (googleOAuth *GoogleOAuth) Exchange(ctx context.Context, code string) (domain.GoogleProfile, error) {
	if !googleOAuth.ready() {
		return domain.GoogleProfile{}, domain.ErrGoogleAuthUnavailable
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return domain.GoogleProfile{}, fmt.Errorf("%w: google code is required", domain.ErrInvalidAuthRequest)
	}

	token, err := googleOAuth.config.Exchange(ctx, code)
	if err != nil {
		return domain.GoogleProfile{}, fmt.Errorf("%w: exchange google code: %w", domain.ErrGoogleProfileUnavailable, err)
	}

	client := googleOAuth.config.Client(ctx, token)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserInfoURL, nil)
	if err != nil {
		return domain.GoogleProfile{}, err
	}

	response, err := client.Do(request)
	if err != nil {
		return domain.GoogleProfile{}, fmt.Errorf("%w: fetch google profile: %w", domain.ErrGoogleProfileUnavailable, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return domain.GoogleProfile{}, err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return domain.GoogleProfile{}, fmt.Errorf("%w: google userinfo status %d", domain.ErrGoogleProfileUnavailable, response.StatusCode)
	}

	var profile googleUserInfo
	if err := json.Unmarshal(body, &profile); err != nil {
		return domain.GoogleProfile{}, fmt.Errorf("%w: decode google profile: %w", domain.ErrGoogleProfileUnavailable, err)
	}

	return domain.GoogleProfile{
		Subject:       profile.Subject,
		Email:         profile.Email,
		EmailVerified: profile.EmailVerified,
		Name:          profile.Name,
		Picture:       profile.Picture,
	}, nil
}

func (googleOAuth *GoogleOAuth) ready() bool {
	return googleOAuth != nil &&
		googleOAuth.config != nil &&
		googleOAuth.config.ClientID != "" &&
		googleOAuth.config.ClientSecret != "" &&
		googleOAuth.config.RedirectURL != ""
}

type googleUserInfo struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}
