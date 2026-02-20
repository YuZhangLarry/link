package service

import (
	"context"
	"fmt"
	"strings"

	"link/internal/types"
	"link/internal/types/interfaces"
)

// TenantService 租户服务
type TenantService struct {
	tenantRepo interfaces.TenantRepository
}

// NewTenantService 创建租户服务
func NewTenantService(tenantRepo interfaces.TenantRepository) *TenantService {
	return &TenantService{
		tenantRepo: tenantRepo,
	}
}

// CreateTenant 创建租户
func (s *TenantService) CreateTenant(ctx context.Context, req *types.CreateTenantRequest) (*types.TenantResponse, error) {
	// 检查租户名称是否已存在
	_, err := s.tenantRepo.FindByName(ctx, req.Name)
	if err == nil {
		return nil, fmt.Errorf("租户名称已存在")
	}

	// 设置默认存储配额 (10GB)
	storageQuota := req.StorageQuota
	if storageQuota == 0 {
		storageQuota = 10 * 1024 * 1024 * 1024 // 10GB
	}

	tenant := &types.Tenant{
		Name:         strings.TrimSpace(req.Name),
		Description:  strings.TrimSpace(req.Description),
		Business:     strings.TrimSpace(req.Business),
		Status:       "active",
		StorageQuota: storageQuota,
		StorageUsed:  0,
	}

	err = s.tenantRepo.Create(ctx, tenant)
	if err != nil {
		return nil, fmt.Errorf("创建租户失败: %w", err)
	}

	return tenant.ToResponse(), nil
}

// GetTenantByID 根据ID获取租户
func (s *TenantService) GetTenantByID(ctx context.Context, id int64) (*types.Tenant, error) {
	tenant, err := s.tenantRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return tenant, nil
}

// GetTenantByName 根据名称获取租户
func (s *TenantService) GetTenantByName(ctx context.Context, name string) (*types.TenantResponse, error) {
	tenant, err := s.tenantRepo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return tenant.ToResponse(), nil
}

// ListTenants 获取租户列表
func (s *TenantService) ListTenants(ctx context.Context, page, pageSize int) ([]*types.TenantResponse, int64, error) {
	tenants, total, err := s.tenantRepo.Find(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*types.TenantResponse, len(tenants))
	for i, tenant := range tenants {
		responses[i] = tenant.ToResponse()
	}

	return responses, total, nil
}

// UpdateTenant 更新租户
func (s *TenantService) UpdateTenant(ctx context.Context, id int64, req *types.UpdateTenantRequest) (*types.TenantResponse, error) {
	tenant, err := s.tenantRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.Name != "" {
		tenant.Name = strings.TrimSpace(req.Name)
	}
	if req.Description != "" {
		tenant.Description = strings.TrimSpace(req.Description)
	}
	if req.Business != "" {
		tenant.Business = strings.TrimSpace(req.Business)
	}
	if req.Status != "" {
		tenant.Status = req.Status
	}

	err = s.tenantRepo.Update(ctx, tenant)
	if err != nil {
		return nil, fmt.Errorf("更新租户失败: %w", err)
	}

	return tenant.ToResponse(), nil
}

// DeleteTenant 删除租户
func (s *TenantService) DeleteTenant(ctx context.Context, id int64) error {
	return s.tenantRepo.Delete(ctx, id)
}

// RegenerateAPIKey 重新生成API Key
func (s *TenantService) RegenerateAPIKey(ctx context.Context, id int64) (string, error) {
	apiKey, err := s.tenantRepo.GenerateAPIKey(ctx, id)
	if err != nil {
		return "", fmt.Errorf("重新生成 API Key 失败: %w", err)
	}
	return apiKey, nil
}

// ValidateAPIKey 验证 API Key 并返回租户ID
func (s *TenantService) ValidateAPIKey(ctx context.Context, apiKey string) (*types.Tenant, error) {
	tenant, err := s.tenantRepo.FindByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	// 检查租户状态
	if tenant.Status != "active" {
		return nil, fmt.Errorf("租户已被暂停或删除")
	}

	return tenant, nil
}

// ExtractTenantIDFromAPIKey 从 API Key 提取租户ID
// 这个方法用于中间件快速验证，不进行完整的 API Key 验证
// 完整验证应该使用 ValidateAPIKey
func (s *TenantService) ExtractTenantIDFromAPIKey(apiKey string) (int64, error) {
	if apiKey == "" {
		return 0, fmt.Errorf("API Key 不能为空")
	}

	// API Key 格式: tenant_<hex>
	// 这里只是解析格式，实际验证在 FindByAPIKey
	if !strings.HasPrefix(apiKey, "tenant_") {
		return 0, fmt.Errorf("无效的 API Key 格式")
	}

	// 需要通过数据库查询来验证并获取租户ID
	// 这里暂时返回错误，强制使用完整验证
	return 0, fmt.Errorf("请使用 ValidateAPIKey 进行完整验证")
}

// UpdateStorageUsed 更新存储使用量
func (s *TenantService) UpdateStorageUsed(ctx context.Context, tenantID int64, delta int64) error {
	return s.tenantRepo.UpdateStorageUsed(ctx, tenantID, delta)
}

// GetStorageUsage 获取存储使用情况
func (s *TenantService) GetStorageUsage(ctx context.Context, id int64) (quota int64, used int64, percentage float64, err error) {
	tenant, err := s.tenantRepo.FindByID(ctx, id)
	if err != nil {
		return 0, 0, 0, err
	}

	if tenant.StorageQuota <= 0 {
		return tenant.StorageQuota, tenant.StorageUsed, 0, nil
	}

	percentage = float64(tenant.StorageUsed) / float64(tenant.StorageQuota) * 100
	return tenant.StorageQuota, tenant.StorageUsed, percentage, nil
}
