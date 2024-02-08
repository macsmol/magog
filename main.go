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
	gener, err := engine.NewGeneratorFromFen("rn1qkbnr/pppb1ppp/8/1B1pp3/P3P3/8/1PPP1PPP/RNBQK1NR b KQkq - 0 4")
	if err != nil {
		panic(fmt.Sprintf("Cannot parse FEN: %v", err))
	}
	fmt.Println("from FEN", gener)

	moves := gener.GenerateMoves()
	fmt.Println("FEN moves:", moves)

	gener.PushMove(engine.NewMove(engine.D7, engine.C6))
	gener.PopMove()
	gener.PushMove(engine.NewMove(engine.D7, engine.C8))

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
