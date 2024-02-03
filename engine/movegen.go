package engine

import "fmt"

type Move struct {
	from, to square
}

func (move Move) String() string {
	return move.from.String() + move.to.String()
}

func (pos *Position) GenerateMoves() []Move {
	// TODO - move this to Position - and recycle it - benchmark speed difference
	var outputMoves []Move = make([]Move, 0, 60)
	currentPieces, currentKing, pawnAdvanceDirection,
		currColorBit, enemyColorBit, pawnStartRank := pos.GetCurrentContext()
	for _, from := range currentPieces {
		p := pos.board[from]
		//IDEA table of functions indexed by piece? Benchmark it
		//IDEA No piece lists? just iterate over all fields. Perhaps add list once material gone
		switch p {
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
			dirs := []Direction{DirN, DirS, DirE, DirW, DirNE, DirNW, DirSW, DirSE}
			pos.generateSlidingPieceMoves(from, currColorBit, enemyColorBit, dirs, &outputMoves)
		default:
			panic(fmt.Sprintf("Unexpected piece found: %v", byte(p)))
		}
	}

	fmt.Printf("King at square: %v\n", currentKing)
	return outputMoves
}

func (pos *Position) generateSlidingPieceMoves(from square, currColorBit, enemyColorBit piece, dirs []Direction, outputMoves *[]Move) {
	for _, dir := range dirs {
		for to := from + square(dir); to&InvalidSquare == 0 ; to = to + square(dir) {
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
