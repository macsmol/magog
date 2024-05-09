package engine

type Line struct {
	moves                []Move
	sublineLengthMatched int
}

func (line *Line)isMoveOnLine(m Move, depth int) bool {
	// Is the move at desired depth?
	if depth < len(line.moves) && line.moves[depth] == m && 
	// Are we looking for the move at the right depth?
	depth == line.sublineLengthMatched {
		line.sublineLengthMatched++
		return true
	}
	return false
}
