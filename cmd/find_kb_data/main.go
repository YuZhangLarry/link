package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

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
	kbShort := "4b856e03"

	fmt.Printf("=== 查询知识库 %s 的数据 ===\n", kbId)
	fmt.Println()

	// 1. 查询节点（按 knowledge_id）
	fmt.Println("1. 按 knowledge_id 查询节点:")
	result1, err := session.Run(ctx, `
		MATCH (n:ENTITY)
		WHERE n.knowledge_id = $kb_id
		RETURN count(n) as count
	`, map[string]interface{}{"kb_id": kbId})
	if err != nil {
		log.Fatalf("Failed: %v", err)
	}
	if result1.Next(ctx) {
		record := result1.Record()
		countVal, _ := record.Get("count")
		fmt.Printf("   完整KB ID匹配: %d 个节点\n", int(countVal.(int64)))
	}

	// 2. 查询节点（按 KB 前缀标签）
	result2, err := session.Run(ctx, `
		MATCH (n:ENTITY:KB_4b856e03)
		RETURN count(n) as count
	`, nil)
	if err != nil {
		log.Fatalf("Failed: %v", err)
	}
	if result2.Next(ctx) {
		record := result2.Record()
		countVal, _ := record.Get("count")
		fmt.Printf("   KB前缀标签匹配: %d 个节点\n", int(countVal.(int64)))
	}

	// 3. 列出所有不同的 knowledge_id
	fmt.Println("\n2. 所有不同的 knowledge_id:")
	result3, err := session.Run(ctx, `
		MATCH (n:ENTITY)
		WHERE n.knowledge_id IS NOT NULL
		RETURN DISTINCT n.knowledge_id as kb_id, count(n) as node_count
		ORDER BY node_count DESC
		LIMIT 10
	`, nil)
	if err != nil {
		log.Fatalf("Failed: %v", err)
	}

	fmt.Println("   KB ID                                    | 节点数")
	fmt.Println("   " + strings.Repeat("-", 50))
	for result3.Next(ctx) {
		record := result3.Record()
		kbVal, _ := record.Get("kb_id")
		countVal, _ := record.Get("node_count")

		kbIdStr := fmt.Sprintf("%v", kbVal)
		count := int(countVal.(int64))

		// 高亮显示匹配的 KB
		if strings.HasPrefix(kbIdStr, kbShort) {
			fmt.Printf(" * %s | %d ⬅\n", kbIdStr, count)
		} else {
			fmt.Printf("   %s | %d\n", kbIdStr, count)
		}
	}

	// 4. 检查前端显示的那些具体关系ID
	fmt.Println("\n3. 检查前端显示的关系ID:")
	sampleIDs := []string{"5", "9", "10", "8", "12", "6"}
	for _, id := range sampleIDs {
		result4, err := session.Run(ctx, `
			MATCH ()-[r:RELATES_TO {id: $id}]->()
			RETURN r.id as id, r.type as type, r.source as source, r.target as target, r.description as description
		`, map[string]interface{}{"id": id})
		if err != nil {
			log.Printf("Failed to query relation %s: %v", id, err)
			continue
		}

		if result4.Next(ctx) {
			record := result4.Record()
			idVal, _ := record.Get("id")
			typeVal, _ := record.Get("type")
			sourceVal, _ := record.Get("source")
			targetVal, _ := record.Get("target")
			descVal, _ := record.Get("description")

			fmt.Printf("   ID=%s: type='%v', source='%v', target='%v', desc='%v'\n",
				fmt.Sprintf("%v", idVal), typeVal, sourceVal, targetVal, descVal)
		} else {
			fmt.Printf("   ID=%s: 不存在\n", id)
		}
	}
}
