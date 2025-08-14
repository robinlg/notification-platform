package dao

import (
	"context"
	"errors"
	"fmt"

	"github.com/ego-component/egorm"
	"github.com/robinlg/notification-platform/internal/errs"
	"gorm.io/gorm"
)

// ChannelTemplate 渠道模板表
type ChannelTemplate struct {
	ID              int64  `gorm:"primaryKey;autoIncrement;comment:'渠道模版ID'"`
	OwnerID         int64  `gorm:"type:BIGINT;NOT NULL;comment:'用户ID或部门ID'"`
	OwnerType       string `gorm:"type:ENUM('person', 'organization');NOT NULL;comment:'业务方类型{person:个人,organization:组织}'"`
	Name            string `gorm:"type:VARCHAR(128);NOT NULL;comment:'模板名称'"`
	Description     string `gorm:"type:VARCHAR(512);NOT NULL;comment:'模板描述'"`
	Channel         string `gorm:"type:ENUM('SMS','EMAIL','IN_APP');NOT NULL;comment:'渠道类型'"`
	BusinessType    int64  `gorm:"type:BIGINT;NOT NULL;DEFAULT:1;comment:'业务类型{1:推广营销,2:通知,3:验证码等}'"`
	ActiveVersionID int64  `gorm:"type:BIGINT;DEFAULT:0;index:idx_active_version;comment:'当前启用的版本ID,0表示无活跃版本'"`
	Ctime           int64
	Utime           int64
}

func (ChannelTemplate) TableName() string {
	return "channel_templates"
}

// ChannelTemplateVersion 渠道模板版本表
type ChannelTemplateVersion struct {
	ID                int64  `gorm:"primaryKey;autoIncrement;comment:'渠道模版版本ID'"`
	ChannelTemplateID int64  `gorm:"type:BIGINT;NOT NULL;index:idx_channel_template_id;comment:'关联渠道模版ID'"`
	Name              string `gorm:"type:VARCHAR(32);NOT NULL;comment:'版本名称，如v1.0.0'"`
	Signature         string `gorm:"type:VARCHAR(64);comment:'已通过所有供应商审核的短信签名/邮件发件人'"`
	Content           string `gorm:"type:TEXT;NOT NULL;comment:'原始模板内容，使用平台统一变量格式，如${name}'"`
	Remark            string `gorm:"type:TEXT;NOT NULL;comment:'申请说明,描述使用短信的业务场景，并提供短信完整示例（填入变量内容），信息完整有助于提高模板审核通过率。'"`
	// 审核相关信息，AuditID之后的为冗余的信息
	AuditID                  int64  `gorm:"type:BIGINT;NOT NULL;DEFAULT:0;comment:'审核表ID, 0表示尚未提交审核或者未拿到审核结果'"`
	AuditorID                int64  `gorm:"type:BIGINT;comment:'审核人ID'"`
	AuditTime                int64  `gorm:"comment:'审核时间'"`
	AuditStatus              string `gorm:"type:ENUM('PENDING','IN_REVIEW','REJECTED','APPROVED');NOT NULL;DEFAULT:'PENDING';comment:'内部审核状态，PENDING表示未提交审核；IN_REVIEW表示已提交审核；APPROVED表示审核通过；REJECTED表示审核未通过'"`
	RejectReason             string `gorm:"type:VARCHAR(512);comment:'拒绝原因'"`
	LastReviewSubmissionTime int64  `gorm:"comment:'上一次提交审核时间'"`
	Ctime                    int64
	Utime                    int64
}

// TableName 重命名表
func (ChannelTemplateVersion) TableName() string {
	return "channel_template_versions"
}

// ChannelTemplateProvider 渠道模版供应商表
type ChannelTemplateProvider struct {
	ID                       int64  `gorm:"primaryKey;autoIncrement;comment:'渠道模版-供应商关联ID'"`
	TemplateID               int64  `gorm:"type:BIGINT;NOT NULL;uniqueIndex:idx_template_version_provider,priority:1;uniqueIndex:idx_tmpl_ver_name_chan,priority:1;comment:'渠道模版ID'"`
	TemplateVersionID        int64  `gorm:"type:BIGINT;NOT NULL;uniqueIndex:idx_template_version_provider,priority:2;uniqueIndex:idx_tmpl_ver_name_chan,priority:2;comment:'渠道模版版本ID'"`
	ProviderID               int64  `gorm:"type:BIGINT;NOT NULL;uniqueIndex:idx_template_version_provider,priority:3;comment:'供应商ID'"`
	ProviderName             string `gorm:"type:VARCHAR(64);NOT NULL;uniqueIndex:idx_tmpl_ver_name_chan,priority:3;comment:'供应商名称'"`
	ProviderChannel          string `gorm:"type:ENUM('SMS','EMAIL','IN_APP');NOT NULL;uniqueIndex:idx_tmpl_ver_name_chan,priority:4;comment:'渠道类型'"`
	RequestID                string `gorm:"type:VARCHAR(256);index:idx_request_id;comment:'审核请求在供应商侧的ID，用于排查问题'"`
	ProviderTemplateID       string `gorm:"type:VARCHAR(256);comment:'当前版本模版在供应商侧的ID，审核通过后才会有值'"`
	AuditStatus              string `gorm:"type:ENUM('PENDING','IN_REVIEW','REJECTED','APPROVED');NOT NULL;DEFAULT:'PENDING';index:idx_audit_status;comment:'供应商侧模版审核状态，PENDING表示未提交审核；IN_REVIEW表示已提交审核；APPROVED表示审核通过；REJECTED表示审核未通过'"`
	RejectReason             string `gorm:"type:VARCHAR(512);comment:'供应商侧拒绝原因'"`
	LastReviewSubmissionTime int64  `gorm:"comment:'上一次提交审核时间'"`
	Ctime                    int64
	Utime                    int64
}

// TableName 重命名表
func (ChannelTemplateProvider) TableName() string {
	return "channel_template_providers"
}

// ChannelTemplateDAO 提供模板数据访问对象接口
type ChannelTemplateDAO interface {
	// 模版相关方法

	// GetTemplateByID 根据ID获取模板
	GetTemplateByID(ctx context.Context, id int64) (ChannelTemplate, error)

	// 模版版本相关方法

	// GetTemplateVersionsByTemplateIDs 根据模板ID列表获取对应的版本列表
	GetTemplateVersionsByTemplateIDs(ctx context.Context, templateIDs []int64) ([]ChannelTemplateVersion, error)

	// 供应商关联相关方法

	// GetProvidersByVersionIDs 根据版本ID列表获取供应商列表
	GetProvidersByVersionIDs(ctx context.Context, versionIDs []int64) ([]ChannelTemplateProvider, error)
}

// channelTemplateDAO 实现了ChannelTemplateDAO接口，提供对模板数据的数据库访问实现
type channelTemplateDAO struct {
	db *egorm.Component
}

// NewChannelTemplateDAO 创建模板DAO实例
func NewChannelTemplateDAO(db *egorm.Component) ChannelTemplateDAO {
	return &channelTemplateDAO{
		db: db,
	}
}

// 模版相关方法

// GetTemplateByID 根据ID获取模板
func (d *channelTemplateDAO) GetTemplateByID(ctx context.Context, id int64) (ChannelTemplate, error) {
	var template ChannelTemplate
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&template).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ChannelTemplate{}, fmt.Errorf("%w", errs.ErrTemplateNotFound)
		}
		return ChannelTemplate{}, err
	}
	return template, nil
}

// 模版版本相关方法

// GetTemplateVersionsByTemplateIDs 根据模板IDs获取版本列表
func (d *channelTemplateDAO) GetTemplateVersionsByTemplateIDs(ctx context.Context, templateIDs []int64) ([]ChannelTemplateVersion, error) {
	if len(templateIDs) == 0 {
		return []ChannelTemplateVersion{}, nil
	}

	var versions []ChannelTemplateVersion
	result := d.db.WithContext(ctx).Where("channel_template_id IN ?", templateIDs).Find(&versions)
	if result.Error != nil {
		return nil, result.Error
	}
	return versions, nil
}

// 供应商关联相关方法

// GetProvidersByVersionIDs 根据版本IDs获取供应商关联
func (d *channelTemplateDAO) GetProvidersByVersionIDs(ctx context.Context, versionIDs []int64) ([]ChannelTemplateProvider, error) {
	if len(versionIDs) == 0 {
		return []ChannelTemplateProvider{}, nil
	}

	var providers []ChannelTemplateProvider
	result := d.db.WithContext(ctx).Where("template_version_id IN (?)", versionIDs).Find(&providers)
	if result.Error != nil {
		return nil, result.Error
	}
	return providers, nil
}
