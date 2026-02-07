package container

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"link/internal/config"
)

var MilvusClient client.Client

// InitMilvus 初始化Milvus连接
func InitMilvus(cfg *config.MilvusConfig) error {
	if cfg.Host == "" || cfg.Token == "" {
		return fmt.Errorf("Milvus配置不完整: host或token为空")
	}

	// 配置 Milvus 客户端
	milvusCfg := client.Config{
		Address:  cfg.Host,
		APIKey:   cfg.Token,
		Username: "",
		Password: "",
	}

	// 创建客户端连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := client.NewClient(ctx, milvusCfg)
	if err != nil {
		return fmt.Errorf("创建Milvus客户端失败: %w", err)
	}

	// 测试连接 - 通过列出collections来验证
	_, err = c.ListCollections(ctx)
	if err != nil {
		return fmt.Errorf("Milvus连接测试失败: %w", err)
	}

	MilvusClient = c
	log.Printf("✅ Milvus连接成功: %s\n", cfg.Host)

	return nil
}

// CloseMilvus 关闭Milvus连接
func CloseMilvus() error {
	if MilvusClient != nil {
		return MilvusClient.Close()
	}
	return nil
}

// GetMilvus 获取Milvus客户端
func GetMilvus() client.Client {
	return MilvusClient
}
