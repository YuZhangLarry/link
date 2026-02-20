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

	fmt.Println("=== 知识库标签统计 ===")
	result, err := session.Run(ctx, `
		MATCH (n:ENTITY)
		RETURN labels(n) as labels, count(n) as count
		ORDER BY count DESC
	`, nil)
	if err != nil {
		log.Fatalf("Failed: %v", err)
	}

	fmt.Println("标签分布:")
	for result.Next(ctx) {
		record := result.Record()
		labelsVal, _ := record.Get("labels")
		countVal, _ := record.Get("count")

		labels := fmt.Sprintf("%v", labelsVal)
		count := int(countVal.(int64))

		fmt.Printf("  %s: %d 个节点\n", labels, count)
	}

	fmt.Println("\n=== 按 knowledge_id 分组统计 ===")
	result2, err := session.Run(ctx, `
		MATCH (n:ENTITY)
		RETURN n.knowledge_id as kb_id, count(n) as node_count,
		       size([(n)-[r:RELATES_TO]->())]) as rel_count
		ORDER BY node_count DESC
	`, nil)
	if err != nil {
		log.Fatalf("Failed: %v", err)
	}

	fmt.Println("知识库节点数:")
	for result2.Next(ctx) {
		record := result2.Record()
		kbVal, _ := record.Get("kb_id")
		nodeCountVal, _ := record.Get("node_count")
		relCountVal, _ := record.Get("rel_count")

		kbId := fmt.Sprintf("%v", kbVal)
		nodeCount := int(nodeCountVal.(int64))
		relCount := int(relCountVal.(int64))

		fmt.Printf("  KB: %s | 节点: %d | 关系: %d\n", kbId, nodeCount, relCount)
	}
}
