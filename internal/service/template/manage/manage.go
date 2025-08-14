package manage

import (
	"context"

	"github.com/robinlg/notification-platform/internal/domain"
	"github.com/robinlg/notification-platform/internal/repository"
	providersvc "github.com/robinlg/notification-platform/internal/service/provider/manage"
	"github.com/robinlg/notification-platform/internal/service/provider/sms/client"
)

// ChannelTemplateService 提供模版管理的服务接口
//
//go:generate mockgen -source=./manage.go -destination=../mocks/mange.mock.go -package=templatemocks -typed ChannelTemplateService
type ChannelTemplateService interface {
	// GetTemplateByID 根据ID获取模板
	GetTemplateByID(ctx context.Context, templateID int64) (domain.ChannelTemplate, error)
}

// templateService 实现了ChannelTemplateService接口，提供模板管理的具体实现
type templateService struct {
	repo        repository.ChannelTemplateRepository
	providerSvc providersvc.Service
	smsClients  map[string]client.Client
}

// NewChannelTemplateService 创建模板服务实例
func NewChannelTemplateService(
	repo repository.ChannelTemplateRepository,
	providerSvc providersvc.Service,
	smsClients map[string]client.Client,
) ChannelTemplateService {
	return &templateService{
		repo:        repo,
		providerSvc: providerSvc,
		smsClients:  smsClients,
	}
}

func (t *templateService) GetTemplateByID(ctx context.Context, templateID int64) (domain.ChannelTemplate, error) {
	return t.repo.GetTemplateByID(ctx, templateID)
}
