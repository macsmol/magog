package engine

import (
	"fmt"
	"os"
	"runtime/pprof"
	"slices"
	"strings"
	"time"
)

// structure for keeping best line (Search) found and retrieving it at the end of the search.
// Two dimensional array accumulates shorter lines from greater depth into longer line towards the
// lower depth (towards starting position)
// [depth]  -> line
// [0]     ->  mov01 mov02 mov03 mov04 mov05
// [1]     ->  mov11 mov12 mov13 mov14
// [2]     ->  mov21 mov22 mov23
// [3]     ->  mov31 mov32
// [4]     ->  mov41
// [5]	   -> empty line - call Evaluate()
// The algorithm for collecting PV is based on the one described here:
// https://web.archive.org/web/20070808093935/http://www.brucemo.com/compchess/programming/pv.htm
type Search struct {
	bestLineAtDepth [MaxSearchDepth][]Move
	stop            chan bool
	interrupted     bool
}

var evaluatedNodes int64
var ProfileFile *os.File

func NewSearch() *Search {
	pv := &Search{}
	for i := 0; i < len(pv.bestLineAtDepth); i++ {
		pv.bestLineAtDepth[i] = make([]Move, MaxSearchDepth-i)
	}
	pv.stop = make(chan bool)
	return pv
}

func (search *Search) StartIterativeDeepening(endtime time.Time, maxDepth int) {
	if ProfileFile != nil {
		fmt.Println("starting profiling")

		pprof.StartCPUProfile(ProfileFile)
		defer stopProfiling()
	}
	var bestLine *Line = &Line{}
	search.interrupted = false
	startTime := time.Now()
	evaluatedNodes = 0
	var bestScore int
	var depthCompleted int

	for currDepth := 1; currDepth <= maxDepth; currDepth++ {
		scoreAtDepth, oneLegalMove := search.startAlphaBeta(posGen, currDepth, &search.bestLineAtDepth[0],
			bestLine, startTime, endtime)

		if time.Now().After(endtime) {
			break
		}
		if search.interrupted {
			break
		}

		copyBestLine(bestLine, search.bestLineAtDepth[0])
		printInfo(scoreAtDepth, currDepth, bestLine.moves, time.Since(startTime), "")
		depthCompleted = currDepth
		bestScore = scoreAtDepth

		// shortest mating line found - no need to go deeper
		if pliesToMate(scoreAtDepth) == currDepth {
			break
		}
		// skip deeper searches when only one move is possible
		if oneLegalMove {
			break
		}
	}
	printInfo(bestScore, depthCompleted, bestLine.moves, time.Since(startTime), "")
	fmt.Println("bestmove", bestLine.moves[0])
}

func copyBestLine(bestLine *Line, bestLineAtDepth []Move) {
	bestLine.moves = append(bestLine.moves[:0], bestLineAtDepth...)
	bestLine.sublineLengthMatched = 0
}

func (pv *Search) PVString() string {
	var sb strings.Builder
	bestLine := pv.bestLineAtDepth[0]
	sb.WriteString(bestLine[0].String())
	for i := 1; i < len(bestLine); i++ {
		sb.WriteRune(' ')
		sb.WriteString(bestLine[i].String())
	}
	return sb.String()
}

func (pv *Search) getBestLine() []Move {
	return pv.bestLineAtDepth[0]
}

// Searches for best move at target depth and returns it's score. Best line found by this
// function is stored currBestLine.
// Param candidateLine stores line that should be evaluated first by the search
// (TODO - make candidateLine a list of candidate lines - so far it seems to be duplicating
// work of currBestLine
func (search *Search) alphaBeta(aPosGen *Generator, targetDepth, depth, alpha, beta int,
	currBestLine *[]Move, candidateLine *Line, startTime, endtime time.Time) int {
	bestSubline := search.bestLineAtDepth[depth+1]
	if targetDepth == depth {
		// Evaluate(posGen.pos, depth)
		return search.quiescence(aPosGen, alpha, beta, depth, currBestLine, startTime)
	}

	moves := aPosGen.GenerateMoves()

	if len(moves) == 0 {
		*currBestLine = (*currBestLine)[:0]
		terminalNodeScore(aPosGen.getTopPos(), depth)
	}

	applyPvMoveBonus(moves, candidateLine, depth)
	sortMoves(moves)
	for _, move := range moves {
		if search.interrupted {
			break
		}
		aPosGen.PushMove(move.mov)
		currScore := -search.alphaBeta(aPosGen, targetDepth, depth+1, -beta, -alpha, &bestSubline,
			candidateLine, startTime, endtime)
		aPosGen.PopMove()

		if currScore > beta {
			return beta
		}
		if currScore > alpha {
			updateBestLine(currBestLine, bestSubline, move.mov)
			alpha = currScore
		}
		if time.Now().After(endtime) {
			break
		}
		if search.interrupted {
			break
		}

		select {
		case <-search.stop:
			search.interrupted = true
		default:
		}
	}

	return alpha
}

func (search *Search) startAlphaBeta(aPosGen *Generator, targetDepth int, currBestLine *[]Move,
	pvLine *Line, starttime, endtime time.Time) (score int, oneLegalMove bool) {
	bestSubline := search.bestLineAtDepth[1]
	moves := aPosGen.GenerateMoves()
	alpha := MinusInfinityScore
	beta := InfinityScore

	if len(moves) == 0 {
		*currBestLine = (*currBestLine)[:0]
		terminalNodeScore(aPosGen.getTopPos(), 0)
	}

	applyPvMoveBonus(moves, pvLine, 0)
	sortMoves(moves)
	for _, move := range moves {
		if search.interrupted {
			break
		}
		aPosGen.PushMove(move.mov)
		currScore := -search.alphaBeta(aPosGen, targetDepth, 1, -beta, -alpha, &bestSubline,
			pvLine, starttime, endtime)
		aPosGen.PopMove()

		if currScore > alpha {
			updateBestLine(currBestLine, bestSubline, move.mov)
			alpha = currScore

			printInfo(alpha, targetDepth, search.getBestLine(), time.Duration(time.Since(starttime)), "")
			// printInfo( alpha, targetDepth, search.getBestLine(), time.Duration(time.Since(starttime)), "in startAB:")
		}

		if search.interrupted {
			break
		}
		if nextMoveWins(currScore) {
			break
		}

		select {
		case <-search.stop:
			search.interrupted = true
		default:
		}
	}

	return alpha, len(moves) == 1
}

func nextMoveWins(score int) bool {
	onePly := 1
	return score == -LostScore-onePly
}

func applyPvMoveBonus(moves []rankedMove, candidateLine *Line, depth int) {
	for i, m := range moves {
		if candidateLine.isMoveOnLine(m.mov, depth) {
			moves[i].ranking += rankingBonusPvMove
			break
		}
	}
}
func sortMoves(moves []rankedMove) {
	slices.SortFunc(moves,
		//desc sort by ranking
		func(a, b rankedMove) int {
			return b.ranking - a.ranking
		})
}

func updateBestLine(currBestLine *[]Move, betterSubline []Move, betterMove Move) {
	(*currBestLine) = (*currBestLine)[:len(betterSubline)+1]
	(*currBestLine)[0] = betterMove
	copy((*currBestLine)[1:], betterSubline)
}

func (search *Search) quiescence(aPosGen *Generator, alpha, beta, depth int,
	currBestLine *[]Move, startTime time.Time) int {
	bestSubline := search.bestLineAtDepth[depth+1]
	score := Evaluate(aPosGen.getTopPos(), depth)

	if evaluatedNodes%LogEveryNNodes == 0 {
		timeElapsed := time.Since(startTime)
		fmt.Println("info",
			"nodes", evaluatedNodes,
			"time", timeElapsed.Milliseconds(),
			"nps", nps(evaluatedNodes, timeElapsed))
	}

	if score >= beta {
		return beta
	}
	if score > alpha {
		//TODO should add updateBestLine() here?
		*currBestLine = (*currBestLine)[:0]
		alpha = score
	}
	tacticalMoves := aPosGen.GenerateTacticalMoves()
	sortMoves(tacticalMoves)
	for _, mov := range tacticalMoves {
		aPosGen.PushMove(mov.mov)
		score = -search.quiescence(aPosGen, -beta, -alpha, depth+1, &bestSubline, startTime)
		aPosGen.PopMove()

		if score >= beta {
			return beta
		}
		if score > alpha {
			updateBestLine(currBestLine, bestSubline, mov.mov)
			alpha = score
		}
	}
	return alpha
}

func stopProfiling() {
	fmt.Println("iterative deepening: stopping profiling.....")
	pprof.StopCPUProfile()
}
