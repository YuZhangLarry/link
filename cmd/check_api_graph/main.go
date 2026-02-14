package main

import (
	"context"
	"fmt"
	"log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"link/internal/config"
	"link/internal/container"
)

func main() {
	cfg := config.LoadNeo4jConfig()

	neo4jCfg := container.Config{
		URI:      cfg.URI,
		Username: cfg.Username,
		Password: cfg.Password,
	}

	ctx := context.Background()
	driver, err := container.CreateDriver(ctx, neo4jCfg)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	tenantID := "80"
	kbID := "4b856e03-953a-4221-8d7e-b2ee7b0b30b3"

	cypher := `
    MATCH (n)-[r:RELATES_TO]->(m)
    WHERE n.tenant_id = $tenant_id AND n.kb_id = $kb_id
    RETURN n.id AS source_id, n.name AS source_name,
           m.id AS target_id, m.name AS target_name,
           r.id AS rel_id, r.type AS rel_type,
           r.strength AS rel_strength, r.description AS rel_desc
    LIMIT 10
    `

	params := map[string]interface{}{
		"tenant_id": tenantID,
		"kb_id":     kbID,
	}

	result, err := session.Run(ctx, cypher, params)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	fmt.Printf("=== Graph Query Results (tenant=%s, kb=%s) ===\n", tenantID, kbID)
	count := 0
	for result.Next(ctx) {
		record := result.Record()
		sourceName := record.Values[1].(string)
		targetName := record.Values[3].(string)
		relID := record.Values[4].(string)
		relType := record.Values[5].(string)
		relStrength := record.Values[6]
		relDesc := record.Values[7]

		if relID == "rel-004" {
			fmt.Printf("✅ rel-004: %s -> %s | Type=%s Strength=%v Desc=%v\n",
				sourceName, targetName, relType, relStrength, relDesc)
		}
		count++
	}

	if err := result.Err(); err != nil {
		log.Fatalf("Error iterating: %v", err)
	}

	fmt.Printf("Total relationships found: %d\n", count)

	if count == 0 {
		fmt.Println("\n⚠️ No relationships found!")
		fmt.Println("Let's check what's actually in Neo4j...")

		checkCypher := `
        MATCH (n)-[r:RELATES_TO]->(m)
        RETURN n.id, n.name, n.tenant_id, n.kb_id,
               m.id, m.name, m.tenant_id, m.kb_id,
               r.id, r.type, r.strength
        LIMIT 20
        `

		result2, _ := session.Run(ctx, checkCypher, nil)
		fmt.Println("\nRaw Neo4j data:")
		for result2.Next(ctx) {
			rec := result2.Record()
			fmt.Printf("  %v(%s/%s) --[%s: %s strength=%v]--> %v(%s/%s)\n",
				rec.Values[1], rec.Values[2], rec.Values[3],
				rec.Values[8], rec.Values[9], rec.Values[10],
				rec.Values[5], rec.Values[6], rec.Values[7])
		}
	}
}
