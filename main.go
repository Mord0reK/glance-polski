package main

import (
	"os"

	"github.com/glanceapp/glance/internal/glance"
)

var commitSHA = "dev"

func main() {
	os.Exit(glance.Main(commitSHA))
}
