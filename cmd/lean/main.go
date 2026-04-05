// lean is a CLI for the lean document rendering engine.
package main

import (
	"fmt"

	"github.com/lean-docs/lean"
)

func main() {
	fmt.Printf("lean %s\n", lean.Version())
}
