package local

import (
	"context"
	"errors"
	"time"

	"github.com/gotomicro/ego/core/elog"
	ca "github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
	"github.com/robinlg/notification-platform/internal/domain"
	"github.com/robinlg/notification-platform/internal/repository/cache"
)

const (
	defaultTimeout = 3 * time.Second
)

type Cache struct {
	rdb    *redis.Client
	logger *elog.Component
	c      *ca.Cache
}

func (l *Cache) Get(_ context.Context, bizID int64) (domain.BusinessConfig, error) {
	key := cache.ConfigKey(bizID)
	v, ok := l.c.Get(key)
	if !ok {
		return domain.BusinessConfig{}, cache.ErrKeyNotFound
	}
	vv, ok := v.(domain.BusinessConfig)
	if !ok {
		return domain.BusinessConfig{}, errors.New("数据类型不正确")
	}
	return vv, nil
}

func (l *Cache) SetConfigs(_ context.Context, configs []domain.BusinessConfig) error {
	for _, config := range configs {
		l.c.Set(cache.ConfigKey(config.ID), config, cache.DefaultExpiredTime)
	}
	return nil
}

func (l *Cache) Set(_ context.Context, cfg domain.BusinessConfig) error {
	key := cache.ConfigKey(cfg.ID)
	l.c.Set(key, cfg, cache.DefaultExpiredTime)
	return nil
}
