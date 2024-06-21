package engine

// Maximum number of plies we expect to reach while searching the game tree in any practical scenario.
// This also means max line length
const MaxSearchDepth = 40

// Used in calculation of time dedicated to the next move in time-controlled games
const ExpectedFullMovesToBePlayed = 30

// Arena adjudicates game as a draw after 250 moves. So hopefully we will never need to realloc that.
const killerMovesMaxPly = 300