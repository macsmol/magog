package engine

import (
	"fmt"
)

const (
	InfinityScore      = 100_000_000
	MinusInfinityScore = -InfinityScore
	LostScore = -100_000
	DrawScore = 0
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

// Returns static evaluation score for Position pos. It's given relative to the currently playing side (negamax score)
func Evaluate(pos *Position, debug ...bool) int {
	currentPieces, enemyPieces := pos.evaluationContext()

	currMaterial := materialScore(currentPieces, &pos.board)
	enemyMaterial := materialScore(enemyPieces, &pos.board)
	materialScore := currMaterial - enemyMaterial

	// mobility
	currentMobilityScore := pos.countMoves() * MobilityScoreFactor
	pos.flags = pos.flags ^ FlagWhiteTurn
	enemyMobilityScore := pos.countMoves() * MobilityScoreFactor
	pos.flags = pos.flags ^ FlagWhiteTurn
	mobilityScore := currentMobilityScore - enemyMobilityScore
	
	if len(debug)>0 {
		fmt.Println("currMaterial: ", currMaterial, "; enemyMaterial:", enemyMaterial)
		fmt.Println("currMobility: ", currentMobilityScore, "; enemyMobility: ", enemyMobilityScore)
	}

	var score int = materialScore + mobilityScore
	return score
}

func (pos *Position) evaluationContext() (currPieces, enemyPieces []square) {
	if pos.flags&FlagWhiteTurn != 0 {
		return pos.whitePieces, pos.blackPieces
	} else {
		return pos.blackPieces, pos.whitePieces
	}
}

// GenerateMoves returns legal moves from position that this generator is currently holding
func (pos *Position) countMoves() int {
	var movesCount int = 0
	currentPieces, enemyPieces,
		currentKing, enemyKing,
		pawnAdvanceDirection,
		currColorBit, enemyColorBit,
		queensideCastlePossible, kingsideCastlePossible,
		pawnStartRank, promotionRank := pos.GetCurrentContext()
	for _, from := range currentPieces {
		piece := pos.board[from]
		switch piece {
		case WPawn, BPawn:
			// queenside take
			to := from + square(pawnAdvanceDirection) - 1
			if to&InvalidSquare == 0 && (pos.board[to]&enemyColorBit != 0 || (to == pos.enPassSquare && 
				// fix for bug where friendly ep-square take is possible while calculating mobility
				from.getRank() != pawnStartRank)) {
				movesCount += pos.countPawnMoves(from, to, promotionRank)
			}
			// kingside take
			to = from + square(pawnAdvanceDirection) + 1
			if pos.board[to]&enemyColorBit != 0 || (to == pos.enPassSquare && 
				// fix for bug where friendly ep-square take is possible while calculating mobility
				from.getRank() != pawnStartRank)  {
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
		case WKnight, BKnight:
			dirs := [...]Direction{DirNNE, DirSSW, DirNNW, DirSSE, DirNEE, DirSWW, DirNWW, DirSEE}
			for _, dir := range dirs {
				to := from + square(dir)
				if to&InvalidSquare == 0 && pos.board[to]&currColorBit == 0 {
					if pos.isLegal(NewMove(from, to)) {
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
			movesCount += pos.countSlidingPieceMoves(from, currColorBit, enemyColorBit, kingDirections)
		default:
			panic(fmt.Sprintf("Unexpected piece found: %v at %v pos %v", byte(piece), from, pos))
		}
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
			!pos.isUnderCheck(enemyPieces, enemyKing, currentKing) &&
			!pos.isUnderCheck(enemyPieces, enemyKing, square(kingAsByte+dirAsByte)) &&
			!pos.isUnderCheck(enemyPieces, enemyKing, square(kingDest)) {
			movesCount++
		}
	}
	if kingsideCastlePossible {
		var kingAsByte, dirAsByte int8 = int8(currentKing), int8(DirE)
		kingDest := kingAsByte + dirAsByte*2
		if pos.board[kingAsByte+dirAsByte] == NullPiece && pos.board[kingDest] == NullPiece &&
			!pos.isUnderCheck(enemyPieces, enemyKing, currentKing) &&
			!pos.isUnderCheck(enemyPieces, enemyKing, square(kingAsByte+dirAsByte)) &&
			!pos.isUnderCheck(enemyPieces, enemyKing, square(kingDest)) {
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

func (pos *Position) countSlidingPieceMoves(from square, currColorBit, enemyColorBit piece, dirs []Direction) int {
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

func (pos *Position) isLegal(pseudolegal Move) bool {
	undo := pos.MakeMove(pseudolegal)
	toReturn := undo.move != Move{}
	if toReturn {
		pos.UnmakeMove(undo)
	}
	return toReturn
}

func materialScore(pieces []square, board *[128]piece) int {
	score := 0
	for _, square := range pieces {
		//Is it faster to just switch over pairs of cases?
		switch board[square] & ColorlessPiece {
		case Pawn:
			score += MaterialPawnScore
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
