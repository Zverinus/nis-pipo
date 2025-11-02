package user

import (
	"context"
)

type Service struct{ 
	repo Repository 
}

func NewService(repo Repository) *Service { 
	return &Service{repo: repo} 
}

func (service *Service) Create(ctx context.Context, username, email, passwordHash, firstName, lastName string) (User, error) { 
	return service.repo.Create(ctx, username, email, passwordHash, firstName, lastName) 
}

func (service *Service) GetByID(ctx context.Context, id string) (User, error) { 
	return service.repo.GetByID(ctx, id) 
}

func (service *Service) List(ctx context.Context, limit, offset int) ([]User, error) { 
	return service.repo.List(ctx, limit, offset) 
}