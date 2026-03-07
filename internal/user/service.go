package user

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailExists  = errors.New("email already exists")
	ErrInvalidCreds = errors.New("invalid email or password")
	ErrNotFound     = errors.New("not found")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByID(ctx context.Context, id string) (User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetByEmail(ctx context.Context, email string) (User, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *Service) Register(ctx context.Context, email, password string) (User, error) {
	_, err := s.repo.GetByEmail(ctx, email)
	if err == nil {
		return User{}, ErrEmailExists
	}
	if !errors.Is(err, ErrNotFound) {
		return User{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	username := email
	if len(username) > 255 {
		username = username[:255]
	}
	return s.repo.Create(ctx, username, email, string(hash), "", "")
}

func (s *Service) Login(ctx context.Context, email, password string) (User, error) {
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return User{}, ErrInvalidCreds
		}
		return User{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return User{}, ErrInvalidCreds
	}
	return u, nil
}