package engine

import (
	"fmt"
)

// Square on 0x88 board -> https://www.chessprogramming.org/0x88
type Square uint8

const (
	A1 Square = iota
	A2
	A3
	A4
	A5
	A6
	A7
	A8
)
const (
	B1 Square = iota + 0x10
	B2
	B3
	B4
	B5
	B6
	B7
	B8
)
const (
	C1 Square = iota + 0x20
	C2
	C3
	C4
	C5
	C6
	C7
	C8
)
const (
	D1 Square = iota + 0x30
	D2
	D3
	D4
	D5
	D6
	D7
	D8
)
const (
	E1 Square = iota + 0x40
	E2
	E3
	E4
	E5
	E6
	E7
	E8
)
const (
	F1 Square = iota + 0x50
	F2
	F3
	F4
	F5
	F6
	F7
	F8
)
const (
	G1 Square = iota + 0x60
	G2
	G3
	G4
	G5
	G6
	G7
	G8
)
const (
	H1 Square = iota + 0x70
	H2
	H3
	H4
	H5
	H6
	H7
	H8
)
const (
	InvalidSquare Square = 0x88
)

func (s Square) String() string {
	if s&InvalidSquare != 0 {
		return "InvalidSquare"
	}
	var rank rune = rune(s&0x0F) + '1'
	var file rune = rune((s&0xF0)>>4) + 'a'
	return fmt.Sprintf("%c%c", file, rank)
}

// ----wppp; w - isWhite; ppp - piece type
type piece uint8

const (
	None        piece = iota
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
