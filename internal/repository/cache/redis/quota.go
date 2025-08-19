package redis

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/gotomicro/ego/core/elog"
	"github.com/redis/go-redis/v9"
	"github.com/robinlg/notification-platform/internal/domain"
	"github.com/robinlg/notification-platform/internal/repository/cache"
)

var (
	ErrQuotaLessThenZero = errors.New("额度小于0")
	//go:embed lua/quota.lua
	quotaScript string
	//go:embed lua/batch_decr_quota.lua
	batchDecrQuotaScript string
	//go:embed lua/batch_incr_quota.lua
	batchIncrQuotaScript string
)

type quotaCache struct {
	client redis.Cmdable
	logger *elog.Component
}

func NewQuotaCache(client redis.Cmdable) cache.QuotaCache {
	return &quotaCache{
		client: client,
		logger: elog.DefaultLogger,
	}
}

func (q *quotaCache) Incr(ctx context.Context, bizID int64, channel domain.Channel, quota int32) error {
	return q.client.Eval(ctx, quotaScript, []string{q.key(domain.Quota{
		BizID:   bizID,
		Channel: channel,
	})}, quota).Err()
}

func (q *quotaCache) Decr(ctx context.Context, bizID int64, channel domain.Channel, quota int32) error {
	res, err := q.client.DecrBy(ctx, q.key(domain.Quota{
		BizID:   bizID,
		Channel: channel,
	}), int64(quota)).Result()
	if err != nil {
		return err
	}
	if res < 0 {
		elog.Error("库存不足", elog.Int("biz_id", int(bizID)), elog.String("channel", channel.String()))
		return ErrQuotaLessThenZero
	}
	return nil
}

func (q *quotaCache) MutiIncr(ctx context.Context, items []cache.IncrItem) error {
	if len(items) == 0 {
		return nil
	}
	keys, quotas := q.getKeysAndQuotas(items)
	err := q.client.Eval(ctx, batchIncrQuotaScript, keys, quotas).Err()
	if err != nil {
		return err
	}
	return nil
}

func (q *quotaCache) getKeysAndQuotas(items []cache.IncrItem) (keys []string, quotas []any) {
	keys = make([]string, 0, len(items))
	quotas = make([]any, 0, len(items))
	for idx := range items {
		item := items[idx]
		keys = append(keys, q.key(domain.Quota{
			BizID:   item.BizID,
			Channel: item.Channel,
		}))
		quotas = append(quotas, item.Val)
	}
	return keys, quotas
}

func (q *quotaCache) MutiDecr(ctx context.Context, items []cache.IncrItem) error {
	keys, quotas := q.getKeysAndQuotas(items)
	res, err := q.client.Eval(ctx, batchDecrQuotaScript, keys, quotas).Result()
	if err != nil {
		return err
	}
	resStr, ok := res.(string)
	if !ok {
		return errors.New("返回值不正确")
	}
	if resStr != "" {
		return fmt.Errorf("%s不足 %w", resStr, ErrQuotaLessThenZero)
	}
	return nil
}

func (q *quotaCache) key(quota domain.Quota) string {
	return fmt.Sprintf("quota:%d:%s", quota.BizID, quota.Channel)
}
