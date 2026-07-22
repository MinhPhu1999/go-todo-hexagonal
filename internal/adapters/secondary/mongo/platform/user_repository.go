package platform

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain "go-crud-db-p2/internal/core/domain/platform"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(database *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: database.Collection("users"),
	}
}

func (repository *UserRepository) Save(ctx context.Context, user *domain.User) error {
	document, err := newUserDocument(user)
	if err != nil {
		return err
	}

	_, err = repository.collection.ReplaceOne(
		ctx,
		bson.M{"_id": document.ID},
		document,
		options.Replace().SetUpsert(true),
	)
	if mongo.IsDuplicateKeyError(err) {
		return domain.ErrEmailAlreadyExists
	}
	return err
}

func (repository *UserRepository) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	objectID, err := userObjectID(id)
	if err != nil {
		return nil, err
	}
	return repository.findOne(ctx, bson.M{"_id": objectID})
}

func (repository *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	normalized, err := domain.NormalizeEmail(email)
	if err != nil {
		return nil, err
	}
	return repository.findOne(ctx, bson.M{"email": normalized})
}

func (repository *UserRepository) FindByGoogleID(ctx context.Context, googleID string) (*domain.User, error) {
	if googleID == "" {
		return nil, domain.ErrUnauthorized
	}
	return repository.findOne(ctx, bson.M{"google_id": googleID})
}

func (repository *UserRepository) findOne(ctx context.Context, filter bson.M) (*domain.User, error) {
	var document userDocument
	err := repository.collection.FindOne(ctx, filter).Decode(&document)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}

	return document.toDomain()
}

type userDocument struct {
	ID                 bson.ObjectID `bson:"_id"`
	Email              string        `bson:"email"`
	Name               string        `bson:"name"`
	Picture            string        `bson:"picture,omitempty"`
	PasswordHash       string        `bson:"password_hash,omitempty"`
	GoogleID           string        `bson:"google_id,omitempty"`
	Providers          []string      `bson:"providers"`
	CreatedAt          time.Time     `bson:"created_at"`
	UpdatedAt          time.Time     `bson:"updated_at"`
	LastLoginAt        *time.Time    `bson:"last_login_at,omitempty"`
	AccountLockedUntil *time.Time    `bson:"account_locked_until,omitempty"`
	ResetOTPHash       string        `bson:"reset_otp_hash,omitempty"`
	ResetOTPExpiresAt  *time.Time    `bson:"reset_otp_expires_at,omitempty"`
}

func newUserDocument(user *domain.User) (userDocument, error) {
	if user == nil {
		return userDocument{}, fmt.Errorf("%w: user is required", domain.ErrInvalidAuthRequest)
	}

	objectID, err := userObjectID(user.ID)
	if err != nil {
		return userDocument{}, err
	}

	return userDocument{
		ID:                 objectID,
		Email:              user.Email,
		Name:               user.Name,
		Picture:            user.Picture,
		PasswordHash:       user.PasswordHash,
		GoogleID:           user.GoogleID,
		Providers:          user.Providers,
		CreatedAt:          user.CreatedAt,
		UpdatedAt:          user.UpdatedAt,
		LastLoginAt:        user.LastLoginAt,
		AccountLockedUntil: user.AccountLockedUntil,
		ResetOTPHash:       user.ResetOTPHash,
		ResetOTPExpiresAt:  user.ResetOTPExpiresAt,
	}, nil
}

func (document userDocument) toDomain() (*domain.User, error) {
	return domain.RehydrateUser(
		domain.UserID(document.ID.Hex()),
		document.Email,
		document.Name,
		document.Picture,
		document.PasswordHash,
		document.GoogleID,
		document.Providers,
		document.CreatedAt,
		document.UpdatedAt,
		document.LastLoginAt,
		document.AccountLockedUntil,
		document.ResetOTPHash,
		document.ResetOTPExpiresAt,
	)
}

func userObjectID(id domain.UserID) (bson.ObjectID, error) {
	objectID, err := bson.ObjectIDFromHex(id.String())
	if err != nil {
		return bson.NilObjectID, fmt.Errorf("%w: user id must be a mongodb object id", domain.ErrInvalidAuthRequest)
	}
	return objectID, nil
}
