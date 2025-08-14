package send_strategy

import (
	"context"
	"fmt"

	"github.com/gotomicro/ego/core/elog"
	"github.com/robinlg/notification-platform/internal/domain"
	"github.com/robinlg/notification-platform/internal/repository"
	configsvc "github.com/robinlg/notification-platform/internal/service/config"
)

// DefaultSendStrategy 延迟发送策略
type DefaultSendStrategy struct {
	repo      repository.NotificationRepository
	configSvc configsvc.BusinessConfigService
	logger    *elog.Component
}

// NewDefaultStrategy 创建延迟发送策略
func NewDefaultStrategy(repo repository.NotificationRepository, configSvc configsvc.BusinessConfigService) *DefaultSendStrategy {
	return &DefaultSendStrategy{
		repo:      repo,
		configSvc: configSvc,
		logger:    elog.DefaultLogger,
	}
}

// Send 单条发送通知
func (s *DefaultSendStrategy) Send(ctx context.Context, notification domain.Notification) (domain.SendResponse, error) {
	notification.SetSendTime()
	// 创建通知记录
	created, err := s.create(ctx, notification)
	if err != nil {
		return domain.SendResponse{}, fmt.Errorf("创建延迟通知失败: %w", err)
	}

	return domain.SendResponse{
		NotificationID: created.ID,
		Status:         created.Status,
	}, nil
}

func (s *DefaultSendStrategy) create(ctx context.Context, notification domain.Notification) (domain.Notification, error) {
	if !s.needCreateCallbackLog(ctx, notification) {
		return s.repo.CreateWithCallbackLog(ctx, notification)
	}
	return s.repo.Create(ctx, notification)
}

func (s *DefaultSendStrategy) needCreateCallbackLog(ctx context.Context, notification domain.Notification) bool {
	bizConfig, err := s.configSvc.GetByID(ctx, notification.BizID)
	if err != nil {
		s.logger.Error("查找 biz config 失败", elog.FieldErr(err))
		return false
	}
	return bizConfig.CallbackConfig != nil
}
