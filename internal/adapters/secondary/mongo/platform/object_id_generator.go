package platform

import (
	domain "go-crud-db-p2/internal/core/domain/platform"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ObjectIDGenerator struct{}

func NewObjectIDGenerator() *ObjectIDGenerator {
	return &ObjectIDGenerator{}
}

func (ObjectIDGenerator) NewID() domain.TodoID {
	return domain.TodoID(bson.NewObjectID().Hex())
}

func (ObjectIDGenerator) NewUserID() domain.UserID {
	return domain.UserID(bson.NewObjectID().Hex())
}
