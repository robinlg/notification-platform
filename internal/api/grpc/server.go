package grpc

import notificationv1 "github.com/robinlg/notification-platform/api/proto/gen/notification/v1"

type NotificationServer struct {
	notificationv1.UnimplementedNotificationServiceServer
	notificationv1.UnimplementedNotificationQueryServiceServer
}
