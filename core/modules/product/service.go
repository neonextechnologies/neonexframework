package product

import (
	"context"
	"fmt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetAll(ctx context.Context) ([]Product, error) {
	return s.repo.FindAll()
}

func (s *Service) GetByID(ctx context.Context, id uint) (*Product, error) {
	entity, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("product not found")
	}
	return entity, nil
}

func (s *Service) Create(ctx context.Context, entity *Product) error {
	return s.repo.Create(entity)
}

func (s *Service) Update(ctx context.Context, id uint, entity *Product) error {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("product not found")
	}

	existing.Name = entity.Name
	existing.Description = entity.Description
	existing.IsActive = entity.IsActive

	return s.repo.Update(existing)
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	entity, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("product not found")
	}
	return s.repo.Delete(entity)
}

func (s *Service) Search(ctx context.Context, query string) ([]Product, error) {
	return s.repo.FindActive()
}
