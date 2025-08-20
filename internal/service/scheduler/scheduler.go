package scheduler

import (
	"context"
	"time"

	"github.com/meoying/dlock-go"
	"github.com/robinlg/notification-platform/internal/pkg/loopjob"

	notificationsvc "github.com/robinlg/notification-platform/internal/service/notification"
	"github.com/robinlg/notification-platform/internal/service/sender"
)

// NotificationScheduler 通知调度服务接口
type NotificationScheduler interface {
	// Start 启动调度服务
	Start(ctx context.Context)
}

// staticScheduler 通知调度服务实现
type staticScheduler struct {
	notificationSvc notificationsvc.Service
	sender          sender.NotificationSender
	dclient         dlock.Client

	batchSize int
}

// NewScheduler 创建通知调度服务
func NewScheduler(
	notificationSvc notificationsvc.Service,
	dispatcher sender.NotificationSender,
	dclient dlock.Client,
) NotificationScheduler {
	const defaultBatchSize = 10
	return &staticScheduler{
		notificationSvc: notificationSvc,
		sender:          dispatcher,
		batchSize:       defaultBatchSize,
		dclient:         dclient,
	}
}

// Start 启动调度服务
// 当 ctx 被取消的或者关闭的时候，就会结束循环
func (s *staticScheduler) Start(ctx context.Context) {
	const key = "notification_platform_async_scheduler"
	lj := loopjob.NewInfiniteLoop(s.dclient, s.processPendingNotifications, key)
	lj.Run(ctx)
}

// processPendingNotifications 处理待发送的通知
func (s *staticScheduler) processPendingNotifications(ctx context.Context) error {
	const defaultTimeout = 3 * time.Second
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	const offset = 0
	notifications, err := s.notificationSvc.FindReadyNotifications(ctx, offset, s.batchSize)
	if err != nil {
		return err
	}
	if len(notifications) == 0 {
		time.Sleep(time.Second)
		return nil
	}
	_, err = s.sender.BatchSend(ctx, notifications)
	return err
}
