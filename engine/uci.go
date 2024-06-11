package engine

import (
	"fmt"
	"strconv"
	"strings"
	"time"
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
	uMoveTime  string = "movetime"
)

var posGen *Generator
var search *Search
var Quit bool

func ParseInputLine(inputLine string) {
	if inputLine == uIsReady {
		search = NewSearch()
		fmt.Println("readyok")
	} else if inputLine == "eval" {
		fmt.Println(Evaluate(posGen.getTopPos(), 0, true))
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
			//TODO this uses up ply buffer for long move list. Compressing that would be good.
			posGen.ApplyUciMove(parseMoveString(moveStr))
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
	// if specified - search exactly this numer of millis
	moveTimeMillis := -1
	blackMillisLeft := 100_000_000_000
	whiteMillisLeft := 100_000_000_000
	blackMillisIncrement := 0
	whiteMillisIncrement := 0
	fullMovesToGo := ExpectedFullMovesToBePlayed
	targetDepth := MaxSearchDepth

	var err error

out:
	for i, token := range tokens {
		switch token {
		case uMoveTime:
			moveTimeMillis, err = strconv.Atoi(tokens[i+1])
			if err != nil {
				return
			}
			//ignore rest of params
			break out
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
	var endtime time.Time
	if moveTimeMillis != -1 {
		endtime = time.Now().Add(time.Duration(moveTimeMillis * int(time.Millisecond)))
	} else {
		endtime = calcEndtime(blackMillisLeft, blackMillisIncrement, whiteMillisLeft, whiteMillisIncrement,
			fullMovesToGo)
	}
	go search.StartIterativeDeepening(endtime, targetDepth)
}

func calcEndtime(blackMillisLeft, blackMillisIncrement, whiteMillisLeft, whiteMillisIncrement,
	givenMovesToGo int) time.Time {
	isBlackTurn := posGen.getTopPos().flags&FlagWhiteTurn == 0
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

func maybePrintNewPvInfo(score, depth int, bestLine []Move, timeElapsed time.Duration, debugSuffix string) {
	if timeElapsed < time.Duration(200*time.Millisecond) {
		return
	}
	printInfo(score, depth, bestLine, timeElapsed, debugSuffix)
}

func printInfo(score, depth int, bestLine []Move, timeElapsed time.Duration, debugSuffix string) {
	line := Line{moves: bestLine}
	fmt.Println("info pv", line.String(),
		"score", formatScore(score),
		"depth", depth,
		"nodes", evaluatedNodes,
		"time", timeElapsed.Milliseconds(),
		"nps", nps(evaluatedNodes, timeElapsed),
		debugSuffix)
}

func printInfoAfterDepth(score, depth int, bestLine []Move, timeElapsed time.Duration, debugSuffix string) {
	if depth < 2 {
		return
	}
	line := Line{moves: bestLine}
	fmt.Println("info depth", depth,
		"pv", line.String(),
		"score", formatScore(score),
		"nodes", evaluatedNodes,
		"time", timeElapsed.Milliseconds(),
		"nps", nps(evaluatedNodes, timeElapsed),
		debugSuffix)
}

func nps(evaluatedNodes int64, timeElapsed time.Duration) int {
	return int(evaluatedNodes * 1000_000 / int64(timeElapsed.Microseconds()+1))
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

	return sign * (pliesToMate + 1) / 2
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
		newPosGen, err := NewGeneratorFromFen(positionWithoutMoves)
		if err != nil {
			fmt.Println("invalid FEN: ", err)
		} else {
			posGen = newPosGen
		}
	}
}
