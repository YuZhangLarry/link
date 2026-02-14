package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func main() {
	neo4jURI := os.Getenv("NEO4J_URI")
	neo4jUser := os.Getenv("NEO4J_USER")
	neo4jPassword := os.Getenv("NEO4J_PASSWORD")

	if neo4jURI == "" {
		neo4jURI = "bolt://localhost:7687"
	}
	if neo4jUser == "" {
		neo4jUser = "neo4j"
	}
	if neo4jPassword == "" {
		neo4jPassword = "larry12345"
	}

	driver, err := neo4j.NewDriverWithContext(neo4jURI, neo4j.BasicAuth(neo4jUser, neo4jPassword, ""))
	if err != nil {
		log.Fatalf("Failed to create Neo4j driver: %v", err)
	}
	defer driver.Close(context.Background())

	ctx := context.Background()

	fmt.Println("=== 修复 knowledge_id 问题 ===")
	fmt.Println()

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	// 1. 首先检查：有 KB 标签但没有 knowledge_id 的节点
	fmt.Println("1. 查找有 KB_4b856e03 标签但 knowledge_id 为 NULL 的节点...")
	result1, err := session.Run(ctx, `
		MATCH (n:ENTITY:KB_4b856e03)
		WHERE n.knowledge_id IS NULL
		RETURN count(n) as count
	`, nil)
	if err != nil {
		log.Fatalf("Failed: %v", err)
	}
	if result1.Next(ctx) {
		record := result1.Record()
		countVal, _ := record.Get("count")
		fmt.Printf("   找到 %d 个节点需要设置 knowledge_id\n", int(countVal.(int64)))
	}

	// 2. 为这些节点设置 knowledge_id（从任意一个有值的节点复制，或者使用默认值）
	_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 先获取一个有效的 knowledge_id 作为参考
		query1 := `
			MATCH (n:ENTITY:KB_4b856e03)
			WHERE n.knowledge_id IS NOT NULL
			RETURN n.knowledge_id as ref_kb_id
			LIMIT 1
		`
		result, err := tx.Run(ctx, query1, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get ref knowledge_id: %w", err)
		}

		refKBID := "4b856e03-953a-4221-8d7e-b2ee7b0b30b3" // 默认值
		if result.Next(ctx) {
			record := result.Record()
			val, _ := record.Get("ref_kb_id")
			if val != nil {
				refKBID = fmt.Sprintf("%v", val)
			}
		}

		// 更新所有 NULL knowledge_id
		query2 := `
			MATCH (n:ENTITY:KB_4b856e03)
			WHERE n.knowledge_id IS NULL
			SET n.knowledge_id = $kb_id
			RETURN count(n) as updated_count
		`
		result, err := tx.Run(ctx, query2, map[string]interface{}{"kb_id": refKBID})
		if err != nil {
			return nil, fmt.Errorf("failed to update knowledge_id: %w", err)
		}

		if result.Next(ctx) {
			record := result.Record()
			countVal, _ := record.Get("updated_count")
			fmt.Printf("   已更新 %d 个节点\n", int(countVal.(int64)))
		}

		return nil, nil
	})

	if err != nil {
		log.Printf("更新失败: %v", err)
	} else {
		fmt.Println("   ✅ knowledge_id 修复完成")
	}

	// 3. 验证修复结果
	fmt.Println("\n2. 验证修复结果...")
	session2 := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session2.Close(ctx)

	result2, err := session2.Run(ctx, `
		MATCH (n:ENTITY:KB_4b856e03)
		RETURN n.knowledge_id as kb_id, count(n) as count
		GROUP BY n.knowledge_id
		ORDER BY count DESC
		LIMIT 10
	`, nil)
	if err != nil {
		log.Fatalf("Failed: %v", err)
	}

	fmt.Println("   knowledge_id 分布：")
	totalCount := 0
	for result2.Next(ctx) {
		record := result2.Record()
		kbVal, _ := record.Get("kb_id")
		countVal, _ := record.Get("count")

		kbStr := "NULL"
		if kbVal != nil {
			kbStr = fmt.Sprintf("%v", kbVal)
		}
		count := int(countVal.(int64))
		totalCount += count

		fmt.Printf("     %s: %d 个节点\n", kbStr, count)
	}
	fmt.Printf("   总计: %d 个节点\n", totalCount)

	fmt.Println("\n=== 修复完成 ===")
	fmt.Println("现在可以重启后端服务了！")
}
