package engine

import "fmt"

type Move struct {
	from, to square
	promoteTo piece
	enPassant square
}

func NewMove(from, to square) Move{
	return Move{from, to, NullPiece, InvalidSquare}
}

func NewPromotionMove(from, to square, promoteTo piece) Move{
	return Move{from, to, promoteTo, InvalidSquare}
}

var kingDirections = []Direction{DirN, DirS, DirE, DirW, DirNE, DirNW, DirSW, DirSE}

func (move Move) String() string {
	if move.promoteTo != NullPiece {
		return move.from.String() + move.to.String() + move.promoteTo.String()
	}
	return move.from.String() + move.to.String()
}

func (pos *Position) GenerateMoves() []Move {
	// TODO - move this to Position - and recycle it - benchmark speed difference
	var outputMoves []Move = make([]Move, 0, 60)
	currentPieces, currentKing, pawnAdvanceDirection,
		currColorBit, enemyColorBit, 
		queensideCastlePossible, kingsideCastlePossible,
		pawnStartRank, promotionRank := pos.GetCurrentContext()
	for _, from := range currentPieces {
		piece := pos.board[from]
		//IDEA table of functions indexed by piece? Benchmark it
		//IDEA No piece lists? just iterate over all fields. Perhaps add list once material gone
		switch piece {
		case WPawn, BPawn:
			// queenside take
			to := from + square(pawnAdvanceDirection) - 1
			if pos.board[to]&enemyColorBit != 0 || to==pos.enPassSquare {
				appendPawnMoves(from, to, promotionRank, &outputMoves)
			}
			// kingside take
			to = from + square(pawnAdvanceDirection) + 1
			if pos.board[to]&enemyColorBit != 0 || to==pos.enPassSquare {
				appendPawnMoves(from, to, promotionRank, &outputMoves)
			}
			//pushes
			to = from + square(pawnAdvanceDirection)
			if pos.board[to] == NullPiece {
				appendPawnMoves(from, to, promotionRank, &outputMoves)
				enPassantSquare := to
				to = to + square(pawnAdvanceDirection)
				if from.getRank() == pawnStartRank && pos.board[to] == NullPiece {
					outputMoves = append(outputMoves, Move{from, to, NullPiece, enPassantSquare})
				}
			}
		case WKnight, BKnight:
			dirs := [...]Direction{DirNNE, DirSSW, DirNNW, DirSSE, DirNEE, DirSWW, DirNWW, DirSEE}
			for _, dir := range dirs {
				to := from + square(dir)
				if to&InvalidSquare == 0 && pos.board[to]&currColorBit == 0 {
					outputMoves = append(outputMoves, NewMove(from, to))
				}
			}
		case WBishop, BBishop:
			dirs := []Direction{DirNE, DirSE, DirNW, DirSW}
			pos.appendSlidingPieceMoves(from, currColorBit, enemyColorBit, dirs, &outputMoves)
		case WRook, BRook:
			dirs := []Direction{DirN, DirS, DirE, DirW}
			pos.appendSlidingPieceMoves(from, currColorBit, enemyColorBit, dirs, &outputMoves)
		case WQueen, BQueen:
			pos.appendSlidingPieceMoves(from, currColorBit, enemyColorBit, kingDirections, &outputMoves)
		default:
			panic(fmt.Sprintf("Unexpected piece found: %v at %v", byte(piece), from))
		}
	}
	// king moves
	for _, dir := range kingDirections {
		to := currentKing + square(dir)
		if to&InvalidSquare == 0 && pos.board[to]&currColorBit == 0 {
			outputMoves = append(outputMoves, NewMove(currentKing, to))
		}
	}
	if queensideCastlePossible {
		//so much casting.. could it be modelled better
		var kingAsByte, dirAsByte int8 = int8(currentKing), int8(DirW)
		kingDest := kingAsByte + dirAsByte*2
		if pos.board[kingAsByte+dirAsByte] == NullPiece &&
			pos.board[kingDest] == NullPiece &&
			pos.board[kingAsByte+dirAsByte*3] == NullPiece {
			// TODO Position.MakeMove() should recognize castling, or add a field to Move type?
			outputMoves = append(outputMoves, NewMove(currentKing, square(kingDest)))
		}
	}
	if kingsideCastlePossible {
		var kingAsByte, dirAsByte int8 = int8(currentKing), int8(DirE)
		kingDest := kingAsByte + dirAsByte*2
		if pos.board[kingAsByte+dirAsByte] == NullPiece && pos.board[kingDest] == NullPiece {
			outputMoves = append(outputMoves, NewMove(currentKing, square(kingDest)))
		}
	}

	return outputMoves
}

func appendPawnMoves(from, to square, promotionRank rank, outputMoves *[]Move) {
	if to.getRank() == promotionRank {
		*outputMoves = append(*outputMoves, NewPromotionMove(from, to, Queen))
		*outputMoves = append(*outputMoves, NewPromotionMove(from, to, Rook))
		*outputMoves = append(*outputMoves, NewPromotionMove(from, to, Bishop))
		*outputMoves = append(*outputMoves, NewPromotionMove(from, to, Knight))
	} else {
		*outputMoves = append(*outputMoves, NewMove(from, to))
	}
}

func (pos *Position) appendSlidingPieceMoves(from square, currColorBit, enemyColorBit piece, dirs []Direction, outputMoves *[]Move) {
	for _, dir := range dirs {
		for to := from + square(dir); to&InvalidSquare == 0; to = to + square(dir) {
			toContent := pos.board[to]
			if toContent&currColorBit != 0 {
				break
			}
			*outputMoves = append(*outputMoves, NewMove(from, to))
			if toContent&enemyColorBit != 0 {
				break
			}
		}
	}
}
