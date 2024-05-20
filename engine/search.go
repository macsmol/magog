package engine

import (
	"fmt"
	"strings"
	"time"
)

// func doIterativeDeepening(posGen *Generator, finishTime) {

// }

// structure for keeping best line (Search) found and retrieving it at the end of the search.
// Two dimensional array accumulates shorter lines from greater depth into longer line towards the lower depth (towards starting position)
// [depth]  -> line
// [0]     ->  mov01 mov02 mov03 mov04 mov05
// [1]     ->  mov11 mov12 mov13 mov14
// [2]     ->  mov21 mov22 mov23
// [3]     ->  mov31 mov32
// [4]     ->  mov41
// [5]	   -> empty line - call Evaluate()
// The algorithm for collecting PV is based on the one described here: https://web.archive.org/web/20070808093935/http://www.brucemo.com/compchess/programming/pv.htm
type Search struct {
	bestLineAtDepth [MaxSearchDepth][]Move
	stop            chan bool
	interrupted     bool
}

func NewSearch() *Search {
	pv := &Search{}
	for i := 0; i < len(pv.bestLineAtDepth); i++ {
		pv.bestLineAtDepth[i] = make([]Move, MaxSearchDepth-i)
	}
	pv.stop = make(chan bool)
	return pv
}

func (search *Search) StartIterativeDeepening(endtime time.Time, maxDepth int) {
	var score, currDepth int
	var oneLegalMove bool
	var bestLine *Line = &Line{}
	search.interrupted = false
	starttime := time.Now()
	for currDepth = 1; currDepth <= maxDepth; currDepth++ {
		score, oneLegalMove = search.startAlphaBeta(posGen, currDepth, &search.bestLineAtDepth[0], bestLine, starttime, endtime)
		search.updateBestLine(bestLine)

		if time.Now().After(endtime) {
			break
		}
		if search.interrupted {
			break
		}
		// shortest mating line found - no need to go deeper
		if pliesToMate(score) == currDepth {
			break
		}
		// skip deeper searches when only one move is possible
		if oneLegalMove {
			break
		}

	}
	fmt.Println("bestmove", search.getBestMove())
}

func (pv *Search) updateBestLine(bestLine *Line) {
	bestLine.moves = pv.bestLineAtDepth[0]
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

func (pv *Search) getBestMove() Move {
	return pv.bestLineAtDepth[0][0]
}

func (pv *Search) getBestLine() []Move {
	return pv.bestLineAtDepth[0]
}

// Searches for best move at target depth and returns it's score. Best line found by this function is stored currBestLine.
// Param candidateLine stores line that should be evaluated first by the search
// (TODO - make candidateLine a list of candidate lines - so far it seems to be duplicating work of currBestLine
func (search *Search) alphaBeta(posGen *Generator, targetDepth, depth, alpha, beta int, currBestLine *[]Move, candidateLine *Line, starttime, endtime time.Time) int {
	bestSubline := search.bestLineAtDepth[depth+1]
	if targetDepth == depth {
		return search.quiescence(posGen, alpha, beta, depth, currBestLine)
	}

	moves := posGen.GenerateMoves()

	if len(moves) == 0 {
		*currBestLine = (*currBestLine)[:0]
		terminalNodeScore(posGen.pos, depth)
	}

	reorderMoves(moves, candidateLine, depth)
	for _, move := range moves {
		if search.interrupted {
			break
		}
		posGen.PushMove(move)
		currScore := -search.alphaBeta(posGen, targetDepth, depth+1, -beta, -alpha, &bestSubline, candidateLine, starttime, endtime)
		posGen.PopMove()

		if currScore > beta {
			return beta
		}
		if currScore > alpha {
			updateBestLine(currBestLine, bestSubline, move)
			alpha = currScore
		}
		if time.Now().After(endtime) && searchedEnoughAtThisDepth() {
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

func (search *Search) startAlphaBeta(posGen *Generator, targetDepth int, currBestLine *[]Move, candidateLine *Line, starttime, endtime time.Time) (score int, oneLegalMove bool) {
	bestSubline := search.bestLineAtDepth[1]

	moves := posGen.GenerateMoves()
	alpha := MinusInfinityScore
	beta := InfinityScore

	if len(moves) == 0 {
		*currBestLine = (*currBestLine)[:0]
		terminalNodeScore(posGen.pos, 0)
	}

	reorderMoves(moves, candidateLine, 0)
	for _, move := range moves {
		if search.interrupted {
			break
		}
		posGen.PushMove(move)
		currScore := -search.alphaBeta(posGen, targetDepth, 1, -beta, -alpha, &bestSubline, candidateLine, starttime, endtime)
		posGen.PopMove()

		if currScore > alpha {
			updateBestLine(currBestLine, bestSubline, move)
			alpha = currScore

			//less noisy alternative
			//if depth == 0 && time.Now().After(starttime.Add(time.Millisecond*time.Duration(200))) {
			printInfo(alpha, targetDepth, search.getBestLine())
			//}
		}
		if time.Now().After(endtime) && searchedEnoughAtThisDepth() {
			break
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

func reorderMoves(moves []Move, candidateLine *Line, depth int) {
	for i, m := range moves {
		if candidateLine.isMoveOnLine(m, depth) {
			tmpMove := moves[0]
			moves[0] = moves[i]
			moves[i] = tmpMove
			break
		}
	}
}

// When doing iterative deepening we don't want to allow engine to stop thinking right after descending one depth deeper because these results will be bad (uninitialized)
// We should only stop searching after at least some variations have been evaluated. For example these that seemed best by search at previous depth.
// See notest.txt entry from 09.05.2024 for an example.
func searchedEnoughAtThisDepth() bool {
	// TODO make multiline - return true after searching through several top lines
	return true
}

func terminalNodeScore(position *Position, depth int) int {
	if position.isCurrentKingUnderCheck() {
		return LostScore + depth
	}
	return DrawScore
}

func updateBestLine(currBestLine *[]Move, betterSubline []Move, betterMove Move) {
	(*currBestLine) = (*currBestLine)[:len(betterSubline)+1]
	(*currBestLine)[0] = betterMove
	copy((*currBestLine)[1:], betterSubline)
}

func (search *Search) quiescence(posGen *Generator, alpha, beta, depth int, currBestLine *[]Move) int {
	bestSubline := search.bestLineAtDepth[depth+1]
	score := Evaluate(posGen.pos, depth)

	if score >= beta {
		return beta
	}
	if score > alpha {
		*currBestLine = (*currBestLine)[:0]
		alpha = score
	}
	tacticalMoves := posGen.GenerateTacticalMoves()
	for _, mov := range tacticalMoves {
		posGen.PushMove(mov)
		score = -search.quiescence(posGen, -beta, -alpha, depth+1, &bestSubline)
		posGen.PopMove()

		if score >= beta {
			return beta
		}
		if score > alpha {
			updateBestLine(currBestLine, bestSubline, mov)
			alpha = score
		}
	}
	return alpha
}
