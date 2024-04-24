package engine

import (
	"fmt"
	"os"
	"strings"
)

const (
	UCI_POSITION string = "position"
	UCI_MOVES    string = "moves"
	UCI_STARTPOS string = "startpos"
)

var posGen *Generator

func ParseInputLine(inputLine string) {
	if inputLine == "isready" {
		fmt.Println("readyok")
	} else if inputLine == "eval" {
		fmt.Println(Evaluate(NewPosition()))
	} else if inputLine == "tostr" {
		fmt.Printf("%v\n", posGen)
	} else if inputLine == "quit" {
		os.Exit(0)
	} else if strings.HasPrefix(inputLine, UCI_POSITION) {
		setPosition(strings.TrimSpace(strings.TrimPrefix(inputLine, UCI_POSITION)))
	} else if inputLine == "go" {
		fmt.Println("we go?")
		if posGen != nil {
			targetDepth := 1
			fmt.Println(AlphaBeta(posGen, targetDepth, 0, MinusInfinityScore, InfinityScore))
		}
	}
}

func setPosition(positionCommand string) {
	movesIdx := strings.Index(positionCommand, UCI_MOVES)
	if movesIdx == -1 {
		parsePosition(positionCommand)
	} else {
		parsePosition(strings.TrimSpace(positionCommand[:movesIdx]))

		movesString := strings.TrimSpace(positionCommand[movesIdx+len(UCI_MOVES):])
		moveStrings := strings.Split(movesString, " ")
		for _, moveStr := range moveStrings {
			posGen.PushMove(parseMoveString(moveStr))
		}
	}
}

func parseMoveString(moveStr string) Move {
	moveStr = strings.ToLower(moveStr)

	var from square = square((moveStr[0] - 'a') + (moveStr[1]-'1')<<4)
	var to square = square((moveStr[2] - 'a') + (moveStr[3]-'1')<<4)
	if len(moveStr) == 5 {
		var promoteTo piece
		switch moveStr[4] {
		case 'n':
			promoteTo = Knight
		case 'b':
			promoteTo = Bishop
		case 'r':
			promoteTo = Rook
		case 'q':
			promoteTo = Queen
		}
		return NewPromotionMove(from, to, promoteTo)
	}
	return NewMove(from, to)
}

func parsePosition(positionWithoutMoves string) {
	if strings.HasPrefix(positionWithoutMoves, UCI_STARTPOS) {
		posGen = NewGenerator()
	} else {
		pos, err := NewGeneratorFromFen(positionWithoutMoves)
		if err != nil {
			fmt.Println("errorr", err)
		} else {
			posGen = pos
		}
	}
}
