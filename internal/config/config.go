package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
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

// Neo4jConfig Neo4j图数据库配置
type Neo4jConfig struct {
	URI         string // Neo4j连接URI，如: bolt://localhost:7687
	Username    string // 用户名，默认: neo4j
	Password    string // 密码
	MaxPoolSize int    // 连接池最大连接数
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret             string
	AccessTokenExpire  int
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

// TenantConfig 租户配置
type TenantConfig struct {
	EnableMultiTenant       bool  // 是否启用多租户
	EnableCrossTenantAccess bool  // 是否启用跨租户访问
	DefaultStorageQuota     int64 // 默认存储配额 (bytes)
}

// ServerConfig HTTP服务配置
type ServerConfig struct {
	Port string // HTTP服务端口
	Mode string // 运行模式: debug/release
	Host string // 监听地址
}

// Config 总配置
type Config struct {
	Database  *DatabaseConfig
	Milvus    *MilvusConfig
	Neo4j     *Neo4jConfig
	JWT       *JWTConfig
	Tenant    *TenantConfig
	Chat      *ChatConfig
	Search    *SearchConfig
	Embedding *EmbeddingConfig
	Server    *ServerConfig
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

// LoadNeo4jConfig 从环境变量加载Neo4j配置
func LoadNeo4jConfig() *Neo4jConfig {
	// 尝试加载 .env 文件
	projectRoot, _ := os.Getwd()
	envPath := filepath.Join(projectRoot, ".env")
	_ = godotenv.Load(envPath)

	return &Neo4jConfig{
		URI:         getEnv("NEO4J_URI", "bolt://localhost:7687"),
		Username:    getEnv("NEO4J_USERNAME", "neo4j"),
		Password:    getEnv("NEO4J_PASSWORD", ""),
		MaxPoolSize: getEnvAsInt("NEO4J_MAX_POOL_SIZE", 50),
	}
}

// LoadJWTConfig 从环境变量加载JWT配置
func LoadJWTConfig() *JWTConfig {
	// 尝试加载 .env 文件
	projectRoot, _ := os.Getwd()
	envPath := filepath.Join(projectRoot, ".env")
	_ = godotenv.Load(envPath)

	return &JWTConfig{
		Secret:             getEnv("JWT_SECRET", "your-secret-key"),
		AccessTokenExpire:  getEnvAsInt("JWT_ACCESS_TOKEN_EXPIRE", 86400),   // 24小时
		RefreshTokenExpire: getEnvAsInt("JWT_REFRESH_TOKEN_EXPIRE", 604800), // 7天
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
		Model:    getEnv("EMBEDDING_MODEL", "text-embedding-v3"),
		BaseURL:  getEnv("EMBEDDING_BASE_URL", "https://dashscope.aliyuncs.com/compatible-mode/v1/embeddings"),
	}
}

// LoadTenantConfig 从环境变量加载租户配置
func LoadTenantConfig() *TenantConfig {
	// 尝试加载 .env 文件
	projectRoot, _ := os.Getwd()
	envPath := filepath.Join(projectRoot, ".env")
	_ = godotenv.Load(envPath)

	return &TenantConfig{
		EnableMultiTenant:       getEnvAsBool("TENANT_ENABLED", false),
		EnableCrossTenantAccess: getEnvAsBool("TENANT_CROSS_ACCESS", false),
		DefaultStorageQuota:     getEnvAsInt64("TENANT_DEFAULT_QUOTA", 10*1024*1024*1024), // 10GB
	}
}

// LoadServerConfig 从环境变量加载服务配置
func LoadServerConfig() *ServerConfig {
	// 尝试加载 .env 文件
	projectRoot, _ := os.Getwd()
	envPath := filepath.Join(projectRoot, ".env")
	_ = godotenv.Load(envPath)

	return &ServerConfig{
		Port: getEnv("SERVER_PORT", "8080"),
		Mode: getEnv("GIN_MODE", "debug"),
		Host: getEnv("SERVER_HOST", "0.0.0.0"),
	}
}

// LoadConfig 加载完整配置
func LoadConfig() *Config {
	return &Config{
		Database:  LoadDatabaseConfig(),
		Milvus:    LoadMilvusConfig(),
		Neo4j:     LoadNeo4jConfig(),
		JWT:       LoadJWTConfig(),
		Tenant:    LoadTenantConfig(),
		Chat:      LoadChatConfig(),
		Search:    LoadSearchConfig(),
		Embedding: LoadEmbeddingConfig(),
		Server:    LoadServerConfig(),
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

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		var intValue int64
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
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

// PromptTemplate 提示词模板
type PromptTemplate struct {
	Templates []struct {
		ID      string `yaml:"id"`
		Content string `yaml:"content"`
	} `yaml:"templates"`
}

// LoadPromptTemplate 加载提示词模板
func LoadPromptTemplate(templateName string) (string, error) {
	projectRoot, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	templatePath := filepath.Join(projectRoot, "config", "prompt_templates", templateName+".yaml")
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	var pt PromptTemplate
	if err := yaml.Unmarshal(content, &pt); err != nil {
		return "", fmt.Errorf("failed to parse template YAML: %w", err)
	}

	if len(pt.Templates) == 0 {
		return "", fmt.Errorf("no templates found in %s", templatePath)
	}

	return pt.Templates[0].Content, nil
}
