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
	currentPieces, currentPawns,
	enemyPieces, enemyPawns := pos.evaluationContext()

	currMaterial := materialScore(currentPieces, currentPawns, &pos.board)
	enemyMaterial := materialScore(enemyPieces, enemyPawns, &pos.board)
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
	currPieces pieceList, currentPawns pawnList,
	 enemyPieces pieceList, enemyPawns pawnList) {
	if pos.flags&FlagWhiteTurn == 0 {
		return pos.blackPieces, pos.blackPawns, pos.whitePieces, pos.whitePawns
	} else {
		return pos.whitePieces, pos.whitePawns, pos.blackPieces, pos.blackPawns
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
	for i := int8(0) ; i < currentPawns.size; i++ {
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
	for i := int8(0) ; i < currentPieces.size; i++ {
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

// Returns true if pseudolegal is legal in pos. False otherwise. 
func isLegal(pos *Position, pseudolegal Move) bool {
	tested_position := *pos
	return (&tested_position).MakeMove(pseudolegal)
}

func materialScore(pieces pieceList, pawns pawnList, board *[128]piece) int {
	score := 0
	for i := int8(0) ; i < pieces.size; i++ {
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
	score += int(pawns.size) * MaterialPawnScore
	return score
}
