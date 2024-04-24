package main

import (
	"bufio"
	"macsmol/magog/engine"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		engine.ParseInputLine(scanner.Text())
	}
}

