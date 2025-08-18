package client

import "errors"

const (
	OK = "OK"
)

// 通用错误定义
var (
	ErrCreateTemplateFailed = errors.New("创建模版失败")
	ErrQueryTemplateStatus  = errors.New("查询模版状态失败")
	ErrSendFailed           = errors.New("发送短信失败")
	ErrQuerySendDetails     = errors.New("查询发送详情失败")
	ErrInvalidParameter     = errors.New("参数无效")
)

type (
	TemplateType int32
)

const (
	TemplateTypeInternational TemplateType = 0 // 国际/港澳台消息 仅阿里云使用
	TemplateTypeMarketing     TemplateType = 1 // 营销短信
	TemplateTypeNotification  TemplateType = 2 // 通知短信
	TemplateTypeVerification  TemplateType = 3 // 验证码
)

// Client 短信客户端接口 (抽象)
//
//go:generate mockgen -source=./types.go -destination=./mocks/sms.mock.go -package=smsmocks -typed Client
type Client interface {
	// Send 发送短信
	Send(req SendReq) (SendResp, error)
}

// CreateTemplateReq 创建短信模板请求参数
type CreateTemplateReq struct {
	TemplateName    string       // 模板名称
	TemplateContent string       // 模板内容
	TemplateType    TemplateType // 短信类型
	Remark          string       // 备注
}

// CreateTemplateResp 创建短信模板响应参数
type CreateTemplateResp struct {
	RequestID  string // 请求 ID,   阿里云、腾讯云共用
	TemplateID string // 模板 ID, 阿里云、腾讯云共用 (阿里云返回 TemplateCode, 腾讯云返回处理过的 TemplateID)
}

// SendReq 发送短信请求参数
type SendReq struct {
	PhoneNumbers  []string          // 手机号码, 阿里云、腾讯云共用
	SignName      string            // 签名名称, 阿里云、腾讯云共用
	TemplateID    string            // 模板 ID, 阿里云、腾讯云共用
	TemplateParam map[string]string // 模板参数, 阿里云、腾讯云共用, key-value 形式
}

// SendResp 发送短信响应参数
type SendResp struct {
	RequestID    string                    // 请求 ID,      阿里云、腾讯云共用
	PhoneNumbers map[string]SendRespStatus // 去掉+86后的手机号
}

type SendRespStatus struct {
	Code    string
	Message string
}
