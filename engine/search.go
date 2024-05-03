package engine

import "strings"

// func doIterativeDeepening(posGen *Generator, finishTime) {

// }

// Maximum number of plies we expect to reach while searching the game tree in any practical scenario.
// This also means max line length
const MaxSearchDepth = 40

// structure for keeping best line (Search) found and retrieving it at the end of the search.
//Two dimensional array accumulates shorter lines from greater depth into longer line towards the lower depth (towards starting position)
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
}

func NewPV() *Search {
	pv := &Search{}
	for i := 0; i < len(pv.bestLineAtDepth); i++ {
		pv.bestLineAtDepth[i] = make([]Move, MaxSearchDepth-i)
	}
	return pv
}

func (pv *Search) StartAlphaBeta(targetDepth int) int {
	return pv.alphaBeta(posGen, targetDepth, 0, MinusInfinityScore, InfinityScore, &pv.bestLineAtDepth[0])
}

func (pv* Search) PVString() string {
	var sb strings.Builder
	bestLine := pv.bestLineAtDepth[0]
	sb.WriteString(bestLine[0].String())
	for i := 1; i < len(bestLine); i++ {
		sb.WriteRune(' ')
		sb.WriteString(bestLine[i].String())
	}
	return sb.String()
}

// Searches for best move at target depth and returns it's score. Best line is stored posGen.bestLine and starts at posGen.plyIdx
func (pv *Search) alphaBeta(posGen *Generator, targetDepth, depth, alpha, beta int, currBestLine *[]Move) int {
	bestSubline := pv.bestLineAtDepth[depth+1]
	if targetDepth == depth {
		*currBestLine = (*currBestLine)[:0]
		return Evaluate(posGen.pos, depth)
	}
	
	moves := posGen.GenerateMoves()
	
	if len(moves) == 0 {
		*currBestLine = (*currBestLine)[:0]
		terminalNodeScore(posGen.pos, depth)
	}

	for _, move := range moves {

		posGen.PushMove(move)
		currScore := -pv.alphaBeta(posGen, targetDepth, depth+1, -beta, -alpha, &bestSubline)
		posGen.PopMove()

		if currScore > beta {
			return beta
		}
		if currScore > alpha {
			(*currBestLine) = (*currBestLine)[:len(bestSubline)+1]
			(*currBestLine)[0] = move
			copy((*currBestLine)[1:], bestSubline)

			alpha = currScore
		}
	}

	return alpha
}

func terminalNodeScore(position *Position, depth int) int {
	if position.isCurrentKingUnderCheck() {
		return LostScore + depth
	}
	return DrawScore
}

// func quiescence(posGen Generator, alpha, beta, )
