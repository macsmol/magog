package engine

type position struct {
	// 0x88 board
	board       [128]piece
	blackPieces []square
	blackKing   square
	whitePieces []square
	whiteKing   square
	flags       byte
	enPassSq    square
}

// used for position.flags
const (
	FlagWhiteTurn byte = 1 << iota
	FlagWhiteCanCastleKside
	FlagWhiteCanCastleQside
	FlagBlackCanCastleKside
	FlagBlackCanCastleQside
)
