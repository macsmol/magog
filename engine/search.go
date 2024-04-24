package engine

func AlphaBeta(posGen *Generator, targetDepth, depth, alpha, beta int) int {
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
