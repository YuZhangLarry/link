# RBAC 权限系统测试指南

## ✅ 已完成的工作

### 1. 数据库表结构
- ✅ 执行 `migrations/rbac_rebuild_users.sql` 初始化RBAC表结构
- ✅ 删除了 `tenant_users` 表
- ✅ `users` 表添加 `tenant_id` 字段
- ✅ 创建了权限、角色相关表

### 2. 类型定义
- ✅ `internal/types/rbac.go` - 权限、角色相关类型
- ✅ `internal/types/user.go` - 用户类型添加 `tenant_id`
- ✅ `internal/models/entity.go` - 更新模型定义

### 3. 仓储层
- ✅ `internal/application/repository/role.go` - 角色仓储
- ✅ `internal/application/repository/user_role.go` - 用户角色仓储
- ✅ `internal/application/repository/permission.go` - 权限相关仓储
- ✅ `internal/application/repository/user.go` - 更新用户仓储

### 4. 服务层
- ✅ `internal/application/service/permission.go` - 权限服务

### 5. 中间件
- ✅ `internal/middleware/permission.go` - 权限检查中间件
- ✅ `internal/middleware/auth.go` - 更新认证中间件

### 6. Handler
- ✅ `internal/handler/permission.go` - 权限管理Handler
- ✅ `internal/handler/tenant.go` - 添加 `TenantRequired` 方法

### 7. 容器和路由
- ✅ `internal/container/permission.go` - 权限系统初始化
- ✅ `internal/router/routes.go` - 完整路由配置
- ✅ `cmd/server/main.go` - 启动时初始化权限系统

---

## 🧪 测试步骤

### 第一步：初始化数据库

```bash
# 1. 执行RBAC升级脚本
mysql -u root -p link_go < migrations/rbac_rebuild_users.sql

# 2. 验证表是否创建成功
mysql -u root -p link_go -e "SHOW TABLES;"

# 3. 检查初始化数据
mysql -u root -p link_go -e "SELECT * FROM permissions;"
mysql -u root -p link_go -e "SELECT * FROM roles;"
mysql -u root -p link_go -e "SELECT * FROM users;"
mysql -u root -p link_go -e "SELECT * FROM user_roles;"
```

### 第二步：启动服务

```bash
# 编译并启动服务
go run cmd/server/main.go
```

预期输出：
```
✅ 数据库初始化成功
🔧 正在初始化 Repository...
✅ 权限系统初始化成功
...
🚀 服务器启动在 http://localhost:8080
```

### 第三步：测试认证和租户验证

```bash
# 1. 登录（需要提供 tenant_id）
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": 1,
    "email": "admin@link.com",
    "password": "admin123"
  }'

# 预期响应：
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": 1234567890,
  "tenant_id": 1,
  "user": {
    "id": 1,
    "username": "admin",
    "email": "admin@link.com",
    "tenant_id": 1
  }
}
```

### 第四步：测试缺少租户ID

```bash
# 不提供 X-Tenant-ID 的请求应该失败
curl -X GET http://localhost:8080/api/v1/sessions \
  -H "Authorization: Bearer <access_token>"

# 预期响应：
{
  "code": 40000,
  "message": "缺少租户ID，请在请求头中添加 X-Tenant-ID"
}
```

### 第五步：测试正确的请求

```bash
# 提供租户ID的请求应该成功
curl -X GET http://localhost:8080/api/v1/sessions \
  -H "Authorization: Bearer <access_token>" \
  -H "X-Tenant-ID: 1"

# 预期响应：
{
  "sessions": [],
  "total": 0
}
```

### 第六步：测试权限系统

```bash
# 1. 获取用户角色
curl -X GET http://localhost:8080/api/v1/role \
  -H "Authorization: Bearer <access_token>" \
  -H "X-Tenant-ID: 1"

# 2. 获取用户权限
curl -X GET http://localhost:8080/api/v1/permissions \
  -H "Authorization: Bearer <access_token>" \
  -H "X-Tenant-ID: 1"

# 3. 检查权限
curl -X POST http://localhost:8080/api/v1/permissions/check \
  -H "Authorization: Bearer <access_token>" \
  -H "X-Tenant-ID: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "resource_type": "kb",
    "action": "create"
  }'

# 4. 获取角色列表
curl -X GET http://localhost:8080/api/v1/roles \
  -H "Authorization: Bearer <access_token>" \
  -H "X-Tenant-ID: 1"
```

---

## 📋 API 端点列表

### 公开接口（无需认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/register` | 用户注册（需要 tenant_id） |
| POST | `/api/v1/auth/login` | 用户登录（需要 tenant_id） |
| POST | `/api/v1/auth/refresh` | 刷新Token |
| GET | `/health` | 健康检查 |

### 需要认证的接口

| 方法 | 路径 | 说明 | 需要租户ID |
|------|------|------|-----------|
| GET | `/api/v1/tenants` | 获取租户列表 | ❌ |
| POST | `/api/v1/tenants` | 创建租户 | ❌ |
| GET | `/api/v1/users/profile` | 获取用户信息 | ❌ |
| GET | `/api/v1/sessions` | 获取会话列表 | ✅ |
| POST | `/api/v1/sessions` | 创建会话 | ✅ |
| GET | `/api/v1/sessions/:id` | 获取会话详情 | ✅ |
| PUT | `/api/v1/sessions/:id` | 更新会话 | ✅ |
| DELETE | `/api/v1/sessions/:id` | 删除会话 | ✅ |
| POST | `/api/v1/chat` | 聊天 | ✅ |
| POST | `/api/v1/chat/stream` | 流式聊天 | ✅ |

### 权限管理接口（需要认证 + 租户ID）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/roles` | 获取角色列表 |
| POST | `/api/v1/roles` | 创建角色 |
| PUT | `/api/v1/roles/:id` | 更新角色 |
| DELETE | `/api/v1/roles/:id` | 删除角色 |
| POST | `/api/v1/roles/assign` | 分配角色 |
| DELETE | `/api/v1/users/:user_id/role` | 撤销角色 |
| GET | `/api/v1/role` | 获取用户角色 |
| GET | `/api/v1/permissions` | 获取用户权限 |
| POST | `/api/v1/permissions/check` | 检查权限 |

---

## 🔍 故障排查

### 问题 1: 编译错误

```bash
# 清理缓存后重新编译
go clean -cache
go mod tidy
go build ./...
```

### 问题 2: 数据库连接失败

检查配置文件中的数据库连接信息：
- Host
- Port
- User
- Password
- Database

### 问题 3: 权限系统未初始化

查看日志中是否有：
```
✅ 权限系统初始化成功
```

如果看到警告信息：
```
⚠️  权限系统初始化失败
```
检查：
1. 数据库表是否已创建
2. 是否有初始数据

### 问题 4: Token 验证失败

确保：
1. Token 未过期
2. Token 包含 tenant_id
3. 使用 Bearer 认证方式

---

## 📝 默认测试数据

### 默认租户
- ID: 1
- Name: 默认租户

### 默认用户
- Username: admin
- Email: admin@link.com
- Password: admin123
- Tenant ID: 1

### 默认角色
- owner (所有者) - 拥有所有权限
- admin (管理员) - 管理权限
- user (普通用户) - 基本权限

---

## 🎯 下一步

1. **实现认证 Handler** - 完善登录、注册接口
2. **添加权限检查** - 在关键接口使用权限中间件
3. **编写单元测试** - 测试权限系统逻辑
4. **完善错误处理** - 统一错误响应格式
5. **添加日志记录** - 记录权限变更和审计日志
