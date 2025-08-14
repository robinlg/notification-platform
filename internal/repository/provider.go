package repository

import (
	"context"

	"github.com/robinlg/notification-platform/internal/domain"
	"github.com/robinlg/notification-platform/internal/repository/dao"
)

// ProviderRepository 供应商仓储接口
type ProviderRepository interface {
	// Create 创建供应商
	Create(ctx context.Context, provider domain.Provider) (domain.Provider, error)
}

type providerRepository struct {
	dao dao.ProviderDAO
}

func NewProviderRepository(d dao.ProviderDAO) ProviderRepository {
	return &providerRepository{dao: d}
}

func (p *providerRepository) Create(ctx context.Context, provider domain.Provider) (domain.Provider, error) {
	created, err := p.dao.Create(ctx, p.toEntity(provider))
	if err != nil {
		return domain.Provider{}, err
	}
	return p.toDomain(created), nil
}

func (p *providerRepository) toDomain(d dao.Provider) domain.Provider {
	return domain.Provider{
		ID:               d.ID,
		Name:             d.Name,
		Channel:          domain.Channel(d.Channel),
		Endpoint:         d.Endpoint,
		APIKey:           d.APIKey,
		APISecret:        d.APISecret,
		Weight:           d.Weight,
		QPSLimit:         d.QPSLimit,
		DailyLimit:       d.DailyLimit,
		AuditCallbackURL: d.AuditCallbackURL,
		Status:           domain.ProviderStatus(d.Status),
	}
}

func (p *providerRepository) toEntity(provider domain.Provider) dao.Provider {
	daoProvider := dao.Provider{
		ID:               provider.ID,
		Name:             provider.Name,
		Channel:          provider.Channel.String(),
		Endpoint:         provider.Endpoint,
		APIKey:           provider.APIKey,
		APISecret:        provider.APISecret,
		Weight:           provider.Weight,
		QPSLimit:         provider.QPSLimit,
		DailyLimit:       provider.DailyLimit,
		AuditCallbackURL: provider.AuditCallbackURL,
		Status:           provider.Status.String(),
	}
	return daoProvider
}
