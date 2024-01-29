package engine

type position struct {
	// 0x88 board
	board       [128]piece
	blackPieces []Square
	blackKing   Square
	whitePieces []Square
	whiteKing   Square
	flags       uint8
	enPassSq    Square
}

// used for position.flags
const (
	WhiteTurnFlag byte = iota
	WhiteKingsideCastlePossible
	WhiteQsideCastlePossible
	BlackKsideCastlePossible
	BlackQsideCastlePossible
)
