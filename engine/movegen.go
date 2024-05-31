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
	rankedMoves []rankedMove
	undo        backtrackInfo
}

// move together with a score suggesting how soon should it be taken while searching the game tree
type rankedMove struct {
	mov Move
	// in alphaBeta search moves are taken starting from the highest ranking
	ranking int
}

// ranking bonud for move that is on line returned by previous iteration of iterative deepening
const (
	rankingBonusPvMove = 20000
	rankingBonusCapture = 1000
)

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
		newPlies[ply] = PlyContext{rankedMoves: make([]rankedMove, 0, moveBufferCapacity)}
	}
	return newPlies
}

func (gen *Generator) PushMove(move Move) (success bool) {
	undo := gen.pos.MakeMove(move)
	if (undo.move == Move{}) {
		return false
	}
	gen.plies[gen.plyIdx].undo = undo
	gen.plyIdx++
	return true
}

func (gen *Generator) PushMoveSafely(move Move) (success bool) {
	if gen.pos.board[move.from]&ColorlessPiece == Pawn &&
		(move.from.getRank() == Rank7 && move.to.getRank() == Rank5 ||
			move.from.getRank() == Rank2 && move.to.getRank() == Rank4) {
		move.enPassant = (move.from + move.to) / 2
	}
	return gen.PushMove(move)
}

func (gen *Generator) PopMove() {
	gen.plyIdx--
	gen.pos.UnmakeMove(gen.plies[gen.plyIdx].undo)
}

// From position that gen generator is currently holding returns all legal moves.
func (gen *Generator) GenerateMoves() []rankedMove {
	return gen.generateLegalMoves(gen.generatePseudoLegalMoves)
}

// From position that gen generator is currently holding returns all legal moves that change material (captures and promotions)
func (gen *Generator) GenerateTacticalMoves() []rankedMove {
	return gen.generateLegalMoves(gen.generatePseudoLegalTacticalMoves)
}

func (gen *Generator) generateLegalMoves(generateSthPseudolegal func()) []rankedMove {
	generateSthPseudolegal()
	plyContext := &gen.plies[gen.plyIdx]
	i := 0
	for _, pseudoMove := range plyContext.rankedMoves {

		undo := gen.pos.MakeMove(pseudoMove.mov)
		// move is valid
		if (undo.move != Move{}) {
			plyContext.rankedMoves[i] = pseudoMove
			i++
			gen.pos.UnmakeMove(undo)
		}
	}
	plyContext.rankedMoves = plyContext.rankedMoves[:i]
	return plyContext.rankedMoves
}

func (gen *Generator) Perft(depth int) int64 {

	var movesCount int64 = 0
	if depth <= 1 {
		return int64(gen.pos.countMoves())
	}

	moves := gen.GenerateMoves()
	for _, move := range moves {
		gen.PushMove(move.mov)
		movesCount += gen.Perft(depth - 1)
		gen.PopMove()
	}
	return movesCount
}

func (gen *Generator) PerftTactical(depth int) int64 {

	var movesCount int64 = 0
	if depth <= 1 {
		return int64(gen.pos.countTacticalMoves())
	}

	moves := gen.GenerateMoves()
	for _, move := range moves {
		gen.PushMove(move.mov)
		movesCount += gen.PerftTactical(depth - 1)
		gen.PopMove()
	}
	return movesCount
}

func (gen *Generator) Perftd(depth int) {
	if depth <= 1 {
		return
	}
	for _, rankedMove := range gen.GenerateMoves() {
		move := rankedMove.mov
		gen.PushMove(move)
		fmt.Printf("%v %d\n", move, gen.Perft(depth-1))
		gen.PopMove()
	}
}

func (gen *Generator) Perftdd(depth int) {
	if depth <= 1 {
		return
	}
	for _, rankedMove := range gen.GenerateMoves() {
		move := rankedMove.mov
		gen.PushMove(move)
		fmt.Printf("Pushed %v: \n", move)

		var sumOfPerft2LevsDown int64 = 0
		if depth <= 1 {
			return
		}
		for _, rankedMovePrime := range gen.GenerateMoves() {
			movePrime := rankedMovePrime.mov
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
	var outputMoves *[]rankedMove = &gen.plies[gen.plyIdx].rankedMoves
	*outputMoves = (*outputMoves)[:0]

	currentPieces, enemyPieces,
		currentPawns, enemyPawns,
		currentKingSq, enemyKingSq,
		pawnAdvanceDirection,
		currColorBit, enemyColorBit,
		queensideCastlePossible, kingsideCastlePossible,
		pawnStartRank, promotionRank := pos.GetCurrentContext()
	for _, from := range currentPawns {
		// queenside take
		to := from + square(pawnAdvanceDirection) - 1

		if to&InvalidSquare == 0 && pos.board[to]&enemyColorBit != NullPiece {
			pos.appendPawnCaptures(from, to, promotionRank, pos.board[to]&ColorlessPiece, outputMoves)
		} else if to == pos.enPassSquare {
			pos.appendPawnCaptures(from, to, promotionRank, Pawn, outputMoves)
		}
		// kingside take
		// this will require boundscheck if I ever decide to implement
		// https://www.talkchess.com/forum/viewtopic.php?p=696431&sid=0f2d2d56c1fed62bbf4d2b793617857f#p696431
		to = from + square(pawnAdvanceDirection) + 1
		enemyPiece := pos.board[to] & enemyColorBit
		if enemyPiece != NullPiece {
			pos.appendPawnCaptures(from, to, promotionRank, enemyPiece&ColorlessPiece, outputMoves)
		} else if to == pos.enPassSquare {
			pos.appendPawnCaptures(from, to, promotionRank, Pawn, outputMoves)
		}
	}
	for _, from := range currentPawns {
		//pushes
		to := from + square(pawnAdvanceDirection)
		if pos.board[to] == NullPiece {
			appendPawnPushes(from, to, promotionRank, outputMoves)
			enPassantSquare := to
			to = to + square(pawnAdvanceDirection)
			if from.getRank() == pawnStartRank && pos.board[to] == NullPiece {
				*outputMoves = append(*outputMoves, rankedMove{Move{from, to, NullPiece, enPassantSquare}, 0})
			}
		}
	}
	for _, from := range currentPieces {
		piece := pos.board[from]
		switch piece {
		case WKnight, BKnight:
			dirs := [...]Direction{DirNNE, DirSSW, DirNNW, DirSSE, DirNEE, DirSWW, DirNWW, DirSEE}
			for _, dir := range dirs {
				to := from + square(dir)
				if to&InvalidSquare == 0 && pos.board[to]&currColorBit == 0 {
					gen.pos.appendRankedMove(outputMoves, from, to)
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
		to := currentKingSq + square(dir)
		if to&InvalidSquare == 0 && pos.board[to]&currColorBit == 0 &&
			// IDEA same check done later in MakeMove. Skip here?
			!pos.isUnderCheck(enemyPieces, enemyPawns, enemyKingSq, to) {
			gen.pos.appendRankedMove(outputMoves, currentKingSq, to)
		}
	}
	if queensideCastlePossible {
		//so much casting.. could it be modelled better?
		var kingAsByte, dirAsByte int8 = int8(currentKingSq), int8(DirW)
		kingDest := kingAsByte + dirAsByte*2
		if pos.board[kingAsByte+dirAsByte] == NullPiece &&
			pos.board[kingDest] == NullPiece &&
			pos.board[kingAsByte+dirAsByte*3] == NullPiece &&
			!pos.isUnderCheck(enemyPieces, enemyPawns, enemyKingSq, currentKingSq) &&
			!pos.isUnderCheck(enemyPieces, enemyPawns, enemyKingSq, square(kingAsByte+dirAsByte)) &&
			// IDEA same check done later in MakeMove. Skip here?
			!pos.isUnderCheck(enemyPieces, enemyPawns, enemyKingSq, square(kingDest)) {
			gen.pos.appendRankedMove(outputMoves, currentKingSq, square(kingDest))
		}
	}
	if kingsideCastlePossible {
		var kingAsByte, dirAsByte int8 = int8(currentKingSq), int8(DirE)
		kingDest := kingAsByte + dirAsByte*2
		if pos.board[kingAsByte+dirAsByte] == NullPiece && pos.board[kingDest] == NullPiece &&
			!pos.isUnderCheck(enemyPieces, enemyPawns, enemyKingSq, currentKingSq) &&
			!pos.isUnderCheck(enemyPieces, enemyPawns, enemyKingSq, square(kingAsByte+dirAsByte)) &&
			// IDEA same check done later in MakeMove. Skip here?
			!pos.isUnderCheck(enemyPieces, enemyPawns, enemyKingSq, square(kingDest)) {
			gen.pos.appendRankedMove(outputMoves, currentKingSq, square(kingDest))
		}
	}
}

func (pos *Position) appendRankedMove(outputMoves *[]rankedMove, from, to square) {
	mov := NewMove(from, to)
	attacked := pos.board[mov.to] & ColorlessPiece
	if attacked == NullPiece {
		*outputMoves = append(*outputMoves, rankedMove{mov, 0})
		return
	}
	attacker := pos.board[mov.from] & ColorlessPiece
	*outputMoves = append(*outputMoves,
		rankedMove{mov, pieceToScore(attacked) - pieceToScore(attacker) + rankingBonusCapture})
}

func appendRankedMove(outputMoves *[]rankedMove, from, to square, attacker, attacked piece) {
	mov := NewMove(from, to)
	if attacked == NullPiece {
		*outputMoves = append(*outputMoves, rankedMove{mov, 0})
		return
	}
	*outputMoves = append(*outputMoves,
		rankedMove{mov, pieceToScore(attacked) - pieceToScore(attacker) + rankingBonusCapture})
}

func (gen *Generator) generatePseudoLegalTacticalMoves() {
	pos := gen.pos
	var outputMoves *[]rankedMove = &gen.plies[gen.plyIdx].rankedMoves
	*outputMoves = (*outputMoves)[:0]

	currentPieces, enemyPieces,
		currentPawns, enemyPawns,
		currentKingSq, enemyKing,
		pawnAdvanceDirection,
		currColorBit, enemyColorBit,
		promotionRank := pos.GetCurrentTacticalMoveContext()

	for _, from := range currentPawns {
		// queenside take
		to := from + square(pawnAdvanceDirection) - 1
		if to&InvalidSquare == 0 && pos.board[to]&enemyColorBit != 0 {
			pos.appendPawnCaptures(from, to, promotionRank, pos.board[to]&ColorlessPiece, outputMoves)
		} else if to == pos.enPassSquare {
			pos.appendPawnCaptures(from, to, promotionRank, Pawn, outputMoves)
		}
		// kingside take -
		// this will require boundscheck if I ever decide to implement https://www.talkchess.com/forum/viewtopic.php?p=696431&sid=0f2d2d56c1fed62bbf4d2b793617857f#p696431
		to = from + square(pawnAdvanceDirection) + 1
		if pos.board[to]&enemyColorBit != 0 {
			pos.appendPawnCaptures(from, to, promotionRank, pos.board[to]&ColorlessPiece, outputMoves)
		} else if to == pos.enPassSquare {
			pos.appendPawnCaptures(from, to, promotionRank, Pawn, outputMoves)
		}
		// promoting pushes
		to = from + square(pawnAdvanceDirection)
		if pos.board[to] == NullPiece && to.getRank() == promotionRank {
			appendPawnPushes(from, to, promotionRank, outputMoves)
		}
	}
	for _, from := range currentPieces {
		piece := pos.board[from]
		switch piece {
		case WKnight, BKnight:
			dirs := [...]Direction{DirNNE, DirSSW, DirNNW, DirSSE, DirNEE, DirSWW, DirNWW, DirSEE}
			for _, dir := range dirs {
				to := from + square(dir)
				if to&InvalidSquare == 0 && pos.board[to]&enemyColorBit != 0 {
					appendRankedMove(outputMoves, from, to, Knight, pos.board[to]&ColorlessPiece)
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
		to := currentKingSq + square(dir)
		if to&InvalidSquare == 0 && pos.board[to]&enemyColorBit != 0 &&
			!pos.isUnderCheck(enemyPieces, enemyPawns, enemyKing, to) {
			pos.appendRankedMove(outputMoves, currentKingSq, to)
		}
	}
}

// Counts all tactical moves possible from pos position
func (pos *Position) countTacticalMoves() int {
	var movesCount int = 0
	currentPieces, enemyPieces,
		currentPawns, enemyPawns,
		currentKingSq, enemyKing,
		pawnAdvanceDirection,
		currColorBit, enemyColorBit,
		promotionRank := pos.GetCurrentTacticalMoveContext()
	for _, from := range currentPawns {
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
		if pos.board[to] == NullPiece && to.getRank() == promotionRank {
			movesCount += pos.countPawnMoves(from, to, promotionRank)
		}
	}
	for _, from := range currentPieces {
		piece := pos.board[from]
		switch piece {
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
		to := currentKingSq + square(dir)
		if to&InvalidSquare == 0 && pos.board[to]&enemyColorBit != 0 &&
			!pos.isUnderCheck(enemyPieces, enemyPawns, enemyKing, to) {
			movesCount++
		}
	}
	return movesCount
}

func (gen *Generator) String() string {
	return gen.pos.String()
}

func appendPawnPushes(from, to square, promotionRank rank, outputMoves *[]rankedMove) {
	if to.getRank() == promotionRank {
		*outputMoves = append(*outputMoves,
			rankedMove{NewPromotionMove(from, to, Queen), MaterialQueenScore - MaterialPawnScore},
			rankedMove{NewPromotionMove(from, to, Rook), MaterialRookScore - MaterialPawnScore},
			rankedMove{NewPromotionMove(from, to, Bishop), MaterialBishopScore - MaterialPawnScore},
			rankedMove{NewPromotionMove(from, to, Knight), MaterialKnightScore - MaterialPawnScore},
		)
	} else {
		*outputMoves = append(*outputMoves, rankedMove{NewMove(from, to), 0})
	}
}

func (pos *Position) appendPawnCaptures(from, to square, promotionRank rank, captured piece, outputMoves *[]rankedMove) {
	captureRanking := pieceToScore(captured) - MaterialPawnScore
	if to.getRank() == promotionRank {
		*outputMoves = append(*outputMoves,
			rankedMove{NewPromotionMove(from, to, Queen), rankingBonusCapture + captureRanking +
				MaterialQueenScore - MaterialPawnScore},
			rankedMove{NewPromotionMove(from, to, Rook), rankingBonusCapture + captureRanking +
				MaterialRookScore - MaterialPawnScore},
			rankedMove{NewPromotionMove(from, to, Bishop), rankingBonusCapture + captureRanking +
				MaterialBishopScore - MaterialPawnScore},
			rankedMove{NewPromotionMove(from, to, Knight), rankingBonusCapture + captureRanking +
				MaterialKnightScore - MaterialPawnScore},
		)
	} else {
		*outputMoves = append(*outputMoves, rankedMove{NewMove(from, to), captureRanking})
	}
}

func (pos *Position) appendSlidingPieceMoves(from square, currColorBit, enemyColorBit piece, dirs []Direction, outputMoves *[]rankedMove) {
	attacker := pos.board[from] & ColorlessPiece
	for _, dir := range dirs {
		for to := from + square(dir); to&InvalidSquare == 0; to = to + square(dir) {
			toContent := pos.board[to]
			if toContent&currColorBit != 0 {
				break
			}
			appendRankedMove(outputMoves, from, to, attacker, toContent&ColorlessPiece)
			if toContent&enemyColorBit != 0 {
				break
			}
		}
	}
}

func (pos *Position) appendSlidingPieceTacticalMoves(from square, currColorBit, enemyColorBit piece, dirs []Direction, outputMoves *[]rankedMove) {
	for _, dir := range dirs {
		for to := from + square(dir); to&InvalidSquare == 0; to = to + square(dir) {
			toContent := pos.board[to]
			if toContent&currColorBit != 0 {
				break
			}
			if toContent&enemyColorBit != 0 {
				appendRankedMove(outputMoves, from, to, pos.board[from]&ColorlessPiece,
					toContent&ColorlessPiece)
				break
			}
		}
	}
}
