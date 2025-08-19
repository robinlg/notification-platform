package cache

import (
	"context"

	"github.com/robinlg/notification-platform/internal/domain"
)

type IncrItem struct {
	BizID   int64
	Channel domain.Channel
	Val     int32
}

type QuotaCache interface {
	Incr(ctx context.Context, bizID int64, channel domain.Channel, quota int32) error
	Decr(ctx context.Context, bizID int64, channel domain.Channel, quota int32) error
	MutiIncr(ctx context.Context, items []IncrItem) error
	MutiDecr(ctx context.Context, items []IncrItem) error
}
