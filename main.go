package main

import (
	"bufio"
	"fmt"
	"macsmol/magog/engine"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("-----------------")
	// r3k2r/8/8/8/8/8/8/R3K1R1_w_Qkq_-_0_1
	fenStr := "r3k2r/8/8/8/8/2B5/8/R3K1R1 w Qkq - 0 1"
	fmt.Println("----------- " + fenStr)
	
	
	
	pos, err := engine.NewPositionFromFen(fenStr)
	if err != nil {
		panic(fmt.Sprintf("Cannot parse FEN: %v", err))
	}
	fmt.Println("evaluation is: ", engine.Evaluate(pos))
	
	for {
		scanner.Scan()
		line := scanner.Text()
		switch line {
		case "isready":
			isReady()
		case "eval":
			fmt.Println(engine.Evaluate(engine.NewPosition()))
		case "quit":
			os.Exit(0)
		}
	}
}

func isReady() {
	fmt.Println("readyok")
}
