package engine

import (
	"fmt"
	"strings"
)

func NewPositionFromFEN(fen string) (*Position, error) {
	fields := strings.Split(fen, " ")
	if len(fields) != 6 {
		return nil, fmt.Errorf("FEN string does not have 6 fields separated by spaces", fen)
	}
	boardStr := fields[0]
	// enPassantStr := fields[3]
	// halfmoveClockStr := fields[4]
	// halfmoveClockStr := fields[5]

	rankStrings := strings.Split(boardStr, "/")
	if len(rankStrings) != 8 {
		return nil, fmt.Errorf("number of ranks different than 8: %v", rankStrings)
	}

	var pos *Position = &Position{enPassSquare: InvalidSquare}

	for fenRankIdx, rankStr := range rankStrings {
		var r rank = rankFrom07Number(7 - fenRankIdx)
		var f file = A
		for _, c := range rankStr {
			if c >= '1' && c <= '8' {
				f += file(c - '0')
				continue
			}
			var sq square = square(r + rank(f))
			piece := charToPiece(c)
			if piece == NullPiece {
				return nil, fmt.Errorf("unknown piece in FEN: %v", c)
			}
			pos.board[sq] = piece
			if piece == BKing {
				pos.blackKing = sq
			} else if piece == WKing {
				pos.whiteKing = sq
			} else if piece&WhitePieceBit == 0 {
				pos.blackPieces = append(pos.blackPieces, sq)
			} else {
				pos.whitePieces = append(pos.whitePieces, sq)
			}
			f++
		}
	}

	turnStr := fields[1]
	if turnStr == "w" {
		pos.flags |= FlagWhiteTurn
	} else if turnStr != "b" {
		return nil, fmt.Errorf("'side to move' is neither 'b' or 'w' : %v", turnStr)
	}

	castleStr := fields[2]
	if strings.Contains(castleStr,"K") {
		pos.flags|=FlagWhiteCanCastleKside
	}
	if strings.Contains(castleStr,"Q") {
		pos.flags|=FlagWhiteCanCastleQside
	}
	if strings.Contains(castleStr,"k") {
		pos.flags|=FlagBlackCanCastleKside
	}
	if strings.Contains(castleStr,"q") {
		pos.flags|=FlagBlackCanCastleQside
	}

	// rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1
	return pos, nil
}


func charToPiece(c rune) piece {
	switch c {
	case 'p':
		return BPawn
	case 'n':
		return BKnight
	case 'b':
		return BBishop
	case 'r':
		return BRook
	case 'q':
		return BQueen
	case 'k':
		return BKing

	case 'P':
		return WPawn
	case 'N':
		return WKnight
	case 'B':
		return WBishop
	case 'R':
		return WRook
	case 'Q':
		return WQueen
	case 'K':
		return WKing
	default:
		return NullPiece
	}
}
