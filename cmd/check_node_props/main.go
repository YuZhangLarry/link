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

	fmt.Println("=== 检查 Entity 节点的属性 ===")

	// 查询 Entity 节点及其属性
	cypher := `
		MATCH (n:Entity)
		RETURN n
		LIMIT 10
	`

	result, err := session.Run(ctx, cypher, nil)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}

	fmt.Println("\nEntity 节点示例：")
	count := 0
	for result.Next(ctx) {
		record := result.Record()
		if nodeValue, ok := record.Get("n"); ok {
			if node, ok := nodeValue.(neo4j.Node); ok {
				count++
				props := node.GetProperties()
				fmt.Printf("\n[节点 %d]\n", count)
				fmt.Printf("  ElementId: %s\n", node.GetElementId())
				fmt.Printf("  Name: %v\n", props["name"])
				fmt.Printf("  Entity Type: %v\n", props["entity_type"])
				fmt.Printf("  Tenant ID: %v\n", props["tenant_id"])
				fmt.Printf("  KB ID: %v\n", props["kb_id"])
			}
		}
	}

	// 统计有 tenant_id 和 kb_id 的节点
	fmt.Println("\n=== 统计节点属性 ===")
	countCypher := `
		MATCH (n:Entity)
		RETURN
			count(n) as total,
			count(n.tenant_id) as with_tenant,
			count(n.kb_id) as with_kb,
			count(CASE WHEN n.tenant_id IS NOT NULL AND n.kb_id IS NOT NULL THEN 1 END) as with_both
	`

	countResult, err := session.Run(ctx, countCypher, nil)
	if err != nil {
		log.Fatalf("Failed to count: %v", err)
	}

	if countResult.Next(ctx) {
		record := countResult.Record()
		total, _ := record.Get("total")
		withTenant, _ := record.Get("with_tenant")
		withKb, _ := record.Get("with_kb")
		withBoth, _ := record.Get("with_both")
		fmt.Printf("\nEntity 节点统计:\n")
		fmt.Printf("  总节点数: %v\n", total)
		fmt.Printf("  有 tenant_id: %v\n", withTenant)
		fmt.Printf("  有 kb_id: %v\n", withKb)
		fmt.Printf("  同时有两者: %v\n", withBoth)
	}

	// 检查不同标签的节点
	fmt.Println("\n=== 按标签统计节点属性 ===")
	labelStatsCypher := `
		MATCH (n)
		WITH labels(n) as labels, count(n) as count
		RETURN labels, count
		ORDER BY count DESC
	`

	labelResult, err := session.Run(ctx, labelStatsCypher, nil)
	if err != nil {
		log.Fatalf("Failed to query labels: %v", err)
	}

	fmt.Println("\n标签分布:")
	for labelResult.Next(ctx) {
		record := labelResult.Record()
		labels, _ := record.Get("labels")
		countVal, _ := record.Get("count")
		fmt.Printf("  Labels: %v, Count: %v\n", labels, countVal)
	}

	// 检查关系的属性
	fmt.Println("\n=== 检查关系属性 ===")
	relStatsCypher := `
		MATCH ()-[r:RELATES_TO]->()
		RETURN
			count(r) as total,
			count(r.tenant_id) as with_tenant,
			count(r.kb_id) as with_kb,
			count(CASE WHEN r.tenant_id IS NOT NULL AND r.kb_id IS NOT NULL THEN 1 END) as with_both
	`

	relResult, err := session.Run(ctx, relStatsCypher, nil)
	if err != nil {
		log.Fatalf("Failed to count relations: %v", err)
	}

	if relResult.Next(ctx) {
		record := relResult.Record()
		total, _ := record.Get("total")
		withTenant, _ := record.Get("with_tenant")
		withKb, _ := record.Get("with_kb")
		withBoth, _ := record.Get("with_both")
		fmt.Printf("\nRELATES_TO 关系统计:\n")
		fmt.Printf("  总关系数: %v\n", total)
		fmt.Printf("  有 tenant_id: %v\n", withTenant)
		fmt.Printf("  有 kb_id: %v\n", withKb)
		fmt.Printf("  同时有两者: %v\n", withBoth)
	}

	// 测试 GetGraph 的查询
	fmt.Println("\n=== 测试 GetGraph 查询 (tenant_id=1, kb_id=4b856e03-953a-4221-8d7e-b2ee7b0b30b3) ===")
	getGraphCypher := `
		MATCH (n:Entity {tenant_id: $tenant_id, kb_id: $kb_id})
		OPTIONAL MATCH (n)-[r:RELATES_TO]->(m:Entity {tenant_id: $tenant_id, kb_id: $kb_id})
			WHERE n.tenant_id = $tenant_id
			  AND n.kb_id = $kb_id
		RETURN count(DISTINCT n) as nodes, count(r) as relations
	`

	getGraphParams := map[string]interface{}{
		"tenant_id": "1",
		"kb_id":     "4b856e03-953a-4221-8d7e-b2ee7b0b30b3",
	}

	graphResult, err := session.Run(ctx, getGraphCypher, getGraphParams)
	if err != nil {
		log.Fatalf("Failed to query graph: %v", err)
	}

	if graphResult.Next(ctx) {
		record := graphResult.Record()
		nodes, _ := record.Get("nodes")
		relations, _ := record.Get("relations")
		fmt.Printf("  找到节点: %v\n", nodes)
		fmt.Printf("  找到关系: %v\n", relations)
	}

	// 也测试标签匹配的查询
	fmt.Println("\n=== 测试标签匹配查询 ===")
	labelMatchCypher := `
		MATCH (n)
		WHERE 'Entity' IN labels(n) OR 'ENTITY' IN labels(n) OR ANY(l IN labels(n) WHERE l STARTS WITH 'ENTITY')
		RETURN count(DISTINCT n) as nodes
	`

	labelRes, err := session.Run(ctx, labelMatchCypher, nil)
	if err != nil {
		log.Fatalf("Failed to query labels: %v", err)
	}

	if labelRes.Next(ctx) {
		record := labelRes.Record()
		nodes, _ := record.Get("nodes")
		fmt.Printf("  找到节点: %v\n", nodes)
	}

	// 测试原始的GetGraph查询（使用KB前缀匹配）
	fmt.Println("\n=== 测试原始 GetGraph 查询 (KB前缀匹配) ===")
	kbPrefix := "KB_4b856e03"
	oldGetGraphCypher := fmt.Sprintf(`
		MATCH (n)
		WHERE 'Entity' IN labels(n) OR 'ENTITY' IN labels(n) OR ANY(l IN labels(n) WHERE l CONTAINS '%s')
		OPTIONAL MATCH (n)-[r:RELATES_TO]->(m)
		WHERE 'Entity' IN labels(m) OR 'ENTITY' IN labels(m) OR ANY(l IN labels(m) WHERE l CONTAINS '%s')
		RETURN count(DISTINCT n) as nodes, count(r) as relations
	`, kbPrefix, kbPrefix)

	oldResult, err := session.Run(ctx, oldGetGraphCypher, nil)
	if err != nil {
		log.Fatalf("Failed to query old graph: %v", err)
	}

	if oldResult.Next(ctx) {
		record := oldResult.Record()
		nodes, _ := record.Get("nodes")
		relations, _ := record.Get("relations")
		fmt.Printf("  找到节点: %v\n", nodes)
		fmt.Printf("  找到关系: %v\n", relations)
	}
}
