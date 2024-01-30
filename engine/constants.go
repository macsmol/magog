package engine

import (
	"fmt"
)

// square on 0x88 board -> https://www.chessprogramming.org/0x88
type square byte

const (
	A1, A2, A3, A4, A5, A6, A7, A8 square = iota * 0x10, iota*0x10 + 1, iota*0x10 + 2, iota*0x10 + 3, iota*0x10 + 4, iota*0x10 + 5, iota*0x10 + 6, iota*0x10 + 7
	B1, B2, B3, B4, B5, B6, B7, B8
	C1, C2, C3, C4, C5, C6, C7, C8
	D1, D2, D3, D4, D5, D6, D7, D8
	E1, E2, E3, E4, E5, E6, E7, E8
	F1, F2, F3, F4, F5, F6, F7, F8
	G1, G2, G3, G4, G5, G6, G7, G8
	H1, H2, H3, H4, H5, H6, H7, H8
	InvalidSquare square = 0x88
)

func (s square) String() string {
	if s&InvalidSquare != 0 {
		return "InvalidSquare"
	}
	var rank rune = rune(s&0x0F) + '1'
	var file rune = rune((s&0xF0)>>4) + 'a'
	return fmt.Sprintf("%c%c", file, rank)
}

// ----wppp; w - isWhite; ppp - piece type
type piece byte

const (
	NullPiece   piece = iota
	BlackPawn         //0b0001
	BlackKnight       //0b0010
	BlackBishop       //0b0011
	BlackRook         //0b0100
	BlackQueen        //0b0101
	BlackKing         //0b0110
)
const (
	WhitePawn piece = iota + 0b1001
	WhiteKnight
	WhiteBishop
	WhiteRook
	WhiteQueen
	WhiteKing
)

func (p piece) String() string {
	switch p {
	case NullPiece:
		return "- "
	case BlackPawn:
		return "pp"
	case BlackKnight:
		return "NN"
	case BlackBishop:
		return "BB"
	case BlackRook:
		return "RR"
	case BlackQueen:
		return "QQ"
	case BlackKing:
		return "KK"

	case WhitePawn:
		return "p "
	case WhiteKnight:
		return "N "
	case WhiteBishop:
		return "B "
	case WhiteRook:
		return "R "
	case WhiteQueen:
		return "Q "
	case WhiteKing:
		return "K "
	}
	panic(fmt.Sprintf("Unknown piece %X", byte(p)))
}
