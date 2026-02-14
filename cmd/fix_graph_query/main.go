package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	filePath := "D:\\link\\internal\\application\\repository\\graph.go"

	// Read the file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Find and replace the cypher query section
	inCypher := false
	newLines := make([]string, 0, len(lines))

	for _, line := range lines {
		// Skip lines that are part of the old cypher query
		if strings.Contains(line, "MATCH (n:Entity {tenant_id") {
			inCypher = true
			// Replace with new query
			newLines = append(newLines, "\tcypher := `")
			newLines = append(newLines, "\t\tMATCH (n)")
			newLines = append(newLines, "\t\tWHERE 'Entity' IN labels(n) OR 'ENTITY' IN labels(n) OR ANY(l IN labels(n) WHERE l CONTAINS $kb_prefix)")
			newLines = append(newLines, "\t\tOPTIONAL MATCH (n)-[r:RELATES_TO]->(m)")
			newLines = append(newLines, "\t\tWHERE 'Entity' IN labels(m) OR 'ENTITY' IN labels(m) OR ANY(l IN labels(m) WHERE l CONTAINS $kb_prefix)")
			newLines = append(newLines, "\t\tRETURN n, r, m")
			newLines = append(newLines, "\t\tLIMIT 5000")
			newLines = append(newLines, "\t`")
			continue
		}

		if inCypher {
			if strings.Contains(line, "result, err := session.Run(ctx, cypher, nil)") {
				// Replace with new params call
				newLines = append(newLines, "")
				newLines = append(newLines, "\tparams := map[string]interface{}{")
				newLines = append(newLines, "\t\t\"kb_prefix\": kbPrefix,")
				newLines = append(newLines, "\t}")
				newLines = append(newLines, "")
				newLines = append(newLines, "\tresult, err := session.Run(ctx, cypher, params)")
				inCypher = false
				continue
			}
			// Skip all lines in the old cypher query
			continue
		}

		newLines = append(newLines, line)
	}

	// Write the modified content back
	outFile, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer outFile.Close()

	for _, line := range newLines {
		fmt.Fprintln(outFile, line)
	}

	fmt.Printf("File updated successfully. Line %d\n", len(newLines))
}
