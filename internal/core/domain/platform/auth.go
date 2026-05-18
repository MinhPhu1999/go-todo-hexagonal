package platform

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidAuthRequest       = errors.New("invalid auth request")
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrEmailAlreadyExists       = errors.New("email already exists")
	ErrUnauthorized             = errors.New("unauthorized")
	ErrGoogleAuthUnavailable    = errors.New("google auth unavailable")
	ErrGoogleEmailNotVerified   = errors.New("google email is not verified")
	ErrGoogleInvalidOAuthState  = errors.New("invalid google oauth state")
	ErrGoogleProfileUnavailable = errors.New("google profile unavailable")
)

type UserID string

func ParseUserID(value string) (UserID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%w: user id is required", ErrInvalidAuthRequest)
	}
	return UserID(value), nil
}

func (id UserID) String() string {
	return string(id)
}

type User struct {
	ID           UserID     `json:"id"`
	Email        string     `json:"email"`
	Name         string     `json:"name"`
	Picture      string     `json:"picture,omitempty"`
	PasswordHash string     `json:"-"`
	GoogleID     string     `json:"-"`
	Providers    []string   `json:"providers"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
}

func NewEmailUser(id UserID, email string, name string, passwordHash string, now time.Time) (*User, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: user id is required", ErrInvalidAuthRequest)
	}
	email, err := normalizeEmail(email)
	if err != nil {
		return nil, err
	}
	name = normalizeName(name, email)
	if passwordHash == "" {
		return nil, fmt.Errorf("%w: password hash is required", ErrInvalidAuthRequest)
	}

	now = now.UTC()
	return &User{
		ID:           id,
		Email:        email,
		Name:         name,
		PasswordHash: passwordHash,
		Providers:    []string{"email"},
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func NewGoogleUser(id UserID, profile GoogleProfile, now time.Time) (*User, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: user id is required", ErrInvalidAuthRequest)
	}
	if profile.Subject == "" {
		return nil, fmt.Errorf("%w: google subject is required", ErrInvalidAuthRequest)
	}
	email, err := normalizeEmail(profile.Email)
	if err != nil {
		return nil, err
	}

	now = now.UTC()
	return &User{
		ID:        id,
		Email:     email,
		Name:      normalizeName(profile.Name, email),
		Picture:   strings.TrimSpace(profile.Picture),
		GoogleID:  profile.Subject,
		Providers: []string{"google"},
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func RehydrateUser(
	id UserID,
	email string,
	name string,
	picture string,
	passwordHash string,
	googleID string,
	providers []string,
	createdAt time.Time,
	updatedAt time.Time,
	lastLoginAt *time.Time,
) (*User, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: user id is required", ErrInvalidAuthRequest)
	}
	email, err := normalizeEmail(email)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:           id,
		Email:        email,
		Name:         normalizeName(name, email),
		Picture:      strings.TrimSpace(picture),
		PasswordHash: passwordHash,
		GoogleID:     strings.TrimSpace(googleID),
		Providers:    normalizeProviders(providers),
		CreatedAt:    createdAt.UTC(),
		UpdatedAt:    updatedAt.UTC(),
		LastLoginAt:  utcTimePointer(lastLoginAt),
	}, nil
}

func (user *User) AttachGoogleProfile(profile GoogleProfile, now time.Time) error {
	if profile.Subject == "" {
		return fmt.Errorf("%w: google subject is required", ErrInvalidAuthRequest)
	}
	email, err := normalizeEmail(profile.Email)
	if err != nil {
		return err
	}
	if user.Email != email {
		return fmt.Errorf("%w: google email does not match user email", ErrInvalidAuthRequest)
	}

	user.GoogleID = profile.Subject
	user.Name = normalizeName(profile.Name, user.Email)
	user.Picture = strings.TrimSpace(profile.Picture)
	user.addProvider("google")
	user.UpdatedAt = now.UTC()
	return nil
}

func (user *User) HasPassword() bool {
	return user.PasswordHash != ""
}

func (user *User) MarkLoggedIn(now time.Time) {
	loginAt := now.UTC()
	user.LastLoginAt = &loginAt
	user.UpdatedAt = loginAt
}

func (user *User) addProvider(provider string) {
	provider = strings.TrimSpace(provider)
	if provider == "" {
		return
	}
	for _, current := range user.Providers {
		if current == provider {
			return
		}
	}
	user.Providers = append(user.Providers, provider)
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	TokenType   string    `json:"token_type"`
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
	User        *User     `json:"user"`
}

type AuthToken struct {
	Token     string
	ExpiresAt time.Time
}

type TokenClaims struct {
	UserID    UserID
	Email     string
	Name      string
	Issuer    string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

type GoogleProfile struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
}

func ValidatePassword(value string) error {
	if len(value) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", ErrInvalidAuthRequest)
	}
	if len(value) > 72 {
		return fmt.Errorf("%w: password must be at most 72 characters", ErrInvalidAuthRequest)
	}
	return nil
}

func NormalizeEmail(value string) (string, error) {
	return normalizeEmail(value)
}

func normalizeEmail(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "", fmt.Errorf("%w: email is required", ErrInvalidAuthRequest)
	}
	if !strings.Contains(value, "@") {
		return "", fmt.Errorf("%w: email is invalid", ErrInvalidAuthRequest)
	}
	return value, nil
}

func normalizeName(value string, email string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	index := strings.Index(email, "@")
	if index > 0 {
		return email[:index]
	}
	return email
}

func normalizeProviders(providers []string) []string {
	seen := make(map[string]struct{}, len(providers))
	normalized := make([]string, 0, len(providers))
	for _, provider := range providers {
		provider = strings.TrimSpace(provider)
		if provider == "" {
			continue
		}
		if _, ok := seen[provider]; ok {
			continue
		}
		seen[provider] = struct{}{}
		normalized = append(normalized, provider)
	}
	return normalized
}

func utcTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}
