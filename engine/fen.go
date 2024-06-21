package engine

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func NewPositionFromFen(fen string) (Position, error) {
	if !isASCII(fen) {
		return Position{}, fmt.Errorf("FEN string should contain only ASCII characters: %v", fen)
	}
	fields := strings.Split(fen, " ")
	if len(fields) != 6 {
		return Position{}, fmt.Errorf("FEN string does not have 6 fields separated by spaces: %v", fen)
	}
	boardStr := fields[0]

	rankStrings := strings.Split(boardStr, "/")
	if len(rankStrings) != 8 {
		return Position{}, fmt.Errorf("number of ranks different than 8: %v", rankStrings)
	}

	var pos Position = Position{enPassSquare: InvalidSquare}

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
				return Position{}, fmt.Errorf("uknown piece: %q", c)
			}
			pos.board[sq] = piece
			if piece == BKing {
				pos.blackKing = sq
			} else if piece == WKing {
				pos.whiteKing = sq
			} else if piece&WhitePieceBit == 0 {
				if piece == BPawn {
					pos.blackPawns.appendPawn(sq)
					} else {
					pos.blackPieces.appendPiece(sq)
				}
			} else {
				if piece == WPawn {
					pos.whitePawns.appendPawn(sq)
				} else {
					pos.whitePieces.appendPiece(sq)
				}
			}
			f++
		}
		if f != H+1 {
			return Position{}, error(fmt.Errorf("this rank does not have 8 files: %q", rankStr))
		}
	}

	turnStr := fields[1]
	if turnStr == "w" {
		pos.flags |= FlagWhiteTurn
	} else if turnStr != "b" {
		return Position{}, fmt.Errorf("'side to move' is neither 'b' or 'w': %v", turnStr)
	}

	castleStr := fields[2]
	if strings.Contains(castleStr, "K") {
		pos.flags |= FlagWhiteCanCastleKside
	}
	if strings.Contains(castleStr, "Q") {
		pos.flags |= FlagWhiteCanCastleQside
	}
	if strings.Contains(castleStr, "k") {
		pos.flags |= FlagBlackCanCastleKside
	}
	if strings.Contains(castleStr, "q") {
		pos.flags |= FlagBlackCanCastleQside
	}

	enPassantStr := fields[3]
	if len(enPassantStr) > 2 {
		return Position{}, fmt.Errorf("invalid en passant square: %v", enPassantStr)
	} else if len(enPassantStr) == 2 {
		fileChar := enPassantStr[0]
		rankChar := enPassantStr[1]
		if fileChar < 'a' || fileChar > 'h' || (rankChar != '3' && rankChar != '6') {
			return Position{}, fmt.Errorf("invalid en passant square: %v", enPassantStr)
		}
		file := fileChar - 'a'
		rank := (rankChar - '1') << 4
		pos.enPassSquare = square(file) + square(rank)
	}

	//TODO read rest of the fields
	// halfmoveClockStr := fields[4]

	fullMoveCounterStr := fields[5]
	fullMoveCounter, err := strconv.Atoi(fullMoveCounterStr)
	if err != nil {
		return Position{}, fmt.Errorf("invalid full move counter: %v", fullMoveCounterStr)
	}
	if fullMoveCounter < 1 {
		return Position{}, fmt.Errorf("full move counter is not 1-based: %d", fullMoveCounter)
	}

	pos.ply = int16((fullMoveCounter-1)*2)
	if pos.flags & FlagWhiteTurn == 0 {
		pos.ply++
	}
	

	return pos, nil
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
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
