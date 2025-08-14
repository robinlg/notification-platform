package callback

import (
	"context"
	"fmt"
	"time"

	"github.com/ecodeclub/ekit/syncx"
	"github.com/gotomicro/ego/client/egrpc"
	"github.com/gotomicro/ego/core/elog"
	clientv1 "github.com/robinlg/notification-platform/api/proto/gen/client/v1"
	notificationv1 "github.com/robinlg/notification-platform/api/proto/gen/notification/v1"
	"github.com/robinlg/notification-platform/internal/domain"
	"github.com/robinlg/notification-platform/internal/errs"
	"github.com/robinlg/notification-platform/internal/pkg/grpc"
	"github.com/robinlg/notification-platform/internal/pkg/retry"
	"github.com/robinlg/notification-platform/internal/repository"
	"github.com/robinlg/notification-platform/internal/service/config"
)

var _ Service = (*service)(nil)

type Service interface {
	SendCallbackByNotification(ctx context.Context, notification domain.Notification) error
}

type service struct {
	configSvc    config.BusinessConfigService
	bizID2Config syncx.Map[int64, *domain.CallbackConfig]
	clients      *grpc.Clients[clientv1.CallbackServiceClient]
	repo         repository.CallbackLogRepository
	logger       *elog.Component
}

func NewService(configSvc config.BusinessConfigService, repo repository.CallbackLogRepository) Service {
	return &service{
		configSvc:    configSvc,
		bizID2Config: syncx.Map[int64, *domain.CallbackConfig]{},
		repo:         repo,
		clients: grpc.NewClients(func(conn *egrpc.Component) clientv1.CallbackServiceClient {
			return clientv1.NewCallbackServiceClient(conn)
		}),
		logger: elog.DefaultLogger.With(elog.FieldComponent("callback")),
	}
}

func (c *service) SendCallbackByNotification(ctx context.Context, notification domain.Notification) error {
	logs, err := c.repo.FindByNotificationIDs(ctx, []uint64{notification.ID})
	if err != nil {
		return err
	}
	return c.sendCallbackAndUpdateCallbackLogs(ctx, logs)
}

func (c *service) sendCallbackAndUpdateCallbackLogs(ctx context.Context, logs []domain.CallbackLog) error {
	needUpdate := make([]domain.CallbackLog, 0, len(logs))
	for i := range logs {
		changed, err := c.sendCallbackAndSetChangedFields(ctx, &logs[i])
		if err != nil {
			c.logger.Warn("业务方回调失败",
				elog.FieldKey("Callback.ID"),
				elog.FieldValueAny(logs[i].ID),
				elog.FieldErr(err))
			continue
		}
		if changed {
			needUpdate = append(needUpdate, logs[i])
		}
	}
	return c.repo.Update(ctx, needUpdate)
}

func (c *service) sendCallbackAndSetChangedFields(ctx context.Context, log *domain.CallbackLog) (changed bool, err error) {
	resp, err := c.sendCallback(ctx, log.Notification)
	if err != nil {
		return false, err
	}

	// 拿到业务方对回调的处理结果
	if resp.Success {
		log.Status = domain.CallbackLogStatusSuccess
		return true, nil
	}

	// 业务方对回调的处理失败，需要重试，此时业务方必定有配置
	cfg, _ := c.getConfig(ctx, log.Notification.BizID)
	retryStrategy, _ := retry.NewRetry(*cfg.RetryPolicy)
	interval, ok := retryStrategy.NextWithRetries(log.RetryCount)
	if ok {
		// 未达到最大重试次数，状态不变但要更新下次重试时间和重试次数
		log.NextRetryTime = time.Now().Add(interval).UnixMilli()
		log.RetryCount++
	} else {
		// 达到最大重试次数限制，不再重试，更新状态为失败
		log.Status = domain.CallbackLogStatusFailed
	}
	return true, nil
}

func (c *service) sendCallback(ctx context.Context, notification domain.Notification) (*clientv1.HandleNotificationResultResponse, error) {
	cfg, err := c.getConfig(ctx, notification.BizID)
	if err != nil {
		c.logger.Warn("获取业务配置失败",
			elog.FieldKey("BizID"),
			elog.FieldValueAny(notification.BizID),
			elog.FieldErr(err))
		return nil, err
	}
	if cfg == nil {
		// 业务方未提供配置
		return nil, fmt.Errorf("%w", errs.ErrConfigNotFound)
	}
	return c.clients.Get(cfg.ServiceName).HandleNotificationResult(ctx, c.buildRequest(notification))
}

func (c *service) getConfig(ctx context.Context, bizID int64) (*domain.CallbackConfig, error) {
	cfg, ok := c.bizID2Config.Load(bizID)
	if ok {
		return cfg, nil
	}
	bizConfig, err := c.configSvc.GetByID(ctx, bizID)
	if err != nil {
		return nil, err
	}
	if bizConfig.CallbackConfig != nil {
		c.bizID2Config.Store(bizID, bizConfig.CallbackConfig)
	}
	return bizConfig.CallbackConfig, nil
}

func (c *service) buildRequest(notification domain.Notification) *clientv1.HandleNotificationResultRequest {
	templateParams := make(map[string]string)
	if notification.Template.Params != nil {
		templateParams = notification.Template.Params
	}
	return &clientv1.HandleNotificationResultRequest{
		NotificationId: notification.ID,
		OriginalRequest: &notificationv1.SendNotificationRequest{
			Notification: &notificationv1.Notification{
				Key:            notification.Key,
				Receivers:      notification.Receivers,
				Channel:        c.getChannel(notification),
				TemplateId:     fmt.Sprintf("%d", notification.Template.ID),
				TemplateParams: templateParams,
			},
		},
		Result: &notificationv1.SendNotificationResponse{
			NotificationId: notification.ID,
			Status:         c.getStatus(notification),
		},
	}
}

func (c *service) getChannel(notification domain.Notification) notificationv1.Channel {
	var channel notificationv1.Channel
	switch notification.Channel {
	case domain.ChannelSMS:
		channel = notificationv1.Channel_SMS
	case domain.ChannelEmail:
		channel = notificationv1.Channel_EMAIL
	case domain.ChannelInApp:
		channel = notificationv1.Channel_IN_APP
	default:
		channel = notificationv1.Channel_CHANNEL_UNSPECIFIED
	}
	return channel
}

func (c *service) getStatus(notification domain.Notification) notificationv1.SendStatus {
	var status notificationv1.SendStatus
	switch notification.Status {
	case domain.SendStatusSucceeded:
		status = notificationv1.SendStatus_SUCCEEDED
	case domain.SendStatusFailed:
		status = notificationv1.SendStatus_FAILED
	case domain.SendStatusPrepare:
		status = notificationv1.SendStatus_PREPARE
	case domain.SendStatusCanceled:
		status = notificationv1.SendStatus_CANCELED
	case domain.SendStatusPending:
		status = notificationv1.SendStatus_PENDING
	case domain.SendStatusSending:
		status = notificationv1.SendStatus_SEND_STATUS_UNSPECIFIED
	default:
		status = notificationv1.SendStatus_SEND_STATUS_UNSPECIFIED
	}
	return status
}
