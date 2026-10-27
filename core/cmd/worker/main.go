package main

import (
	"os"

	"github.com/Onurlulardan/harpy/core/internal/composition/worker"
)

func main() {
	os.Exit(worker.Run(os.Args[1:]))
}
