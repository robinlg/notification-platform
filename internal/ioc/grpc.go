package ioc

import (
	"github.com/ego-component/eetcd"
	"github.com/ego-component/eetcd/registry"
	"github.com/gotomicro/ego/client/egrpc/resolver"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/server/egrpc"
	notificationv1 "github.com/robinlg/notification-platform/api/proto/gen/notification/v1"
	grpcapi "github.com/robinlg/notification-platform/internal/api/grpc"
	"github.com/robinlg/notification-platform/internal/api/grpc/interceptor/jwt"
	"github.com/robinlg/notification-platform/internal/api/grpc/interceptor/log"
	"github.com/robinlg/notification-platform/internal/api/grpc/interceptor/metrics"
	"github.com/robinlg/notification-platform/internal/api/grpc/interceptor/tracing"
)

func InitGrpc(noserver *grpcapi.NotificationServer, etcdClint *eetcd.Component) *egrpc.Component {
	// 注册全局的注册中心
	type Config struct {
		Key string `yaml:"key"`
	}
	var cfg Config

	err := econf.UnmarshalKey("jwt", &cfg)
	if err != nil {
		panic("config err:" + err.Error())
	}

	reg := registry.Load("").Build(registry.WithClientEtcd(etcdClint))
	resolver.Register("etcd", reg)

	// 创建指标数据拦截器
	metricsInterceptor := metrics.New().Build()
	// 创建日志拦截器
	logInterceptor := log.New().Build()
	// 创建跟踪(全链路日志)拦截器
	traceInterceptor := tracing.New().Build()
	// 创建token拦截器
	tokenInterceptor := jwt.New(cfg.Key).Build()
	server := egrpc.Load("server.grpc").Build(
		egrpc.WithUnaryInterceptor(metricsInterceptor, logInterceptor, traceInterceptor, tokenInterceptor),
	)

	notificationv1.RegisterNotificationServiceServer(server.Server, noserver)
	notificationv1.RegisterNotificationQueryServiceServer(server.Server, noserver)

	return server
}
