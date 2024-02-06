package main

import (
	"bufio"
	"fmt"
	"macsmol/magog/engine"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println(engine.A1, engine.A2, engine.B2, engine.InvalidSquare, engine.H8)
	fmt.Println(engine.BBishop, engine.WBishop, engine.NullPiece)

	var pos *engine.Position = engine.NewPosition()
	fmt.Println("pos1:", pos)
	// ply 1-----------
	moves := pos.GenerateMoves()
	fmt.Println("moves1:", moves)

	// fenPos, err := engine.NewPositionFromFEN("rnbqkb1r/pppp1ppp/8/4P3/2B1n3/8/PPP2PPP/RNBQK1NR b KQkq - 0 4")
	fenPos, err := engine.NewPositionFromFEN("r3kbnr/ppp1pppp/2nq4/3p1b2/3P1B2/2NQ4/PPP1PPPP/R3KBNR w KQkq - 6 5")
	if err != nil {
		panic(fmt.Sprintf("Cannot parse FEN: %v", err))
	}
	fmt.Println("from FEN", fenPos)

	moves = fenPos.GenerateMoves()
	fmt.Println("FEN moves:", moves)
	daMove := moves[len(moves)-1]
	fenPos.MakeMove(daMove)
	fmt.Println("fenPos+1:", fenPos)

	fenPos.UnmakeMove(daMove)
	fmt.Println("fenPos+1 unmade:", fenPos)

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
