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
	// fenPos, err := engine.NewPositionFromFEN("rnbqkb1r/pppp1ppp/8/4P3/2B1n3/8/PPP2PPP/RNBQK1NR b KQkq - 0 4")
	// fenPos, err := engine.NewPositionFromFEN("r3kbnr/ppp1pppp/2nq4/3p1b2/3P1B2/2NQ4/PPP1PPPP/R3KBNR w KQkq - 6 5")
	fenPos, err := engine.NewPositionFromFEN("rnbqkbnr/ppp1pp1p/6p1/3pP3/8/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 3")
	if err != nil {
		panic(fmt.Sprintf("Cannot parse FEN: %v", err))
	}
	fmt.Println("from FEN", fenPos)

	moves := fenPos.GenerateMoves()
	fmt.Println("FEN moves:", moves)
	daMove := moves[0]
	fmt.Println("Taking EP")
	undo := fenPos.MakeMove(daMove)
	fmt.Println("fen+1 undo info", undo)
	fmt.Println("fenPos+1 after EP", fenPos)

	fenPos.UnmakeMove(daMove, undo)
	fmt.Println("fenPos+1 unmade (before EP):", fenPos)

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
