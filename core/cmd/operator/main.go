package main

import (
	"os"

	"github.com/Onurlulardan/harpy/core/internal/composition/operator"
)

func main() {
	os.Exit(operator.Run(os.Args[1:]))
}
