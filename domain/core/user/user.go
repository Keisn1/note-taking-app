package user

import (
	"github.com/google/uuid"
)

type User struct {
	ID uuid.UUID
}

type Service interface {
	QueryByID(userID uuid.UUID) (User, error)
}

type UserSvc struct{}

func (us UserSvc) QueryByID(userID uuid.UUID) (User, error) {
	return User{}, nil
}