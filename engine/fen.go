package engine

import (
	"errors"
	"fmt"
	"strings"
)

func NewPositionFromFEN(fen string) (*Position, error) {
	fields := strings.Split(fen, " ")
	if len(fields) != 6 {
		return nil, errors.New("Parse error! Expecting FEN string with 6 columns separated by spaces, eg.: 7k/8/p7/8/8/1P6/8/7K b - - 0 1")
	}
	boardStr := fields[0]
	// turnStr := fields[1]
	// castleStr := fields[2]
	// enPassantStr := fields[3]
	// halfmoveClockStr := fields[4]
	// halfmoveClockStr := fields[5]

	rankStrings := strings.Split(boardStr,"/")
	if len(rankStrings) != 8 {
		return nil, errors.New("Parse error! Expecting Piece placement with exacly 8 Ranks (separated by 7 '/' chars), eg: 7k/8/8/8/8/1P6/8/7K")
	}

	// var board [128]piece
	// blackPieces := []square{}
	whitePieces := []square{}
	// var blackKing square
	// var whiteKing square

	fmt.Printf("white pieces sliceeee:%v, len: %v, cap: %v", len(whitePieces), cap(whitePieces), whitePieces)

	for fenRankIdx, rankStr := range rankStrings {
		var r rank = rankFrom07Number(7-fenRankIdx)
		var f file = A
		for _, c := range rankStr {
			if c >= '1' && c <= '8' {
				f += file(c -'0')
				continue
			}
			var _ square = square(r + rank(f)) 
			
		}
		
	}

	// rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1
	return nil, nil
}
