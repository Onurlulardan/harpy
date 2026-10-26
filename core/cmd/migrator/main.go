package main

import (
	"os"

	"github.com/Onurlulardan/harpy/core/internal/composition/migrator"
)

func main() {
	os.Exit(migrator.Run(os.Args[1:]))
}
