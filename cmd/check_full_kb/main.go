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
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	kbId := "4b856e03-953a-4221-8d7e-b2ee7b0b30b3"

	// 查询该知识库的所有关系
	fmt.Printf("=== 查询知识库 %s 的关系 ===\n", kbId)
	result, err := session.Run(ctx, `
		MATCH (n)-[r:RELATES_TO]->(m)
		WHERE n.knowledge_id = $kb_id AND m.knowledge_id = $kb_id
		RETURN r.id as id, n.name as source, m.name as target, r.type as type, r.description as description, r.strength as strength
		ORDER BY r.type
		LIMIT 100
	`, map[string]interface{}{"kb_id": kbId})
	if err != nil {
		log.Fatalf("Failed to query relations: %v", err)
	}

	emptyCount := 0
	nonEmptyCount := 0
	typeStats := make(map[string]int)

	fmt.Println("\n所有关系：")
	for result.Next(ctx) {
		record := result.Record()
		idVal, _ := record.Get("id")
		sourceVal, _ := record.Get("source")
		targetVal, _ := record.Get("target")
		typeVal, _ := record.Get("type")
		descVal, _ := record.Get("description")
		strengthVal, _ := record.Get("strength")

		id := fmt.Sprintf("%v", idVal)
		source := fmt.Sprintf("%v", sourceVal)
		target := fmt.Sprintf("%v", targetVal)
		var typeStr string
		if typeVal != nil {
			typeStr = fmt.Sprintf("%v", typeVal)
		}
		desc := ""
		if descVal != nil {
			desc = fmt.Sprintf("%v", descVal)
		}
		strength := 0.0
		if strengthVal != nil {
			strength = strengthVal.(float64)
		}

		if typeStr == "" || typeStr == "<nil>" {
			emptyCount++
			fmt.Printf("[空] ID=%s: %s -> %s (type='%s', desc: %s, strength: %.1f)\n", id, source, target, typeStr, desc, strength)
		} else {
			nonEmptyCount++
			typeStats[typeStr]++
			fmt.Printf("[√] ID=%s: %s -> %s (type='%s', desc: %s, strength: %.1f)\n", id, source, target, typeStr, desc, strength)
		}
	}

	fmt.Printf("\n=== 统计 ===\n")
	fmt.Printf("  有 type: %d 条\n", nonEmptyCount)
	fmt.Printf("  无 type (空): %d 条\n", emptyCount)
	fmt.Printf("  总计: %d 条\n\n", emptyCount+nonEmptyCount)

	fmt.Println("=== Type 分布 ===")
	for t, count := range typeStats {
		fmt.Printf("  %s: %d 条\n", t, count)
	}

	// 如果有空 type，批量更新
	if emptyCount > 0 {
		fmt.Println("\n发现空的 type 字段！开始批量更新...")
		updateSession := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})

		_, err := updateSession.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
			updateQuery := `
				MATCH (n)-[r:RELATES_TO]->(m)
				WHERE n.knowledge_id = $kb_id AND m.knowledge_id = $kb_id AND (r.type IS NULL OR r.type = '')
				SET r.type = '关联'
				RETURN count(r) as updated_count
			`
			result, err := tx.Run(ctx, updateQuery, map[string]interface{}{"kb_id": kbId})
			if err != nil {
				return nil, fmt.Errorf("failed to update: %w", err)
			}

			if result.Next(ctx) {
				record := result.Record()
				countVal, _ := record.Get("updated_count")
				fmt.Printf("✅ 已更新 %d 条关系\n", countVal)
			}

			return nil, nil
		})
		updateSession.Close(ctx)

		if err != nil {
			log.Printf("更新失败: %v", err)
		} else {
			fmt.Println("✅ 批量更新完成！")
		}
	} else {
		fmt.Println("\n✅ 所有关系都有 type 字段。")
		fmt.Println("\n如果前端仍显示空值，请尝试：")
		fmt.Println("1. 强制刷新浏览器（Ctrl+Shift+R）")
		fmt.Println("2. 清除浏览器缓存")
		fmt.Println("3. 重启后端服务")
	}
}
