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
	currentPieces, currentKing, pawnAdvanceDirection, currPieceColorBit, enemyPieceBit, pawnStartRank := pos.GetCurrentContext()
	for _, from := range currentPieces {
		p := pos.board[from]
		fmt.Printf("\t%v at square: %v\n", p, from)
		//IDEA table of functions indexed by piece? Benchmark it
		//IDEA No piece lists? just iterate over all fields. Perhaps add list once material gone
		switch p {
		case WPawn, BPawn:
			// queenside take
			to := from + square(pawnAdvanceDirection) - 1
			fmt.Println("toSquare (qs take)", to)
			if pos.board[to]&enemyPieceBit != 0 {
				outputMoves = append(outputMoves, Move{from, to})
			}
			// kingside take
			to = to + 2
			fmt.Println("toSquare (ks take)", to)
			if pos.board[to]&enemyPieceBit != 0 {
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
		case WBishop, BBishop:
			fmt.Println("it's a bishop!")
		case WKnight, BKnight:
			dirs := [...]Direction{DirNNE, DirSSW, DirNNW, DirSSE, DirNEE, DirSWW, DirNWW, DirSEE}
			for _, dir := range dirs {
				to := from + square(dir)
				if to&InvalidSquare == 0 && pos.board[to]&currPieceColorBit == 0 {
					outputMoves = append(outputMoves, Move{from, to})
				}
			}
		case WRook, BRook:
			fmt.Println("it's a rook!")
		case WQueen, BQueen:
			fmt.Println("it's a Queen!")
		default:
			panic(fmt.Sprintf("Unexpected piece found: %v", byte(p)))
		}
	}

	fmt.Printf("King at square: %v\n", currentKing)
	return outputMoves
}
