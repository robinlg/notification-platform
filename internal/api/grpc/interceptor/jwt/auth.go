package jwt

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/golang-jwt/jwt/v4"
)

const BizIDName = "biz_id"

// Builder token拦截器构建器
type Builder struct {
	key string
}

// New 创建token拦截器构建器
func New(key string) *Builder {
	return &Builder{
		key: key,
	}
}

// Decode 解码Token，并验证有效性
func (a *Builder) Decode(tokenStr string) (jwt.MapClaims, error) {
	// 去除可能的Bearer前缀（兼容不同客户端实现）
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

	// 解析Token
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("不支持的签名算法: %v", token.Header["alg"])
		}
		return []byte(a.key), nil
	})
	if err != nil {
		return nil, fmt.Errorf("令牌解析失败: %w", err)
	}

	// 验证Token有效性
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("无效的令牌")
}

// Encode 生成JWT Token，支持自定义声明和自动添加标准声明
func (a *Builder) Encode(customClaims jwt.MapClaims) (string, error) {
	// 默认声明
	claims := jwt.MapClaims{
		"iat": time.Now().Unix(),
		"iss": "notification-platform",
	}

	// 合并用户自定义声明（覆盖默认声明）
	for k, v := range customClaims {
		claims[k] = v
	}

	// 自动处理过期时间
	if _, ok := claims["exp"]; !ok {
		claims["exp"] = time.Now().Add(24 * time.Hour).Unix() // 默认24小时过期
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(a.key))
}

// Build 构建gRPC一元拦截器
func (a *Builder) Build() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 提取metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
		}

		// 获取Authorization头
		authHeaders := md.Get("Authorization")
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization token is required")
		}

		// 解码认证
		tokenStr := authHeaders[0]
		val, err := a.Decode(tokenStr)
		if err != nil {
			// 细化错误类型处理
			if errors.Is(err, jwt.ErrTokenExpired) {
				return nil, status.Error(codes.Unauthenticated, "token expired")
			}
			if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
				return nil, status.Error(codes.Unauthenticated, "invalid signature")
			}
			return nil, status.Error(codes.Unauthenticated, "invalid token: "+err.Error())
		}
		v, ok := val[BizIDName]
		if ok {
			bizId := v.(float64)
			ctx = context.WithValue(ctx, BizIDName, int64(bizId))
		}

		v, ok = val["Priority"]
		if ok {
			ctx = context.WithValue(ctx, "Priority", v)
		}

		return handler(ctx, req)
	}
}
