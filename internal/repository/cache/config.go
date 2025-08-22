package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/robinlg/notification-platform/internal/domain"
)

var ErrKeyNotFound = errors.New("key not found")

const (
	ConfigPrefix       = "config"
	DefaultExpiredTime = 10 * time.Minute
)

type ConfigCache interface {
	Get(ctx context.Context, bizID int64) (domain.BusinessConfig, error)
	Set(ctx context.Context, cfg domain.BusinessConfig) error
	GetConfigs(ctx context.Context, bizIDs []int64) (map[int64]domain.BusinessConfig, error)
	SetConfigs(ctx context.Context, configs []domain.BusinessConfig) error
}

func ConfigKey(bizID int64) string {
	return fmt.Sprintf("%s:%d", ConfigPrefix, bizID)
}
