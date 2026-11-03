package main

import (
	"os"

	"github.com/Onurlulardan/harpy/core/internal/composition/recovery"
)

func main() {
	os.Exit(recovery.Run(os.Args[1:]))
}
