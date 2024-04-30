package engine

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

const (
	VERSION_STRING string = "0.1"
)

const (
	uUci      string = "uci"
	uIsReady  string = "isready"
	uPosition string = "position"
	uStartpos string = "startpos"
	uMoves    string = "moves"

	uGo       string = "go"
	uWtime    string = "wtime"
	uBtime    string = "btime"
	uWinc     string = "winc"
	uBinc     string = "binc"
	uDepth    string = "depth"
	uInfinite string = "infinite"
)

var posGen *Generator

func ParseInputLine(inputLine string) {
	if inputLine == uIsReady {
		fmt.Println("readyok")
	} else if inputLine == "eval" {
		fmt.Println(Evaluate(posGen.pos, true))
	} else if inputLine == "tostr" {
		fmt.Printf("%v\n", posGen)
	} else if inputLine == "quit" {
		os.Exit(0)
	} else if strings.HasPrefix(inputLine, uPosition) {
		doPosition(strings.TrimSpace(strings.TrimPrefix(inputLine, uPosition)))
	} else if inputLine == uUci {
		fmt.Println("id name Magog " + VERSION_STRING)
		fmt.Println("id author Maciej Smolczewski")
		fmt.Println("uciok")
	} else if strings.HasPrefix(inputLine, uGo) {
		doGo(strings.TrimSpace(strings.TrimPrefix(inputLine, uGo)))
	}
}

func doPosition(positionCommand string) {
	movesIdx := strings.Index(positionCommand, uMoves)
	if movesIdx == -1 {
		parsePosition(positionCommand)
	} else {
		parsePosition(strings.TrimSpace(positionCommand[:movesIdx]))

		movesString := strings.TrimSpace(positionCommand[movesIdx+len(uMoves):])
		moveStrings := strings.Split(movesString, " ")
		for _, moveStr := range moveStrings {
			posGen.PushMove(parseMoveString(moveStr))
		}
	}
}

func doGo(goCommand string) {
	if posGen == nil {
		return
	}

	tokens := strings.Split(goCommand, " ")

	blackMillisLeft      := math.MaxInt32
	whiteMillisLeft      := math.MaxInt32
	blackMillisIncrement := 0
	whiteMillisIncrement := 0

	var err error

	for i, token := range tokens {
		switch token {
		case uWtime:
			whiteMillisLeft, err = strconv.Atoi(tokens[i+1])
			if err != nil {
				return
			}
		case uBtime:
			blackMillisLeft, err = strconv.Atoi(tokens[i+1])
			if err != nil {
				return
			}
		case uWinc:
			whiteMillisIncrement, err = strconv.Atoi(tokens[i+1])
			if err != nil {
				return
			}
		case uBinc:
			blackMillisIncrement, err = strconv.Atoi(tokens[i+1])
			if err != nil {
				return
			}
		case uDepth:
			targetDepth, err := strconv.Atoi(tokens[i+1])
			if err != nil {
				return
			}
			fmt.Println("target depth", targetDepth)
			// TODO move to goroutine
			fmt.Println("info", AlphaBeta(posGen, targetDepth, 0, MinusInfinityScore, InfinityScore))
			
			fmt.Println("info pv ", posGen.bestLine[:int(posGen.plyIdx)+targetDepth])
		}
	}
	fmt.Printf("btime: %d; binc: %d; wtime: %d; winc: %d\n", blackMillisLeft, blackMillisIncrement, whiteMillisLeft, whiteMillisIncrement)

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
	if strings.HasPrefix(positionWithoutMoves, uStartpos) {
		posGen = NewGenerator()
	} else {
		pos, err := NewGeneratorFromFen(positionWithoutMoves)
		if err != nil {
			fmt.Println("invalid FEN: ", err)
		} else {
			posGen = pos
		}
	}
}
