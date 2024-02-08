package engine

const (
	_ byte = 1 << iota
	RookAttacks
	BishopAttacks
	QueenAttacks
	KnightAttacks
	KingAttacks
	WhitePawnAttacks
	BlackPawnAttacks
)

const lastValidSquare = int16(H8)

var attackTable [239]byte

func init() {
	//knight
	attackTable[lastValidSquare+int16(DirNNE)] |= KnightAttacks
	attackTable[lastValidSquare+int16(DirSSW)] |= KnightAttacks
	attackTable[lastValidSquare+int16(DirNNW)] |= KnightAttacks
	attackTable[lastValidSquare+int16(DirSSE)] |= KnightAttacks
	attackTable[lastValidSquare+int16(DirNEE)] |= KnightAttacks
	attackTable[lastValidSquare+int16(DirSWW)] |= KnightAttacks
	attackTable[lastValidSquare+int16(DirNWW)] |= KnightAttacks
	attackTable[lastValidSquare+int16(DirSEE)] |= KnightAttacks
	//pawns
	attackTable[lastValidSquare+int16(DirNE)] |= WhitePawnAttacks
	attackTable[lastValidSquare+int16(DirNW)] |= WhitePawnAttacks
	attackTable[lastValidSquare+int16(DirSE)] |= BlackPawnAttacks
	attackTable[lastValidSquare+int16(DirSW)] |= BlackPawnAttacks
	//king
	attackTable[lastValidSquare+int16(DirN)] |= KingAttacks
	attackTable[lastValidSquare+int16(DirNE)] |= KingAttacks
	attackTable[lastValidSquare+int16(DirE)] |= KingAttacks
	attackTable[lastValidSquare+int16(DirSE)] |= KingAttacks
	attackTable[lastValidSquare+int16(DirS)] |= KingAttacks
	attackTable[lastValidSquare+int16(DirSW)] |= KingAttacks
	attackTable[lastValidSquare+int16(DirW)] |= KingAttacks
	attackTable[lastValidSquare+int16(DirNW)] |= KingAttacks
	//rook and queen moving north
	for rankChange := Rank2; rankChange <= Rank8; rankChange += 0x10 {
		attackTable[lastValidSquare+int16(rankChange)] |= RookAttacks | QueenAttacks
	}
	//rook and queen moving south
	startingRank := Rank8
	for destinationRank := Rank7; destinationRank >= Rank1; destinationRank -= 0x10 {
		rankChange := destinationRank - startingRank
		attackTable[lastValidSquare+int16(rankChange)] |= RookAttacks | QueenAttacks
	}
	// rook and queen moving east
	for file := B; file <= H; file++ {
		attackTable[lastValidSquare+int16(file)] |= RookAttacks | QueenAttacks
	}
	// rook and queen moving west
	startingFile := H
	for destFile := G; destFile >= A; destFile-- {
		fileDifference := destFile-startingFile
		attackTable[lastValidSquare+int16(fileDifference)] |= RookAttacks | QueenAttacks
	}
	//bishop and queen moving NE
	attackTable[attackIndex(A1, B2)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(A1, C3)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(A1, D4)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(A1, E5)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(A1, F6)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(A1, G7)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(A1, H8)] |= BishopAttacks | QueenAttacks
	//bishop and queen moving SW
	attackTable[attackIndex(H8, G7)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(H8, F6)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(H8, E5)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(H8, D4)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(H8, C3)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(H8, B2)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(H8, A1)] |= BishopAttacks | QueenAttacks
	//bishop and queen moving NW
	attackTable[attackIndex(H1, G2)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(H1, F3)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(H1, E4)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(H1, D5)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(H1, C6)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(H1, B7)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(H1, A8)] |= BishopAttacks | QueenAttacks
	//bishop and queen moving SE
	attackTable[attackIndex(A8, B7)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(A8, C6)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(A8, D5)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(A8, E4)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(A8, F3)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(A8, G2)] |= BishopAttacks | QueenAttacks
	attackTable[attackIndex(A8, H1)] |= BishopAttacks | QueenAttacks
}

// from - attacker, to - attacked
func attackIndex(from, to square) int16 {
	return lastValidSquare+int16(to)- int16(from)
}

