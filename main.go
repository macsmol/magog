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
	var moveOk bool
	// gener.Pos, err := engine.NewPositionFromFEN("r3kbnr/ppp1pppp/2nq4/3p1b2/3P1B2/2NQ4/PPP1PPPP/R3KBNR w KQkq - 6 5")
	// gener, err := engine.NewGeneratorFromFen("rn1qkbnr/pppb1ppp/8/1B1pp3/P3P3/8/1PPP1PPP/RNBQK1NR b KQkq - 0 4")
	gener, err := engine.NewGeneratorFromFen("1nbqkbnr/pppppppp/8/8/4P3/8/PPr2KPP/RNBQ1BNR w k - 0 1")
	if err != nil {
		panic(fmt.Sprintf("Cannot parse FEN: %v", err))
	}
	fmt.Println("from FEN", gener)

	moves := gener.GenPseudoLegalMoves()
	fmt.Println("FEN moves:", moves)

	mov1 := engine.NewMove(engine.F2, engine.E2)
	moveOk = gener.PushMove(mov1)
	if moveOk {
		fmt.Println("possss after ", mov1, gener)

		gener.PopMove()
		fmt.Println("pos after pop", gener)

		mov2 := engine.NewMove(engine.D1, engine.E2)
		gener.PushMove(mov2)
		fmt.Println("pos after", mov2, gener)
	} else {
		fmt.Println("that move was illegal: ", mov1)
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
