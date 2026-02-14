package main

import (
	"context"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"link/internal/config"
	"link/internal/container"
)

func main() {
	cfg := config.LoadNeo4jConfig()
	neo4jCfg := container.Config{URI: cfg.URI, Username: cfg.Username, Password: cfg.Password}
	ctx := context.Background()
	driver, _ := container.CreateDriver(ctx, neo4jCfg)
	defer driver.Close(ctx)
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)
	kbID := "4b856e03-953a-4221-8d7e-b2ee7b0b30b3"
	fmt.Println("=== 检查真实 KB 数据 ===")
	fmt.Printf("KB ID: %s\n\n", kbID)
	result1, _ := session.Run(ctx, "MATCH (n) WHERE n.kb_id IS NOT NULL RETURN DISTINCT n.kb_id LIMIT 10", nil)
	fmt.Println("1. 所有不同的 kb_id:")
	for result1.Next(ctx) {
		fmt.Printf("  - %v\n", result1.Record().Values[0])
	}
	result2, _ := session.Run(ctx, "MATCH (n {kb_id: $kb_id}) RETURN count(n)", map[string]interface{}{"kb_id": kbID})
	if result2.Next(ctx) {
		fmt.Printf("\n2. 指定 kb_id 节点数: %v\n", result2.Record().Values[0])
	}
	result3, _ := session.Run(ctx, "MATCH ()-[r:RELATES_TO {kb_id: $kb_id}]->() RETURN count(r)", map[string]interface{}{"kb_id": kbID})
	if result3.Next(ctx) {
		fmt.Printf("3. 指定 kb_id 关系数: %v\n", result3.Record().Values[0])
	}
	result4, _ := session.Run(ctx, "MATCH (n:ENTITY:KB_4b856e03) RETURN count(n)", nil)
	if result4.Next(ctx) {
		fmt.Printf("4. ENTITY:KB_4b856e03 节点数: %v\n", result4.Record().Values[0])
	}
}
