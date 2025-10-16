package main

import (
	"startdb/internal/cli"
)

func main() {
	defer cli.Cleanup()
	cli.Execute()
}
