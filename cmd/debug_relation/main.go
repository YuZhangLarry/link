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

	fmt.Println("=== 检查关系 rel-001 的实际数据 ===")

	// 查询这个关系
	cypher := `
		MATCH ()-[r:RELATES_TO {id: 'rel-001'}]->()
		RETURN r
	`

	result, err := session.Run(ctx, cypher, nil)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}

	fmt.Println("\n关系原始数据:")
	count := 0
	for result.Next(ctx) {
		record := result.Record()
		if relValue, ok := record.Get("r"); ok {
			if rel, ok := relValue.(neo4j.Relationship); ok {
				count++
				props := rel.GetProperties()
				fmt.Printf("\n[关系 %d]\n", count)
				fmt.Printf("  ID: %s\n", rel.GetElementId())
				fmt.Printf("  Type: %v\n", props["type"])
				fmt.Printf("  Description: %v\n", props["description"])
				fmt.Printf("  Strength: %v\n", props["strength"])
				fmt.Printf("  Weight: %v\n", props["weight"])
				fmt.Printf("  Combined Degree: %v\n", props["combined_degree"])
				if src, ok := props["source"].(string); ok {
					fmt.Printf("  Source: %s\n", src)
				}
				if tgt, ok := props["target"].(string); ok {
					fmt.Printf("  Target: %s\n", tgt)
				}
			}
		}
	}

	fmt.Printf("\n总计: %d 条关系\n", count)
}
