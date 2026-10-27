// Command archguard checks import direction for the remaining harpy tree.
// Run from the harpy module: go run ./tools/archguard
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	root := flag.String("root", ".", "repository root")
	flag.Parse()

	violations, err := inspect(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "archguard:", err)
		os.Exit(2)
	}
	if len(violations) == 0 {
		fmt.Println("archguard: OK")
		return
	}
	for _, violation := range violations {
		fmt.Fprintln(os.Stderr, "archguard:", violation)
	}
	os.Exit(1)
}
