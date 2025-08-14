package send_strategy

import (
	"context"
	"errors"
	"fmt"

	"github.com/robinlg/notification-platform/internal/domain"
	"github.com/robinlg/notification-platform/internal/errs"
	"github.com/robinlg/notification-platform/internal/repository"
	"github.com/robinlg/notification-platform/internal/service/sender"
)

// ImmediateSendStrategy 立即发送策略
// 同步立刻发送，异步接口选择了立即发送策略也不会生效。
type ImmediateSendStrategy struct {
	repo   repository.NotificationRepository
	sender sender.NotificationSender
}

// NewImmediateStrategy 创建立即发送策略
func NewImmediateStrategy(repo repository.NotificationRepository, sender sender.NotificationSender) *ImmediateSendStrategy {
	return &ImmediateSendStrategy{
		repo:   repo,
		sender: sender,
	}
}

func (s *ImmediateSendStrategy) Send(ctx context.Context, notification domain.Notification) (domain.SendResponse, error) {
	notification.SetSendTime()
	created, err := s.repo.Create(ctx, notification)
	if err == nil {
		// 立即发送
		return s.sender.Send(ctx, notification)
	}

	// 非唯一索引冲突直接返回错误
	if !errors.Is(err, errs.ErrNotificationDuplicate) {
		return domain.SendResponse{}, fmt.Errorf("创建通知失败: %w", err)
	}

	// 唯一索引冲突表示业务方重试
	found, err := s.repo.GetByKey(ctx, created.BizID, created.Key)
	if err != nil {
		return domain.SendResponse{}, fmt.Errorf("获取通知失败: %w", err)
	}

	if found.Status == domain.SendStatusSucceeded {
		return domain.SendResponse{
			NotificationID: found.ID,
			Status:         found.Status,
		}, nil
	}

	if found.Status == domain.SendStatusSending {
		return domain.SendResponse{}, fmt.Errorf("发送失败 %w", errs.ErrSendNotificationFailed)
	}

	// 更新通知状态为SENDING同时获取乐观锁（版本号）
	found.Status = domain.SendStatusSending
	err = s.repo.CASStatus(ctx, found)
	if err != nil {
		return domain.SendResponse{}, fmt.Errorf("并发竞争失败: %w", err)
	}
	found.Version++
	// 再次立即发送
	return s.sender.Send(ctx, found)
}
