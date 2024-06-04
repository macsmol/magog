package engine

import (
	"fmt"
)

type Move struct {
	from, to  square
	promoteTo piece
	enPassant square
}

// move together with a score suggesting how soon should it be taken while searching the game tree
// in alphaBeta search moves are taken starting from the highest ranking
type rankedMove struct {
	mov     Move
	ranking int
}

// ranking bonud for move that is on line returned by previous iteration of iterative deepening
const (
	rankingBonusPvMove  = 20000
	rankingBonusCapture = 1000
)

type Generator struct {
	// indexed [plyIdx]
	posStack []Position
	// indexed [plyIdx][moveIdx]
	movStack  [][]rankedMove
	// ply meaning turn of one side AKA half move
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

func (move rankedMove) String() string {
	return fmt.Sprintf("(%v, %d)",move.mov, move.ranking)
}

func NewGenerator() *Generator {
	newPosStack := make([]Position, plyBufferCapacity)
	newPosStack[0] = NewPosition()
	return &Generator{
		posStack: newPosStack,
		movStack:    newMoveStack(),
		plyIdx:   0,
	}
}

func NewGeneratorFromFen(fen string) (*Generator, error) {
	fenPos, err := NewPositionFromFen(fen)
	if err != nil {
		return nil, err
	}
	newPosStack := make([]Position, plyBufferCapacity)
	newPosStack[0] = fenPos
	
	return &Generator{
		posStack: newPosStack,
		movStack:    newMoveStack(),
		plyIdx:   0,
	}, nil
}

func newMoveStack() [][]rankedMove {
	// IDEA probably will experiment with something that does not realloc whole
	// thing when exceeding max
	newMoveStack := make([][]rankedMove, plyBufferCapacity)
	for ply := 0; ply < plyBufferCapacity; ply++ {
		newMoveStack[ply] = make([]rankedMove, 0, moveBufferCapacity)
	}
	return newMoveStack
}

// Pushes legalMove on top of posStack. Panics if the move is illegal
func (gen *Generator) PushMove(legalMove Move) {
	gen.posStack[gen.plyIdx + 1] = gen.posStack[gen.plyIdx]
	gen.plyIdx++
	// MakeMove is defined on *Position. So the call below will change value at the top of the stack, right?
	success := gen.posStack[gen.plyIdx].MakeMove(legalMove)
	if !success {
		panic(fmt.Sprintf("Applying move %v resulted in illegal position %v", legalMove, gen.getTopPos()))
	}
}

func (gen *Generator) ApplyUciMove(moveFromUci Move) {
	if gen.getTopPos().board[moveFromUci.from]&ColorlessPiece == Pawn &&
		(moveFromUci.from.getRank() == Rank7 && moveFromUci.to.getRank() == Rank5 ||
			moveFromUci.from.getRank() == Rank2 && moveFromUci.to.getRank() == Rank4) {
		moveFromUci.enPassant = (moveFromUci.from + moveFromUci.to) / 2
	}
	success := gen.posStack[gen.plyIdx].MakeMove(moveFromUci)
	if !success {
		panic(fmt.Sprintf("Applying uci move %v resulted in illegal position %v", moveFromUci, gen.getTopPos()))
	}
}

func (gen Generator)getTopPos() *Position {
	return &gen.posStack[gen.plyIdx]
}

func (gen Generator) getMovesFromTopPos() *[]rankedMove {
	return &gen.movStack[gen.plyIdx]
}

func (gen *Generator) PopMove() {
	gen.plyIdx--
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
	rankedMoves := &gen.movStack[gen.plyIdx]
	i := 0
	for _, pseudoMove := range *rankedMoves {
		success := isLegal(gen.getTopPos(), pseudoMove.mov)
		// move is valid
		if success {
			(*rankedMoves)[i] = pseudoMove
			i++
		}
	}
	(*rankedMoves) = (*rankedMoves)[:i]
	return *rankedMoves
}

func (gen *Generator) Perft(depth int) int64 {
	var movesCount int64 = 0
	if depth <= 1 {
		return int64(gen.getTopPos().countMoves())
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
		return int64(gen.getTopPos().countTacticalMoves())
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
	pos := gen.getTopPos()
	var outputMoves *[]rankedMove = gen.getMovesFromTopPos()
	
	*outputMoves = (*outputMoves)[:0]

	currentPieces, enemyPieces,
		currentPawns, enemyPawns,
		currentKingSq, enemyKingSq,
		pawnAdvanceDirection,
		currColorBit, enemyColorBit,
		queensideCastlePossible, kingsideCastlePossible,
		pawnStartRank, promotionRank := pos.GetCurrentContext()
	for i := int8(0); i < currentPawns.size; i++ {
		from := currentPawns.squares[i]
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
		//pushes
		to = from + square(pawnAdvanceDirection)
		if pos.board[to] == NullPiece {
			appendPawnPushes(from, to, promotionRank, outputMoves)
			enPassantSquare := to
			to = to + square(pawnAdvanceDirection)
			if from.getRank() == pawnStartRank && pos.board[to] == NullPiece {
				*outputMoves = append(*outputMoves, rankedMove{Move{from, to, NullPiece, enPassantSquare}, 0})
			}
		}
	}
	for i := int8(0); i < currentPieces.size; i++ {
		from := currentPieces.squares[i]

		piece := pos.board[from]
		switch piece {
		case WKnight, BKnight:
			dirs := [...]Direction{DirNNE, DirSSW, DirNNW, DirSSE, DirNEE, DirSWW, DirNWW, DirSEE}
			for _, dir := range dirs {
				to := from + square(dir)
				if to&InvalidSquare == 0 && pos.board[to]&currColorBit == 0 {
					appendRankedMoveOrCapture(outputMoves, from, to, pos)
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
			appendRankedMoveOrCapture(outputMoves, currentKingSq, to, pos)
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
				appendRankedMoveOrCapture(outputMoves, currentKingSq, square(kingDest), pos)
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
			appendRankedMoveOrCapture(outputMoves, currentKingSq, square(kingDest), pos)
		}
	}
}

//TODO I only take board from pos. How would it affect the speed if I passed it as a table, or ref to table.
func appendRankedMoveOrCapture(outputMoves *[]rankedMove, from, to square, pos *Position) {
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
	pos := gen.getTopPos()
	var outputMoves *[]rankedMove = gen.getMovesFromTopPos()
	*outputMoves = (*outputMoves)[:0]

	currentPieces, enemyPieces,
		currentPawns, enemyPawns,
		currentKingSq, enemyKing,
		pawnAdvanceDirection,
		currColorBit, enemyColorBit,
		promotionRank := pos.GetCurrentTacticalMoveContext()

	for i := int8(0); i < currentPawns.size; i++ {
		from := currentPawns.squares[i]
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
	for i := int8(0); i < currentPieces.size; i++ {
		from := currentPieces.squares[i]
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
			appendRankedMoveOrCapture(outputMoves, currentKingSq, to, pos)
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
	for i := int8(0); i < currentPawns.size; i++ {
		from := currentPawns.squares[i]
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
	for i := int8(0); i < currentPieces.size; i++ {
		from := currentPieces.squares[i]
		piece := pos.board[from]
		switch piece {
		case WKnight, BKnight:
			dirs := [...]Direction{DirNNE, DirSSW, DirNNW, DirSSE, DirNEE, DirSWW, DirNWW, DirSEE}
			for _, dir := range dirs {
				to := from + square(dir)
				if to&InvalidSquare == 0 && pos.board[to]&enemyColorBit != 0 {
					if isLegal(pos, NewMove(from, to)) {
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

//TODO this method shows almost nothing about gen. add more info
func (gen *Generator) String() string {
	return gen.getTopPos().String()
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
