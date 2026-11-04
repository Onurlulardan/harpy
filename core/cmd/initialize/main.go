package main

import (
	"os"

	"github.com/Onurlulardan/harpy/core/internal/composition/initialize"
)

func main() {
	os.Exit(initialize.Run(os.Args[1:]))
}
