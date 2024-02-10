package engine

import "fmt"

type Move struct {
	from, to  square
	promoteTo piece
	enPassant square
}

type backtrackInfo struct {
	move          Move
	lastFlags     byte
	lastEnPassant square
	takenPiece    piece
}

type Generator struct {
	pos     *Position
	history []backtrackInfo
	outputMoves []Move
}

func NewMove(from, to square) Move {
	return Move{from, to, NullPiece, InvalidSquare}
}

func NewPromotionMove(from, to square, promoteTo piece) Move {
	return Move{from, to, promoteTo, InvalidSquare}
}

var kingDirections = []Direction{DirN, DirS, DirE, DirW, DirNE, DirNW, DirSW, DirSE}

func (move Move) String() string {
	if move.promoteTo != NullPiece {
		return move.from.String() + move.to.String() + move.promoteTo.String()
	}
	return move.from.String() + move.to.String()
}

func NewGenerator() *Generator {
	return &Generator{
		pos:     NewPosition(),
		history: make([]backtrackInfo, 0, 20),
		outputMoves: make([]Move, 0, 60),
	}
}

func NewGeneratorFromFen(fen string) (*Generator, error) {
	//new pos allocation for every generator, worthwhile reusing?
	fenPos, err := NewPositionFromFen(fen)
	if err != nil {
		return nil, err
	}
	return &Generator{
		pos:     fenPos,
		history: make([]backtrackInfo, 0, 20),
	}, nil
}

// TODO add a method AssertAndPushMove() 
func (gen *Generator) PushMove(move Move) (success bool) {
	undo := gen.pos.MakeMove(move)
	if (undo.move == Move{}) {
		return false
	}
	gen.history = append(gen.history, undo)
	return true
}

func (gen *Generator) PopMove() {
	lastIdx := len(gen.history) - 1
	gen.pos.UnmakeMove(gen.history[lastIdx])
	gen.history = gen.history[:lastIdx]
}

// GenerateMoves returns legal moves from position that this generator is currently holding
func (gen *Generator) GenerateMoves() []Move {
	gen.generatePseudoLegalMoves()
	fmt.Println("Pseudo legal movessssssss ", gen.outputMoves)
	i := 0
	for _, pseudoMove := range gen.outputMoves {
		
		undo := gen.pos.MakeMove(pseudoMove)
		// move is valid
		if (undo.move != Move{}) {
			gen.outputMoves[i] = pseudoMove
			i++
			gen.pos.UnmakeMove(undo)
		}
	}
	gen.outputMoves = gen.outputMoves[:i]
	return gen.outputMoves
}

func(gen *Generator) Perft(depth byte) int {
	var movesCount int = 0
	if depth == 1 {
		//TODO implement method that only counts the moves
		return len(gen.GenerateMoves())
	}

	for _, move := range gen.GenerateMoves() {
		gen.PushMove(move)
		fmt.Printf("Depth: %v Just pushed %v. position is: %v\n", depth, move, gen)
		movesCount += gen.Perft(depth-1)
		fmt.Printf("movesCount: %v \n", movesCount)

		gen.PopMove()
	}
	return movesCount
}

func (gen *Generator) generatePseudoLegalMoves() {
	pos := gen.pos
	gen.outputMoves = gen.outputMoves[:0]

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
			if pos.board[to]&enemyColorBit != 0 || to == pos.enPassSquare {
				appendPawnMoves(from, to, promotionRank, &gen.outputMoves)
			}
			// kingside take
			to = from + square(pawnAdvanceDirection) + 1
			if pos.board[to]&enemyColorBit != 0 || to == pos.enPassSquare {
				appendPawnMoves(from, to, promotionRank, &gen.outputMoves)
			}
			//pushes
			to = from + square(pawnAdvanceDirection)
			if pos.board[to] == NullPiece {
				appendPawnMoves(from, to, promotionRank, &gen.outputMoves)
				enPassantSquare := to
				to = to + square(pawnAdvanceDirection)
				if from.getRank() == pawnStartRank && pos.board[to] == NullPiece {
					gen.outputMoves = append(gen.outputMoves, Move{from, to, NullPiece, enPassantSquare})
				}
			}
		case WKnight, BKnight:
			dirs := [...]Direction{DirNNE, DirSSW, DirNNW, DirSSE, DirNEE, DirSWW, DirNWW, DirSEE}
			for _, dir := range dirs {
				to := from + square(dir)
				if to&InvalidSquare == 0 && pos.board[to]&currColorBit == 0 {
					gen.outputMoves = append(gen.outputMoves, NewMove(from, to))
				}
			}
		case WBishop, BBishop:
			dirs := []Direction{DirNE, DirSE, DirNW, DirSW}
			pos.appendSlidingPieceMoves(from, currColorBit, enemyColorBit, dirs, &gen.outputMoves)
		case WRook, BRook:
			dirs := []Direction{DirN, DirS, DirE, DirW}
			pos.appendSlidingPieceMoves(from, currColorBit, enemyColorBit, dirs, &gen.outputMoves)
		case WQueen, BQueen:
			pos.appendSlidingPieceMoves(from, currColorBit, enemyColorBit, kingDirections, &gen.outputMoves)
		default:
			panic(fmt.Sprintf("Unexpected piece found: %v at %v pos %v", byte(piece), from, gen.pos))
		}
	}
	// king moves
	for _, dir := range kingDirections {
		to := currentKing + square(dir)
		if to&InvalidSquare == 0 && pos.board[to]&currColorBit == 0 {
			gen.outputMoves = append(gen.outputMoves, NewMove(currentKing, to))
		}
	}
	if queensideCastlePossible {
		//so much casting.. could it be modelled better?
		var kingAsByte, dirAsByte int8 = int8(currentKing), int8(DirW)
		kingDest := kingAsByte + dirAsByte*2
		if pos.board[kingAsByte+dirAsByte] == NullPiece &&
			pos.board[kingDest] == NullPiece &&
			pos.board[kingAsByte+dirAsByte*3] == NullPiece {
			gen.outputMoves = append(gen.outputMoves, NewMove(currentKing, square(kingDest)))
		}
	}
	if kingsideCastlePossible {
		var kingAsByte, dirAsByte int8 = int8(currentKing), int8(DirE)
		kingDest := kingAsByte + dirAsByte*2
		if pos.board[kingAsByte+dirAsByte] == NullPiece && pos.board[kingDest] == NullPiece {
			gen.outputMoves = append(gen.outputMoves, NewMove(currentKing, square(kingDest)))
		}
	}
}

func (gen *Generator) String() string {
	return gen.pos.String()
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
