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
	ranking int16
	flags   byte
}

const (
	// set for move that changes material (is capture or promotion)
	mFlagTactical byte = 1 << iota
)

// ranking bonud for move that is on line returned by previous iteration of iterative deepening
// TODO there's some overlap between some sacrificing captures and killer moves.. run some games to fine-tune it
// Watch out for overlflow - rankedMove.rakning is int16!
const (
	rankingBonusPvMove    int16 = 10000
	rankingBonusTactical  int16 = 9000
	rankingBonusKiller1st int16 = 8000
	rankingBonusKiller2nd       = 7000
)

type Generator struct {
	// indexed [plyIdx]
	posStack []Position
	// indexed [plyIdx][moveIdx]
	movStack [][]rankedMove
	// ply meaning turn of one side AKA half move
	plyIdx int16
	//index of the first move in currently searched line (so it can be print in quiescence search)
	firstMoveIdx int
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
	return fmt.Sprintf("(%v, %d)", move.mov, move.ranking)
}

func NewGenerator() *Generator {
	newPosStack := make([]Position, plyBufferCapacity)
	newPosStack[0] = NewPosition()
	return &Generator{
		posStack: newPosStack,
		movStack: newMoveStack(),
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
		movStack: newMoveStack(),
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
	gen.posStack[gen.plyIdx+1] = gen.posStack[gen.plyIdx]
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

func (gen Generator) getTopPos() *Position {
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

var tested_for_legality Position

// Returns true if pseudolegal is legal in pos. False otherwise.
func isLegal(pos *Position, pseudolegal Move) bool {
	// it's important to assign this to global var, otherwise a lot of GC kicks in for some reason
	tested_for_legality = *pos
	return (&tested_for_legality).MakeMove(pseudolegal)
}

func (gen *Generator) Perft(depth int) int64 {
	var movesCount int64 = 0
	if depth == 1 {
		return int64(gen.getTopPos().countMoves())
	} else if depth == 0 {
		return 1
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

func (gen *Generator) PerftDivTactical(depth int)  {
	var total int64 = 0
	if depth <= 1 {
		return
	}

	for _, move := range gen.GenerateMoves() {
		gen.PushMove(move.mov)
		subTotal := gen.PerftTactical(depth - 1)
		total += subTotal
		gen.PopMove()
		fmt.Printf("%v: %d\n", move.mov, subTotal)
	}
	fmt.Println("total material-changing moves:", total)
}

func (gen *Generator) Perftd(depth int) {
	if depth == 0 {
		return
	}
	var total int64 = 0
	for _, rankedMove := range gen.GenerateMoves() {
		move := rankedMove.mov
		gen.PushMove(move)
		subTotal := gen.Perft(depth - 1)
		total += subTotal
		fmt.Printf("%v: %d\n", move, subTotal)
		gen.PopMove()
	}
	fmt.Println("total:", total)
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
		enemyPiece := pos.board[to]
		if enemyPiece&enemyColorBit != NullPiece {
			pos.appendPawnCaptures(from, to, promotionRank, enemyPiece&ColorlessPiece, outputMoves)
		} else if to == pos.enPassSquare {
			pos.appendPawnCaptures(from, to, promotionRank, Pawn, outputMoves)
		}
		//pushes
		to = from + square(pawnAdvanceDirection)
		if pos.board[to] == NullPiece {
			appendPawnPushes(from, to, promotionRank, pos.ply, outputMoves)
			enPassantSquare := to
			to = to + square(pawnAdvanceDirection)
			if from.getRank() == pawnStartRank && pos.board[to] == NullPiece {
				*outputMoves = append(*outputMoves, rankedMove{Move{from, to, NullPiece, enPassantSquare}, 0, 0})
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
					appendMoveOrCapture(outputMoves, from, to, pos)
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
			appendMoveOrCapture(outputMoves, currentKingSq, to, pos)
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
			// TODO same expensive check as done for normal king move - reuse prev result
			!pos.isUnderCheck(enemyPieces, enemyPawns, enemyKingSq, square(kingAsByte+dirAsByte)) &&
			!pos.isUnderCheck(enemyPieces, enemyPawns, enemyKingSq, square(kingDest)) {
			mov := NewMove(currentKingSq, square(kingDest))
			*outputMoves = append(*outputMoves, rankedMove{mov, probeKillerMoves(mov, pos.ply), 0})
		}
	}
	if kingsideCastlePossible {
		var kingAsByte, dirAsByte int8 = int8(currentKingSq), int8(DirE)
		kingDest := kingAsByte + dirAsByte*2
		if pos.board[kingAsByte+dirAsByte] == NullPiece && pos.board[kingDest] == NullPiece &&
			!pos.isUnderCheck(enemyPieces, enemyPawns, enemyKingSq, currentKingSq) &&
			// TODO same expensive check as done for normal king move - reuse prev result
			!pos.isUnderCheck(enemyPieces, enemyPawns, enemyKingSq, square(kingAsByte+dirAsByte)) &&
			!pos.isUnderCheck(enemyPieces, enemyPawns, enemyKingSq, square(kingDest)) {
			mov := NewMove(currentKingSq, square(kingDest))
			*outputMoves = append(*outputMoves, rankedMove{mov, probeKillerMoves(mov, pos.ply), 0})
		}
	}
}

// TODO I only take board from pos. How would it affect the speed if I passed it as a table, or ref to table.
func appendMoveOrCapture(outputMoves *[]rankedMove, from, to square, pos *Position) {
	mov := NewMove(from, to)
	attacked := pos.board[mov.to] & ColorlessPiece
	if attacked == NullPiece {
		*outputMoves = append(*outputMoves, rankedMove{mov, probeKillerMoves(mov, pos.ply), 0})
		return
	}
	attacker := pos.board[mov.from] & ColorlessPiece
	captureRanking := int16(pieceToScore(attacked)-pieceToScore(attacker)) + rankingBonusTactical
	*outputMoves = append(*outputMoves,
		rankedMove{mov, captureRanking, mFlagTactical})
}

func appendSlidingPieceMoveOrCapture(outputMoves *[]rankedMove, from, to square, attacker, attacked piece, ply int16) {
	mov := NewMove(from, to)
	if attacked == NullPiece {
		*outputMoves = append(*outputMoves, rankedMove{mov, probeKillerMoves(mov, ply), 0})
		return
	}
	captureRanking := int16(pieceToScore(attacked)-pieceToScore(attacker)) + rankingBonusTactical
	*outputMoves = append(*outputMoves, rankedMove{mov, captureRanking, mFlagTactical})
}

func appendCapture(outputMoves *[]rankedMove, from, to square, attacker, attacked piece) {
	mov := NewMove(from, to)
	captureRanking := int16(pieceToScore(attacked)-pieceToScore(attacker)) + rankingBonusTactical
	*outputMoves = append(*outputMoves, rankedMove{mov, captureRanking, mFlagTactical})
}

func probeKillerMoves(mov Move, ply int16) int16 {
	killers := killerMoves[ply]
	if mov == killers[0] {
		return rankingBonusKiller1st
	} else if mov == killers[1] {
		return rankingBonusKiller2nd
	}
	return 0
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
			appendPawnPushes(from, to, promotionRank, gen.getTopPos().ply, outputMoves)
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
					appendCapture(outputMoves, from, to, Knight, pos.board[to]&ColorlessPiece)
				}
			}
		case WBishop, BBishop:
			dirs := []Direction{DirNE, DirSE, DirNW, DirSW}
			pos.appendSlidingPieceCaptures(from, currColorBit, enemyColorBit, dirs, outputMoves)
		case WRook, BRook:
			dirs := []Direction{DirN, DirS, DirE, DirW}
			pos.appendSlidingPieceCaptures(from, currColorBit, enemyColorBit, dirs, outputMoves)
		case WQueen, BQueen:
			pos.appendSlidingPieceCaptures(from, currColorBit, enemyColorBit, kingDirections, outputMoves)
		default:
			panic(fmt.Sprintf("Unexpected piece found: %v at %v pos %v", byte(piece), from, pos))
		}
	}
	// king moves
	for _, dir := range kingDirections {
		to := currentKingSq + square(dir)
		if to&InvalidSquare == 0 && pos.board[to]&enemyColorBit != 0 &&
			!pos.isUnderCheck(enemyPieces, enemyPawns, enemyKing, to) {
			appendCapture(outputMoves, currentKingSq, to, King, pos.board[to]&ColorlessPiece)
		}
	}
}

// Counts all tactical moves possible from pos position
func (pos *Position) countTacticalMoves() int {
	var movesCount int = 0
	currentPieces, _,
		currentPawns, _,
		currentKingSq, _,
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
			isLegal(pos, NewMove(currentKingSq, to)) {
			movesCount++
		}
	}
	return movesCount
}

// TODO this method shows almost nothing about gen. add more info
func (gen *Generator) String() string {
	return gen.getTopPos().String()
}

func appendPawnPushes(from, to square, promotionRank rank, ply int16, outputMoves *[]rankedMove) {
	if to.getRank() == promotionRank {
		var commonPart int16 = rankingBonusTactical - MaterialPawnScore
		*outputMoves = append(*outputMoves,
			rankedMove{NewPromotionMove(from, to, Queen), MaterialQueenScore + commonPart, mFlagTactical},
			rankedMove{NewPromotionMove(from, to, Rook), MaterialRookScore + commonPart, mFlagTactical},
			rankedMove{NewPromotionMove(from, to, Bishop), MaterialBishopScore + commonPart, mFlagTactical},
			rankedMove{NewPromotionMove(from, to, Knight), MaterialKnightScore + commonPart, mFlagTactical},
		)
	} else {
		mov := NewMove(from, to)
		*outputMoves = append(*outputMoves, rankedMove{mov, probeKillerMoves(mov, ply), 0})
	}
}

func (pos *Position) appendPawnCaptures(from, to square, promotionRank rank, captured piece, outputMoves *[]rankedMove) {
	captureRanking := int16(pieceToScore(captured)-MaterialPawnScore) + rankingBonusTactical
	if to.getRank() == promotionRank {
		var promoCaptureRanking int16 = captureRanking - MaterialPawnScore
		*outputMoves = append(*outputMoves,
			rankedMove{NewPromotionMove(from, to, Queen), MaterialQueenScore + promoCaptureRanking, mFlagTactical},
			rankedMove{NewPromotionMove(from, to, Rook), MaterialRookScore + promoCaptureRanking, mFlagTactical},
			rankedMove{NewPromotionMove(from, to, Bishop), MaterialBishopScore + promoCaptureRanking, mFlagTactical},
			rankedMove{NewPromotionMove(from, to, Knight), MaterialKnightScore + promoCaptureRanking, mFlagTactical},
		)
	} else {
		*outputMoves = append(*outputMoves, rankedMove{NewMove(from, to), captureRanking, mFlagTactical})
	}
}

func (pos *Position) appendSlidingPieceMoves(from square, currColorBit, enemyColorBit piece,
	dirs []Direction, outputMoves *[]rankedMove) {
	attacker := pos.board[from] & ColorlessPiece
	for _, dir := range dirs {
		for to := from + square(dir); to&InvalidSquare == 0; to = to + square(dir) {
			toContent := pos.board[to]
			if toContent&currColorBit != 0 {
				break
			}
			appendSlidingPieceMoveOrCapture(outputMoves, from, to, attacker, toContent&ColorlessPiece, pos.ply)
			if toContent&enemyColorBit != 0 {
				break
			}
		}
	}
}

func (pos *Position) appendSlidingPieceCaptures(from square, currColorBit, enemyColorBit piece,
	dirs []Direction, outputMoves *[]rankedMove) {
	for _, dir := range dirs {
		for to := from + square(dir); to&InvalidSquare == 0; to = to + square(dir) {
			toContent := pos.board[to]
			if toContent&currColorBit != 0 {
				break
			}
			if toContent&enemyColorBit != 0 {
				appendCapture(outputMoves, from, to, pos.board[from]&ColorlessPiece,
					toContent&ColorlessPiece)
				break
			}
		}
	}
}
