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

	fmt.Println("=== 关系属性检查 ===")

	result, _ := session.Run(ctx, `MATCH ()-[r:RELATES_TO]->() RETURN r.id, r.type, r.strength, r.kb_id LIMIT 5`, nil)
	for result.Next(ctx) {
		rec := result.Record()
		fmt.Printf("id=%s type=%s strength=%v kb_id=%v\n", rec.Values[0], rec.Values[1], rec.Values[2], rec.Values[3])
	}
}
