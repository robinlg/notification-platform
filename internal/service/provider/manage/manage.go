package manage

import (
	"context"

	"github.com/robinlg/notification-platform/internal/domain"
	"github.com/robinlg/notification-platform/internal/repository"
)

// Service 供应商服务接口
//
//go:generate mockgen -source=./manage.go -destination=../mocks/manage.mock.go -package=providermocks -typed Service
type Service interface {
	// Create 创建供应商
	Create(ctx context.Context, provider domain.Provider) (domain.Provider, error)
}

// providerService 供应商服务实现
type providerService struct {
	repo repository.ProviderRepository
}

// NewProviderService 创建供应商服务
func NewProviderService(repo repository.ProviderRepository) Service {
	return &providerService{
		repo: repo,
	}
}

// Create 创建供应商
func (s *providerService) Create(ctx context.Context, provider domain.Provider) (domain.Provider, error) {
	if err := provider.Validate(); err != nil {
		return domain.Provider{}, err
	}
	return s.repo.Create(ctx, provider)
}
