package main

import (
	"bufio"
	"fmt"
	"macsmol/magog/engine"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("%v: %v;\n%v; %v;\n%v\n", engine.A1, engine.A2, engine.B1, engine.B3, engine.C3)
	fmt.Println("-------")
	fmt.Println(engine.A1, engine.A2, engine.B2, engine.InvalidSquare, engine.H8)
	fmt.Println(engine.BlackBishop, engine.NullPiece)
	for {

		scanner.Scan()
		line := scanner.Text()
		switch line {
		case "isready":
			isReady()
		case "quit":
			os.Exit(0)
		}
	}
}

func isReady() {
	fmt.Println("readyok")
}
