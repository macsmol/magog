package engine

import (
	"fmt"
	"strings"
)

// max capacities of piece lists
const (
	// pawns - equal to no of pawns in start position
	pawnCap = 8
	// other pieces - equal to no of pieces in start position + upper bound of promotions
	pieceCap = 8 + 7
)

type pawnList struct {
	squares [pawnCap]square
	size    int8
}

func (list pawnList) String() string {
	var sb strings.Builder
	sb.WriteRune('[')
	for i, sq := range list.squares {
		if i == int(list.size) {
			sb.WriteString(" / ")
		}
		sb.WriteString(fmt.Sprintf("%v ",sq))
	}
	sb.WriteRune(']')
	return sb.String()
}

// everything but pawns and king
type pieceList struct {
	squares [pieceCap]square
	size    int8
}

func (list pieceList) String() string {
	var sb strings.Builder
	sb.WriteRune('[')
	for i, sq := range list.squares {
		if i == int(list.size) {
			sb.WriteString(" / ")
		}
		sb.WriteString(fmt.Sprintf("%v ",sq))
	}
	sb.WriteRune(']')
	return sb.String()
}

// Struct representing current state of the game.
// Implementation note: No slices are used here. I want this struct to be a contigous block of memory.
// This way MakeMove() is just a copy+modify of previous position and pushing it on a stack and unmakeMove.
// You don't need any unmakeMove then (just stack pop -> stackTop--).
type Position struct {
	// 0x88 board
	board        [128]piece
	blackPieces  pieceList
	whitePieces  pieceList
	blackPawns   pawnList
	whitePawns   pawnList
	blackKing    square
	whiteKing    square
	flags        byte
	enPassSquare square
	// zero based halfmove counter
	ply			 int16
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
func NewPosition() Position {
	// &Position{}  - shorthand for new Position on heap + return a pointer to it
	z := InvalidSquare
	return Position{
		board: [128]piece{
			// FFS how do you turn off whitespace formatting in VSCode?
			A1: WRook, B1: WKnight, C1: WBishop, D1: WQueen, E1: WKing, F1: WBishop, G1: WKnight, H1: WRook,
			A2: WPawn, B2: WPawn, C2: WPawn, D2: WPawn, E2: WPawn, F2: WPawn, G2: WPawn, H2: WPawn,
			A7: BPawn, B7: BPawn, C7: BPawn, D7: BPawn, E7: BPawn, F7: BPawn, G7: BPawn, H7: BPawn,
			A8: BRook, B8: BKnight, C8: BBishop, D8: BQueen, E8: BKing, F8: BBishop, G8: BKnight, H8: BRook,
		},
		blackPieces: pieceList{[...]square{A8, B8, C8, D8, F8, G8, H8, z, z, z, z, z, z, z, z}, 7},
		blackPawns:  pawnList{[...]square{A7, B7, C7, D7, E7, F7, G7, H7}, 8},
		blackKing:   E8,
		whitePieces: pieceList{[...]square{A1, B1, C1, D1, F1, G1, H1, z, z, z, z, z, z, z, z}, 7},
		whitePawns:  pawnList{[...]square{A2, B2, C2, D2, E2, F2, G2, H2}, 8},
		whiteKing:   E1,
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
	}
	appendFlagsString(&sb,
		pos.flags&FlagBlackCanCastleQside != 0,
		pos.flags&FlagBlackCanCastleKside != 0,
		pos.flags&FlagWhiteTurn == 0)
	sb.WriteString(fmt.Sprintf("BlackKing: %v; BlackPieces: %v; BlackPawns: %v\n", pos.blackKing, pos.blackPieces, pos.blackPawns))
	appendFlagsString(&sb,
		pos.flags&FlagWhiteCanCastleQside != 0,
		pos.flags&FlagWhiteCanCastleKside != 0,
		pos.flags&FlagWhiteTurn != 0)
	sb.WriteString(fmt.Sprintf("WhiteKing: %v; WhitePieces: %v; WhitePawns: %v\n", pos.whiteKing, pos.whitePieces, pos.whitePawns))
	sb.WriteString(fmt.Sprintf("En passant square: %v; ply: %d", pos.enPassSquare, pos.ply))
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

// TODO return a pointer? With the array this is a rather fat param on the stack.
// Or maybe not - every slice is 24 bytes (excluding memory block)
func (pos *Position) GetCurrentContext() (
	currPieces, enemyPieces pieceList,
	currPawns, enemyPawns pawnList,
	currKing square, enemyKing square,
	pawnAdvance Direction,
	currColorBit piece, enemyColorBit piece,
	queensideCastlePossible, kingsideCastlePossible bool,
	currPawnsStartRank, promotionRank rank) {
	if pos.flags&FlagWhiteTurn == 0 {
		return pos.blackPieces, pos.whitePieces,
			pos.blackPawns, pos.whitePawns,
			pos.blackKing, pos.whiteKing,
			DirS, BlackPieceBit, WhitePieceBit,
			pos.flags&FlagBlackCanCastleQside != 0, pos.flags&FlagBlackCanCastleKside != 0,
			Rank7, Rank1
	}
	return pos.whitePieces, pos.blackPieces,
		pos.whitePawns, pos.blackPawns,
		pos.whiteKing, pos.blackKing,
		DirN, WhitePieceBit, BlackPieceBit,
		pos.flags&FlagWhiteCanCastleQside != 0, pos.flags&FlagWhiteCanCastleKside != 0,
		Rank2, Rank8
}

func (pos *Position) GetCurrentTacticalMoveContext() (
	currPieces, enemyPieces pieceList,
	currPawns, enemyPawns pawnList,
	currKing square, enemyKing square,
	pawnAdvance Direction,
	currColorBit piece, enemyColorBit piece,
	promotionRank rank) {
	if pos.flags&FlagWhiteTurn == 0 {
		return pos.blackPieces, pos.whitePieces,
			pos.blackPawns, pos.whitePawns,
			pos.blackKing, pos.whiteKing,
			DirS, BlackPieceBit, WhitePieceBit,
			Rank1
	}
	return pos.whitePieces, pos.blackPieces,
		pos.whitePawns, pos.blackPawns,
		pos.whiteKing, pos.blackKing,
		DirN, WhitePieceBit, BlackPieceBit,
		Rank8
}

func (pos *Position) MakeMove(mov Move) (isLegal bool) {
	currPieces, currPawnsPtr, currKingSq,
		enemyPieces, enemyPawns, enemyKingSq,
		currCastleRank, currKingSideCastleFlag, currQueenSideCastleFlag,
		enemyCastleRank, enemyKingSideCastleFlag, enemyQueenSideCastleFlag,
		currColorBit, enemyColorBit := pos.getCurrentMakeMoveContext()
	// pos.AssertConsistency("make" + mov.String())
	pos.ply++

	// one of thre possibilities - pawn move, king move, other piece move
	if pos.board[mov.from] == Pawn|currColorBit {
		// normal move - just update entry
		if mov.promoteTo == NullPiece {
			for i := int8(0); i < currPawnsPtr.size; i++ {
				if mov.from == currPawnsPtr.squares[i] {
					currPawnsPtr.squares[i] = mov.to
					break
				}
			}
		} else {
			// promotion move - remove from currPawns and add to currPieces
			for i := int8(0); i < currPawnsPtr.size; i++ {
				if mov.from == currPawnsPtr.squares[i] {
					currPawnsPtr.remove(i)
					currPieces.appendPiece(mov.to)
					break
				}
			}
		}
	} else if mov.from == *currKingSq {
		pos.flags &= ^(currKingSideCastleFlag | currQueenSideCastleFlag)
		*currKingSq = mov.to
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
	} else {
		for i := int8(0); i < currPieces.size; i++ {
			if mov.from == currPieces.squares[i] {
				currPieces.squares[i] = mov.to
				break
			}
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

	if pos.board[mov.to] != NullPiece {
		// when calculating enemy mobility it is possible to kill enemy king.
		// Kings are not on piece lists so we don't modify piece lists in MakeMove() and UnmakeMove()
		if pos.board[mov.to] != King|enemyColorBit {
			if pos.board[mov.to] == Pawn|enemyColorBit {
				killPawn(enemyPawns, mov.to, pos)
			} else {
				killPiece(enemyPieces, mov.to)
			}
		}
	}
	if mov.promoteTo == NullPiece {
		pos.board[mov.to] = pos.board[mov.from]
		//en passant take
		if pos.enPassSquare == mov.to && pos.board[mov.from] == Pawn|currColorBit {
			killSquare := square(mov.to.getFile() + file(mov.from.getRank()))
			killPawn(enemyPawns, killSquare, nil)
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
	return !pos.isUnderCheck(*enemyPieces, *enemyPawns, enemyKingSq, *currKingSq)
}

// // like bubbleSort but moves only one entry up to a correct place in an otherwise sorted list
// func bubbleUp(i int, pawns []square) {
// 	for ; i < len(pawns)-1; i++ {
// 		if pawns[i] <= pawns[i+1] {
// 			break
// 		}
// 		pawns[i], pawns[i+1] = pawns[i+1], pawns[i]
// 	}
// }

// // like bubbleSort but moves only one entry down to a correct place in an otherwise sorted list
// func bubbleDown(i int, pawns []square) {
// 	for ; i > 0; i-- {
// 		if pawns[i-1] <= pawns[i] {
// 			break
// 		}
// 		pawns[i], pawns[i-1] = pawns[i-1], pawns[i]
// 	}
// }

// removes element from the slice while preserving order. We want to keep pawn lists ordered
// func removeOrdered(pieceList []square, idxToRemove int) []square {
// 	return append(pieceList[:idxToRemove], pieceList[idxToRemove+1:]...)
// }

func (pawnList *pawnList) appendPawn(pawnSq square) {
	pawnList.squares[pawnList.size] = pawnSq
	pawnList.size++
}

// removes pawn without preserving order.
func (pawnList *pawnList) remove(idxToRemove int8) {
	pawnList.squares[idxToRemove] = pawnList.squares[pawnList.size-1]
	pawnList.size--
}

func (pieceList *pieceList) appendPiece(sq square) {
	pieceList.squares[pieceList.size] = sq
	pieceList.size++
}

func (pos *Position) AssertConsistency(prefix string) {
	// piece lists to board
	for i := int8(0); i < pos.blackPieces.size; i++ {
		pieceSquare := pos.blackPieces.squares[i]
		pieceOnBoard := pos.board[pieceSquare]
		if pieceOnBoard&BlackPieceBit == 0 {
			panic(fmt.Sprintf("%v Piece on board should be black but was: %v", prefix, pieceOnBoard))
		}
	}
	for i := int8(0); i < pos.blackPawns.size; i++ {
		pawnSquare := pos.blackPawns.squares[i]
		pawnOnBoard := pos.board[pawnSquare]
		if pawnOnBoard != BPawn {
			panic(fmt.Sprintf("%v Black pawn should be on board but was: %v", prefix, pawnOnBoard))
		}
	}
	for i := int8(0); i < pos.whitePieces.size; i++ {
		pieceSquare := pos.whitePieces.squares[i]
		pieceOnBoard := pos.board[pieceSquare]
		if pieceOnBoard&WhitePieceBit == 0 {
			panic(fmt.Sprintf("%v Piece on board should be white but was: %v", prefix, pieceOnBoard))
		}
	}
	for i := int8(0); i < pos.whitePawns.size; i++ {
		pawnSquare := pos.whitePawns.squares[i]
		pawnOnBoard := pos.board[pawnSquare]
		if pawnOnBoard != WPawn {
			panic(fmt.Sprintf("%v White pawn should be on board but was: %v", prefix, pawnOnBoard))
		}
	}
	// board to piece lists
	for i, piece := range pos.board {
		sqOnBoard := square(i)
		if sqOnBoard == InvalidSquare {
			continue
		}
		if piece&BlackPieceBit != 0 {
			matchFound := false
			for i := int8(0); i < pos.blackPieces.size; i++ {
				if pos.blackPieces.squares[i] == sqOnBoard {
					matchFound = true
					break
				}
			}
			for i := int8(0); i < pos.blackPawns.size; i++ {
				if pos.blackPawns.squares[i] == sqOnBoard {
					if matchFound {
						panic(fmt.Sprintf("%v Square %v appears on both blackPieces and blackPawns list. It has %v on board", prefix, sqOnBoard, piece))
					}
					matchFound = true
					break
				}
			}
			if !matchFound && pos.blackKing != sqOnBoard {
				panic(fmt.Sprintf("%v Square %v has %v that's not on the black pieces list", prefix, sqOnBoard, piece))
			}

		} else if piece&WhitePieceBit != 0 {
			matchFound := false
			for i := int8(0); i < pos.whitePieces.size; i++ {
				if pos.whitePieces.squares[i] == sqOnBoard {
					matchFound = true
					break
				}
			}
			for i := int8(0); i < pos.whitePawns.size; i++ {
				if pos.whitePawns.squares[i] == sqOnBoard {
					if matchFound {
						panic(fmt.Sprintf("%v Square %v appears on both whitePieces and whitePawns list. It has %v on board", prefix, sqOnBoard, piece))
					}
					matchFound = true
					break
				}
			}
			if !matchFound && pos.whiteKing != sqOnBoard {
				panic(fmt.Sprintf("%v Square %v has %v that's not on the white pieces list", prefix, sqOnBoard, piece))
			}
		}
	}
	// TODO verify that pawn list sorted
}

func (pos *Position) isCurrentKingUnderCheck() bool {
	var currentKing, enemyKing square
	var enemyPieces pieceList
	var enemyPawns pawnList
	if pos.flags&FlagWhiteTurn == 0 {
		currentKing = pos.blackKing
		enemyKing = pos.whiteKing
		enemyPieces = pos.whitePieces
		enemyPawns = pos.whitePawns
	} else {
		currentKing = pos.whiteKing
		enemyKing = pos.blackKing
		enemyPieces = pos.blackPieces
		enemyPawns = pos.blackPawns
	}
	return pos.isUnderCheck(enemyPieces, enemyPawns, enemyKing, currentKing)
}

// Returns true if the destSquare is under check by anything on enemyPieces square or enemy king on
// enemyKing square.
func (pos *Position) isUnderCheck(enemyPieces pieceList, enemyPawns pawnList, enemyKingSq square, destSquare square) bool {
	var moveIdx int16
	var PawnAttackFlag byte
	if pos.board[enemyKingSq]&BlackPieceBit == 0 {
		PawnAttackFlag = WPawnAttacks
	} else {
		PawnAttackFlag = BPawnAttacks
	}

	for i := int8(0); i < enemyPawns.size; i++ {
		attackFrom := enemyPawns.squares[i]
		moveIdx = moveIndex(attackFrom, destSquare)
		if attackTable[moveIdx]&PawnAttackFlag != 0 {
			return true
		}
	}

	for i := int8(0); i < enemyPieces.size; i++ {
		attackFrom := enemyPieces.squares[i]
		moveIdx = moveIndex(attackFrom, destSquare)
		attacker := pos.board[attackFrom] & ColorlessPiece

		if attackTable[moveIdx]&byte(attacker) == 0 {
			continue
		}
		if attacker&Knight != 0 {
			return true
		}
		if pos.checkedBySlidingPiece(attackFrom, destSquare, moveIdx) {
			return true
		}
	}
	moveIdx = moveIndex(enemyKingSq, destSquare)
	kingAttack := attackTable[moveIdx]&KingAttacks != 0
	return kingAttack
}

func (pos *Position) checkedBySlidingPiece(slidingPieceSquare, destSquare square, moveIndex int16) bool {
	direction := directionTable[moveIndex]
	for sq := slidingPieceSquare + square(direction); sq != destSquare; sq += square(direction) {
		if pos.board[sq] != NullPiece {
			return false
		}
	}
	return true
}

func killPiece(enemyPieces *pieceList, killSquare square) {
	for i := int8(0); i < enemyPieces.size; i++ {
		if enemyPieces.squares[i] == killSquare {
			enemyPieces.squares[i] = enemyPieces.squares[enemyPieces.size-1]
			enemyPieces.size--
			return
		}
	}
	panic(fmt.Sprintf("Didn't find square: %v on enemyPieces: %v", killSquare, enemyPieces))
}

func killPawn(enemyPawns *pawnList, killSquare square, debugPos *Position) {
	for i := int8(0); i < enemyPawns.size; i++ {
		if enemyPawns.squares[i] == killSquare {
			enemyPawns.squares[i] = enemyPawns.squares[enemyPawns.size-1]
			enemyPawns.size--
			return
		}
	}
	panic(fmt.Sprintf("Didn't find square: %v on enemyPieces: %v in position: %v", killSquare, enemyPawns, debugPos))
}

// func killPieceOrdered(pieceList []square, killSquare square) []square {
// 	for i := range pieceList {
// 		if killSquare == pieceList[i] {
// 			return removeOrdered(pieceList, i)
// 		}
// 	}
// 	panic(fmt.Sprintf("Didn't find a piece to kill on: %v", killSquare))
// }

// used to castle/undo castle
func (pos *Position) moveRook(rookFrom, rookTo square, pieces *pieceList, colorBit piece) {
	for i := int8(0); i < pieces.size; i++ {
		if pieces.squares[i] == rookFrom {
			pieces.squares[i] = rookTo
			break
		}
	}
	pos.board[rookFrom] = NullPiece
	pos.board[rookTo] = Rook | colorBit
}

func (pos *Position) getCurrentMakeMoveContext() (
	currPieces *pieceList, currPawns *pawnList, currKing *square,
	enemyPieces *pieceList, enemyPawns *pawnList, enemyKing square,
	currCastleRank rank, currKingSideCastleFlag, currQueenSideCastleFlag byte,
	enemyCastleRank rank, enemyKingSideCastleFlag, enemyQueenSideCastleFlag byte,
	currColorBit, enemyColorBit piece,
) {
	if pos.flags&FlagWhiteTurn == 0 {
		return &pos.blackPieces, &pos.blackPawns, &pos.blackKing,
			&pos.whitePieces, &pos.whitePawns, pos.whiteKing,
			Rank8, FlagBlackCanCastleKside, FlagBlackCanCastleQside,
			Rank1, FlagWhiteCanCastleKside, FlagWhiteCanCastleQside,
			BlackPieceBit, WhitePieceBit
	}
	return &pos.whitePieces, &pos.whitePawns, &pos.whiteKing,
		&pos.blackPieces, &pos.blackPawns, pos.blackKing,
		Rank1, FlagWhiteCanCastleKside, FlagWhiteCanCastleQside,
		Rank8, FlagBlackCanCastleKside, FlagBlackCanCastleQside,
		WhitePieceBit, BlackPieceBit
}

func (pos *Position) GetAtSquare(s square) piece {
	return pos.board[s]
}

func (pos *Position) GetAtFileRank(f file, r rank) piece {
	// cast to file is kindof dodgy but it must be faster than two casts to byte, right?
	var index file = f + file(r)
	return pos.board[index]
}
