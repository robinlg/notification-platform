package notification

import (
	"context"
	"fmt"

	"github.com/robinlg/notification-platform/internal/domain"
	"github.com/robinlg/notification-platform/internal/errs"
	"github.com/robinlg/notification-platform/internal/service/template/manage"

	idgen "github.com/robinlg/notification-platform/internal/pkg/id_generator"
	sendstrategy "github.com/robinlg/notification-platform/internal/service/send_strategy"
)

// SendService 负责处理发送
//
//go:generate mockgen -source=./send_notification.go -destination=./mocks/send_notification.mock.go -package=notificationmocks -typed SendService
type SendService interface {
	// SendNotification 同步单条发送
	SendNotification(ctx context.Context, n domain.Notification) (domain.SendResponse, error)
	// SendNotificationAsync 异步单条发送
	SendNotificationAsync(ctx context.Context, n domain.Notification) (domain.SendResponse, error)
}

// sendService 执行器实现
type sendService struct {
	notificationSvc Service
	templateSvc     manage.ChannelTemplateService
	idGenerator     *idgen.Generator
	sendStrategy    sendstrategy.SendStrategy
}

// NewSendService 创建执行器实例
func NewSendService(templateSvc manage.ChannelTemplateService, notificationSvc Service, sendStrategy sendstrategy.SendStrategy) SendService {
	return &sendService{
		notificationSvc: notificationSvc,
		templateSvc:     templateSvc,
		idGenerator:     idgen.NewGenerator(),
		sendStrategy:    sendStrategy,
	}
}

// SendNotification 同步单条发送
func (e *sendService) SendNotification(ctx context.Context, n domain.Notification) (domain.SendResponse, error) {
	resp := domain.SendResponse{
		Status: domain.SendStatusFailed,
	}

	// 参数校验
	if err := n.Validate(); err != nil {
		return resp, err
	}

	// 生成通知ID(后续考虑分库分表)
	id := e.idGenerator.GenerateID(n.BizID, n.Key)
	n.ID = uint64(id)

	// 发送通知
	response, err := e.sendStrategy.Send(ctx, n)
	if err != nil {
		// 通用的发送失败错误
		return resp, fmt.Errorf("%w, 发送通知失败，原因：%w", errs.ErrSendNotificationFailed, err)
	}

	return response, nil
}

// SendNotificationAsync 异步单条发送
func (e *sendService) SendNotificationAsync(ctx context.Context, n domain.Notification) (domain.SendResponse, error) {
	// 参数校验
	if err := n.Validate(); err != nil {
		return domain.SendResponse{}, err
	}
	// 生成通知ID
	id := e.idGenerator.GenerateID(n.BizID, n.Key)
	n.ID = uint64(id)

	// 使用异步接口但要立即发送，修改为延时发送
	// 本质上这是一个不怎好的用法，但是业务方可能不清楚，所以我们兼容一下
	n.ReplaceAsyncImmediate()
	return e.sendStrategy.Send(ctx, n)
}
