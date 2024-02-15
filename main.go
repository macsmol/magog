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
	fenStr := "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 0"
	fmt.Println("----------- " + fenStr)
	// bug and drilldown
	gener, err := engine.NewGeneratorFromFen(fenStr) 
	if err != nil {
		panic(fmt.Sprintf("Cannot parse FEN: %v", err))
	}

	fmt.Println("perft1", gener.Perft(1))
	fmt.Println("perft2", gener.Perft(2))
	fmt.Println("perft3",gener.Perft(3))
	fmt.Println("perft4",gener.Perft(4))
	// fmt.Println("perft5",gener.Perft(5))
	// fmt.Println("perft6",gener.Perft(6))
	
	fenStr = "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 0"
	fmt.Println("----------- " + fenStr)
	gener, err = engine.NewGeneratorFromFen(fenStr) 
	if err != nil {
		panic(fmt.Sprintf("Cannot parse FEN: %v", err))
	}
	fmt.Println("perft1", gener.Perft(1))
	fmt.Println("perft2", gener.Perft(2))
	fmt.Println("perft3",gener.Perft(3))
	fmt.Println("perft4",gener.Perft(4))
	fmt.Println("perft5",gener.Perft(5))
	fmt.Println("perft6",gener.Perft(6))
	
	fenStr = "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1"
	fmt.Println("----------- " + fenStr)
	gener, err = engine.NewGeneratorFromFen(fenStr) 
	if err != nil {
		panic(fmt.Sprintf("Cannot parse FEN: %v", err))
	}
	fmt.Println("perft1", gener.Perft(1))
	fmt.Println("perft2", gener.Perft(2))
	fmt.Println("perft3",gener.Perft(3))
	fmt.Println("perft4",gener.Perft(4))
	fmt.Println("perft5",gener.Perft(5))
	fmt.Println("perft6",gener.Perft(6))

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
