package channel

import (
	"context"
	"fmt"

	"github.com/robinlg/notification-platform/internal/domain"
	"github.com/robinlg/notification-platform/internal/errs"
)

// Channel 渠道接口
//
//go:generate mockgen -source=./channel.go -destination=./mocks/channel.mock.go -package=channelmocks -typed Channel
type Channel interface {
	// Send 发送通知
	Send(ctx context.Context, notification domain.Notification) (domain.SendResponse, error)
}

// Dispatcher 渠道分发器，对外伪装成Channel，作为统一入口
type Dispatcher struct {
	channels map[domain.Channel]Channel
}

// NewDispatcher 创建渠道分发器
func NewDispatcher(channels map[domain.Channel]Channel) *Dispatcher {
	return &Dispatcher{
		channels: channels,
	}
}

func (d *Dispatcher) Send(ctx context.Context, notification domain.Notification) (domain.SendResponse, error) {
	channel, ok := d.channels[notification.Channel]
	if !ok {
		return domain.SendResponse{}, fmt.Errorf("%w: %s", errs.ErrNoAvailableChannel, notification.Channel)
	}
	return channel.Send(ctx, notification)
}
