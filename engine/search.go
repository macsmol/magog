package engine


func startAlphaBeta(posGen *Generator, depth int, ) int {

	alpha := MinusInfinity
	beta := Infinity
	return AlphaBeta(posGen, depth, alpha, beta)
}

func AlphaBeta(posGen *Generator, targetDepth, depth, alpha, beta int) int {
	if depth == 0 {
		return Evaluate(posGen.pos)
	}

	moves := posGen.GenerateMoves()

 	if len(moves) == 0 {
		terminalNodeScore(posGen.pos)
	}

	for _, move := range moves {
		
		posGen.PushMove(move)
		currScore := -AlphaBeta(posGen, depth-1, -beta, -alpha)
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

// func quiescence(posGen Generator, alpha, beta, )
