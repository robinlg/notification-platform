package manage

import (
	"context"
	"fmt"

	"github.com/robinlg/notification-platform/internal/domain"
	"github.com/robinlg/notification-platform/internal/errs"
	"github.com/robinlg/notification-platform/internal/repository"
	providersvc "github.com/robinlg/notification-platform/internal/service/provider/manage"
	"github.com/robinlg/notification-platform/internal/service/provider/sms/client"
)

// ChannelTemplateService 提供模版管理的服务接口
//
//go:generate mockgen -source=./manage.go -destination=../mocks/mange.mock.go -package=templatemocks -typed ChannelTemplateService
type ChannelTemplateService interface {
	// 模版相关方法

	// GetTemplateByID 根据ID获取模板
	GetTemplateByID(ctx context.Context, templateID int64) (domain.ChannelTemplate, error)

	// 供应商相关方法

	// GetTemplateByIDAndProviderInfo 根据模板ID和供应商信息获取模板
	GetTemplateByIDAndProviderInfo(ctx context.Context, templateID int64, providerName string, channel domain.Channel) (domain.ChannelTemplate, error)
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

func (t *templateService) GetTemplateByIDAndProviderInfo(ctx context.Context, templateID int64, providerName string, channel domain.Channel) (domain.ChannelTemplate, error) {
	// 1. 获取模板基本信息
	template, err := t.repo.GetTemplateByID(ctx, templateID)
	if err != nil {
		return domain.ChannelTemplate{}, err
	}

	if template.ID == 0 {
		return domain.ChannelTemplate{}, fmt.Errorf("%w: templateID=%d", errs.ErrTemplateNotFound, templateID)
	}

	// 2. 获取指定的版本信息
	version, err := t.repo.GetTemplateVersionByID(ctx, template.ActiveVersionID)
	if err != nil {
		return domain.ChannelTemplate{}, err
	}

	if version.AuditStatus != domain.AuditStatusApproved {
		return domain.ChannelTemplate{}, fmt.Errorf("%w: versionID=%d", errs.ErrTemplateVersionNotApprovedByPlatform, version.ID)
	}

	// 3. 获取指定供应商信息
	providers, err := t.repo.GetProviderByNameAndChannel(ctx, templateID, version.ID, providerName, channel)
	if err != nil {
		return domain.ChannelTemplate{}, err
	}

	if len(providers) == 0 {
		return domain.ChannelTemplate{}, fmt.Errorf("%w: providerName=%s, channel=%s", errs.ErrProviderNotFound, providerName, channel)
	}

	// 4. 组装完整模板
	version.Providers = providers
	template.Versions = []domain.ChannelTemplateVersion{version}

	return template, nil
}
