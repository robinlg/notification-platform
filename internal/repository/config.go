package repository

import (
	"context"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	"github.com/robinlg/notification-platform/internal/domain"
	"github.com/robinlg/notification-platform/internal/pkg/sqlx"
	"github.com/robinlg/notification-platform/internal/repository/cache"
	"github.com/robinlg/notification-platform/internal/repository/cache/local"
	"github.com/robinlg/notification-platform/internal/repository/cache/redis"
	"github.com/robinlg/notification-platform/internal/repository/dao"
)

type BusinessConfigRepository interface {
	GetByID(ctx context.Context, id int64) (domain.BusinessConfig, error)
}

type businessConfigRepository struct {
	dao        dao.BusinessConfigDAO
	localCache cache.ConfigCache
	redisCache cache.ConfigCache
	logger     *elog.Component
}

// NewBusinessConfigRepository 创建业务配置仓库实例
func NewBusinessConfigRepository(
	configDao dao.BusinessConfigDAO,
	localCache *local.Cache,
	redisCache *redis.Cache,
) BusinessConfigRepository {
	res := &businessConfigRepository{
		dao:        configDao,
		localCache: localCache,
		redisCache: redisCache,
		logger:     elog.DefaultLogger,
	}
	// 复杂系统里面，启动非常慢，可以考虑开 goroutine
	go func() {
		const preloadTimeout = time.Minute
		ctx, cancel := context.WithTimeout(context.Background(), preloadTimeout)
		defer cancel()
		err := res.LoadCache(ctx)
		if err != nil {
			// 缓存预热失败，你可以中断
			res.logger.Error("缓存预热失败", elog.FieldErr(err))
		}
	}()
	return res
}

// LoadCache 加载缓存，用 DB 中的数据，填充本地缓存
func (b *businessConfigRepository) LoadCache(ctx context.Context) error {
	offset := 0
	const (
		limit       = 10
		loopTimeout = time.Second * 3
	)
	for {
		ctx, cancel := context.WithTimeout(ctx, loopTimeout)
		cnt, err := b.loadCacheBatch(ctx, offset, limit)
		cancel()
		if err != nil {
			// 继续下一轮
			// 精细处理：比如说三个循环都是 error，你就判定数据库不可挽回了，你就中断
			b.logger.Error("分批加载缓存失败", elog.FieldErr(err))
			continue
		}
		if cnt < limit {
			// 说明没了
			return nil
		}
		offset += cnt
	}
}

func (b *businessConfigRepository) loadCacheBatch(ctx context.Context, offset, limit int) (int, error) {
	res, err := b.Find(ctx, offset, limit)
	if err != nil {
		return 0, err
	}
	err = b.localCache.SetConfigs(ctx, res)
	return len(res), err
}

func (b *businessConfigRepository) Find(ctx context.Context, offset, limit int) ([]domain.BusinessConfig, error) {
	res, err := b.dao.Find(ctx, offset, limit)
	return slice.Map(res, func(_ int, src dao.BusinessConfig) domain.BusinessConfig {
		return b.toDomain(src)
	}), err
}

// GetByID 根据ID获取业务配置
func (b *businessConfigRepository) GetByID(ctx context.Context, id int64) (domain.BusinessConfig, error) {
	// 从数据库获取配置

	cfg, localErr := b.localCache.Get(ctx, id)
	if localErr == nil {
		return cfg, nil
	}
	cfg, redisErr := b.redisCache.Get(ctx, id)
	if redisErr == nil {
		// 刷新本地缓存
		lerr := b.localCache.Set(ctx, cfg)
		if lerr != nil {
			b.logger.Error("刷新本地缓存失败", elog.Any("err", lerr), elog.Int("bizId", int(id)))
		}
		return cfg, nil
	}

	c, err := b.dao.GetByID(ctx, id)
	if err != nil {
		return domain.BusinessConfig{}, err
	}
	domainConfig := b.toDomain(c)
	// 刷新本地缓存+redis
	lerr := b.localCache.Set(ctx, domainConfig)
	if lerr != nil {
		b.logger.Error("刷新本地缓存失败", elog.Any("err", lerr), elog.Int("bizId", int(id)))
	}
	rerr := b.redisCache.Set(ctx, domainConfig)
	if rerr != nil {
		b.logger.Error("刷新redis缓存失败", elog.Any("err", rerr), elog.Int("bizId", int(id)))
	}
	// 将DAO对象转换为领域对象
	return domainConfig, nil
}

func (b *businessConfigRepository) toDomain(config dao.BusinessConfig) domain.BusinessConfig {
	domainCfg := domain.BusinessConfig{
		ID:        config.ID,
		OwnerID:   config.OwnerID,
		OwnerType: config.OwnerType,
		RateLimit: config.RateLimit,
		Ctime:     config.Ctime,
		Utime:     config.Utime,
	}
	if config.ChannelConfig.Valid {
		domainCfg.ChannelConfig = &config.ChannelConfig.Val
	}
	if config.TxnConfig.Valid {
		domainCfg.TxnConfig = &config.TxnConfig.Val
	}
	if config.Quota.Valid {
		domainCfg.Quota = &config.Quota.Val
	}
	if config.CallbackConfig.Valid {
		domainCfg.CallbackConfig = &config.CallbackConfig.Val
	}
	return domainCfg
}

func (b *businessConfigRepository) toEntity(config domain.BusinessConfig) dao.BusinessConfig {
	businessConfig := dao.BusinessConfig{
		ID:        config.ID,
		OwnerID:   config.OwnerID,
		OwnerType: config.OwnerType,
		RateLimit: config.RateLimit,
		Ctime:     config.Ctime,
		Utime:     config.Utime,
	}

	if config.ChannelConfig != nil {
		businessConfig.ChannelConfig = sqlx.JSONColumn[domain.ChannelConfig]{
			Val:   *config.ChannelConfig,
			Valid: true,
		}
	}

	if config.TxnConfig != nil {
		businessConfig.TxnConfig = sqlx.JSONColumn[domain.TxnConfig]{
			Val:   *config.TxnConfig,
			Valid: true,
		}
	}

	if config.Quota != nil {
		businessConfig.Quota = sqlx.JSONColumn[domain.QuotaConfig]{
			Val:   *config.Quota,
			Valid: true,
		}
	}

	if config.CallbackConfig != nil {
		businessConfig.CallbackConfig = sqlx.JSONColumn[domain.CallbackConfig]{
			Val:   *config.CallbackConfig,
			Valid: true,
		}
	}

	return businessConfig
}
