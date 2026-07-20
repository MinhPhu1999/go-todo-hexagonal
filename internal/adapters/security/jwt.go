package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-crud-db-p2/config"
	domain "go-crud-db-p2/internal/core/domain/platform"
)

type JWTManager struct {
	secret    []byte
	issuer    string
	expiresIn time.Duration
}

func NewJWTManager(cfg config.Config) *JWTManager {
	secret := cfg.JWT.Secret
	issuer := cfg.JWT.Issuer
	expiresIn := cfg.JWT.ExpiresIn

	if expiresIn <= 0 {
		expiresIn = 24 * time.Hour
	}
	if issuer == "" {
		issuer = "go-crud-db-p2"
	}

	return &JWTManager{
		secret:    []byte(secret),
		issuer:    issuer,
		expiresIn: expiresIn,
	}
}

func (manager *JWTManager) IssueToken(user *domain.User) (domain.AuthToken, error) {
	if user == nil {
		return domain.AuthToken{}, fmt.Errorf("%w: user is required", domain.ErrUnauthorized)
	}
	if len(manager.secret) < 16 {
		return domain.AuthToken{}, fmt.Errorf("%w: jwt secret is too short", domain.ErrUnauthorized)
	}

	now := time.Now().UTC()
	expiresAt := now.Add(manager.expiresIn)
	claims := jwtClaims{
		Subject:   user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		Issuer:    manager.issuer,
		IssuedAt:  now.Unix(),
		ExpiresAt: expiresAt.Unix(),
	}

	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return domain.AuthToken{}, err
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return domain.AuthToken{}, err
	}

	unsigned := encodeSegment(headerJSON) + "." + encodeSegment(claimsJSON)
	signature := manager.sign(unsigned)

	return domain.AuthToken{
		Token:     unsigned + "." + signature,
		ExpiresAt: expiresAt,
	}, nil
}

func (manager *JWTManager) VerifyToken(token string) (domain.TokenClaims, error) {
	if len(manager.secret) < 16 {
		return domain.TokenClaims{}, domain.ErrUnauthorized
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return domain.TokenClaims{}, domain.ErrUnauthorized
	}

	unsigned := parts[0] + "." + parts[1]
	expectedSignature := manager.sign(unsigned)
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[2])) {
		return domain.TokenClaims{}, domain.ErrUnauthorized
	}

	payload, err := decodeSegment(parts[1])
	if err != nil {
		return domain.TokenClaims{}, domain.ErrUnauthorized
	}

	var claims jwtClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return domain.TokenClaims{}, domain.ErrUnauthorized
	}
	if claims.Issuer != manager.issuer {
		return domain.TokenClaims{}, domain.ErrUnauthorized
	}

	expiresAt := time.Unix(claims.ExpiresAt, 0).UTC()
	if time.Now().UTC().After(expiresAt) {
		return domain.TokenClaims{}, domain.ErrUnauthorized
	}

	userID, err := domain.ParseUserID(claims.Subject)
	if err != nil {
		return domain.TokenClaims{}, domain.ErrUnauthorized
	}

	return domain.TokenClaims{
		UserID:    userID,
		Email:     claims.Email,
		Name:      claims.Name,
		Issuer:    claims.Issuer,
		IssuedAt:  time.Unix(claims.IssuedAt, 0).UTC(),
		ExpiresAt: expiresAt,
	}, nil
}

func (manager *JWTManager) sign(value string) string {
	mac := hmac.New(sha256.New, manager.secret)
	_, _ = mac.Write([]byte(value))
	return encodeSegment(mac.Sum(nil))
}

func encodeSegment(value []byte) string {
	return base64.RawURLEncoding.EncodeToString(value)
}

func decodeSegment(value string) ([]byte, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, errors.New("invalid jwt segment")
	}
	return decoded, nil
}

type jwtClaims struct {
	Subject   string `json:"sub"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Issuer    string `json:"iss"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}
