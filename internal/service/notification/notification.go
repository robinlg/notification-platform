package notification

// Service 通知服务接口
//
//go:generate mockgen -source=./notification.go -destination=./mocks/notification.mock.go -package=notificationmocks -typed Service
type Service interface {
}
