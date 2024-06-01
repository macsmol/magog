package engine

import (
	"fmt"
)

const (
	InfinityScore      = 100_000_000
	MinusInfinityScore = -InfinityScore
	LostScore          = -100_000
	DrawScore          = 0
	ScoreCloseToMate   = 2 * (9*MaterialQueenScore + 2*MaterialRookScore + 2*
		MaterialBishopScore + 2*MaterialKnightScore)
)

const (
	MaterialPawnScore   = 100
	MaterialKnightScore = 300
	MaterialBishopScore = 300
	MaterialRookScore   = 500
	MaterialQueenScore  = 900
)
const (
	MobilityScoreFactor = 5
)

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
	return 0
}

// Returns static evaluation score for Position pos. It's given relative to the currently playing
// side (negamax score)
func Evaluate(pos *Position, depth int, debug ...bool) int {
	currentSliders, currentKnights, currentPawns,
	enemySliders, enemyKnights, enemyPawns := pos.evaluationContext()

	currMaterial := materialScore(currentSliders, currentKnights, currentPawns, &pos.board)
	enemyMaterial := materialScore(enemySliders, enemyKnights, enemyPawns, &pos.board)
	materialScore := currMaterial - enemyMaterial

	// mobility
	currentMobilityScore := pos.countMoves() * MobilityScoreFactor
	if currentMobilityScore == 0 {
		return terminalNodeScore(pos, depth)
	}
	pos.flags = pos.flags ^ FlagWhiteTurn
	enemyMobilityScore := pos.countMoves() * MobilityScoreFactor
	pos.flags = pos.flags ^ FlagWhiteTurn
	mobilityScore := currentMobilityScore - enemyMobilityScore

	if len(debug) > 0 {
		fmt.Println("currMaterial: ", currMaterial, "; enemyMaterial:", enemyMaterial)
		fmt.Println("currMobility: ", currentMobilityScore, "; enemyMobility: ", enemyMobilityScore)
	}

	var score int = materialScore + mobilityScore
	evaluatedNodes++
	return score
}

func terminalNodeScore(position *Position, depth int) int {
	evaluatedNodes++
	if position.isCurrentKingUnderCheck() {
		return LostScore + depth
	}
	return DrawScore
}

func (pos *Position) evaluationContext() (
	currPieces, currentKnights, currentPawns,
	 enemyPieces, enemyKnights, enemyPawns []square) {
	if pos.flags&FlagWhiteTurn == 0 {
		return pos.blackSliders,pos.blackKnights, pos.blackPawns,
		 pos.whiteSliders, pos.whiteKnights, pos.whitePawns
	} else {
		return pos.whiteSliders, pos.whiteKnights, pos.whitePawns,
		 pos.blackSliders, pos.blackKnights, pos.blackPawns
	}
}

// Counts all possible moves from pos position
func (pos *Position) countMoves() int {
	var movesCount int = 0
	currentSliders, enemySliders,
		currentKnights, enemyKnights,
		currentPawns, enemyPawns,
		currentKing, enemyKing,
		pawnAdvanceDirection,
		currColorBit, enemyColorBit,
		queensideCastlePossible, kingsideCastlePossible,
		pawnStartRank, promotionRank := pos.GetCurrentContext()
	for _, from := range currentPawns {
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
				if pos.isLegal(Move{from, to, NullPiece, enPassantSquare}) {
					movesCount++
				}
			}
		}
	}
	for _, from := range currentKnights {
		for _, dir := range knightDirections {
			to := from + square(dir)
			if to&InvalidSquare == 0 && pos.board[to]&currColorBit == 0 {
				if pos.isLegal(NewMove(from, to)) {
					movesCount++
				}
			}
		}
	}
	for _, from := range currentSliders {
		piece := pos.board[from]
		var dirs []Direction
		switch piece {
		case WBishop, BBishop:
			dirs = bishopDirections
		case WRook, BRook:
			dirs = rookDirections
		case WQueen, BQueen:
			dirs = kingDirections
		default:
			panic(fmt.Sprintf("Unexpected piece found: %v at %v pos %v", byte(piece), from, pos))
		}
		movesCount += pos.countSlidingPieceMoves(from, currColorBit, enemyColorBit, dirs)
	}
	// king moves
	for _, dir := range kingDirections {
		to := currentKing + square(dir)
		if to&InvalidSquare == 0 && pos.board[to]&currColorBit == 0 {
			if pos.isLegal(NewMove(currentKing, to)) {
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
			!pos.isUnderCheck(enemySliders, enemyKnights, enemyPawns, enemyKing, currentKing) &&
			!pos.isUnderCheck(enemySliders, enemyKnights, enemyPawns, enemyKing, square(kingAsByte+dirAsByte)) &&
			!pos.isUnderCheck(enemySliders, enemyKnights, enemyPawns, enemyKing, square(kingDest)) {
			movesCount++
		}
	}
	if kingsideCastlePossible {
		var kingAsByte, dirAsByte int8 = int8(currentKing), int8(DirE)
		kingDest := kingAsByte + dirAsByte*2
		if pos.board[kingAsByte+dirAsByte] == NullPiece && pos.board[kingDest] == NullPiece &&
			!pos.isUnderCheck(enemySliders, enemyKnights, enemyPawns, enemyKing, currentKing) &&
			!pos.isUnderCheck(enemySliders, enemyKnights, enemyPawns, enemyKing, square(kingAsByte+dirAsByte)) &&
			!pos.isUnderCheck(enemySliders, enemyKnights, enemyPawns, enemyKing, square(kingDest)) {
			movesCount++
		}
	}
	return movesCount
}

func (pos *Position) countPawnMoves(from, to square, promotionRank rank) int {
	if !pos.isLegal(NewMove(from, to)) {
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
			if pos.isLegal(NewMove(from, to)) {
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
			if toContent&enemyColorBit != 0 && pos.isLegal(NewMove(from, to)) {
				movesCount++
				break
			}
		}
	}
	return movesCount
}

func (pos *Position) isLegal(pseudolegal Move) bool {
	undo := pos.MakeMove(pseudolegal)
	toReturn := undo.move != Move{}
	if toReturn {
		pos.UnmakeMove(undo)
	}
	return toReturn
}

func materialScore(sliders, knights, pawns []square, board *[128]piece) int {
	score := 0
	for _, square := range sliders {
		switch board[square] & ColorlessPiece {
		case Bishop:
			score += MaterialBishopScore
		case Rook:
			score += MaterialRookScore
		case Queen:
			score += MaterialQueenScore
		}
	}
	score += len(knights) * MaterialKnightScore
	score += len(pawns) * MaterialPawnScore
	return score
}
