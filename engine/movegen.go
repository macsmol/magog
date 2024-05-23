package engine

import (
	"fmt"
)

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

// Ply meaning: https://www.chessprogramming.org/Ply
type PlyContext struct {
	moves []Move
	undo  backtrackInfo
}

type Generator struct {
	pos    *Position
	plies  []PlyContext
	plyIdx int16
}

const (
	plyBufferCapacity  int = 200
	moveBufferCapacity int = 60
)

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
		pos:    NewPosition(),
		plies:  newPlies(),
		plyIdx: 0,
	}
}

func NewGeneratorFromFen(fen string) (*Generator, error) {
	//new pos allocation for every generator, worthwhile reusing?
	fenPos, err := NewPositionFromFen(fen)
	if err != nil {
		return nil, err
	}
	return &Generator{
		pos:    fenPos,
		plies:  newPlies(),
		plyIdx: 0,
	}, nil
}

func newPlies() []PlyContext {

	// IDEA probably will experiment with something that does not realloc whole
	// thing when exceeding max
	newPlies := make([]PlyContext, plyBufferCapacity)
	for ply := 0; ply < plyBufferCapacity; ply++ {
		newPlies[ply] = PlyContext{moves: make([]Move, 0, moveBufferCapacity)}
	}
	return newPlies
}

// TODO add a method AssertAndPushMove()
func (gen *Generator) PushMove(move Move) (success bool) {
	undo := gen.pos.MakeMove(move)
	if (undo.move == Move{}) {
		return false
	}
	gen.plies[gen.plyIdx].undo = undo
	gen.plyIdx++
	return true
}

func (gen *Generator) PopMove() {
	gen.plyIdx--
	gen.pos.UnmakeMove(gen.plies[gen.plyIdx].undo)
}

// From position that gen generator is currently holding returns all legal moves.
func (gen *Generator) GenerateMoves() []Move {
	return gen.generateLegalMoves(gen.generatePseudoLegalMoves)
}

// From position that gen generator is currently holding returns all legal moves that change material (captures and promotions)
func (gen *Generator) GenerateTacticalMoves() []Move {
	return gen.generateLegalMoves(gen.generatePseudoLegalTacticalMoves)
}

func (gen *Generator)generateLegalMoves(generateSthPseudolegal func()) []Move {
	generateSthPseudolegal()
	plyContext := &gen.plies[gen.plyIdx]
	i := 0
	for _, pseudoMove := range plyContext.moves {

		// gen.pos.AssertConsistency("before making: "+ pseudoMove.String() + gen.pos.String())
		undo := gen.pos.MakeMove(pseudoMove)
		// move is valid
		if (undo.move != Move{}) {
			plyContext.moves[i] = pseudoMove
			i++
			// todo add string for generator to print the current line/move sequence
			// gen.pos.AssertConsistency("before Unmake: " + pseudoMove.String() + gen.pos.String())
			gen.pos.UnmakeMove(undo)
		}
		// gen.pos.AssertConsistency("after making: "+ pseudoMove.String() + gen.pos.String())
	}
	plyContext.moves = plyContext.moves[:i]
	return plyContext.moves
}

func (gen *Generator) Perft(depth int) int64 {

	var movesCount int64 = 0
	if depth <= 1 {
		//TODO implement method that only counts the moves
		return int64(gen.pos.countMoves())
	}

	moves := gen.GenerateMoves()
	for _, move := range moves {
		gen.PushMove(move)
		movesCount += gen.Perft(depth - 1)
		gen.PopMove()
	}
	return movesCount
}

func (gen *Generator) PerftTactical(depth int) int64 {

	var movesCount int64 = 0
	if depth <= 1 {
		//TODO implement method that only counts the moves
		return int64(gen.pos.countTacticalMoves())
	}

	moves := gen.GenerateMoves()
	for _, move := range moves {
		gen.PushMove(move)
		movesCount += gen.PerftTactical(depth - 1)
		gen.PopMove()
	}
	return movesCount
}

func (gen *Generator) Perftd(depth int) {
	if depth <= 1 {
		return
	}
	for _, move := range gen.GenerateMoves() {
		gen.PushMove(move)
		fmt.Printf("%v %d\n", move, gen.Perft(depth-1))
		gen.PopMove()
	}
}

func (gen *Generator) Perftdd(depth int) {
	if depth <= 1 {
		return
	}
	for _, move := range gen.GenerateMoves() {
		gen.PushMove(move)
		fmt.Printf("Pushed %v: \n", move)

		var sumOfPerft2LevsDown int64 = 0
		if depth <= 1 {
			return
		}
		for _, movePrime := range gen.GenerateMoves() {
			gen.PushMove(movePrime)
			fmt.Printf("\tPushed %v: \n", movePrime)
			perft2 := gen.Perft(depth - 2)
			sumOfPerft2LevsDown += perft2
			fmt.Printf("\t%v: %d\n", movePrime, perft2)
			gen.PopMove()
		}
		fmt.Printf("%v: %d\n", move, sumOfPerft2LevsDown)
		gen.PopMove()
	}
}

func (gen *Generator) generatePseudoLegalMoves() {
	pos := gen.pos
	var outputMoves *[]Move = &gen.plies[gen.plyIdx].moves
	*outputMoves = (*outputMoves)[:0]

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
			if to&InvalidSquare == 0 && pos.board[to]&enemyColorBit != 0 || to == pos.enPassSquare {
				appendPawnMoves(from, to, promotionRank, outputMoves)
			}
			// kingside take
			to = from + square(pawnAdvanceDirection) + 1
			if pos.board[to]&enemyColorBit != 0 || to == pos.enPassSquare {
				appendPawnMoves(from, to, promotionRank, outputMoves)
			}
			//pushes
			to = from + square(pawnAdvanceDirection)
			if pos.board[to] == NullPiece {
				appendPawnMoves(from, to, promotionRank, outputMoves)
				enPassantSquare := to
				to = to + square(pawnAdvanceDirection)
				if from.getRank() == pawnStartRank && pos.board[to] == NullPiece {
					*outputMoves = append(*outputMoves, Move{from, to, NullPiece, enPassantSquare})
				}
			}
		case WKnight, BKnight:
			dirs := [...]Direction{DirNNE, DirSSW, DirNNW, DirSSE, DirNEE, DirSWW, DirNWW, DirSEE}
			for _, dir := range dirs {
				to := from + square(dir)
				if to&InvalidSquare == 0 && pos.board[to]&currColorBit == 0 {
					*outputMoves = append(*outputMoves, NewMove(from, to))
				}
			}
		case WBishop, BBishop:
			dirs := []Direction{DirNE, DirSE, DirNW, DirSW}
			pos.appendSlidingPieceMoves(from, currColorBit, enemyColorBit, dirs, outputMoves)
		case WRook, BRook:
			dirs := []Direction{DirN, DirS, DirE, DirW}
			pos.appendSlidingPieceMoves(from, currColorBit, enemyColorBit, dirs, outputMoves)
		case WQueen, BQueen:
			pos.appendSlidingPieceMoves(from, currColorBit, enemyColorBit, kingDirections, outputMoves)
		default:
			panic(fmt.Sprintf("Unexpected piece found: %v at %v pos %v", byte(piece), from, pos))
		}
	}
	// king moves
	for _, dir := range kingDirections {
		to := currentKing + square(dir)
		if to&InvalidSquare == 0 && pos.board[to]&currColorBit == 0 &&
		!pos.isUnderCheck(enemyPieces,enemyKing, to) {
			*outputMoves = append(*outputMoves, NewMove(currentKing, to))
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
			*outputMoves = append(*outputMoves, NewMove(currentKing, square(kingDest)))
		}
	}
	if kingsideCastlePossible {
		var kingAsByte, dirAsByte int8 = int8(currentKing), int8(DirE)
		kingDest := kingAsByte + dirAsByte*2
		if pos.board[kingAsByte+dirAsByte] == NullPiece && pos.board[kingDest] == NullPiece &&
			!pos.isUnderCheck(enemyPieces, enemyKing, currentKing) &&
			!pos.isUnderCheck(enemyPieces, enemyKing, square(kingAsByte+dirAsByte)) &&
			!pos.isUnderCheck(enemyPieces, enemyKing, square(kingDest)) {
			*outputMoves = append(*outputMoves, NewMove(currentKing, square(kingDest)))
		}
	}
}

func (gen *Generator) generatePseudoLegalTacticalMoves() {
	pos := gen.pos
	var outputMoves *[]Move = &gen.plies[gen.plyIdx].moves
	*outputMoves = (*outputMoves)[:0]

	currentPieces, enemyPieces,
		currentKing, enemyKing,
		pawnAdvanceDirection,
		currColorBit, enemyColorBit,
		promotionRank := pos.GetCurrentTacticalMoveContext()
	for _, from := range currentPieces {
		piece := pos.board[from]
		switch piece {
		case WPawn, BPawn:
			// queenside take
			to := from + square(pawnAdvanceDirection) - 1
			if to&InvalidSquare == 0 && pos.board[to]&enemyColorBit != 0 || to == pos.enPassSquare {
				appendPawnMoves(from, to, promotionRank, outputMoves)
			}
			// kingside take
			to = from + square(pawnAdvanceDirection) + 1
			if pos.board[to]&enemyColorBit != 0 || to == pos.enPassSquare {
				appendPawnMoves(from, to, promotionRank, outputMoves)
			}
			// promoting pushes
			to = from + square(pawnAdvanceDirection)
			if pos.board[to] == NullPiece && to.getRank() == promotionRank {
				appendPawnMoves(from, to, promotionRank, outputMoves)
			}
		case WKnight, BKnight:
			dirs := [...]Direction{DirNNE, DirSSW, DirNNW, DirSSE, DirNEE, DirSWW, DirNWW, DirSEE}
			for _, dir := range dirs {
				to := from + square(dir)
				if to&InvalidSquare == 0 && pos.board[to]&enemyColorBit != 0 {
					*outputMoves = append(*outputMoves, NewMove(from, to))
				}
			}
		case WBishop, BBishop:
			dirs := []Direction{DirNE, DirSE, DirNW, DirSW}
			pos.appendSlidingPieceTacticalMoves(from, currColorBit, enemyColorBit, dirs, outputMoves)
		case WRook, BRook:
			dirs := []Direction{DirN, DirS, DirE, DirW}
			pos.appendSlidingPieceTacticalMoves(from, currColorBit, enemyColorBit, dirs, outputMoves)
		case WQueen, BQueen:
			pos.appendSlidingPieceTacticalMoves(from, currColorBit, enemyColorBit, kingDirections, outputMoves)
		default:
			panic(fmt.Sprintf("Unexpected piece found: %v at %v pos %v", byte(piece), from, pos))
		}
	}
	// king moves
	for _, dir := range kingDirections {
		to := currentKing + square(dir)
		if to&InvalidSquare == 0 && pos.board[to]&enemyColorBit != 0 && !pos.isUnderCheck(enemyPieces, enemyKing, to) {
			*outputMoves = append(*outputMoves, NewMove(currentKing, to))
		}
	}
}

// Counts all tactical moves possible from pos position
func (pos *Position) countTacticalMoves() int {
	var movesCount int = 0
	currentPieces, enemyPieces,
		currentKing, enemyKing,
		pawnAdvanceDirection,
		currColorBit, enemyColorBit,
		promotionRank := pos.GetCurrentTacticalMoveContext()
	for _, from := range currentPieces {
		piece := pos.board[from]
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
			if pos.board[to] == NullPiece && to.getRank() == promotionRank{
				movesCount += pos.countPawnMoves(from, to, promotionRank)
			}
		case WKnight, BKnight:
			dirs := [...]Direction{DirNNE, DirSSW, DirNNW, DirSSE, DirNEE, DirSWW, DirNWW, DirSEE}
			for _, dir := range dirs {
				to := from + square(dir)
				if to&InvalidSquare == 0 && pos.board[to]&enemyColorBit != 0 {
					if pos.isLegal(NewMove(from, to)) {
						movesCount++
					}
				}
			}
		case WBishop, BBishop:
			dirs := []Direction{DirNE, DirSE, DirNW, DirSW}
			movesCount += pos.countSlidingPieceTacticalMoves(from, currColorBit, enemyColorBit, dirs)
		case WRook, BRook:
			dirs := []Direction{DirN, DirS, DirE, DirW}
			movesCount += pos.countSlidingPieceTacticalMoves(from, currColorBit, enemyColorBit, dirs)
		case WQueen, BQueen:
			movesCount += pos.countSlidingPieceTacticalMoves(from, currColorBit, enemyColorBit, kingDirections)
		default:
			panic(fmt.Sprintf("Unexpected piece found: %v at %v pos %v", byte(piece), from, pos))
		}
	}
	// king moves
	for _, dir := range kingDirections {
		to := currentKing + square(dir)
		if to&InvalidSquare == 0 && pos.board[to]&enemyColorBit != 0 && !pos.isUnderCheck(enemyPieces, enemyKing, to) {
			movesCount++
		}
	}
	return movesCount
}

func (gen *Generator) String() string {
	return gen.pos.String()
}

func appendPawnMoves(from, to square, promotionRank rank, outputMoves *[]Move) {
	if to.getRank() == promotionRank {
		*outputMoves = append(*outputMoves,
			NewPromotionMove(from, to, Queen),
			NewPromotionMove(from, to, Rook),
			NewPromotionMove(from, to, Bishop),
			NewPromotionMove(from, to, Knight),
		)
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

func (pos *Position) appendSlidingPieceTacticalMoves(from square, currColorBit, enemyColorBit piece, dirs []Direction, outputMoves *[]Move) {
	for _, dir := range dirs {
		for to := from + square(dir); to&InvalidSquare == 0; to = to + square(dir) {
			toContent := pos.board[to]
			if toContent&currColorBit != 0 {
				break
			}
			if toContent&enemyColorBit != 0 {
				*outputMoves = append(*outputMoves, NewMove(from, to))
				break
			}
		}
	}
}
