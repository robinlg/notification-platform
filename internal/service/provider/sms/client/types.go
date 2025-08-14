package client

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
	// CreateTemplate 创建模板
	CreateTemplate(req CreateTemplateReq) (CreateTemplateResp, error)
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
