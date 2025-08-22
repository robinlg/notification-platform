package repository

import (
	"context"

	"github.com/robinlg/notification-platform/internal/domain"
	"github.com/robinlg/notification-platform/internal/repository/dao"
)

type TxNotificationRepository interface {
	Create(ctx context.Context, notification domain.TxNotification) (uint64, error)
	UpdateStatus(ctx context.Context, bizID int64, key string, status domain.TxNotificationStatus, notificationStatus domain.SendStatus) error
}

type txNotificationRepo struct {
	txdao dao.TxNotificationDAO
}

// NewTxNotificationRepository creates a new TxNotificationRepository instance
func NewTxNotificationRepository(txdao dao.TxNotificationDAO) TxNotificationRepository {
	return &txNotificationRepo{
		txdao: txdao,
	}
}

func (t *txNotificationRepo) Create(ctx context.Context, txn domain.TxNotification) (uint64, error) {
	// 转换领域模型到DAO对象
	txnEntity := t.toDao(txn)
	notificationEntity := t.toEntity(txn.Notification)
	// 调用DAO层创建记录
	return t.txdao.Prepare(ctx, txnEntity, notificationEntity)
}

func (t *txNotificationRepo) UpdateStatus(ctx context.Context, bizID int64, key string, status domain.TxNotificationStatus, notificationStatus domain.SendStatus) error {
	// 直接调用DAO层更新状态
	return t.txdao.UpdateStatus(ctx, bizID, key, status, notificationStatus)
}

// toEntity 将领域对象转换为DAO实体
func (t *txNotificationRepo) toEntity(notification domain.Notification) dao.Notification {
	templateParams, _ := notification.MarshalTemplateParams()
	receivers, _ := notification.MarshalReceivers()
	return dao.Notification{
		ID:                notification.ID,
		BizID:             notification.BizID,
		Key:               notification.Key,
		Receivers:         receivers,
		Channel:           string(notification.Channel),
		TemplateID:        notification.Template.ID,
		TemplateVersionID: notification.Template.VersionID,
		TemplateParams:    templateParams,
		Status:            string(notification.Status),
		ScheduledSTime:    notification.ScheduledSTime.UnixMilli(),
		ScheduledETime:    notification.ScheduledETime.UnixMilli(),
		Version:           notification.Version,
	}
}

// toDomain 将DAO对象转换为领域模型
func (t *txNotificationRepo) toDomain(daoNotification dao.TxNotification) domain.TxNotification {
	return domain.TxNotification{
		TxID: daoNotification.TxID,
		Notification: domain.Notification{
			ID: daoNotification.NotificationID,
		},
		Key:           daoNotification.Key,
		BizID:         daoNotification.BizID,
		Status:        domain.TxNotificationStatus(daoNotification.Status),
		CheckCount:    daoNotification.CheckCount,
		NextCheckTime: daoNotification.NextCheckTime,
		Ctime:         daoNotification.Ctime,
		Utime:         daoNotification.Utime,
	}
}

// toDao 将领域模型转换为DAO对象
func (t *txNotificationRepo) toDao(domainNotification domain.TxNotification) dao.TxNotification {
	return dao.TxNotification{
		TxID:           domainNotification.TxID,
		Key:            domainNotification.Key,
		NotificationID: domainNotification.Notification.ID,
		BizID:          domainNotification.BizID,
		Status:         string(domainNotification.Status),
		CheckCount:     domainNotification.CheckCount,
		NextCheckTime:  domainNotification.NextCheckTime,
		Ctime:          domainNotification.Ctime,
		Utime:          domainNotification.Utime,
	}
}
