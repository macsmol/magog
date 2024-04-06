package engine

import (
	"fmt"
)

const (
	MaterialPawn   = 100
	MaterialKnight = 300
	MaterialBishop = 300
	MaterialRook   = 500
	MaterialQueen  = 900
)
const (
	MobilityFactor = 5
)

func AlphaBeta() int {
	fmt.Println("AlfaBeta")
	return 1
}

// Returns static evaluation score for Position pos. It's given relative to the currently playing side (negamax score)
func Evaluate(pos *Position) int {
	currentPieces, enemyPieces := pos.evaluationContext()

	currMaterial := materialScore(currentPieces, &pos.board)
	enemyMaterial := materialScore(enemyPieces, &pos.board)
	fmt.Println("currMaterial: ", currMaterial, " enemyMaterial: ", enemyMaterial)
	materialScore := currMaterial - enemyMaterial
	mobilityScore := pos.mobilityScore()

	var score int = materialScore + mobilityScore
	return score
}

func (pos *Position) mobilityScore() int {
	currentMobilityScore := pos.countMoves()
	pos.flags = pos.flags ^ FlagWhiteTurn
	enemyMobilityScore := pos.countMoves()
	pos.flags = pos.flags ^ FlagWhiteTurn
	fmt.Println("Current mobility: ", currentMobilityScore, " enemy mobility: ", enemyMobilityScore)

	return (currentMobilityScore - enemyMobilityScore) * MobilityFactor
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

	attackersCount, attackerSquare := pos.countPseudolegalChecksOn(enemyPieces, currentKing)
	if attackersCount >= 2 {
		return pos.countNormalKingMoves(currentKing, currColorBit)
	}
	if attackersCount == 1 {
		return pos.countNormalKingMoves(currentKing, currColorBit) + pos.countKingProtectingMoves(currentKing, attackerSquare, currentPieces, enemyColorBit)
	}
	for _, from := range currentPieces {
		piece := pos.board[from]
		//IDEA table of functions indexed by piece? Benchmark it
		//IDEA No piece lists? just iterate over all fields. Perhaps add list once material gone
		switch piece {
		case WPawn, BPawn:
			// queenside take
			to := from + square(pawnAdvanceDirection) - 1
			if to&InvalidSquare == 0 && (pos.board[to]&enemyColorBit != 0 || to == pos.enPassSquare) {
				movesCount += pos.countPawnMoves(from, to, promotionRank)
			}
			// kingside take
			to = from + square(pawnAdvanceDirection) + 1
			if pos.board[to]&enemyColorBit != 0 || to == pos.enPassSquare {
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
	movesCount += pos.countNormalKingMoves(currentKing, currColorBit)

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

func (pos *Position) countNormalKingMoves(currentKing square, currColorBit piece) int {
	// king moves
	movesCount := 0
	for _, dir := range kingDirections {
		to := currentKing + square(dir)
		if to&InvalidSquare == 0 && pos.board[to]&currColorBit == 0 {
			if pos.isLegal(NewMove(currentKing, to)) {
				movesCount++
			}
		}
	}
	return movesCount
}

// Counts moves (captures and interpositions) that can be made to protect king from single attacker. Does not count the moves of king itself
func (pos *Position) countKingProtectingMoves(attackedKing square, attacker square, defendingPieces []square, pinnerColorBit piece) int {
	piece := pos.board[attacker] & ColorlessPiece
	switch piece {
	case Pawn, Knight:
		capturesCount := pos.countLegalChecksOn(defendingPieces, pinnerColorBit, attacker, attackedKing, false)
		return capturesCount
	case Bishop, Rook, Queen:
		capturesCount := pos.countLegalChecksOn(defendingPieces, pinnerColorBit, attacker, attackedKing, false)

		interpositions := 0
		direction := directionTable[moveIndex(attacker, attackedKing)]
		for interposingSquare := attacker + square(direction); interposingSquare != attackedKing; interposingSquare += square(direction) {
			interpositions += pos.countLegalChecksOn(defendingPieces, pinnerColorBit, interposingSquare, attackedKing, true)
		}
		return capturesCount + interpositions
	}
	panic(fmt.Sprintf("No attacker found on %v; piece: piece '%v'; pos: %v", attacker, piece, pos))
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
			score += MaterialPawn
		case Knight:
			score += MaterialKnight
		case Bishop:
			score += MaterialBishop
		case Rook:
			score += MaterialRook
		case Queen:
			score += MaterialQueen
		}
	}
	return score
}
