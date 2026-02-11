package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"

	"link/internal/application/repository"
	"link/internal/application/service"
	"link/internal/config"
	"link/internal/container"
	"link/internal/handler"
)

var testUserService *service.UserService

// TestMain 测试主函数，在所有测试前执行
func TestMain(m *testing.M) {
	// 手动加载 .env 文件（从项目根目录）
	wd, _ := os.Getwd()
	projectRoot := filepath.Join(wd, "../..")
	envPath := filepath.Join(projectRoot, ".env")
	_ = godotenv.Load(envPath)

	// 初始化数据库
	gin.SetMode(gin.TestMode)
	dbConfig := config.LoadDatabaseConfig()
	jwtConfig := config.LoadJWTConfig()

	if err := container.InitDatabase(dbConfig); err != nil {
		panic(fmt.Sprintf("数据库初始化失败: %v", err))
	}

	db := container.GetDB()
	sqlDB, err := db.DB()
	if err != nil {
		panic(fmt.Sprintf("获取 sql.DB 失败: %v", err))
	}
	userRepo := repository.NewUserRepository(sqlDB)
	refreshTokenRepo := repository.NewRefreshTokenRepository(sqlDB)
	tenantRepo := repository.NewTenantRepository(db, true)
	testUserService = service.NewUserService(userRepo, refreshTokenRepo, tenantRepo, jwtConfig)

	// 运行测试
	code := m.Run()

	// 清理资源
	container.CloseDatabase()

	_ = code
}

// setupTestRouter 设置测试环境
func setupTestRouter() *gin.Engine {
	authHandler := handler.NewAuthHandler(testUserService)

	r := gin.New()
	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
		}
	}

	return r
}

// TestRegister_Success 测试成功注册
func TestRegister_Success(t *testing.T) {
	router := setupTestRouter()

	// 准备测试数据
	reqBody := map[string]string{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonData, _ := json.Marshal(reqBody)

	// 创建请求
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// 记录响应
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), response["code"])
	assert.Equal(t, "注册成功", response["message"])

	// 验证返回的数据包含 token 和用户信息
	data, ok := response["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, data["access_token"])
	assert.NotEmpty(t, data["refresh_token"])
}

// TestRegister_DuplicateEmail 测试重复邮箱注册
func TestRegister_DuplicateEmail(t *testing.T) {
	router := setupTestRouter()

	// 准备测试数据
	reqBody := map[string]string{
		"username": "user1",
		"email":    "duplicate@example.com",
		"password": "password123",
	}
	jsonData, _ := json.Marshal(reqBody)

	// 第一次注册
	req1, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// 第二次注册（相同邮箱）
	reqBody["username"] = "user2"
	jsonData2, _ := json.Marshal(reqBody)
	req2, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// 验证第二次注册失败
	assert.Equal(t, http.StatusBadRequest, w2.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w2.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(-1), response["code"])
	assert.Contains(t, response["message"], "邮箱已被注册")
}

// TestRegister_InvalidEmail 测试无效邮箱格式
func TestRegister_InvalidEmail(t *testing.T) {
	router := setupTestRouter()

	testCases := []struct {
		name  string
		email string
	}{
		{"无效邮箱1", "invalid"},
		{"无效邮箱2", "invalid@"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := map[string]string{
				"username": "testuser",
				"email":    tc.email,
				"password": "password123",
			}
			jsonData, _ := json.Marshal(reqBody)

			req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}
