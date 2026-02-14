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

	// 查询最近添加的关系（按时间倒序）
	fmt.Println("=== 查询最近添加的关系 ===")
	result, err := session.Run(ctx, `
		MATCH ()-[r:RELATES_TO]->()
		RETURN r
		ORDER BY r.id DESC
		LIMIT 5
	`, nil)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}

	fmt.Println("\n最近的关系：")
	count := 0
	for result.Next(ctx) {
		record := result.Record()
		rel, ok := record.Get("r")
		if !ok {
			continue
		}
		count++

		// 直接打印整个关系对象
		fmt.Printf("[%d] Relation: %+v\n", count, rel)
	}

	fmt.Printf("\n总计: %d 条关系\n", count)
}
