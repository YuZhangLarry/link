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

	fmt.Println("=== 查看节点标签 ===")
	result, err := session.Run(ctx, `
		MATCH (n)
		RETURN DISTINCT labels(n) AS labels, count(n) AS count
		ORDER BY count DESC
		LIMIT 20
	`, nil)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}

	fmt.Println("\n节点标签统计：")
	count := 0
	for result.Next(ctx) {
		record := result.Record()
		labels, _ := record.Get("labels")
		countVal, _ := record.Get("count")
		count++
		fmt.Printf("[%d] Labels: %+v, Count: %v\n", count, labels, countVal)
	}

	fmt.Printf("\n总计: %d 种标签\n", count)

	// 查看具体节点示例
	fmt.Println("\n=== 查看节点示例 ===")
	result2, err := session.Run(ctx, `
		MATCH (n)
		RETURN n
		LIMIT 5
	`, nil)
	if err != nil {
		log.Fatalf("Failed to query nodes: %v", err)
	}

	nodeCount := 0
	for result2.Next(ctx) {
		record := result2.Record()
		if nodeValue, ok := record.Get("n"); ok {
			if node, ok := nodeValue.(neo4j.Node); ok {
				nodeCount++
				fmt.Printf("\n[节点 %d]\n", nodeCount)
				fmt.Printf("  ElementId: %s\n", node.GetElementId())
				// GetLabels 不存在，从 Pairs 获取
				fmt.Printf("  Labels: (无法直接获取)\n")
				fmt.Printf("  Props: %+v\n", node.GetProperties())
			}
		}
	}

	// 查看关系的 tenant_id 和 kb_id
	fmt.Println("\n=== 查看关系属性 ===")
	result3, err := session.Run(ctx, `
		MATCH ()-[r:RELATES_TO]->()
		RETURN r
		LIMIT 5
	`, nil)
	if err != nil {
		log.Fatalf("Failed to query relations: %v", err)
	}

	relCount := 0
	for result3.Next(ctx) {
		record := result3.Record()
		if relValue, ok := record.Get("r"); ok {
			if rel, ok := relValue.(neo4j.Relationship); ok {
				relCount++
				fmt.Printf("\n[关系 %d]\n", relCount)
				fmt.Printf("  ElementId: %s\n", rel.GetElementId())
				fmt.Printf("  Type: RELATES_TO\n")
				fmt.Printf("  Props: %v\n", rel.GetProperties())
			}
		}
	}
}
