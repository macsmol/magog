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
	blackPawns  []square
	blackKing   square
	// all but king
	whitePieces  []square
	whitePawns   []square
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
		blackPieces: []square{A8, B8, C8, D8, F8, G8, H8},
		blackPawns:  []square{A7, B7, C7, D7, E7, F7, G7, H7},
		blackKing:   E8,
		whitePieces: []square{A1, B1, C1, D1, F1, G1, H1},
		whitePawns:  []square{A2, B2, C2, D2, E2, F2, G2, H2},
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
	currPawns []square, enemyPawns []square,
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
	currPieces []square, enemyPieces []square,
	currPawns []square, enemyPawns []square,
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

// MakeMove applies mov to a position pos. Returns a backtrackInfo that can be used to revert pos back
// to it's original state. In case where applying a mov would result in an illegal Position (i.e. capturing
// a king is possible), pos is left unchanged and backtrackInfo returned is all zeroes.
// Probably will crash when take a move that is either:
// -not possible in this position.
// -not possible according to the rules of chess: a1b8
func (pos *Position) MakeMove(mov Move) (undo backtrackInfo) {
	currPieces, currPawnsPtr, currKing,
		enemyPieces, enemyPawns, enemyKing,
		currCastleRank, currKingSideCastleFlag, currQueenSideCastleFlag,
		enemyCastleRank, enemyKingSideCastleFlag, enemyQueenSideCastleFlag,
		currColorBit, enemyColorBit,
		bubbleFunc := pos.getCurrentMakeMoveContext()
	// pos.AssertConsistency(mov.String())
	undo = backtrackInfo{
		move:          mov,
		lastFlags:     pos.flags,
		lastEnPassant: pos.enPassSquare,
	}
	// one of thre possibilities - pawn move, king move, other piece move
	if pos.board[mov.from] == Pawn|currColorBit {
		currPawns := (*currPawnsPtr)
		// normal move - just update entry
		if mov.promoteTo == NullPiece {
			// it's sorted but binary search slower for len(list) < 8
			for i := range currPawns {
				if mov.from == currPawns[i] {
					currPawns[i] = mov.to
					// keep sorted
					bubbleFunc(i, currPawns)
					break
				}
			}
		} else {
			// promotion move - remove from currPawns and add to currPieces
			for i := range currPawns {
				if mov.from == currPawns[i] {
					(*currPawnsPtr) = removeOrdered(currPawns, i)
					*currPieces = append(*currPieces, mov.to)
					break
				}
			}
		}
	} else if mov.from == *currKing {
		pos.flags &= ^(currKingSideCastleFlag | currQueenSideCastleFlag)
		*currKing = mov.to
		if mov.from.getFile() == E {
			if mov.to.getFile() == C {
				rookFrom := square(A + file(currCastleRank))
				rookTo := square(D + file(currCastleRank))
				pos.moveRook(rookFrom, rookTo, *currPieces, currColorBit)
			} else if mov.to.getFile() == G {
				rookFrom := square(H + file(currCastleRank))
				rookTo := square(F + file(currCastleRank))
				pos.moveRook(rookFrom, rookTo, *currPieces, currColorBit)
			}
		}
	} else {
		for i := range *currPieces {
			if mov.from == (*currPieces)[i] {
				(*currPieces)[i] = mov.to
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
		undo.takenPiece = pos.board[mov.to]
		// when calculating enemy mobility it is possible to kill enemy king.
		// Kings are not on piece lists so we don't modify piece lists in MakeMove() and UnmakeMove()
		if pos.board[mov.to] != King|enemyColorBit {
			if pos.board[mov.to] == Pawn|enemyColorBit {
				*enemyPawns = killPieceOrdered(*enemyPawns, mov.to)
			} else {
				*enemyPieces = killPiece(*enemyPieces, mov.to)
			}
		}
	}
	if mov.promoteTo == NullPiece {
		pos.board[mov.to] = pos.board[mov.from]
		//en passant take
		if pos.enPassSquare == mov.to && pos.board[mov.from] == Pawn|currColorBit {
			killSquare := square(mov.to.getFile() + file(mov.from.getRank()))
			*enemyPawns = killPieceOrdered(*enemyPawns, killSquare)
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
	if pos.isUnderCheck(*enemyPieces, *enemyPawns, enemyKing, *currKing) {
		pos.UnmakeMove(undo)
		return backtrackInfo{}
	}

	return undo
}

// like bubbleSort but moves only one entry up to a correct place in an otherwise sorted list
func bubbleUp(i int, pawns []square) {
	for ; i < len(pawns)-1; i++ {
		if pawns[i] <= pawns[i+1] {
			break
		}
		pawns[i], pawns[i+1] = pawns[i+1], pawns[i]
	}
}

// like bubbleSort but moves only one entry down to a correct place in an otherwise sorted list
func bubbleDown(i int, pawns []square) {
	for ; i > 0; i-- {
		if pawns[i-1] <= pawns[i] {
			break
		}
		pawns[i], pawns[i-1] = pawns[i-1], pawns[i]
	}
}

// removes element from the slice while preserving order. We want to keep pawn lists ordered
func removeOrdered(pieceList []square, idxToRemove int) []square {
	return append(pieceList[:idxToRemove], pieceList[idxToRemove+1:]...)
}

// removes element from the slice without preserving order. We dont care about piece lists ordering
func remove(pieceList []square, idxToRemove int) []square {
	pieceList[idxToRemove] = pieceList[len(pieceList)-1]
	return pieceList[:len(pieceList)-1]
}

func (pos *Position) AssertConsistency(prefix string) {
	// piece lists to board
	for _, pieceSquare := range pos.blackPieces {
		pieceOnBoard := pos.board[pieceSquare]
		if pieceOnBoard&BlackPieceBit == 0 {
			panic(fmt.Sprintf("%v Piece on board should be black but was: %v", prefix, pieceOnBoard))
		}
	}
	for _, pawnSquare := range pos.blackPawns {
		pawnOnBoard := pos.board[pawnSquare]
		if pawnOnBoard != BPawn {
			panic(fmt.Sprintf("%v Black pawn should be on board but was: %v", prefix, pawnOnBoard))
		}
	}
	for _, piece := range pos.whitePieces {
		pieceOnBoard := pos.board[piece]
		if pieceOnBoard&WhitePieceBit == 0 {
			panic(fmt.Sprintf("%v Piece on board should be white but was: %v", prefix, pieceOnBoard))
		}
	}
	for _, pawnSquare := range pos.whitePawns {
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
			for _, sq := range pos.blackPieces {
				if sq == sqOnBoard {
					matchFound = true
					break
				}
			}
			for _, sq := range pos.blackPawns {
				if sq == sqOnBoard {
					if matchFound {
						panic(fmt.Sprintf("%v Square %v appears on both whitePieces and whitePawns list. It has %v on board", prefix, sqOnBoard, piece))
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
			for _, sq := range pos.whitePieces {
				if sq == sqOnBoard {
					matchFound = true
					break
				}
			}
			for _, sq := range pos.whitePawns {
				if sq == sqOnBoard {
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
	var enemyPieces, enemyPawns []square
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
func (pos *Position) isUnderCheck(enemyPieces, enemyPawns []square, enemyKing square, destSquare square) bool {
	var moveIdx int16
//TODO
// -use the sorting

	for _, attackFrom := range enemyPawns {
		moveIdx = moveIndex(attackFrom, destSquare)
		if pos.board[attackFrom] == WPawn && attackTable[moveIdx]&WPawnAttacks != 0 {
			return true
		} else if pos.board[attackFrom] == BPawn && attackTable[moveIdx]&BPawnAttacks != 0 {
			return true
		}
	}

	for _, attackFrom := range enemyPieces {
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
	moveIdx = moveIndex(enemyKing, destSquare)
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

func killPiece(enemyPieces []square, killSquare square) []square {
	for i := range enemyPieces {
		if enemyPieces[i] == killSquare {
			enemyPieces[i] = enemyPieces[len(enemyPieces)-1]
			return enemyPieces[:len(enemyPieces)-1]
		}
	}
	panic(fmt.Sprintf("Didn't find square: %v on enemyPieces: %v", killSquare, enemyPieces))
}

func killPieceOrdered(pieceList []square, killSquare square) []square {
	for i := range pieceList {
		if killSquare == pieceList[i] {
			return removeOrdered(pieceList, i)
		}
	}
	panic(fmt.Sprintf("Didn't find a piece to kill on: %v", killSquare))
}

func (pos *Position) UnmakeMove(undo backtrackInfo) {
	unmadePieces, unmadePawnsPtr, unkilledPieces, unkilledPawns,
		unmadeKing, unmadeColorBit, castleRank, enPassantUnkillRank,
		bubbleFunc := pos.getUnmakeMoveContext()
	// pos.AssertConsistency("undo " + undo.move.String())
	mov := undo.move
	if mov.to == *unmadeKing {
		*unmadeKing = mov.from
		if mov.from.getFile() == E {
			if mov.to.getFile() == C {
				rookFrom := square(A + file(castleRank))
				rookTo := square(D + file(castleRank))
				// just like castling the rook with To/From squares swapped
				pos.moveRook(rookTo, rookFrom, *unmadePieces, unmadeColorBit)
			} else if mov.to.getFile() == G {
				rookFrom := square(H + file(castleRank))
				rookTo := square(F + file(castleRank))
				pos.moveRook(rookTo, rookFrom, *unmadePieces, unmadeColorBit)
			}
		}
	} else if mov.promoteTo != NullPiece {
		//remove from unmadePieces
		// iterate from the end because that's where the promos are appended
		for i := len(*unmadePieces) - 1; i >= 0; i-- {
			// for i := range unmadePieces {
			if mov.to == (*unmadePieces)[i] {
				*unmadePieces = remove(*unmadePieces, i)
				break
			}
		}
		*unmadePawnsPtr = append((*unmadePawnsPtr), mov.from)
		// keep sorted
		bubbleDown(len(*unmadePawnsPtr)-1, (*unmadePawnsPtr))
	} else if pos.board[mov.to] == Pawn|unmadeColorBit {
		for i := range *unmadePawnsPtr {
			if mov.to == (*unmadePawnsPtr)[i] {
				(*unmadePawnsPtr)[i] = mov.from
				bubbleFunc(i, (*unmadePawnsPtr))
				break
			}
		}
	} else {
		for i := range *unmadePieces {
			if mov.to == (*unmadePieces)[i] {
				(*unmadePieces)[i] = mov.from
				break
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
		var killSquare square = mov.to
		// when calculating enemy mobility it is possible to kill enemy king.
		// Kings are not on piece lists so we don't modify piece lists in MakeMove() and UnmakeMove()
		if undo.takenPiece&ColorlessPiece != King {
			//mov was an en passant take
			if undo.lastEnPassant == mov.to && pos.board[mov.from] == Pawn|unmadeColorBit {
				killSquare = square(mov.to.getFile()) + square(enPassantUnkillRank)
				*unkilledPawns = append(*unkilledPawns, killSquare)
				bubbleDown(len(*unkilledPawns)-1, *unkilledPawns)
			} else if undo.takenPiece&ColorlessPiece == Pawn {
				*unkilledPawns = append(*unkilledPawns, killSquare)
				bubbleDown(len(*unkilledPawns)-1, *unkilledPawns)
			} else {
				*unkilledPieces = append(*unkilledPieces, killSquare)
			}
		}
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
	currPieces *[]square, currPawns *[]square, currKing *square,
	enemyPieces *[]square, enemyPawns *[]square, enemyKing square,
	currCastleRank rank, currKingSideCastleFlag, currQueenSideCastleFlag byte,
	enemyCastleRank rank, enemyKingSideCastleFlag, enemyQueenSideCastleFlag byte,
	currColorBit, enemyColorBit piece,
	bubbleFunc func(int, []square),
) {
	if pos.flags&FlagWhiteTurn == 0 {
		return &pos.blackPieces, &pos.blackPawns, &pos.blackKing,
			&pos.whitePieces, &pos.whitePawns, pos.whiteKing,
			Rank8, FlagBlackCanCastleKside, FlagBlackCanCastleQside,
			Rank1, FlagWhiteCanCastleKside, FlagWhiteCanCastleQside,
			BlackPieceBit, WhitePieceBit,
			bubbleDown
	}
	return &pos.whitePieces, &pos.whitePawns, &pos.whiteKing,
		&pos.blackPieces, &pos.blackPawns, pos.blackKing,
		Rank1, FlagWhiteCanCastleKside, FlagWhiteCanCastleQside,
		Rank8, FlagBlackCanCastleKside, FlagBlackCanCastleQside,
		WhitePieceBit, BlackPieceBit,
		bubbleUp
}

// inverse of GetCurrentMakeMoveContext()
func (pos *Position) getUnmakeMoveContext() (
	unmadePieces, unmadePawns *[]square,
	unkilledPieces, unkilledPawns *[]square,
	unmadeKing *square,
	unmadeColorBit piece,
	castleRank rank,
	enPassantUnkillRank rank,
	bubbleFunc func(int, []square),
) {
	if pos.flags&FlagWhiteTurn != 0 {
		return &pos.blackPieces, &pos.blackPawns, &pos.whitePieces, &pos.whitePawns, &pos.blackKing,
			BlackPieceBit, Rank8, Rank4, bubbleUp
	}
	return &pos.whitePieces, &pos.whitePawns, &pos.blackPieces, &pos.blackPawns, &pos.whiteKing,
		WhitePieceBit, Rank1, Rank5, bubbleDown
}

func (pos *Position) GetAtSquare(s square) piece {
	return pos.board[s]
}

func (pos *Position) GetAtFileRank(f file, r rank) piece {
	// cast to file is kindof dodgy but it must be faster than two casts to byte, right?
	var index file = f + file(r)
	return pos.board[index]
}
