package notification

import (
	"context"

	"github.com/robinlg/notification-platform/internal/domain"
	"github.com/robinlg/notification-platform/internal/repository"
)

// Service 通知服务接口
//
//go:generate mockgen -source=./notification.go -destination=./mocks/notification.mock.go -package=notificationmocks -typed Service
type Service interface {
	// FindReadyNotifications 准备好调度发送的通知
	FindReadyNotifications(ctx context.Context, offset, limit int) ([]domain.Notification, error)
}

// notificationService 通知服务实现
type notificationService struct {
	repo repository.NotificationRepository
}

// NewNotificationService 创建通知服务实例
func NewNotificationService(repo repository.NotificationRepository) Service {
	return &notificationService{
		repo: repo,
	}
}

// FindReadyNotifications 准备好调度发送的通知
func (s *notificationService) FindReadyNotifications(ctx context.Context, offset, limit int) ([]domain.Notification, error) {
	return s.repo.FindReadyNotifications(ctx, offset, limit)
}
