package security

import (
	"testing"
	"time"

	"go-crud-db-p2/config"
	domain "go-crud-db-p2/internal/core/domain/platform"
)

func TestJWTManagerIssueAndVerifyToken(t *testing.T) {
	manager := NewJWTManager(
		config.JWTSecret("test-secret-with-enough-length"),
		config.JWTIssuer("go-crud-db-p2-test"),
		config.JWTExpiresIn(time.Hour),
	)

	user := &domain.User{
		ID:    domain.UserID("665000000000000000000201"),
		Email: "demo@example.com",
		Name:  "Demo User",
	}

	token, err := manager.IssueToken(user)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}
	if token.Token == "" {
		t.Fatal("IssueToken() returned an empty token")
	}

	claims, err := manager.VerifyToken(token.Token)
	if err != nil {
		t.Fatalf("VerifyToken() error = %v", err)
	}
	if claims.UserID != user.ID {
		t.Fatalf("UserID = %s, want %s", claims.UserID, user.ID)
	}
	if claims.Email != user.Email {
		t.Fatalf("Email = %s, want %s", claims.Email, user.Email)
	}
}
