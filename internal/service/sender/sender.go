package sender

import (
	"context"

	"github.com/ecodeclub/ekit/pool"
	"github.com/gotomicro/ego/core/elog"
	"github.com/robinlg/notification-platform/internal/domain"
	"github.com/robinlg/notification-platform/internal/repository"
	"github.com/robinlg/notification-platform/internal/service/channel"
	configsvc "github.com/robinlg/notification-platform/internal/service/config"
	"github.com/robinlg/notification-platform/internal/service/notification/callback"
)

// NotificationSender 通知发送接口
//
//go:generate mockgen -source=./sender.go -destination=./mocks/sender.mock.go -package=sendermocks -typed NotificationSender
type NotificationSender interface {
	// Send 单条发送通知
	Send(ctx context.Context, notification domain.Notification) (domain.SendResponse, error)
}

// sender 通知发送器实现
type sender struct {
	repo        repository.NotificationRepository
	configSvc   configsvc.BusinessConfigService
	callbackSvc callback.Service
	channel     channel.Channel
	taskPool    pool.TaskPool

	logger *elog.Component
}

// NewSender 创建通知发送器
func NewSender(
	repo repository.NotificationRepository,
	configSvc configsvc.BusinessConfigService,
	callbackSvc callback.Service,
	channel channel.Channel,
	taskPool pool.TaskPool,
) NotificationSender {
	return &sender{
		repo:        repo,
		configSvc:   configSvc,
		callbackSvc: callbackSvc,
		channel:     channel,
		taskPool:    taskPool,
		logger:      elog.DefaultLogger,
	}
}

// Send 单条发送通知
func (d *sender) Send(ctx context.Context, notification domain.Notification) (domain.SendResponse, error) {
	resp := domain.SendResponse{
		NotificationID: notification.ID,
	}
	_, err := d.channel.Send(ctx, notification)
	if err != nil {
		d.logger.Error("发送失败 %w", elog.FieldErr(err))
		resp.Status = domain.SendStatusFailed
		notification.Status = domain.SendStatusFailed
		// 如果是FAILED，需要把quota加回去
		err = d.repo.MarkFailed(ctx, notification)
	} else {
		resp.Status = domain.SendStatusSucceeded
		notification.Status = domain.SendStatusSucceeded
		err = d.repo.MarkSuccess(ctx, notification)
	}

	// 更新发送状态
	if err != nil {
		return domain.SendResponse{}, err
	}

	// 得到准确的发送结果，发起回调，发送成功和失败都应该回调
	_ = d.callbackSvc.SendCallbackByNotification(ctx, notification)

	return resp, nil
}
