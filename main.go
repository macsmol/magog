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
	
	// ply 3-----------
	moves = pos.GenerateMoves()
	fmt.Println("moves3:", moves)
	move = moves[8]
	fmt.Println("Making move3: ", move)
	pos.MakeMove(move)
	fmt.Println("pos4:", pos)
	// ply 4-----------
	moves = pos.GenerateMoves()
	fmt.Println("moves4:", moves)
	move = moves[8]
	fmt.Println("Making move4: ", move)
	pos.MakeMove(move)
	fmt.Println("pos5:", pos)

	fenPos, err := engine.NewPositionFromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		fmt.Println("Cannot parse FEN", err)
	} else {
		fmt.Println("from FEN", fenPos)
	}

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
