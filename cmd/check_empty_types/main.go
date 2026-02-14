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
	if err := driver.VerifyConnectivity(ctx); err != nil {
		log.Fatalf("Failed to verify Neo4j connectivity: %v", err)
	}

	log.Println("Connected to Neo4j successfully")
	fmt.Println()

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	// 1. 查询所有关系（包括 type 为空的）
	fmt.Println("=== 所有关系统计 ===")
	result, err := session.Run(ctx, `
		MATCH (n)-[r:RELATES_TO]->(m)
		RETURN r.id as id, n.name as source, m.name as target, r.type as type, r.description as description
		ORDER BY r.type
		LIMIT 300
	`, nil)
	if err != nil {
		log.Fatalf("Failed to query relations: %v", err)
	}

	emptyCount := 0
	nonEmptyCount := 0
	typeStats := make(map[string]int)

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
			if emptyCount <= 10 {
				fmt.Printf("[EMPTY] %s: %s -> %s (desc: %s)\n", id, source, target, desc)
			}
		} else {
			nonEmptyCount++
			typeStats[typeStr]++
		}
	}

	fmt.Printf("\n统计：\n")
	fmt.Printf("  有 type: %d 条\n", nonEmptyCount)
	fmt.Printf("  无 type (空): %d 条\n", emptyCount)
	fmt.Printf("  总计: %d 条\n\n", emptyCount+nonEmptyCount)

	fmt.Println("=== Type 分布 ===")
	for t, count := range typeStats {
		fmt.Printf("  %s: %d 条\n", t, count)
	}

	// 2. 查询前端的示例数据
	fmt.Println("\n=== 检查前端显示的关系 ===")
	sampleIDs := []string{"5", "9", "10", "8", "12", "6"}
	for _, id := range sampleIDs {
		result, err := session.Run(ctx, `
			MATCH (n)-[r:RELATES_TO {id: $id}]->(m)
			RETURN r.id as id, n.name as source, m.name as target, r.type as type, r.description as description, r.strength as strength, r.weight as weight
		`, map[string]interface{}{"id": id})
		if err != nil {
			log.Printf("Failed to query relation %s: %v", id, err)
			continue
		}

		if result.Next(ctx) {
			record := result.Record()
			idVal, _ := record.Get("id")
			sourceVal, _ := record.Get("source")
			targetVal, _ := record.Get("target")
			typeVal, _ := record.Get("type")
			descVal, _ := record.Get("description")
			strengthVal, _ := record.Get("strength")
			weightVal, _ := record.Get("weight")

			fmt.Printf("\nID: %v\n", idVal)
			fmt.Printf("  Source: %v\n", sourceVal)
			fmt.Printf("  Target: %v\n", targetVal)
			fmt.Printf("  Type: '%v'\n", typeVal)
			fmt.Printf("  Description: %v\n", descVal)
			fmt.Printf("  Strength: %v\n", strengthVal)
			fmt.Printf("  Weight: %v\n", weightVal)
		} else {
			fmt.Printf("\nID %s: NOT FOUND\n", id)
		}
	}
}
