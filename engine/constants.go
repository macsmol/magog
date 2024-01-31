package engine

import (
	"fmt"
)

type rank int8

const (
	Rank1 rank = iota * 0x10
	Rank2
	Rank3
	Rank4
	Rank5
	Rank6
	Rank7
	Rank8
)

func (r rank) String() string {
	return fmt.Sprintf("%d.", int(r)>>4 + 1)
}

type file byte

const (
	A file = iota
	B
	C
	D
	E
	F
	G
	H
)

// square on 0x88 board -> https://www.chessprogramming.org/0x88
type square byte

const (
	A1, B1, C1, D1, E1, F1, G1, H1 square = iota * 0x10, iota*0x10 + 1, iota*0x10 + 2, iota*0x10 + 3, iota*0x10 + 4, iota*0x10 + 5, iota*0x10 + 6, iota*0x10 + 7
	A2, B2, C2, D2, E2, F2, G2, H2
	A3, B3, C3, D3, E3, F3, G3, H3
	A4, B4, C4, D4, E4, F4, G4, H4
	A5, B5, C5, D5, E5, F5, G5, H5
	A6, B6, C6, D6, E6, F6, G6, H6
	A7, B7, C7, D7, E7, F7, G7, H7
	A8, B8, C8, D8, E8, F8, G8, H8
	InvalidSquare square = 0x88
)

func (s square) String() string {
	if s&InvalidSquare != 0 {
		return "InvalidSquare"
	}
	var file rune = rune(s&0x0F) + 'a'
	var rank rune = rune((s&0xF0)>>4) + '1'
	return fmt.Sprintf("%c%c", file, rank)
}

// ----wppp; w - isWhite; ppp - piece type
type piece byte

const (
	NullPiece piece = iota
	BPawn           //0b0001
	BKnight         //0b0010
	BBishop         //0b0011
	BRook           //0b0100
	BQueen          //0b0101
	BKing           //0b0110
)
const (
	WPawn piece = iota + 0b1001
	WKnight
	WBishop
	WRook
	WQueen
	WKing
)

func (p piece) String() string {
	switch p {
	case NullPiece:
		return "- "
	case BPawn:
		return "pp"
	case BKnight:
		return "NN"
	case BBishop:
		return "BB"
	case BRook:
		return "RR"
	case BQueen:
		return "QQ"
	case BKing:
		return "KK"

	case WPawn:
		return "p "
	case WKnight:
		return "N "
	case WBishop:
		return "B "
	case WRook:
		return "R "
	case WQueen:
		return "Q "
	case WKing:
		return "K "
	}
	panic(fmt.Sprintf("Unknown piece %X", byte(p)))
}
