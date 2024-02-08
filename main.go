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
	gener, err := engine.NewGeneratorFromFen("rnbqkbnr/ppp1pp1p/6p1/3pP3/8/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 3")
	if err != nil {
		panic(fmt.Sprintf("Cannot parse FEN: %v", err))
	}
	fmt.Println("from FEN", gener)

	moves := gener.GenerateMoves()
	fmt.Println("FEN moves:", moves)

	daMove := moves[0]
	fmt.Println("Taking EP")
	gener.PushMove(daMove)
	fmt.Println("gener.Pos+1 after EP", gener)

	gener.PopMove()	
	fmt.Println("gener.Pos+1 unmade (before EP):", gener)

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
