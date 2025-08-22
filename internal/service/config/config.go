package config

import (
	"context"
	"errors"
	"fmt"

	"github.com/ego-component/egorm"
	"github.com/robinlg/notification-platform/internal/domain"
	"github.com/robinlg/notification-platform/internal/errs"
	"github.com/robinlg/notification-platform/internal/repository"
)

//go:generate mockgen -source=./config.go -destination=./mocks/config.mock.go -package=configmocks -typed BusinessConfigService
type BusinessConfigService interface {
	GetByID(ctx context.Context, id int64) (domain.BusinessConfig, error)
	GetByIDs(ctx context.Context, ids []int64) (map[int64]domain.BusinessConfig, error)
}

type BusinessConfigServiceV1 struct {
	repo repository.BusinessConfigRepository
}

// NewBusinessConfigService 创建业务配置服务实例
func NewBusinessConfigService(repo repository.BusinessConfigRepository) BusinessConfigService {
	return &BusinessConfigServiceV1{
		repo: repo,
	}
}

// GetByID 根据ID获取单个业务配置
func (b *BusinessConfigServiceV1) GetByID(ctx context.Context, id int64) (domain.BusinessConfig, error) {
	// 参数校验
	if id <= 0 {
		return domain.BusinessConfig{}, fmt.Errorf("%w", errs.ErrInvalidParameter)
	}

	// 调用仓库层方法
	config, err := b.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, egorm.ErrRecordNotFound) {
			return domain.BusinessConfig{}, fmt.Errorf("%w", errs.ErrConfigNotFound)
		}
		return domain.BusinessConfig{}, err
	}

	return config, nil
}

// GetByIDs 根据多个ID批量获取业务配置
func (b *BusinessConfigServiceV1) GetByIDs(ctx context.Context, ids []int64) (map[int64]domain.BusinessConfig, error) {
	// 参数校验
	if len(ids) == 0 {
		return make(map[int64]domain.BusinessConfig), nil
	}

	// 调用仓库层方法
	return b.repo.GetByIDs(ctx, ids)
}
