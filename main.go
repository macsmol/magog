package main

import (
	"bufio"
	"fmt"
	"macsmol/magog/engine"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("%T; %T; \n%T; %T; \n%T\n", engine.A1, engine.A2, engine.B1, engine.B3, engine.C3)
	fmt.Println("--------------")
	fmt.Printf("%v: %v; \n%X; %X; \n%X\n", engine.A1, engine.A2, engine.B1, engine.B3, engine.C3)
	fmt.Println("-------")
	fmt.Println(engine.A1, engine.A2, engine.B2, engine.InvalidSquare,
		 engine.H8, engine.Square(0x81), engine.Square(0x18))
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
