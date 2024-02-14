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
	// gener.Pos, err := engine.NewPositionFromFEN("r3kbnr/ppp1pppp/2nq4/3p1b2/3P1B2/2NQ4/PPP1PPPP/R3KBNR w KQkq - 6 5")
	// gener, err := engine.NewGeneratorFromFen("rn1qkbnr/pppb1ppp/8/1B1pp3/P3P3/8/1PPP1PPP/RNBQK1NR b KQkq - 0 4")
	gener, err := engine.NewGeneratorFromFen("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 0")
	if err != nil {
		panic(fmt.Sprintf("Cannot parse FEN: %v", err))
	}

	// gener := engine.NewGenerator()
	// fmt.Println("perft1",gener.Perft(1))
	// fmt.Println("gener after perft(1)",gener)

	//bugs to fix
	fmt.Println("perft1",gener.Perft(1))
	fmt.Println("perft2",gener.Perft(2))
	fmt.Println("perft3",gener.Perft(3))
	fmt.Println("perftd4")
	gener.Perftd(4)

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
