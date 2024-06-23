package engine

import (
	"fmt"
	"math"
)

const (
	InfinityScore      = 100_000_000
	MinusInfinityScore = -InfinityScore
	LostScore          = -100_000
	DrawScore          = 0
	ScoreCloseToMate   = 2 * (9*MaterialQueenScore + 2*MaterialRookScore +
		2*MaterialBishopScore + 2*MaterialKnightScore)
	// used to calc interpolation factor between mid/end game king-square tables
	StartingSumOfMaterial = 2 * (MaterialQueenScore + 2*MaterialRookScore +
		2*MaterialBishopScore + 2*MaterialKnightScore)
)

const (
	MaterialPawnScore   = 100
	MaterialKnightScore = 320
	MaterialBishopScore = 330
	MaterialRookScore   = 500
	MaterialQueenScore  = 900
)
const (
	MobilityScoreFactor = 5
)

// controls width of window where full evaluation is done
const fullEvalScoreMargin = MaterialKnightScore

func pieceToScore(p piece) int {
	switch p {
	case Pawn:
		return MaterialPawnScore
	case Knight:
		return MaterialKnightScore
	case Bishop:
		return MaterialBishopScore
	case Rook:
		return MaterialRookScore
	case Queen:
		return MaterialQueenScore
	case King:
		// not sure what should return here. Lower numbers would make captures by king preferrable
		// as king only appears as attacker. TODO Test it
		return 0
	}
	panic(fmt.Sprintf("Should not get score for this piece: %v", p))
}

// Returns fulll static evaluation score for Position pos. It's given relative to the currently playing
// side (negamax score)
func Evaluate(pos *Position, depth int, debug ...bool) int {
	return LazyEvaluate(pos, depth, MinusInfinityScore, InfinityScore, debug...)
}

// Returns static evaluation score for Position pos. It's given relative to the currently playingside (negamax score)
// If the score is outsied <alpha-fullEvalScoreMargin, beta+fullEvalScoreMargin> window it skips costly part of evaluation.
func LazyEvaluate(pos *Position, depth int, alpha, beta int, debug ...bool) int {
	evaluatedNodes++

	if isCheckMate(pos) {
		return LostScore + depth
	}

	gamePhaseFactor := gamePhaseFactor(pos)
	materialSquaresScore := pieceSquareScore(pos, gamePhaseFactor, debug...)

	if materialSquaresScore > beta+fullEvalScoreMargin ||
		materialSquaresScore < alpha-fullEvalScoreMargin {
		return materialSquaresScore
	}

	// mobility
	currentMobilityScore := pos.countMoves() * MobilityScoreFactor
	if currentMobilityScore == 0 {
		return DrawScore
	}
	pos.flags = pos.flags ^ FlagWhiteTurn
	enemyMobilityScore := pos.countMoves() * MobilityScoreFactor
	pos.flags = pos.flags ^ FlagWhiteTurn
	mobilityScore := currentMobilityScore - enemyMobilityScore
	if len(debug) > 0 {
		fmt.Println("gamePhaseFactor:", gamePhaseFactor,
			"materialSquaresScore: ", materialSquaresScore,
			"mobilityScore: ", mobilityScore)
	}
	var score int = materialSquaresScore + mobilityScore
	return score
}

// (endgame) 0.0 <-------------> 1.0 (opening/midgame)
func gamePhaseFactor(pos *Position) float64 {
	whiteMaterial := nonPawnMaterialScore(pos.whitePieces, &pos.board)
	blackMaterial := nonPawnMaterialScore(pos.blackPieces, &pos.board)
	gamePhaseFactor := float64(whiteMaterial+blackMaterial) / StartingSumOfMaterial
	math.Min(gamePhaseFactor, 1.0)
	return gamePhaseFactor
}

// piece-square score from the perspective of side to move
func pieceSquareScore(pos *Position, gamePhaseFactor float64, debug ...bool) int {
	whiteScore := 0
	for i := int8(0); i < pos.whitePieces.size; i++ {
		pieceSquare := pos.whitePieces.squares[i]
		switch pos.board[pieceSquare] & ColorlessPiece {
		case Knight:
			whiteScore += MaterialKnightScore + int(sqTableKnightsWhite[pieceSquare])
		case Bishop:
			whiteScore += MaterialBishopScore + int(sqTableBishopsWhite[pieceSquare])
		case Rook:
			whiteScore += MaterialRookScore + int(sqTableRooksWhite[pieceSquare])
		case Queen:
			whiteScore += MaterialQueenScore + int(sqTableQueensWhite[pieceSquare])
		}
	}
	for i := int8(0); i < pos.whitePawns.size; i++ {
		whiteScore += MaterialPawnScore + int(sqTablePawnsWhite[pos.whitePawns.squares[i]])
	}
	{
		yMid := float64(sqTableKingMidgameWhite[pos.whiteKing])
		yEnd := float64(sqTableKingEndgameWhite[pos.whiteKing])
		kingScore := int(gamePhaseFactor*yMid + (1.0-gamePhaseFactor)*yEnd)
		// if len(debug) > 0 {
		// 	fmt.Println("\twhiteKingScore: ", kingScore)
		// }
		whiteScore += kingScore
	}

	blackScore := 0
	for i := int8(0); i < pos.blackPieces.size; i++ {
		pieceSquare := pos.blackPieces.squares[i]
		switch pos.board[pieceSquare] & ColorlessPiece {
		case Knight:
			blackScore += MaterialKnightScore + int(sqTableKnightsBlack[pieceSquare])
		case Bishop:
			blackScore += MaterialBishopScore + int(sqTableBishopsBlack[pieceSquare])
		case Rook:
			blackScore += MaterialRookScore + int(sqTableRooksBlack[pieceSquare])
		case Queen:
			blackScore += MaterialQueenScore + int(sqTableQueensBlack[pieceSquare])
		}
	}
	for i := int8(0); i < pos.blackPawns.size; i++ {
		blackScore += MaterialPawnScore + int(sqTablePawnsBlack[pos.blackPawns.squares[i]])
	}
	{
		yMid := float64(sqTableKingMidgameBlack[pos.blackKing])
		yEnd := float64(sqTableKingEndgameBlack[pos.blackKing])
		kingScore := int(gamePhaseFactor*yMid + (1.0-gamePhaseFactor)*yEnd)
		// if len(debug) > 0 {
		// 	fmt.Println("\tblackKingScore: ", kingScore)
		// }
		blackScore += kingScore
	}

	score := whiteScore - blackScore
	negamaxFactor := pos.evaluationContext()

	// if len(debug) > 0 {
	// 	fmt.Println("whitePieceSquare: ", whiteScore, "; blackPieceSquare:", blackScore, " negamaxFactor: ", negamaxFactor)
	// }
	return score * negamaxFactor
}

func isCheckMate(position *Position) bool {
	return position.isCurrentKingUnderCheck() && position.countMoves() == 0
}

func terminalNodeScore(position *Position, depth int) int {
	evaluatedNodes++
	if position.isCurrentKingUnderCheck() {
		return LostScore + depth
	}
	return DrawScore
}

func (pos *Position) evaluationContext() (
	negamaxFactor int) {
	if pos.flags&FlagWhiteTurn == 0 {
		return -1
	} else {
		return 1
	}
}

// Counts all possible moves from pos position
func (pos *Position) countMoves() int {
	var movesCount int = 0
	currentPieces, enemyPieces,
		currentPawns, enemyPawns,
		currentKing, enemyKing,
		pawnAdvanceDirection,
		currColorBit, enemyColorBit,
		queensideCastlePossible, kingsideCastlePossible,
		pawnStartRank, promotionRank := pos.GetCurrentContext()
	for i := int8(0); i < currentPawns.size; i++ {
		from := currentPawns.squares[i]
		// queenside take
		to := from + square(pawnAdvanceDirection) - 1
		if to&InvalidSquare == 0 && (pos.board[to]&enemyColorBit != 0 ||
			(to == pos.enPassSquare &&
				// fix for bug where friendly ep-square take is possible while calculating mobility
				from.getRank() != pawnStartRank)) {
			movesCount += pos.countPawnMoves(from, to, promotionRank)
		}
		// kingside take
		to = from + square(pawnAdvanceDirection) + 1
		if pos.board[to]&enemyColorBit != 0 || (to == pos.enPassSquare &&
			// fix for bug where friendly ep-square take is possible while calculating mobility
			from.getRank() != pawnStartRank) {
			movesCount += pos.countPawnMoves(from, to, promotionRank)
		}
		//pushes
		to = from + square(pawnAdvanceDirection)
		if pos.board[to] == NullPiece {
			movesCount += pos.countPawnMoves(from, to, promotionRank)
			enPassantSquare := to
			to = to + square(pawnAdvanceDirection)
			if from.getRank() == pawnStartRank && pos.board[to] == NullPiece {
				if isLegal(pos, Move{from, to, NullPiece, enPassantSquare}) {
					movesCount++
				}
			}
		}
	}
	for i := int8(0); i < currentPieces.size; i++ {
		from := currentPieces.squares[i]
		piece := pos.board[from]
		switch piece {
		case WKnight, BKnight:
			dirs := [...]Direction{DirNNE, DirSSW, DirNNW, DirSSE, DirNEE, DirSWW, DirNWW, DirSEE}
			for _, dir := range dirs {
				to := from + square(dir)
				if to&InvalidSquare == 0 && pos.board[to]&currColorBit == 0 {
					if isLegal(pos, NewMove(from, to)) {
						movesCount++
					}
				}
			}
		case WBishop, BBishop:
			dirs := []Direction{DirNE, DirSE, DirNW, DirSW}
			movesCount += pos.countSlidingPieceMoves(from, currColorBit, enemyColorBit, dirs)
		case WRook, BRook:
			dirs := []Direction{DirN, DirS, DirE, DirW}
			movesCount += pos.countSlidingPieceMoves(from, currColorBit, enemyColorBit, dirs)
		case WQueen, BQueen:
			movesCount += pos.countSlidingPieceMoves(from, currColorBit, enemyColorBit,
				kingDirections)
		default:
			panic(fmt.Sprintf("Unexpected piece found: %v at %v pos %v", byte(piece), from, pos))
		}
	}
	// king moves
	for _, dir := range kingDirections {
		to := currentKing + square(dir)
		if to&InvalidSquare == 0 && pos.board[to]&currColorBit == 0 {
			if isLegal(pos, NewMove(currentKing, to)) {
				movesCount++
			}
		}
	}
	if queensideCastlePossible {
		//so much casting.. could it be modelled better?
		var kingAsByte, dirAsByte int8 = int8(currentKing), int8(DirW)
		kingDest := kingAsByte + dirAsByte*2
		if pos.board[kingAsByte+dirAsByte] == NullPiece &&
			pos.board[kingDest] == NullPiece &&
			pos.board[kingAsByte+dirAsByte*3] == NullPiece &&
			!pos.isUnderCheck(enemyPieces, enemyPawns, enemyKing, currentKing) &&
			!pos.isUnderCheck(enemyPieces, enemyPawns, enemyKing, square(kingAsByte+dirAsByte)) &&
			!pos.isUnderCheck(enemyPieces, enemyPawns, enemyKing, square(kingDest)) {
			movesCount++
		}
	}
	if kingsideCastlePossible {
		var kingAsByte, dirAsByte int8 = int8(currentKing), int8(DirE)
		kingDest := kingAsByte + dirAsByte*2
		if pos.board[kingAsByte+dirAsByte] == NullPiece && pos.board[kingDest] == NullPiece &&
			!pos.isUnderCheck(enemyPieces, enemyPawns, enemyKing, currentKing) &&
			!pos.isUnderCheck(enemyPieces, enemyPawns, enemyKing, square(kingAsByte+dirAsByte)) &&
			!pos.isUnderCheck(enemyPieces, enemyPawns, enemyKing, square(kingDest)) {
			movesCount++
		}
	}
	return movesCount
}

func (pos *Position) countPawnMoves(from, to square, promotionRank rank) int {
	if !isLegal(pos, NewMove(from, to)) {
		return 0
	}
	if to.getRank() == promotionRank {
		return 4
	} else {
		return 1
	}
}

func (pos *Position) countSlidingPieceMoves(from square, currColorBit, enemyColorBit piece,
	dirs []Direction) int {
	movesCount := 0
	for _, dir := range dirs {
		for to := from + square(dir); to&InvalidSquare == 0; to = to + square(dir) {
			toContent := pos.board[to]
			if toContent&currColorBit != 0 {
				break
			}
			if isLegal(pos, NewMove(from, to)) {
				movesCount++
			}
			if toContent&enemyColorBit != 0 {
				break
			}
		}
	}
	return movesCount
}

func (pos *Position) countSlidingPieceTacticalMoves(from square, currColorBit, enemyColorBit piece,
	dirs []Direction) int {
	movesCount := 0
	for _, dir := range dirs {
		for to := from + square(dir); to&InvalidSquare == 0; to = to + square(dir) {
			toContent := pos.board[to]
			if toContent&currColorBit != 0 {
				break
			}
			if toContent&enemyColorBit != 0 && isLegal(pos, NewMove(from, to)) {
				movesCount++
				break
			}
		}
	}
	return movesCount
}

func nonPawnMaterialScore(pieces pieceList, board *[128]piece) int {
	score := 0
	for i := int8(0); i < pieces.size; i++ {
		switch board[pieces.squares[i]] & ColorlessPiece {
		case Knight:
			score += MaterialKnightScore
		case Bishop:
			score += MaterialBishopScore
		case Rook:
			score += MaterialRookScore
		case Queen:
			score += MaterialQueenScore
		}
	}
	return score
}