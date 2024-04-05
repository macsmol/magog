package main

import (
	"bufio"
	"fmt"
	"macsmol/magog/engine"
	"os"
	"time"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("-----------------")
	// r3k2r/8/8/8/8/8/8/R3K1R1_w_Qkq_-_0_1
	fenStr := "8/8/8/2k5/8/8/6Kp/Q7 w - - 3 3"
	
	
	
	generator,err := engine.NewGeneratorFromFen(fenStr)
	if err != nil {
		panic(fmt.Sprintf("Cannot parse FEN: %v", err))
	}

	fmt.Println("position is", generator)
	const depth = 2
	fmt.Printf("perftd(%d)--------------------\n", depth)
	start := time.Now()
	generator.Perftd(depth)
	fmt.Printf("took %v millis\n", time.Since(start).Milliseconds())
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
