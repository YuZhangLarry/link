# 用户认证系统文档

## 📁 项目结构

```
internal/
├── handler/
│   └── auth.go                    # HTTP处理层 - 注册/登录/登出/刷新token
│
├── middleware/
│   └── auth.go                    # JWT认证中间件 - 验证token、跨租户访问
│
├── application/
│   ├── service/
│   │   └── user.go                # 业务逻辑层 - 用户注册、登录、token生成/验证
│   └── repository/
│       └── user.go                # 数据访问层 - 用户CRUD
│
├── types/
│   ├── user.go                    # 用户数据结构定义
│   └── interfaces/
│       └── user.go                # UserService接口定义
│
├── config/
│   └── config.go                  # 配置管理（数据库、JWT）
│
└── database/
    └── mysql.go                   # MySQL连接管理
```

## 🏗️ 各层职责

| 层级 | 文件 | 主要功能 |
|------|------|----------|
| HTTP层 | handler/auth.go | Register(), Login(), Logout(), RefreshToken(), GetProfile() |
| 中间件 | middleware/auth.go | JWT token验证, 认证中间件, 用户信息获取 |
| 服务层 | service/user.go | 密码hash, JWT生成/验证, 用户业务逻辑 |
| 数据层 | repository/user.go | 数据库操作 (MySQL) |
| 类型定义 | types/user.go | User, RegisterRequest, LoginRequest等结构体 |

## 🗄️ 数据库表结构

### users 表
```sql
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    avatar VARCHAR(255),
    status TINYINT DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_login_at DATETIME
);
```

### refresh_tokens 表
```sql
CREATE TABLE refresh_tokens (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

## 🔌 API 路由

### 认证相关（无需Token）

- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/refresh` - 刷新Token

### 用户相关（需要Token）

- `POST /api/v1/auth/logout` - 用户登出
- `GET /api/v1/user/profile` - 获取当前用户信息

### 其他

- `GET /health` - 健康检查

## 📝 API 使用示例

### 1. 用户注册

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
```

**响应：**
```json
{
  "code": 0,
  "message": "注册成功",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": 1770383662,
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "avatar": "",
      "status": 1,
      "created_at": "2026-02-05T21:14:22+08:00"
    }
  }
}
```

### 2. 用户登录

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

**响应：**
```json
{
  "code": 0,
  "message": "登录成功",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": 1770383664,
    "user": { ... }
  }
}
```

### 3. 获取用户信息（需要Token）

```bash
curl -X GET http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**响应：**
```json
{
  "code": 0,
  "message": "获取成功",
  "data": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "avatar": "",
    "status": 1,
    "created_at": "2026-02-05T21:14:22+08:00"
  }
}
```

### 4. 刷新Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

**响应：**
```json
{
  "code": 0,
  "message": "刷新成功",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": 1770383713,
    "user": { ... }
  }
}
```

### 5. 用户登出（需要Token）

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**响应：**
```json
{
  "code": 0,
  "message": "登出成功"
}
```

## 🔐 安全特性

1. **密码加密**：使用 bcrypt 加密存储密码
2. **JWT认证**：基于 JSON Web Token 的无状态认证
3. **Token刷新**：支持刷新Token机制，延长用户会话
4. **Token过期**：AccessToken 24小时过期，RefreshToken 7天过期
5. **数据库存储**：RefreshToken哈希存储在数据库，支持撤销

## ⚙️ 配置说明

在 `.env` 文件中配置：

```env
# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=1234
DB_NAME=link_go

# JWT 配置
JWT_SECRET=your-secret-key-change-this-in-production
JWT_ACCESS_TOKEN_EXPIRE=86400      # 24小时（秒）
JWT_REFRESH_TOKEN_EXPIRE=604800    # 7天（秒）
```

## 🚀 启动服务

```bash
# 执行数据库迁移
mysql -u root -p link_go < migrations/001_create_users_tables.sql

# 启动服务器
go run cmd/base/main.go
```

服务器将在 `http://localhost:8080` 启动。

## 📊 测试结果

所有API端点测试通过：

✅ 健康检查
✅ 用户注册
✅ 用户登录
✅ 获取用户信息
✅ 刷新Token
✅ 用户登出

## 🔧 技术栈

- **Web框架**: Gin v1.11.0
- **数据库**: MySQL
- **认证**: JWT (golang-jwt/jwt/v5)
- **密码加密**: bcrypt
- **配置管理**: godotenv
