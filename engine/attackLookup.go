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
var directionTable [239]Direction

func init() {
	initAttackTable()
	initDirectionsTable()
}

func initAttackTable() {
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
	for rankChange := int16(Rank2); rankChange <= int16(Rank8); rankChange += 0x10 {
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
	startingFile := int16(H)
	for destFile := int16(G); destFile >= int16(A); destFile-- {
		fileDifference := destFile - startingFile
		attackTable[lastValidSquare+int16(fileDifference)] |= RookAttacks | QueenAttacks
	}
	//bishop and queen moving NE
	attackTable[moveIndex(A1, B2)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(A1, C3)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(A1, D4)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(A1, E5)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(A1, F6)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(A1, G7)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(A1, H8)] |= BishopAttacks | QueenAttacks
	//bishop and queen moving SW
	attackTable[moveIndex(H8, G7)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(H8, F6)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(H8, E5)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(H8, D4)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(H8, C3)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(H8, B2)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(H8, A1)] |= BishopAttacks | QueenAttacks
	//bishop and queen moving NW
	attackTable[moveIndex(H1, G2)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(H1, F3)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(H1, E4)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(H1, D5)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(H1, C6)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(H1, B7)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(H1, A8)] |= BishopAttacks | QueenAttacks
	//bishop and queen moving SE
	attackTable[moveIndex(A8, B7)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(A8, C6)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(A8, D5)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(A8, E4)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(A8, F3)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(A8, G2)] |= BishopAttacks | QueenAttacks
	attackTable[moveIndex(A8, H1)] |= BishopAttacks | QueenAttacks
}

func initDirectionsTable() {
	//rook and queen moving north
	for rankChange := int16(Rank2); rankChange <= int16(Rank8); rankChange += 0x10 {
		directionTable[lastValidSquare+int16(rankChange)] = DirN
	}
	//rook and queen moving south
	startingRank := Rank8
	for destinationRank := Rank7; destinationRank >= Rank1; destinationRank -= 0x10 {
		rankChange := destinationRank - startingRank
		directionTable[lastValidSquare+int16(rankChange)] = DirS
	}
	// rook and queen moving east
	for file := B; file <= H; file++ {
		directionTable[lastValidSquare+int16(file)] = DirE
	}
	// rook and queen moving west
	startingFile := int16(H)
	for destFile := int16(G); destFile >= int16(A); destFile-- {
		fileDifference := destFile - startingFile
		directionTable[lastValidSquare+int16(fileDifference)] = DirW
	}
	//bishop and queen moving NE
	directionTable[moveIndex(A1, B2)] = DirNE
	directionTable[moveIndex(A1, C3)] = DirNE
	directionTable[moveIndex(A1, D4)] = DirNE
	directionTable[moveIndex(A1, E5)] = DirNE
	directionTable[moveIndex(A1, F6)] = DirNE
	directionTable[moveIndex(A1, G7)] = DirNE
	directionTable[moveIndex(A1, H8)] = DirNE
	//bishop and queen moving SW
	directionTable[moveIndex(H8, G7)] = DirSW
	directionTable[moveIndex(H8, F6)] = DirSW
	directionTable[moveIndex(H8, E5)] = DirSW
	directionTable[moveIndex(H8, D4)] = DirSW
	directionTable[moveIndex(H8, C3)] = DirSW
	directionTable[moveIndex(H8, B2)] = DirSW
	directionTable[moveIndex(H8, A1)] = DirSW
	//bishop and queen moving NW
	directionTable[moveIndex(H1, G2)] = DirNW
	directionTable[moveIndex(H1, F3)] = DirNW
	directionTable[moveIndex(H1, E4)] = DirNW
	directionTable[moveIndex(H1, D5)] = DirNW
	directionTable[moveIndex(H1, C6)] = DirNW
	directionTable[moveIndex(H1, B7)] = DirNW
	directionTable[moveIndex(H1, A8)] = DirNW
	//bishop and queen moving SE
	directionTable[moveIndex(A8, B7)] = DirSE
	directionTable[moveIndex(A8, C6)] = DirSE
	directionTable[moveIndex(A8, D5)] = DirSE
	directionTable[moveIndex(A8, E4)] = DirSE
	directionTable[moveIndex(A8, F3)] = DirSE
	directionTable[moveIndex(A8, G2)] = DirSE
	directionTable[moveIndex(A8, H1)] = DirSE
}

// from - attacker, to - attacked
func moveIndex(from, to square) int16 {
	return lastValidSquare + int16(to) - int16(from)
}
