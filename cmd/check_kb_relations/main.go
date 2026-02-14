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

	// 查询知识库 4b856e03 的所有关系
	fmt.Println("=== 查询知识库 KB_4b856e03 的关系 ===")
	result, err := session.Run(ctx, `
		MATCH (n:ENTITY:KB_4b856e03)-[r:RELATES_TO]->(m:ENTITY:KB_4b856e03)
		RETURN r.id as id, n.name as source, m.name as target, r.type as type, r.description as description
		ORDER BY r.type
		LIMIT 50
	`, nil)
	if err != nil {
		log.Fatalf("Failed to query relations: %v", err)
	}

	emptyCount := 0
	nonEmptyCount := 0
	typeStats := make(map[string]int)

	fmt.Println("\n前30条关系：")
	for result.Next(ctx) {
		record := result.Record()
		idVal, _ := record.Get("id")
		sourceVal, _ := record.Get("source")
		targetVal, _ := record.Get("target")
		typeVal, _ := record.Get("type")
		descVal, _ := record.Get("description")

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

		if typeStr == "" || typeStr == "<nil>" {
			emptyCount++
			fmt.Printf("[空] ID=%s: %s -> %s (desc: %s)\n", id, source, target, desc)
		} else {
			nonEmptyCount++
			typeStats[typeStr]++
			if nonEmptyCount <= 30 {
				fmt.Printf("[%s] ID=%s: %s -> %s (desc: %s)\n", typeStr, id, source, target, desc)
			}
		}
	}

	fmt.Printf("\n统计：\n")
	fmt.Printf("  有 type: %d 条\n", nonEmptyCount)
	fmt.Printf("  无 type (空): %d 条\n", emptyCount)
	fmt.Printf("  总计: %d 条\n", emptyCount+nonEmptyCount)

	fmt.Println("\n=== Type 分布 ===")
	for t, count := range typeStats {
		fmt.Printf("  %s: %d 条\n", t, count)
	}

	// 检查是否需要更新
	if emptyCount > 0 {
		fmt.Println("\n发现空的 type 字段！需要更新。")
	} else {
		fmt.Println("\n✅ 所有关系都有 type 字段。")
		fmt.Println("\n如果前端仍显示空值，请尝试：")
		fmt.Println("1. 强制刷新浏览器（Ctrl+Shift+R）")
		fmt.Println("2. 清除浏览器缓存")
		fmt.Println("3. 重启后端服务")
	}
}
