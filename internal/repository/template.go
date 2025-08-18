package repository

import (
	"context"

	"github.com/robinlg/notification-platform/internal/domain"
	"github.com/robinlg/notification-platform/internal/repository/dao"
)

// ChannelTemplateRepository 提供模版数据储存的仓库接口
type ChannelTemplateRepository interface {
	// 模版相关方法

	// GetTemplateByID 根据ID获取模板
	GetTemplateByID(ctx context.Context, templateID int64) (domain.ChannelTemplate, error)

	// 模版版本相关方法

	// GetTemplateVersionByID 根据ID获取模板版本
	GetTemplateVersionByID(ctx context.Context, versionID int64) (domain.ChannelTemplateVersion, error)

	// 供应商相关方法

	// GetProviderByNameAndChannel 根据名称和渠道获取供应商
	GetProviderByNameAndChannel(ctx context.Context, templateID, versionID int64, providerName string, channel domain.Channel) ([]domain.ChannelTemplateProvider, error)
}

// channelTemplateRepository 实现了ChannelTemplateRepository接口，提供模板数据的存储实现
type channelTemplateRepository struct {
	dao dao.ChannelTemplateDAO
}

// NewChannelTemplateRepository 创建仓储实例
func NewChannelTemplateRepository(dao dao.ChannelTemplateDAO) ChannelTemplateRepository {
	return &channelTemplateRepository{
		dao: dao,
	}
}

func (r *channelTemplateRepository) getTemplates(ctx context.Context, templates []dao.ChannelTemplate) ([]domain.ChannelTemplate, error) {
	// 提取模板IDs
	templateIDs := make([]int64, len(templates))
	for i := range templates {
		templateIDs[i] = templates[i].ID
	}

	// 获取所有模板关联的版本
	versions, err := r.dao.GetTemplateVersionsByTemplateIDs(ctx, templateIDs)
	if err != nil {
		return nil, err
	}

	// 提取版本IDs
	versionIDs := make([]int64, len(versions))
	for i := range versions {
		versionIDs[i] = versions[i].ID
	}

	// 获取所有版本关联的供应商
	providers, err := r.dao.GetProvidersByVersionIDs(ctx, versionIDs)
	if err != nil {
		return nil, err
	}

	// 构建版本ID到供应商列表的映射
	versionToProviders := make(map[int64][]domain.ChannelTemplateProvider)
	for i := range providers {
		domainProvider := r.toProviderDomain(providers[i])
		versionToProviders[providers[i].TemplateVersionID] = append(versionToProviders[providers[i].TemplateVersionID], domainProvider)
	}

	// 构建模板ID到版本列表的映射
	templateToVersions := make(map[int64][]domain.ChannelTemplateVersion)
	for i := range versions {
		domainVersion := r.toVersionDomain(versions[i])
		// 添加版本关联的供应商
		domainVersion.Providers = versionToProviders[versions[i].ID]
		templateToVersions[versions[i].ChannelTemplateID] = append(templateToVersions[versions[i].ChannelTemplateID], domainVersion)
	}

	// 构建最终的领域模型列表
	result := make([]domain.ChannelTemplate, len(templates))
	for i, t := range templates {
		domainTemplate := r.toTemplateDomain(t)
		// 添加模板关联的版本
		domainTemplate.Versions = templateToVersions[t.ID]
		result[i] = domainTemplate
	}

	return result, nil
}

func (r *channelTemplateRepository) GetTemplateByID(ctx context.Context, templateID int64) (domain.ChannelTemplate, error) {
	templateEntity, err := r.dao.GetTemplateByID(ctx, templateID)
	if err != nil {
		return domain.ChannelTemplate{}, err
	}
	templates, err := r.getTemplates(ctx, []dao.ChannelTemplate{templateEntity})
	if err != nil {
		return domain.ChannelTemplate{}, err
	}
	const first = 0
	return templates[first], nil
}

func (r *channelTemplateRepository) toTemplateDomain(daoTemplate dao.ChannelTemplate) domain.ChannelTemplate {
	return domain.ChannelTemplate{
		ID:              daoTemplate.ID,
		OwnerID:         daoTemplate.OwnerID,
		OwnerType:       domain.OwnerType(daoTemplate.OwnerType),
		Name:            daoTemplate.Name,
		Description:     daoTemplate.Description,
		Channel:         domain.Channel(daoTemplate.Channel),
		BusinessType:    domain.BusinessType(daoTemplate.BusinessType),
		ActiveVersionID: daoTemplate.ActiveVersionID,
		Ctime:           daoTemplate.Ctime,
		Utime:           daoTemplate.Utime,
	}
}

func (r *channelTemplateRepository) toVersionDomain(daoVersion dao.ChannelTemplateVersion) domain.ChannelTemplateVersion {
	return domain.ChannelTemplateVersion{
		ID:                       daoVersion.ID,
		ChannelTemplateID:        daoVersion.ChannelTemplateID,
		Name:                     daoVersion.Name,
		Signature:                daoVersion.Signature,
		Content:                  daoVersion.Content,
		Remark:                   daoVersion.Remark,
		AuditID:                  daoVersion.AuditID,
		AuditorID:                daoVersion.AuditorID,
		AuditTime:                daoVersion.AuditTime,
		AuditStatus:              domain.AuditStatus(daoVersion.AuditStatus),
		RejectReason:             daoVersion.RejectReason,
		LastReviewSubmissionTime: daoVersion.LastReviewSubmissionTime,
		Ctime:                    daoVersion.Ctime,
		Utime:                    daoVersion.Utime,
	}
}

func (r *channelTemplateRepository) toProviderDomain(daoProvider dao.ChannelTemplateProvider) domain.ChannelTemplateProvider {
	return domain.ChannelTemplateProvider{
		ID:                       daoProvider.ID,
		TemplateID:               daoProvider.TemplateID,
		TemplateVersionID:        daoProvider.TemplateVersionID,
		ProviderID:               daoProvider.ProviderID,
		ProviderName:             daoProvider.ProviderName,
		ProviderChannel:          domain.Channel(daoProvider.ProviderChannel),
		RequestID:                daoProvider.RequestID,
		ProviderTemplateID:       daoProvider.ProviderTemplateID,
		AuditStatus:              domain.AuditStatus(daoProvider.AuditStatus),
		RejectReason:             daoProvider.RejectReason,
		LastReviewSubmissionTime: daoProvider.LastReviewSubmissionTime,
		Ctime:                    daoProvider.Ctime,
		Utime:                    daoProvider.Utime,
	}
}

func (r *channelTemplateRepository) GetTemplateVersionByID(ctx context.Context, versionID int64) (domain.ChannelTemplateVersion, error) {
	version, err := r.dao.GetTemplateVersionByID(ctx, versionID)
	if err != nil {
		return domain.ChannelTemplateVersion{}, err
	}
	providers, err := r.dao.GetProvidersByVersionIDs(ctx, []int64{versionID})
	if err != nil {
		return domain.ChannelTemplateVersion{}, err
	}
	domainProviders := make([]domain.ChannelTemplateProvider, 0, len(providers))
	for i := range providers {
		domainProviders = append(domainProviders, r.toProviderDomain(providers[i]))
	}

	domainVersion := r.toVersionDomain(version)
	domainVersion.Providers = domainProviders
	return domainVersion, nil
}

func (r *channelTemplateRepository) GetProviderByNameAndChannel(ctx context.Context, templateID, versionID int64, providerName string, channel domain.Channel) ([]domain.ChannelTemplateProvider, error) {
	providers, err := r.dao.GetProviderByNameAndChannel(ctx, templateID, versionID, providerName, channel.String())
	if err != nil {
		return nil, err
	}
	results := make([]domain.ChannelTemplateProvider, len(providers))
	for i := range providers {
		results[i] = r.toProviderDomain(providers[i])
	}
	return results, nil
}
