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

func (pos *Position) String() string {
	var sb strings.Builder
	sb.WriteRune('\n')
	sb.WriteString("  ┃ a│ b│ c│ d│ e│ f│ g│ h│\n")
	sb.WriteString("━━╋━━┿━━┿━━┿━━┿━━┿━━┿━━┿━━┥\n")
	for r := Rank8; r >= Rank1; r -= (Rank2 - Rank1) {
		sb.WriteString(fmt.Sprintf("%v┃", r))
		for f := A; f <= H; f++ {
			p := pos.GetAtFileRank(f, r)
			sb.WriteString(fmt.Sprintf("%v│", p))
		}
		sb.WriteString("\n──╂──┼──┼──┼──┼──┼──┼──┼──┤\n")
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
	currPieces []square,
	currKing square, pawnAdvance Direction,
	currColorBit piece, enemyColorBit piece,
	queensideCastlePossible, kingsideCastlePossible bool,
	currPawnsStartRank, promotionRank rank) {
	if pos.flags&FlagWhiteTurn == 0 {
		return pos.blackPieces, pos.blackKing, DirS, BlackPieceBit, WhitePieceBit,
			pos.flags&FlagBlackCanCastleQside != 0, pos.flags&FlagBlackCanCastleKside != 0,
			Rank7, Rank1
	}
	return pos.whitePieces, pos.whiteKing, DirN, WhitePieceBit, BlackPieceBit,
		pos.flags&FlagWhiteCanCastleQside != 0, pos.flags&FlagWhiteCanCastleKside != 0,
		Rank2, Rank8
}

func (pos *Position) MakeMove(mov Move) (undo backtrackInfo) {
	currPieces, enemyPieces, currKing, castleRank,
		currColorBit, kingSideCastleFlag, queenSideCastleFlag := pos.getCurrentMakeMoveContext()

	undo = backtrackInfo{
		move: mov,
		lastFlags:     pos.flags,
		lastEnPassant: pos.enPassSquare,
	}

	for i := range currPieces {
		if mov.from == currPieces[i] {
			currPieces[i] = mov.to
		}
	}

	if mov.from.getFile() == A && mov.from.getRank() == castleRank {
		pos.flags &= ^queenSideCastleFlag
	}
	if mov.from.getFile() == H && mov.from.getRank() == castleRank {
		pos.flags &= ^kingSideCastleFlag
	}

	if mov.from == *currKing {
		pos.flags &= ^(kingSideCastleFlag | queenSideCastleFlag)
		*currKing = mov.to
		if mov.from.getFile() == E {
			if mov.to.getFile() == C {
				rookFrom := square(A + file(castleRank))
				rookTo := square(D + file(castleRank))
				pos.moveRook(rookFrom, rookTo, currPieces, currColorBit)
			} else if mov.to.getFile() == G {
				rookFrom := square(H + file(castleRank))
				rookTo := square(F + file(castleRank))
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

	return undo
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
	castleRank rank,
	currColorBit piece,
	kingSideCastleFlag, queenSideCastleFlag byte,
) {
	if pos.flags&FlagWhiteTurn == 0 {
		return pos.blackPieces, &pos.whitePieces, &pos.blackKing, Rank8,
			BlackPieceBit, FlagBlackCanCastleKside, FlagBlackCanCastleQside
	}
	return pos.whitePieces, &pos.blackPieces, &pos.whiteKing, Rank1,
		WhitePieceBit, FlagWhiteCanCastleKside, FlagWhiteCanCastleQside
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
