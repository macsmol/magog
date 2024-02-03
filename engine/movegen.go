package engine

import "fmt"

type Move struct {
	from, to square
}

var kingDirections = []Direction{DirN, DirS, DirE, DirW, DirNE, DirNW, DirSW, DirSE}

func (move Move) String() string {
	return move.from.String() + move.to.String()
}

func (pos *Position) GenerateMoves() []Move {
	// TODO - move this to Position - and recycle it - benchmark speed difference
	var outputMoves []Move = make([]Move, 0, 60)
	currentPieces, currentKing, pawnAdvanceDirection,
		currColorBit, enemyColorBit, pawnStartRank,
		queensideCastlePossible, kingsideCastlePossible := pos.GetCurrentContext()
	// _,_ := pos.GetCurrentContext()
	for _, from := range currentPieces {
		piece := pos.board[from]
		//IDEA table of functions indexed by piece? Benchmark it
		//IDEA No piece lists? just iterate over all fields. Perhaps add list once material gone
		switch piece {
		case WPawn, BPawn:
			// queenside take
			to := from + square(pawnAdvanceDirection) - 1
			if pos.board[to]&enemyColorBit != 0 {
				outputMoves = append(outputMoves, Move{from, to})
			}
			// kingside take
			to = to + 2
			if pos.board[to]&enemyColorBit != 0 {
				outputMoves = append(outputMoves, Move{from, to})
			}
			//pushes
			to = from + square(pawnAdvanceDirection)
			if pos.board[to] == NullPiece {
				outputMoves = append(outputMoves, Move{from, to})
				to = to + square(pawnAdvanceDirection)
				if from.getRank() == pawnStartRank && pos.board[to] == NullPiece {
					outputMoves = append(outputMoves, Move{from, to})
				}
			}
		case WKnight, BKnight:
			dirs := [...]Direction{DirNNE, DirSSW, DirNNW, DirSSE, DirNEE, DirSWW, DirNWW, DirSEE}
			for _, dir := range dirs {
				to := from + square(dir)
				if to&InvalidSquare == 0 && pos.board[to]&currColorBit == 0 {
					outputMoves = append(outputMoves, Move{from, to})
				}
			}
		case WBishop, BBishop:
			dirs := []Direction{DirNE, DirSE, DirNW, DirSW}
			pos.generateSlidingPieceMoves(from, currColorBit, enemyColorBit, dirs, &outputMoves)
		case WRook, BRook:
			dirs := []Direction{DirN, DirS, DirE, DirW}
			pos.generateSlidingPieceMoves(from, currColorBit, enemyColorBit, dirs, &outputMoves)
		case WQueen, BQueen:
			pos.generateSlidingPieceMoves(from, currColorBit, enemyColorBit, kingDirections, &outputMoves)
		default:
			panic(fmt.Sprintf("Unexpected piece found: %v", byte(piece)))
		}
	}
	// king moves
	for _, dir := range kingDirections {
		to := currentKing + square(dir)
		if to&InvalidSquare == 0 && pos.board[to]&currColorBit == 0 {
			outputMoves = append(outputMoves, Move{currentKing, to})
		}
	}
	if queensideCastlePossible {
		kingDest := square(int8(currentKing) + int8(DirW)*2)
		// Position.MakeMove() should recognize castling, or add a field to Move type?
		outputMoves = append(outputMoves, Move{currentKing, kingDest})
	}
	if kingsideCastlePossible {
		kingDest := square(int8(currentKing) + int8(DirE)*2)
		outputMoves = append(outputMoves, Move{currentKing, kingDest})
	}

	return outputMoves
}

func (pos *Position) generateSlidingPieceMoves(from square, currColorBit, enemyColorBit piece, dirs []Direction, outputMoves *[]Move) {
	for _, dir := range dirs {
		for to := from + square(dir); to&InvalidSquare == 0; to = to + square(dir) {
			toContent := pos.board[to]
			if toContent&currColorBit != 0 {
				break
			}
			*outputMoves = append(*outputMoves, Move{from, to})
			if toContent&enemyColorBit != 0 {
				break
			}
		}
	}
}
