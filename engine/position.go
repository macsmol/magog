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
	sb.WriteString("  | a| b| c| d| e| f| g| h|\n")
	sb.WriteString("===========================\n")
	for r := Rank8; r >= Rank1; r -= (Rank2 - Rank1) {
		sb.WriteString(fmt.Sprintf("%v|", r))
		for f := A; f <= H; f++ {
			p := pos.GetAtFileRank(f, r)
			sb.WriteString(fmt.Sprintf("%v|", p))
		}
		sb.WriteRune('\n')
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
	currPieceColorBit piece, enemyColorPieceBit piece,
	currPawnsStartRank rank,
) {
	if pos.flags&FlagWhiteTurn == 0 {
		return pos.blackPieces, pos.blackKing, DirS, BlackPieceBit, WhitePieceBit, Rank7
	}
	return pos.whitePieces, pos.whiteKing, DirN, WhitePieceBit, BlackPieceBit, Rank2
}

func (pos *Position) GetAtSquare(s square) piece {
	return pos.board[s]
}

func (pos *Position) GetAtFileRank(f file, r rank) piece {
	// cast to file is kindof dodgy but it must be faster than two casts to byte, right?
	var index file = f + file(r)
	return pos.board[index]
}
