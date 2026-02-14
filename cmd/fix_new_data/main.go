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
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)
	kbID := "4b856e03-953a-4221-8d7e-b2ee7b0b30b3"
	fmt.Println("=== 给新数据添加 kb_id ===")
	result1, _ := session.Run(ctx, "MATCH (n:ENTITY:KB_4b856e03) WHERE n.kb_id IS NULL SET n.kb_id = $kb_id RETURN count(n)", map[string]interface{}{"kb_id": kbID})
	if result1.Next(ctx) {
		fmt.Printf("1. 节点: %v\n", result1.Record().Values[0])
	}
	result2, _ := session.Run(ctx, "MATCH (n:ENTITY:KB_4b856e03)-[r:RELATES_TO]->(m:ENTITY:KB_4b856e03) WHERE r.kb_id IS NULL SET r.kb_id = $kb_id RETURN count(r)", map[string]interface{}{"kb_id": kbID})
	if result2.Next(ctx) {
		fmt.Printf("2. 关系: %v\n", result2.Record().Values[0])
	}
}
