package engine

// func doIterativeDeepening(posGen *Generator, finishTime) {

// }

// Searches for best move at target depth and returns it's score. Best line is stored posGen.bestLine and starts at posGen.plyIdx
func AlphaBeta(posGen *Generator, targetDepth, depth,
	alpha, beta int) int {
	if targetDepth == depth {
		return Evaluate(posGen.pos)
	}

	moves := posGen.GenerateMoves()

	if len(moves) == 0 {
		terminalNodeScore(posGen.pos, depth)
	}

	for _, move := range moves {

		posGen.PushMove(move)
		currScore := -AlphaBeta(posGen, targetDepth, depth+1, -beta, -alpha)
		posGen.PopMove()

		if currScore > beta {
			return beta
		}
		if currScore > alpha {
			depthLeft := targetDepth-depth
			posGen.updateSubline(depthLeft)
			//nie znajduje matów! :(
			alpha = currScore
		}
	}

	return alpha
}
// tu jest jakiś babok. 
// z pozycji startowej na koniec go depth 2 twierdzi że pv to e2e3 g7g6 
// (skądinąd wiem ze to e2e3 e7e6 bo machess ma takie samo eval)
func (posGen *Generator) updateSubline(sublineLength int) {
	if len(posGen.bestLine) < int(posGen.plyIdx)+sublineLength {
		posGen.bestLine = append(posGen.bestLine, posGen.plies[posGen.plyIdx].undo.move)
	} else {
		posGen.bestLine[posGen.plyIdx] = posGen.plies[posGen.plyIdx].undo.move
	}
}

func terminalNodeScore(position *Position, depth int) int {
	if position.isCurrentKingUnderCheck() {
		return LostScore + depth
	}
	return DrawScore
}

// func quiescence(posGen Generator, alpha, beta, )
