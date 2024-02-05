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
	move := moves[13]
	fmt.Println("Making move: ", move)
	pos.MakeMove(move)
	fmt.Println("pos2:", pos)

	// ply 2-----------
	moves = pos.GenerateMoves()
	fmt.Println("moves2:", moves)
	move = moves[13]
	fmt.Println("Making move2: ", move)
	pos.MakeMove(move)
	fmt.Println("pos3:", pos)
	
	pos.UnmakeMove(move)
	fmt.Println("pos3 unmade:", pos)

	// fenPos, err := engine.NewPositionFromFEN("rnbqkb1r/pppp1ppp/8/4P3/2B1n3/8/PPP2PPP/RNBQK1NR b KQkq - 0 4")
	fenPos, err := engine.NewPositionFromFEN("8/2P3k1/1B3p1p/3B2pP/2P3K1/4p3/8/8 w - - 0 47")
	if err != nil {
		fmt.Println("Cannot parse FEN", err)
	} else {
		fmt.Println("from FEN", fenPos)
	}
	moves = fenPos.GenerateMoves()
	fmt.Println("FEN moves:", moves)
	fenPos.MakeMove(moves[0])
	fmt.Println("fenPos+1:", fenPos)
	fenPos.UnmakeMove(moves[0])
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
