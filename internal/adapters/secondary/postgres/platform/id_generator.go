package platform

import (
	domain "go-crud-db-p2/internal/core/domain/platform"

	"github.com/google/uuid"
)

type UUIDGenerator struct{}

func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

func (UUIDGenerator) NewID() domain.TodoID {
	return domain.TodoID(uuid.New().String())
}

func (UUIDGenerator) NewUserID() domain.UserID {
	return domain.UserID(uuid.New().String())
}
