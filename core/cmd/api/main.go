package main

import (
	"os"

	"github.com/Onurlulardan/harpy/core/internal/composition/api"
)

func main() {
	os.Exit(api.Run(os.Args[1:]))
}
