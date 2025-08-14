package grpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/robinlg/notification-platform/internal/errs"
	templatesvc "github.com/robinlg/notification-platform/internal/service/template/manage"

	notificationv1 "github.com/robinlg/notification-platform/api/proto/gen/notification/v1"
	"github.com/robinlg/notification-platform/internal/api/grpc/interceptor/jwt"
	"github.com/robinlg/notification-platform/internal/domain"
	notificationsvc "github.com/robinlg/notification-platform/internal/service/notification"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NotificationServer struct {
	notificationv1.UnimplementedNotificationServiceServer
	notificationv1.UnimplementedNotificationQueryServiceServer

	sendSvc     notificationsvc.SendService
	templateSvc templatesvc.ChannelTemplateService
}

func (s *NotificationServer) SendNotification(ctx context.Context, req *notificationv1.SendNotificationRequest) (*notificationv1.SendNotificationResponse, error) {
	response := &notificationv1.SendNotificationResponse{}

	// 从metadata中解析Authorization JWT Token
	bizID, err := jwt.GetBizIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	// 构建通知领域对象
	notification, err := s.buildNotification(ctx, req.Notification, bizID)
	if err != nil {
		response.ErrorCode = notificationv1.ErrorCode_INVALID_PARAMETER
		response.ErrorMessage = err.Error()
		response.Status = notificationv1.SendStatus_FAILED
		return response, nil
	}

	// 调用发送服务
	result, err := s.sendSvc.SendNotification(ctx, notification)
	if err != nil {
		if s.isSystemError(err) {
			return nil, status.Errorf(codes.Internal, "%v", err)
		} else {
			response.ErrorCode = s.convertToGRPCErrorCode(err)
			response.ErrorMessage = err.Error()
			response.Status = notificationv1.SendStatus_FAILED
			return response, nil
		}
	}

	response.NotificationId = result.NotificationID
	response.Status = s.convertToGRPCSendStatus(result.Status)
	return response, nil
}

// isSystemError 判断错误是否为系统错误
func (s *NotificationServer) isSystemError(err error) bool {
	return errors.Is(err, errs.ErrDatabaseError) ||
		errors.Is(err, errs.ErrExternalServiceError) ||
		errors.Is(err, errs.ErrNotificationDuplicate) ||
		errors.Is(err, errs.ErrNotificationVersionMismatch)
}

// buildNotification 构建通知
func (s *NotificationServer) buildNotification(ctx context.Context, n *notificationv1.Notification, bizID int64) (domain.Notification, error) {
	notification, err := domain.NewNotificationFromAPI(n)
	if err != nil {
		return domain.Notification{}, err
	}

	tmpl, err := s.templateSvc.GetTemplateByID(ctx, notification.Template.ID)
	if err != nil {
		return domain.Notification{}, fmt.Errorf("%w: 模板ID: %s", errs.ErrInvalidParameter, n.TemplateId)
	}

	if !tmpl.HasPublished() {
		return domain.Notification{}, fmt.Errorf("%w: 模板ID: %s 未发布", errs.ErrInvalidParameter, n.TemplateId)
	}

	notification.BizID = bizID
	notification.Template.VersionID = tmpl.ActiveVersionID
	return notification, nil
}

// convertToGRPCSendStatus 将领域发送状态转换为gRPC发送状态
func (s *NotificationServer) convertToGRPCSendStatus(status domain.SendStatus) notificationv1.SendStatus {
	switch status {
	case domain.SendStatusPrepare:
		return notificationv1.SendStatus_PREPARE
	case domain.SendStatusCanceled:
		return notificationv1.SendStatus_CANCELED
	case domain.SendStatusPending:
		return notificationv1.SendStatus_PENDING
	case domain.SendStatusSucceeded:
		return notificationv1.SendStatus_SUCCEEDED
	case domain.SendStatusFailed:
		return notificationv1.SendStatus_FAILED
	default:
		return notificationv1.SendStatus_SEND_STATUS_UNSPECIFIED
	}
}

// convertToGRPCErrorCode 将错误映射为gRPC错误代码
func (s *NotificationServer) convertToGRPCErrorCode(err error) notificationv1.ErrorCode {
	// 注意：这个函数只处理业务错误，系统错误由isSystemError判断后直接通过gRPC status返回
	switch {
	case errors.Is(err, errs.ErrInvalidParameter):
		return notificationv1.ErrorCode_INVALID_PARAMETER

	case errors.Is(err, errs.ErrTemplateNotFound):
		return notificationv1.ErrorCode_TEMPLATE_NOT_FOUND

	case errors.Is(err, errs.ErrChannelDisabled):
		return notificationv1.ErrorCode_CHANNEL_DISABLED

	case errors.Is(err, errs.ErrRateLimited):
		return notificationv1.ErrorCode_RATE_LIMITED

	case errors.Is(err, errs.ErrBizIDNotFound):
		return notificationv1.ErrorCode_BIZ_ID_NOT_FOUND

	case errors.Is(err, errs.ErrSendNotificationFailed):
		return notificationv1.ErrorCode_SEND_NOTIFICATION_FAILED

	case errors.Is(err, errs.ErrCreateNotificationFailed):
		return notificationv1.ErrorCode_CREATE_NOTIFICATION_FAILED

	case errors.Is(err, errs.ErrNotificationNotFound):
		return notificationv1.ErrorCode_NOTIFICATION_NOT_FOUND

	case errors.Is(err, errs.ErrNoAvailableProvider):
		return notificationv1.ErrorCode_NO_AVAILABLE_PROVIDER

	case errors.Is(err, errs.ErrNoAvailableChannel):
		return notificationv1.ErrorCode_NO_AVAILABLE_CHANNEL

	case errors.Is(err, errs.ErrConfigNotFound):
		return notificationv1.ErrorCode_CONFIG_NOT_FOUND

	case errors.Is(err, errs.ErrNoQuotaConfig):
		return notificationv1.ErrorCode_NO_QUOTA_CONFIG

	case errors.Is(err, errs.ErrNoQuota):
		return notificationv1.ErrorCode_NO_QUOTA

	case errors.Is(err, errs.ErrQuotaNotFound):
		return notificationv1.ErrorCode_QUOTA_NOT_FOUND

	case errors.Is(err, errs.ErrProviderNotFound):
		return notificationv1.ErrorCode_PROVIDER_NOT_FOUND

	case errors.Is(err, errs.ErrUnknownChannel):
		return notificationv1.ErrorCode_UNKNOWN_CHANNEL

	default:
		return notificationv1.ErrorCode_ERROR_CODE_UNSPECIFIED
	}
}
