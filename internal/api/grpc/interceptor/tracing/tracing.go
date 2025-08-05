package tracing

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// 用于OpenTelemetry跟踪的仪表名
	instrumentationName = "internal/api/grpc/interceptor/tracing"
)

// Builder 追踪拦截器构建器
type Builder struct{}

// New 创建追踪拦截器构造器
func New() *Builder {
	return &Builder{}
}

// Build 构建gRPC一元拦截器
// 为每个一元RPC调用创建一个新的跟踪span
// 生成的追踪数据将发送到OTLP收集器，最终在Zipkin中可视化展示
func (b *Builder) Build() grpc.UnaryServerInterceptor {
	tracer := otel.GetTracerProvider().Tracer(instrumentationName)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 从gRPC方法名称中提取服务和方法名
		fullMethod := info.FullMethod
		serviceName, methodName := extractNames(fullMethod)

		// 创建新的span
		spanName := fmt.Sprintf("%s/%s", serviceName, methodName)
		ctx, span := tracer.Start(
			ctx,
			spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("rpc.system", "grpc"),
				attribute.String("grpc.service", serviceName),
				attribute.String("grpc.method", methodName),
			),
		)
		defer span.End()

		// 添加请求元数据作为span的属性
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			for k, v := range md {
				// 仅添加重要的元数据，避免span太大
				if isTracingRelevantMetadata(k) && len(v) > 0 {
					span.SetAttributes(attribute.String("rpc.metadata."+k, v[0]))
				}
			}
		}

		// 执行处理器
		resp, err := handler(ctx, req)

		// 错误记录
		if err != nil {
			s, _ := status.FromError(err)
			span.SetStatus(codes.Error, s.Message())
			span.SetAttributes(attribute.Int64("rpc.grpc.status_code", int64(s.Code())))
		} else {
			span.SetStatus(codes.Ok, "")
		}

		return resp, err
	}
}

// extractNames 从完整的gRPC方法名中提取服务名和方法名
// 例如 "/service.Service/Method" -> "service.Service", "Method"
func extractNames(fullMethod string) (string, string) {
	fullMethod = strings.TrimPrefix(fullMethod, "/")
	if i := strings.LastIndex(fullMethod, "/"); i >= 0 {
		return fullMethod[:i], fullMethod[i+1:]
	}
	return "unknown", fullMethod
}

// isTracingRelevantMetadata 确定哪些元数据键值对应该被添加到跟踪中
func isTracingRelevantMetadata(key string) bool {
	// 仅记录特定的元数据，例如用户ID、请求ID等
	relevantKeys := map[string]bool{
		"user-id":    true,
		"request-id": true,
		"trace-id":   true,
		"x-api-key":  true,
	}

	return relevantKeys[key]
}
