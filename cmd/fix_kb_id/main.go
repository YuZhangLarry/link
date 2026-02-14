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

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	kbID := "4b856e03-953a-4221-8d7e-b2ee7b0b30b3"

	cypher := `
    MATCH (n)
    WHERE n.kb_id IS NULL
    SET n.kb_id = $kb_id
    RETURN count(n) as updated
    `

	params := map[string]interface{}{
		"kb_id": kbID,
	}

	result, err := session.Run(ctx, cypher, params)
	if err != nil {
		log.Fatalf("Update failed: %v", err)
	}

	if result.Next(ctx) {
		count := result.Record().Values[0]
		fmt.Printf("Updated %d nodes with kb_id=%s\n", count, kbID)
	}
}
