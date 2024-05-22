package engine

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	VERSION_STRING string = "0.2"
)

const (
	uUci      string = "uci"
	uIsReady  string = "isready"
	uPosition string = "position"
	uStartpos string = "startpos"
	uMoves    string = "moves"

	uGo        string = "go"
	uWtime     string = "wtime"
	uBtime     string = "btime"
	uWinc      string = "winc"
	uBinc      string = "binc"
	uMovesToGo string = "movestogo"
	uDepth     string = "depth"
	uInfinite  string = "infinite"
)

var posGen *Generator
var search *Search
var Quit bool

func ParseInputLine(inputLine string) {
	if inputLine == uIsReady {
		search = NewSearch()
		fmt.Println("readyok")
	} else if inputLine == "eval" {
		fmt.Println(Evaluate(posGen.pos, 0, true))
	} else if inputLine == "tostr" {
		fmt.Printf("%v\n", posGen)
	} else if inputLine == "quit" {
		Quit = true
	} else if strings.HasPrefix(inputLine, uPosition) {
		doPosition(strings.TrimSpace(strings.TrimPrefix(inputLine, uPosition)))
	} else if inputLine == uUci {
		fmt.Println("id name Magog " + VERSION_STRING)
		fmt.Println("id author Maciej Smolczewski")
		fmt.Println("uciok")
	} else if strings.HasPrefix(inputLine, uGo) {
		doGo(strings.TrimSpace(strings.TrimPrefix(inputLine, uGo)))
	} else if inputLine == "stop" {
		if search != nil {
			search.stop <- true
		}
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
	if search == nil {
		return
	}

	tokens := strings.Split(goCommand, " ")

	blackMillisLeft := 100_000_000_000
	whiteMillisLeft := 100_000_000_000
	blackMillisIncrement := 0
	whiteMillisIncrement := 0
	fullMovesToGo := ExpectedFullMovesToBePlayed
	targetDepth := MaxSearchDepth

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
		case uMovesToGo:
			fullMovesToGo, err = strconv.Atoi(tokens[i+1])
			if err != nil {
				return
			}
		case uDepth:
			targetDepth, err = strconv.Atoi(tokens[i+1])
			if err != nil || targetDepth < 1 {
				return
			}
		}
	}
	endtime := calcEndtime(blackMillisLeft, blackMillisIncrement, whiteMillisLeft, whiteMillisIncrement, 
		fullMovesToGo)
	go search.StartIterativeDeepening(endtime, targetDepth)
}

func calcEndtime(blackMillisLeft, blackMillisIncrement, whiteMillisLeft, whiteMillisIncrement, 
	givenMovesToGo int) time.Time {
	isBlackTurn := posGen.pos.flags&FlagWhiteTurn == 0
	var millisForMove int
	if isBlackTurn {
		millisForMove = blackMillisLeft / givenMovesToGo
	} else {
		millisForMove = whiteMillisLeft / givenMovesToGo
	}
	now := time.Now()
	endtime := now.Add(time.Millisecond * time.Duration(millisForMove))
	return endtime
}

func printInfo(score, depth int, bestLine []Move, timeElapsed time.Duration) {
	if timeElapsed < 300 * time.Millisecond {
		return
	}
	line := Line{moves: bestLine}
	fmt.Println("info pv", line.String(), 
	"score", formatScore(score), 
	"depth", depth,
	"nodes", evaluatedNodes, 
	"time", timeElapsed.Milliseconds(), 
	"nps", nps(evaluatedNodes, timeElapsed))
}

func nps(evaluatedNodes int, timeElapsed time.Duration) int {
	return int(int64(evaluatedNodes) * 1000_000 / int64(timeElapsed.Microseconds()+1))
}

func formatScore(score int) string {
	if closeToMate(score) {
		return fmt.Sprintf("mate %d", fullMovesToMate(score))
	}
	return fmt.Sprintf("cp %d", score)
}

func closeToMate(score int) bool {
	if score < 0 {
		score = -score
	}
	return score > ScoreCloseToMate 
}

func fullMovesToMate(score int) int {
	var sign int
	if score < 0 {
		sign = -1
		score = -score
	} else {
		sign = 1
	}
	pliesToMate := -LostScore - score

	return sign * (pliesToMate + 1) /2
}

func pliesToMate(score int) int {
	if score < 0 {
		score = -score
	}
	pliesToMate := -LostScore - score

	return pliesToMate
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
