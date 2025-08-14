package send_strategy

import (
	"context"

	"github.com/robinlg/notification-platform/internal/domain"
)

// SendStrategy 发送策略接口
//
//go:generate mockgen -source=./types.go -destination=./mocks/send_strategy.mock.go -package=sendstrategymocks -typed SendStrategy
type SendStrategy interface {
	// Send 单条发送通知
	Send(ctx context.Context, notification domain.Notification) (domain.SendResponse, error)
}

// Dispatcher 通知发送分发器
// 根据通知的策略类型选择合适的发送策略
type Dispatcher struct {
	immediate       *ImmediateSendStrategy
	defaultStrategy *DefaultSendStrategy
}

// NewDispatcher 创建通知发送分发器
func NewDispatcher(
	immediate *ImmediateSendStrategy,
	defaultStrategy *DefaultSendStrategy,
) SendStrategy {
	return &Dispatcher{
		immediate:       immediate,
		defaultStrategy: defaultStrategy,
	}
}

// Send 发送通知
func (d *Dispatcher) Send(ctx context.Context, notification domain.Notification) (domain.SendResponse, error) {
	// 执行发送
	return d.selectStrategy(notification).Send(ctx, notification)
}

func (d *Dispatcher) selectStrategy(not domain.Notification) SendStrategy {
	if not.SendStrategyConfig.Type == domain.SendStrategyImmediate {
		return d.immediate
	}
	return d.defaultStrategy
}
