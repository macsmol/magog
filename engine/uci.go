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

	uOptionSet   string = "setoption"
	uOptionName  string = "name"
	uOptionValue string = "value"
)

// anti 'loose on time' duration in case of delays (printing on console, GC kicking in, system clock granularity)
const antiflagMillis int = 50

var posGen *Generator
var search *Search
var Quit bool

func ParseInputLine(inputLine string) {
	if inputLine == uIsReady {
		search = NewSearch()
		fmt.Println("readyok")
	} else if inputLine == "eval" {
		fmt.Println(Evaluate(posGen.getTopPos(), 0, true))
	} else if inputLine == "quit" {
		Quit = true
	} else if strings.HasPrefix(inputLine, uPosition) {
		doPosition(strings.TrimSpace(strings.TrimPrefix(inputLine, uPosition)))
	} else if inputLine == uUci {
		doUci()
	} else if strings.HasPrefix(inputLine, uGo) {
		doGo(strings.TrimSpace(strings.TrimPrefix(inputLine, uGo)))
	} else if inputLine == "stop" {
		if search != nil && !search.interrupted {
			search.stop <- true
		}
	} else if strings.HasPrefix(inputLine, uOptionSet) {
		setOption(strings.TrimSpace(strings.TrimPrefix(inputLine, uOptionSet)))
		// non-uci commands
	} else if inputLine == "tostr" {
		fmt.Printf("%v\n", posGen)
	} else if strings.HasPrefix(inputLine, "perft") {
		doPerftDivide(strings.TrimSpace(strings.TrimPrefix(inputLine, "perft")))
	} else if strings.HasPrefix(inputLine, "tperft") {
		doTacticalPerftDivide(strings.TrimSpace(strings.TrimPrefix(inputLine, "tperft")))
	} else if inputLine == "help" {
		printHelp()
	}
}

func printHelp() {
	fmt.Println(`Available UCI commands:
 * uci - print engine info and options
 * isready - print 'readyok' when the engine is ready
 * setoption name <name> value <value> - set an UCI option
 * position [startpos | fen <fenstring> [moves <move1> ... <movei>]] - set position
 * go [depth <depth> | movetime <time> | wtime <time> | btime <time> | winc <time> | binc <time> | movestogo <moves> | infinite] - start search
 * stop - stop searching
 * quit - quit the engine
Other available commands:
 * perft <depth> - count number of moves possible from current position
 * tperft <depth> - same as perft but at <depth> count only captures and promotions. Useful for testing movegen in quiescence search.
 * tostr - print current position
 * eval - evaluate current position`)
}

func doPerftDivide(perftArg string) {
	depth, err := strconv.Atoi(perftArg)
	if err != nil || depth <= 0 {
		fmt.Println("Invalid depth: ", perftArg)
		return
	}
	if posGen == nil {
		fmt.Println("No position set to count perft from")
		return
	}
	posGen.Perftd(depth)
}

func doTacticalPerftDivide(tperftArg string) {
	depth, err := strconv.Atoi(tperftArg)
	if err != nil || depth <= 0 {
		fmt.Println("Invalid depth: ", tperftArg)
		return
	}
	if posGen == nil {
		fmt.Println("No position set to count perft from")
		return
	}
	posGen.PerftDivTactical(depth)
}

func setOption(setOptionCommand string) {
	// So far this func can only set one option. Will upgrade it if needed.

	tokens := strings.Split(setOptionCommand, " ")
	if len(tokens) != 4 {
		return
	}

	if tokens[0] != uOptionName || tokens[2] != uOptionValue {
		return
	}
	if tokens[1] == currmoveLogIntervalKey {
		val, err := strconv.Atoi(tokens[3])
		if err == nil {
			currmoveLogInterval = val
		}
	}
}

func doUci() {
	fmt.Println("id name Magog " + VERSION_STRING)
	fmt.Println("id author Maciej Smolczewski")
	fmt.Println("option",
		uOptionName, currmoveLogIntervalKey,
		"type", "spin",
		"default", currmoveLogIntervalDefault,
		"min", currmoveLogIntervalMin,
		"max", currmoveLogIntervalMax,
	)
	fmt.Println("uciok")
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
			move, err := parseMoveString(moveStr)
			if err != nil {
				fmt.Println("Invalid position command:", err)
				return
			}
			posGen.ApplyUciMove(move)
		}
	}
	//clear killer moves
	for _, killers := range killerMoves {
		killers[0] = Move{}
		killers[1] = Move{}
	}
}

func doGo(goCommand string) {
	startTime := time.Now()
	if posGen == nil {
		fmt.Println("No position set to start search from")
		return
	}
	if search == nil {
		search = NewSearch()
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
		case uInfinite:
			// Default value of ****MillisLeft should make it search  for few years - good enough.
			// Ignore rest of params
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
		moveTimeMillis -= antiflagMillis
		endtime = startTime.Add(time.Duration(moveTimeMillis * int(time.Millisecond)))
	} else {
		endtime = calcEndtime(startTime, blackMillisLeft, blackMillisIncrement, whiteMillisLeft, whiteMillisIncrement,
			fullMovesToGo)
	}
	go search.StartIterativeDeepening(startTime, endtime, targetDepth)
}

func calcEndtime(startTime time.Time, blackMillisLeft, blackMillisInc, whiteMillisLeft, whiteMillisInc int,
	movesToGo int) time.Time {
	isBlackTurn := posGen.getTopPos().flags&FlagWhiteTurn == 0
	var millisForMove int
	if isBlackTurn {
		if blackMillisLeft > blackMillisInc {
			millisForMove = min(blackMillisLeft/movesToGo + blackMillisInc, blackMillisLeft)
		} else {
			millisForMove = blackMillisLeft
		}
	} else {
		if whiteMillisLeft > whiteMillisInc {
			millisForMove = min(whiteMillisLeft/movesToGo + whiteMillisInc, whiteMillisLeft)
		} else {
			millisForMove = whiteMillisLeft
		}
	}
	millisForMove -= antiflagMillis
	millisForMove = max(millisForMove, 1)

	endtime := startTime.Add(time.Millisecond * time.Duration(millisForMove))
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
	fmt.Println("info score", formatScore(score),
		"depth", depth,
		"nps", nps(evaluatedNodes, timeElapsed),
		"time", timeElapsed.Milliseconds(),
		"nodes", evaluatedNodes,
		"pv", line.String(),
		debugSuffix)
}

func printInfoAfterDepth(score, depth int, bestLine []Move, timeElapsed time.Duration, debugSuffix string) {
	line := Line{moves: bestLine}
	fmt.Println("info depth", depth,
		"score", formatScore(score),
		"nps", nps(evaluatedNodes, timeElapsed),
		"time", timeElapsed.Milliseconds(),
		"nodes", evaluatedNodes,
		"pv", line.String(),
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
	return abs(score) > ScoreCloseToMate
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
	return -LostScore - abs(score)
}

func parseMoveString(moveStr string) (Move, error) {
	moveStr = strings.ToLower(moveStr)

	if len(moveStr) < 4 {
		return Move{}, fmt.Errorf("move in not in long algebraic notation: %s", moveStr)
	}
	if moveStr[0] < 'a' || moveStr[0] > 'h' ||
		moveStr[2] < 'a' || moveStr[2] > 'h' {
		return Move{}, fmt.Errorf("move has invalid file: %s", moveStr)
	}
	if moveStr[1] < '1' || moveStr[1] > '8' ||
		moveStr[3] < '1' || moveStr[3] > '8' {
		return Move{}, fmt.Errorf("move has invalid rank: %s", moveStr)
	}

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
		return NewPromotionMove(from, to, promoteTo), nil
	}
	return NewMove(from, to), nil
}

func parsePosition(positionWithoutMoves string) {
	if strings.HasPrefix(positionWithoutMoves, uStartpos) {
		posGen = NewGenerator()
	} else {
		newPosGen, err := NewGeneratorFromFen(positionWithoutMoves)
		if err != nil {
			fmt.Println("invalid FEN:", err)
		} else {
			posGen = newPosGen
		}
	}
}
