package engine

import (
	"fmt"
	"strings"
)

type Position struct {
	// 0x88 board
	board [128]piece
	// all but king
	blackPieces []square
	blackKing   square
	// all but king
	whitePieces  []square
	whiteKing    square
	flags        byte
	enPassSquare square
}

// used for position.flags
const (
	FlagWhiteTurn byte = 1 << iota
	FlagWhiteCanCastleKside
	FlagWhiteCanCastleQside
	FlagBlackCanCastleKside
	FlagBlackCanCastleQside
)

// returns new starting position
func NewPosition() *Position {
	// &Position{}  - shorthand for new Position on heap + return a pointer to it
	return &Position{
		board: [128]piece{
			// FFS how do you turn off whitespace formatting in VSCode?
			A1: WRook, B1: WKnight, C1: WBishop, D1: WQueen, E1: WKing, F1: WBishop, G1: WKnight, H1: WRook,
			A2: WPawn, B2: WPawn, C2: WPawn, D2: WPawn, E2: WPawn, F2: WPawn, G2: WPawn, H2: WPawn,
			A7: BPawn, B7: BPawn, C7: BPawn, D7: BPawn, E7: BPawn, F7: BPawn, G7: BPawn, H7: BPawn,
			A8: BRook, B8: BKnight, C8: BBishop, D8: BQueen, E8: BKing, F8: BBishop, G8: BKnight, H8: BRook,
		},
		blackPieces: []square{
			A8, B8, C8, D8, F8, G8, H8,
			A7, B7, C7, D7, E7, F7, G7, H7},
		blackKing: E8,
		whitePieces: []square{
			A1, B1, C1, D1, F1, G1, H1,
			A2, B2, C2, D2, E2, F2, G2, H2},
		whiteKing: E1,
		flags: FlagWhiteTurn | FlagWhiteCanCastleKside | FlagWhiteCanCastleQside |
			FlagBlackCanCastleKside | FlagBlackCanCastleQside,
		enPassSquare: InvalidSquare,
	}
}

// BUG/IDEA This string is too long to fit in 'debug watch' in VSCode. Not sure how to change cfg.
func (pos *Position) String() string {
	var sb strings.Builder
	sb.WriteRune('\n')
	sb.WriteString("  ┃ a │ b │ c │ d │ e │ f │ g │ h │\n")
	sb.WriteString("━━╋━━━┿━━━┿━━━┿━━━┿━━━┿━━━┿━━━┿━━━┥\n")
	for r := Rank8; r >= Rank1; r -= (Rank2 - Rank1) {
		sb.WriteString(fmt.Sprintf("%v┃", r))
		for f := A; f <= H; f++ {
			p := pos.GetAtFileRank(f, r)
			sb.WriteString(fmt.Sprintf(" %v│", p))
		}
		sb.WriteRune('\n')
		// sb.WriteString("\n──╂───┼───┼───┼───┼───┼───┼───┼───┤\n")
	}
	appendFlagsString(&sb,
		pos.flags&FlagBlackCanCastleQside != 0,
		pos.flags&FlagBlackCanCastleKside != 0,
		pos.flags&FlagWhiteTurn == 0)
	sb.WriteString(fmt.Sprintf("BlackKing: %v; BlackPieces: %v\n", pos.blackKing, pos.blackPieces))
	appendFlagsString(&sb,
		pos.flags&FlagWhiteCanCastleQside != 0,
		pos.flags&FlagWhiteCanCastleKside != 0,
		pos.flags&FlagWhiteTurn != 0)
	sb.WriteString(fmt.Sprintf("WhiteKing: %v; WhitePieces: %v\n", pos.whiteKing, pos.whitePieces))
	sb.WriteString(fmt.Sprintf("En passant square: %v", pos.enPassSquare))
	return sb.String()
}

func appendFlagsString(sb *strings.Builder, castleQueenside, castleKingside, myTurn bool) {
	if castleQueenside {
		sb.WriteString("<--")
	} else {
		sb.WriteString("   ")
	}
	if myTurn {
		sb.WriteRune('X')
	} else {
		sb.WriteRune(' ')
	}
	if castleKingside {
		sb.WriteString("--> ")
	} else {
		sb.WriteString("    ")
	}
}

func (pos *Position) GetCurrentContext() (
	currPieces []square, enemyPieces []square,
	currKing square, enemyKing square,
	pawnAdvance Direction,
	currColorBit piece, enemyColorBit piece,
	queensideCastlePossible, kingsideCastlePossible bool,
	currPawnsStartRank, promotionRank rank) {
	if pos.flags&FlagWhiteTurn == 0 {
		return pos.blackPieces, pos.whitePieces,
			pos.blackKing, pos.whiteKing,
			DirS, BlackPieceBit, WhitePieceBit,
			pos.flags&FlagBlackCanCastleQside != 0, pos.flags&FlagBlackCanCastleKside != 0,
			Rank7, Rank1
	}
	return pos.whitePieces, pos.blackPieces,
		pos.whiteKing, pos.blackKing,
		DirN, WhitePieceBit, BlackPieceBit,
		pos.flags&FlagWhiteCanCastleQside != 0, pos.flags&FlagWhiteCanCastleKside != 0,
		Rank2, Rank8
}

// MakeMove applies mov to a position pos. Returns a backtrackInfo that can be used to revert pos back
// to it's original state. In case where applying a mov would result in an illegal Position (i.e. capturing
// a king is possible), pos is left unchanged and backtrackInfo returned is all zeroes.
// Probably will crash when take a move that is either:
// -not possible in this position.
// -not possible according to the rules of chess: a1b8
func (pos *Position) MakeMove(mov Move) (undo backtrackInfo) {
	currPieces, enemyPieces, currKing, enemyKing,
		currCastleRank, currKingSideCastleFlag, currQueenSideCastleFlag,
		enemyCastleRank, enemyKingSideCastleFlag, enemyQueenSideCastleFlag,
		currColorBit := pos.getCurrentMakeMoveContext()

	undo = backtrackInfo{
		move:          mov,
		lastFlags:     pos.flags,
		lastEnPassant: pos.enPassSquare,
	}

	for i := range currPieces {
		if mov.from == currPieces[i] {
			currPieces[i] = mov.to
			break
		}
	}

	if mov.from.getFile() == A && mov.from.getRank() == currCastleRank {
		pos.flags &= ^currQueenSideCastleFlag
	}
	if mov.from.getFile() == H && mov.from.getRank() == currCastleRank {
		pos.flags &= ^currKingSideCastleFlag
	}
	if mov.to.getFile() == A && mov.to.getRank() == enemyCastleRank {
		pos.flags &= ^enemyQueenSideCastleFlag
	}
	if mov.to.getFile() == H && mov.to.getRank() == enemyCastleRank {
		pos.flags &= ^enemyKingSideCastleFlag
	}

	if mov.from == *currKing {
		pos.flags &= ^(currKingSideCastleFlag | currQueenSideCastleFlag)
		*currKing = mov.to
		if mov.from.getFile() == E {
			if mov.to.getFile() == C {
				rookFrom := square(A + file(currCastleRank))
				rookTo := square(D + file(currCastleRank))
				pos.moveRook(rookFrom, rookTo, currPieces, currColorBit)
			} else if mov.to.getFile() == G {
				rookFrom := square(H + file(currCastleRank))
				rookTo := square(F + file(currCastleRank))
				pos.moveRook(rookFrom, rookTo, currPieces, currColorBit)
			}
		}
	}

	if pos.board[mov.to] != NullPiece {
		undo.takenPiece = pos.board[mov.to]
		*enemyPieces = killPiece(*enemyPieces, mov.to)
	}
	if mov.promoteTo == NullPiece {
		pos.board[mov.to] = pos.board[mov.from]
		//en passant take
		if pos.enPassSquare == mov.to && pos.board[mov.from] == Pawn|currColorBit {
			killSquare := square(mov.to.getFile() + file(mov.from.getRank()))
			*enemyPieces = killPiece(*enemyPieces, killSquare)
			undo.takenPiece = pos.board[killSquare]
			pos.board[killSquare] = NullPiece
		}
	} else {
		pos.board[mov.to] = mov.promoteTo | currColorBit
	}
	pos.board[mov.from] = NullPiece

	// move mov was a double push
	pos.enPassSquare = mov.enPassant

	pos.flags = pos.flags ^ FlagWhiteTurn

	// everything's been moved to it's place - time to check if it's actually legal
	if pos.isUnderCheck(*enemyPieces, enemyKing, *currKing) {
		pos.UnmakeMove(undo)
		return backtrackInfo{}
	}

	return undo
}

func (pos *Position) AssertConsistency(prefix string) {
	for _, piece := range pos.blackPieces {
		pieceOnBoard := pos.board[piece]
		if pieceOnBoard&BlackPieceBit == 0 {
			panic(fmt.Sprintf("%v Piece on board should be black but was: %v", prefix, pieceOnBoard))
		}
	}
	for _, piece := range pos.whitePieces {
		pieceOnBoard := pos.board[piece]
		if pieceOnBoard&WhitePieceBit == 0 {
			panic(fmt.Sprintf("%v Piece on board should be white but was: %v", prefix, pieceOnBoard))
		}
	}
	for i, p := range pos.board {
		currSquare := square(i)
		if currSquare == InvalidSquare {
			continue
		}
		if p&BlackPieceBit != 0 {
			matchFound := false
			for _, sq := range pos.blackPieces {
				if sq == square(i) {
					matchFound = true
					break
				}
			}
			if !matchFound && pos.blackKing != currSquare {
				panic(fmt.Sprintf("%v Square %v has %v that's not on the black pieces list", prefix, currSquare, p))
			}

		} else if p&WhitePieceBit != 0 {
			matchFound := false
			for _, sq := range pos.whitePieces {
				if sq == square(i) {
					matchFound = true
					break
				}
			}
			if !matchFound && pos.whiteKing != currSquare {
				panic(fmt.Sprintf("%v Square %v has %v that's not on the white pieces list", prefix, currSquare, p))
			}
		}
	}
}

// Returns true if the destSquare is under check by anything on enemyPieces square or enemy king on enemyKing square.
func (pos *Position) isUnderCheck(enemyPieces []square, enemyKing square, destSquare square) bool {
	var moveIdx int16
	for _, attackFrom := range enemyPieces {
		moveIdx = moveIndex(attackFrom, destSquare)
		switch pos.board[attackFrom] {
		case WKnight, BKnight:
			if attackTable[moveIdx]&KnightAttacks == 0 {
				continue
			}
			return true
		case WBishop, BBishop:
			if attackTable[moveIdx]&BishopAttacks == 0 {
				continue
			}
			if pos.checkedBySlidingPiece(attackFrom, destSquare, moveIdx) {
				return true
			}
		case WRook, BRook:
			if attackTable[moveIdx]&RookAttacks == 0 {
				continue
			}
			if pos.checkedBySlidingPiece(attackFrom, destSquare, moveIdx) {
				return true
			}
		case WQueen, BQueen:
			if attackTable[moveIdx]&QueenAttacks == 0 {
				continue
			}
			if pos.checkedBySlidingPiece(attackFrom, destSquare, moveIdx) {
				return true
			}
		case WPawn:
			if attackTable[moveIdx]&WhitePawnAttacks == 0 {
				continue
			}
			return true
		case BPawn:
			if attackTable[moveIdx]&BlackPawnAttacks == 0 {
				continue
			}
			return true
		}
	}
	moveIdx = moveIndex(enemyKing, destSquare)
	kingAttack := attackTable[moveIdx]&KingAttacks != 0
	return kingAttack
}

// Counts checks on destSquare. The second returned value is location of the attacker (useful if theres only one attacker).
// This does not count checks by enemy king. Ignores absolute pins (includes checks by pinned pieces)
func (pos *Position) countPseudolegalChecksOn(attackingPieces []square, destSquare square) (numOfChecks int, checkerSquare square) {
	var moveIdx int16
	var checksCount int = 0
	checkerSquare = InvalidSquare

	for _, attackFrom := range attackingPieces {
		moveIdx = moveIndex(attackFrom, destSquare)
		switch pos.board[attackFrom] {
		case WKnight, BKnight:
			if attackTable[moveIdx]&KnightAttacks == 0 {
				continue
			}
			checksCount++
			checkerSquare = attackFrom
		case WBishop, BBishop:
			if attackTable[moveIdx]&BishopAttacks == 0 {
				continue
			}
			if pos.checkedBySlidingPiece(attackFrom, destSquare, moveIdx) {
				checksCount++
				checkerSquare = attackFrom
			}
		case WRook, BRook:
			if attackTable[moveIdx]&RookAttacks == 0 {
				continue
			}
			if pos.checkedBySlidingPiece(attackFrom, destSquare, moveIdx) {
				checksCount++
				checkerSquare = attackFrom
			}
		case WQueen, BQueen:
			if attackTable[moveIdx]&QueenAttacks == 0 {
				continue
			}
			if pos.checkedBySlidingPiece(attackFrom, destSquare, moveIdx) {
				checksCount++
				checkerSquare = attackFrom
			}
		case WPawn:
			if attackTable[moveIdx]&WhitePawnAttacks == 0 {
				continue
			}
			checksCount++
			checkerSquare = attackFrom
		case BPawn:
			if attackTable[moveIdx]&BlackPawnAttacks == 0 {
				continue
			}
			checksCount++
			checkerSquare = attackFrom
		}
	}
	return checksCount, checkerSquare
}

// Counts legal checks on a destSquare by checkingPieces. Considers absolute pins to king by pinningPieces.
// pushesNotCaptures - is set to true when generating interpositions rather than checks (in such case we generate pawn pushes; not captures)
func (pos *Position) countLegalChecksOn(checkingPieces []square, pinnerColorBit piece, destSquare square, king square, pushesNotCaptures bool) int {
	var checksCount int = 0

	var pinnedPieces uint16 = pos.findIndexesOfPinnedPieces(checkingPieces, pinnerColorBit, king)
	for _, attackFrom := range checkingPieces {
		checksCount += pos.countLegalChecksByPiece(attackFrom, destSquare, pinnedPieces&1 != 0, pushesNotCaptures)
		// var moveIdx = moveIndex(attackFrom, destSquare)
		// switch pos.board[attackFrom] {
		// case WKnight, BKnight:
		// 	if attackTable[moveIdx]&KnightAttacks == 0 {
		// 		pinnedPieces >>= 1
		// 		continue
		// 	}
		// 	if pinnedPieces&1 == 0 {
		// 		checksCount++
		// 	}
		// case WBishop, BBishop:
		// 	if attackTable[moveIdx]&BishopAttacks == 0 || pinnedPieces&1 != 0 {
		// 		pinnedPieces >>= 1
		// 		continue
		// 	}
		// 	if pos.checkedBySlidingPiece(attackFrom, destSquare, moveIdx) {
		// 		checksCount++
		// 	}
		// case WRook, BRook:
		// 	if attackTable[moveIdx]&RookAttacks == 0 || pinnedPieces&1 != 0 {
		// 		pinnedPieces >>= 1
		// 		continue
		// 	}
		// 	if pos.checkedBySlidingPiece(attackFrom, destSquare, moveIdx) {
		// 		checksCount++
		// 	}
		// case WQueen, BQueen:
		// 	if attackTable[moveIdx]&QueenAttacks == 0 || pinnedPieces&1 != 0 {
		// 		pinnedPieces >>= 1
		// 		continue
		// 	}
		// 	if pos.checkedBySlidingPiece(attackFrom, destSquare, moveIdx) {
		// 		checksCount++
		// 	}
		// case WPawn:
		// 	if pushesNotCaptures {
		// 		//on same file
		// 		if attackFrom.getFile() == destSquare.getFile() && (
		// 		//within push
		// 		attackFrom.getRank()+UnitRank == destSquare.getRank() ||
		// 			//within double push and we're on starting rank
		// 			attackFrom.getRank()+2*UnitRank == destSquare.getRank() && attackFrom.getRank() == Rank2) {
		// 			if pinnedPieces&1 == 0 {
		// 				if destSquare.getRank() == Rank8 {
		// 					checksCount += 4
		// 				} else {
		// 					checksCount++
		// 				}
		// 			}
		// 		}
		// 	} else {
		// 		if attackTable[moveIdx]&WhitePawnAttacks == 0 {
		// 			pinnedPieces >>= 1
		// 			continue
		// 		}
		// 		if pinnedPieces&1 == 0 {
		// 			if destSquare.getRank() == Rank8 {
		// 				checksCount += 4
		// 			} else {
		// 				checksCount++
		// 			}
		// 		}
		// 	}
		// case BPawn:
		// 	if pushesNotCaptures {
		// 		//on same file
		// 		if attackFrom.getFile() == destSquare.getFile() && (
		// 		//within push
		// 		attackFrom.getRank()-UnitRank == destSquare.getRank() ||
		// 			//within double push and we're on starting rank
		// 			attackFrom.getRank()-2*UnitRank == destSquare.getRank() && attackFrom.getRank() == Rank7) {
		// 			if pinnedPieces&1 == 0 {
		// 				if destSquare.getRank() == Rank1 {
		// 					checksCount += 4
		// 				} else {
		// 					checksCount++
		// 				}
		// 			}
		// 		}
		// 	} else {
		// 		if attackTable[moveIdx]&BlackPawnAttacks == 0 {
		// 			pinnedPieces >>= 1
		// 			continue
		// 		}
		// 		if pinnedPieces&1 == 0 {
		// 			if destSquare.getRank() == Rank1 {
		// 				checksCount += 4
		// 			} else {
		// 				checksCount++
		// 			}
		// 		}
		// 	}
		// }
		pinnedPieces >>= 1
	}
	return checksCount
}

func (pos *Position) countLegalChecksByPiece(attackFrom, destSquare square, pinned, pushesNotCaptures bool) int {
	var moveIdx = moveIndex(attackFrom, destSquare)
	switch pos.board[attackFrom] {
	case WKnight, BKnight:
		if attackTable[moveIdx]&KnightAttacks == 0 {
			return 0
		}
		if !pinned {
			return 1
		}
	case WBishop, BBishop:
		if attackTable[moveIdx]&BishopAttacks == 0 || pinned {
			return 0
		}
		if pos.checkedBySlidingPiece(attackFrom, destSquare, moveIdx) {
			return 1
		}
	case WRook, BRook:
		if attackTable[moveIdx]&RookAttacks == 0 || pinned {
			return 0
		}
		if pos.checkedBySlidingPiece(attackFrom, destSquare, moveIdx) {
			return 1
		}
	case WQueen, BQueen:
		if attackTable[moveIdx]&QueenAttacks == 0 || pinned {
			return 0
		}
		if pos.checkedBySlidingPiece(attackFrom, destSquare, moveIdx) {
			return 1
		}
	case WPawn:
		if pushesNotCaptures {
			//on same file
			if attackFrom.getFile() == destSquare.getFile() && (
			//within push
			attackFrom.getRank()+UnitRank == destSquare.getRank() ||
				//within double push and we're on starting rank
				attackFrom.getRank()+2*UnitRank == destSquare.getRank() && attackFrom.getRank() == Rank2 &&
				//and double push not obstructed
				 pos.board[attackFrom+square(UnitRank)] == NullPiece) {
				if !pinned {
					if destSquare.getRank() == Rank8 {
						return 4
					} else {
						return 1
					}
				}
			}
		} else {
			if attackTable[moveIdx]&WhitePawnAttacks == 0 {
				return 0
			}
			if !pinned {
				if destSquare.getRank() == Rank8 {
					return 4
				} else {
					return 1
				}
			}
		}
	case BPawn:
		if pushesNotCaptures {
			//on same file
			if attackFrom.getFile() == destSquare.getFile() && (
			//within push
			attackFrom.getRank()-UnitRank == destSquare.getRank() ||
				//within double push and we're on starting rank
				attackFrom.getRank()-2*UnitRank == destSquare.getRank() && attackFrom.getRank() == Rank7 &&
				//and double push not obstructed
				pos.board[attackFrom-square(UnitRank)] == NullPiece) {
				if !pinned {
					if destSquare.getRank() == Rank1 {
						return 4
					} else {
						return 1
					}
				}
			}
		} else {
			if attackTable[moveIdx]&BlackPawnAttacks == 0 {
				return 0
			}
			if !pinned {
				if destSquare.getRank() == Rank1 {
					return 4
				} else {
					return 1
				}
			}
		}
	}
	return 0
}

// Returns indexes into pinnedPieces[] square slice that contained absolutely pinned pieces.
// Could return them as a slice of ints but since we know that that there are never more
// than 16 pieces per side we can just use bit positions in int16. LSB means i==0;
func (pos *Position) findIndexesOfPinnedPieces(pinnedPieces []square, pinnerColorBit piece, king square) uint16 {
	var indexesOfPinnedPieces uint16 = 0
	for i, maybePinned := range pinnedPieces {
		if pos.isAbsolutelyPinned(maybePinned, pinnerColorBit, king) {
			indexesOfPinnedPieces |= 1 << i
			continue
		}
	}
	return indexesOfPinnedPieces
}

func (pos *Position) isAbsolutelyPinned(maybePinned square, pinnerColorBit piece, king square) bool {
	moveIdxToKing := moveIndex(maybePinned, king)
	directionToKing := directionTable[moveIdxToKing]
	// has king on one side
	if directionToKing == 0 || !pos.checkedAlongRay(maybePinned, king, directionToKing) {
		return false
	}
	oppositeDirection := -directionToKing
	// has relevant enemy sliding piece on other side
	for sq := maybePinned + square(oppositeDirection); sq&InvalidSquare == 0; sq += square(oppositeDirection) {
		somePiece := pos.board[sq]
		if somePiece == NullPiece {
			continue
		}
		if somePiece&pinnerColorBit == NullPiece {
			return false
		}
		switch somePiece & ColorlessPiece {
		case Bishop:
			return contains(bishopDirections, oppositeDirection)
		case Rook:
			return contains(rookDirections, oppositeDirection)
		case Queen:
			return contains(kingDirections, oppositeDirection)
		default:
			return false
		}

	}
	return false
}

func contains(directions []Direction, testedDiretion Direction) bool {
	for _, dir := range directions {
		if dir == testedDiretion {
			return true
		}
	}
	return false
}

func (pos *Position) checkedBySlidingPiece(slidingPieceSquare, destSquare square, moveIndex int16) bool {
	direction := directionTable[moveIndex]
	return pos.checkedAlongRay(slidingPieceSquare, destSquare, direction)
}

func (pos *Position) checkedAlongRay(slidingPieceSquare, destSquare square, direction Direction) bool {
	for sq := slidingPieceSquare + square(direction); sq != destSquare; sq += square(direction) {
		if pos.board[sq] != NullPiece {
			return false
		}
	}
	return true
}

func killPiece(enemyPieces []square, killSquare square) []square {
	for i := range enemyPieces {
		if enemyPieces[i] == killSquare {
			enemyPieces[i] = enemyPieces[len(enemyPieces)-1]
			return enemyPieces[:len(enemyPieces)-1]
		}
	}
	panic(fmt.Sprintf("Didn't find square: %v on enemyPieces: %v", killSquare, enemyPieces))
}

func (pos *Position) UnmakeMove(undo backtrackInfo) {
	unmadePieces, unkilledPieces, unmadeKing, unmadeColorBit,
		castleRank, enPassantUnkillRank := pos.getUnmakeMoveContext()

	mov := undo.move
	for i := range unmadePieces {
		if mov.to == unmadePieces[i] {
			unmadePieces[i] = mov.from
		}
	}
	if mov.to == *unmadeKing {
		*unmadeKing = mov.from
		if mov.from.getFile() == E {
			if mov.to.getFile() == C {
				rookFrom := square(A + file(castleRank))
				rookTo := square(D + file(castleRank))
				// just call the castleRook with To/From squares swapped
				pos.moveRook(rookTo, rookFrom, unmadePieces, unmadeColorBit)
			} else if mov.to.getFile() == G {
				rookFrom := square(H + file(castleRank))
				rookTo := square(F + file(castleRank))
				pos.moveRook(rookTo, rookFrom, unmadePieces, unmadeColorBit)
			}
		}
	}

	if mov.promoteTo == NullPiece {
		pos.board[mov.from] = pos.board[mov.to]
	} else {
		pos.board[mov.from] = Pawn | unmadeColorBit
	}
	pos.board[mov.to] = NullPiece

	if undo.takenPiece != NullPiece {
		var killSquare square
		//mov was an en passant take
		if undo.lastEnPassant == mov.to && pos.board[mov.from] == Pawn|unmadeColorBit {
			killSquare = square(mov.to.getFile()) + square(enPassantUnkillRank)
		} else {
			killSquare = mov.to
		}
		*unkilledPieces = append(*unkilledPieces, killSquare)
		pos.board[killSquare] = undo.takenPiece
	}
	pos.enPassSquare = undo.lastEnPassant
	pos.flags = undo.lastFlags
}

// used to castle/undo castle
func (pos *Position) moveRook(rookFrom, rookTo square, pieces []square, colorBit piece) {
	for i := range pieces {
		if pieces[i] == rookFrom {
			pieces[i] = rookTo
			break
		}
	}
	pos.board[rookFrom] = NullPiece
	pos.board[rookTo] = Rook | colorBit
}

func (pos *Position) getCurrentMakeMoveContext() (
	currPieces []square,
	enemyPieces *[]square,
	currKing *square,
	enemyKing square,
	currCastleRank rank, currKingSideCastleFlag, currQueenSideCastleFlag byte,
	enemyCastleRank rank, enemyKingSideCastleFlag, enemyQueenSideCastleFlag byte,
	currColorBit piece,
) {
	if pos.flags&FlagWhiteTurn == 0 {
		return pos.blackPieces, &pos.whitePieces,
			&pos.blackKing, pos.whiteKing,
			Rank8, FlagBlackCanCastleKside, FlagBlackCanCastleQside,
			Rank1, FlagWhiteCanCastleKside, FlagWhiteCanCastleQside,
			BlackPieceBit
	}
	return pos.whitePieces, &pos.blackPieces,
		&pos.whiteKing, pos.blackKing,
		Rank1, FlagWhiteCanCastleKside, FlagWhiteCanCastleQside,
		Rank8, FlagBlackCanCastleKside, FlagBlackCanCastleQside,
		WhitePieceBit
}

// inverse of GetCurrentMakeMoveContext()
func (pos *Position) getUnmakeMoveContext() (
	unmadePieces []square,
	unkilledPieces *[]square,
	unmadeKing *square,
	unmadeColorBit piece,
	castleRank rank,
	enPassantUnkillRank rank,
) {
	if pos.flags&FlagWhiteTurn != 0 {
		return pos.blackPieces, &pos.whitePieces, &pos.blackKing, BlackPieceBit, Rank8, Rank4
	}
	return pos.whitePieces, &pos.blackPieces, &pos.whiteKing, WhitePieceBit, Rank1, Rank5
}

func (pos *Position) GetAtSquare(s square) piece {
	return pos.board[s]
}

func (pos *Position) GetAtFileRank(f file, r rank) piece {
	// cast to file is kindof dodgy but it must be faster than two casts to byte, right?
	var index file = f + file(r)
	return pos.board[index]
}
