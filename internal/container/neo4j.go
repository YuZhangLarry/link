package container

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Config Neo4j 连接配置
type Config struct {
	URI      string
	Username string
	Password string
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		URI:      "bolt://localhost:7687",
		Username: "neo4j",
		Password: "larry12345",
	}
}

// CreateDriver 创建并返回 Neo4j 驱动实例
func CreateDriver(ctx context.Context, config Config) (neo4j.DriverWithContext, error) {
	driver, err := neo4j.NewDriverWithContext(config.URI, neo4j.BasicAuth(config.Username, config.Password, ""))
	if err != nil {
		return nil, err
	}

	// 验证连接
	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		driver.Close(ctx)
		return nil, err
	}

	return driver, nil
}
