package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"link/internal/types"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

// MilvusConfig Milvus配置
type MilvusConfig struct {
	Host  string
	Token string
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret            string
	AccessTokenExpire int
	RefreshTokenExpire int
}

// ChatConfig 聊天配置
type ChatConfig struct {
	Source    types.ModelSource // 模型源: local/remote
	BaseURL   string            // API Base URL
	ModelName string            // 模型名称
	APIKey    string            // API密钥
	Provider  string            // Provider: openai, aliwen, deepseek等
}

// SearchConfig 搜索配置
type SearchConfig struct {
	MetasoAPIKey string // Metaso 搜索 API Key
	APIEndpoint  string // 搜索 API 端点
}

// EmbeddingConfig Embedding 配置
type EmbeddingConfig struct {
	Provider string // 提供商: dashscope, openai, etc
	APIKey   string // API 密钥
	Model    string // 模型名称
	BaseURL  string // API Base URL
}

// LoadDatabaseConfig 从环境变量加载数据库配置
func LoadDatabaseConfig() *DatabaseConfig {
	// 尝试加载 .env 文件
	projectRoot, _ := os.Getwd()
	envPath := filepath.Join(projectRoot, ".env")
	_ = godotenv.Load(envPath)

	return &DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "3306"),
		User:     getEnv("DB_USER", "root"),
		Password: getEnv("DB_PASSWORD", ""),
		Database: getEnv("DB_NAME", "link_go"),
	}
}

// LoadMilvusConfig 从环境变量加载Milvus配置
func LoadMilvusConfig() *MilvusConfig {
	// 尝试加载 .env 文件
	projectRoot, _ := os.Getwd()
	envPath := filepath.Join(projectRoot, ".env")
	_ = godotenv.Load(envPath)

	return &MilvusConfig{
		Host:  getEnv("MILVUS_HOST", ""),
		Token: getEnv("MILVUS_TOKEN", ""),
	}
}

// LoadJWTConfig 从环境变量加载JWT配置
func LoadJWTConfig() *JWTConfig {
	// 尝试加载 .env 文件
	projectRoot, _ := os.Getwd()
	envPath := filepath.Join(projectRoot, ".env")
	_ = godotenv.Load(envPath)

	return &JWTConfig{
		Secret:            getEnv("JWT_SECRET", "your-secret-key"),
		AccessTokenExpire:  getEnvAsInt("JWT_ACCESS_TOKEN_EXPIRE", 86400),    // 24小时
		RefreshTokenExpire: getEnvAsInt("JWT_REFRESH_TOKEN_EXPIRE", 604800),  // 7天
	}
}

// LoadChatConfig 从环境变量加载聊天配置
func LoadChatConfig() *ChatConfig {
	// 尝试加载 .env 文件
	projectRoot, _ := os.Getwd()
	envPath := filepath.Join(projectRoot, ".env")
	_ = godotenv.Load(envPath)

	source := types.ModelSource(getEnv("CHAT_SOURCE", string(types.ModelSourceRemote)))

	return &ChatConfig{
		Source:    source,
		BaseURL:   getEnv("CHAT_BASE_URL", "https://api.openai.com/v1"),
		ModelName: getEnv("CHAT_MODEL_NAME", "gpt-3.5-turbo"),
		APIKey:    getEnv("CHAT_API_KEY", ""),
		Provider:  getEnv("CHAT_PROVIDER", "openai"),
	}
}

// LoadSearchConfig 从环境变量加载搜索配置
func LoadSearchConfig() *SearchConfig {
	// 尝试加载 .env 文件
	projectRoot, _ := os.Getwd()
	envPath := filepath.Join(projectRoot, ".env")
	_ = godotenv.Load(envPath)

	return &SearchConfig{
		MetasoAPIKey: getEnv("METASO_API_KEY", ""),
		APIEndpoint:  getEnv("SEARCH_API_ENDPOINT", "https://metaso.cn/api/v1/search"),
	}
}

// LoadEmbeddingConfig 从环境变量加载 Embedding 配置
func LoadEmbeddingConfig() *EmbeddingConfig {
	// 尝试加载 .env 文件
	projectRoot, _ := os.Getwd()
	envPath := filepath.Join(projectRoot, ".env")
	_ = godotenv.Load(envPath)

	return &EmbeddingConfig{
		Provider: getEnv("EMBEDDING_PROVIDER", "dashscope"),
		APIKey:   getEnv("EMBEDDING_API_KEY", ""),
		Model:    getEnv("EMBEDDING_MODEL", "text-embedding-v4"),
		BaseURL:  getEnv("EMBEDDING_BASE_URL", "https://dashscope.aliyuncs.com/compatible-mode/v1/embeddings"),
	}
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
