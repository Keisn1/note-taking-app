package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidPassword = errors.New("invalid password")
)

type Service interface {
	QueryByID(ctx context.Context, userID uuid.UUID) (User, error)
	Create(ctx context.Context, nu UpdateUser) (User, error)
	Update(ctx context.Context, u User, newU UpdateUser) (User, error)
}

type Svc struct {
	repo Repo
}

func NewSvc(repo Repo) Service {
	return Svc{repo: repo}
}

func (s Svc) Update(ctx context.Context, u User, newU UpdateUser) (User, error) {
	if !newU.Name.IsEmpty() {
		u.Name = newU.Name
	}

	if !newU.Email.IsEmpty() {
		u.Email = newU.Email
	}

	return u, nil
}

func (s Svc) Create(ctx context.Context, newU UpdateUser) (User, error) {
	if len(newU.Password) == 0 {
		return User{}, fmt.Errorf("create: %w", ErrInvalidPassword)
	}

	pwHash, err := bcrypt.GenerateFromPassword([]byte(newU.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("create: %w: %w", ErrInvalidPassword, err)
	}

	u := User{
		ID:           uuid.New(),
		Name:         newU.Name,
		Email:        newU.Email,
		PasswordHash: pwHash,
	}

	s.repo.Create(ctx, u)
	return u, nil
}

func (s Svc) QueryByID(ctx context.Context, userID uuid.UUID) (User, error) {
	u, err := s.repo.QueryByID(ctx, userID)
	if err != nil {
		return User{}, fmt.Errorf("queryByID: %w", err)
	}
	return u, nil
}

type Repo struct {
	users map[uuid.UUID]User
}

func (r Repo) Create(ctx context.Context, u User) error {
	r.users[u.ID] = u
	return nil
}

func (r Repo) QueryByID(ctx context.Context, userID uuid.UUID) (User, error) {
	if user, ok := r.users[userID]; ok {
		return user, nil
	}
	return User{}, errors.New("user not found")
}

func NewRepo(users []User) Repo {
	us := make(map[uuid.UUID]User)
	for _, u := range users {
		us[u.ID] = u
	}
	return Repo{users: us}
}
