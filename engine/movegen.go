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
	currentPieces, currentKing, pawnAdvanceDirection, _, enemyPieceBit := pos.GetCurrentContext()
	for _, from := range currentPieces {
		p := pos.board[from]
		fmt.Printf("\t%v at square: %v\n", p, from)
		//IDEA table of functions indexed by piece? Benchmark it
		//IDEA No piece lists? just iterate over all fields. Perhaps add list once material gone
		switch (p) {
		case WPawn, BPawn:
			to := from+square(pawnAdvanceDirection)-1 // queenside take
			fmt.Println("toSquare (qs take)", to)
			if pos.board[to]&enemyPieceBit != 0 {
				outputMoves = append(outputMoves, Move{from, to})
			}
			to = to + 2 // kingside take
			fmt.Println("toSquare (ks take)", to)
			if pos.board[to]&enemyPieceBit != 0 {
				outputMoves = append(outputMoves, Move{from, to})
			}
			//takes
		case WBishop, BBishop:
			//fmt.Println("it's a bishop!")
		case WKnight, BKnight:
			fmt.Println("it's a knight!")
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


func (pos *Position) GetCurrentContext() ([]square, square, Direction, piece, piece) {
	if pos.flags&FlagWhiteTurn == 0 {
		return pos.blackPieces, pos.blackKing, DirS, BlackPieceBit, WhitePieceBit
	}
	return pos.whitePieces, pos.whiteKing, DirN, WhitePieceBit, BlackPieceBit
}