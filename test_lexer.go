package main

import (
	"fmt"
	"startdb/internal/sql"
)

func main() {
	lexer := sql.NewLexer("INSERT INTO users VALUES (1, 'John', 'john@example.com')")
	
	for {
		token := lexer.Next()
		fmt.Printf("Token: %s (Type: %d)\n", token.Literal, token.Type)
		if token.Type == sql.TokenEOF {
			break
		}
	}
}
