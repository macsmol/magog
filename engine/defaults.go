package engine

// Maximum number of plies we expect to reach while searching the game tree in any practical scenario.
// This also means max line length
const MaxSearchDepth = 40

// Used in calculation of time dedicated to the next move in time-controlled games
const ExpectedFullMovesToBePlayed = 30

// Typically game adjudicated as draws after 300 moves. So 300 + margin for search depth.
const killerMovesMaxPly = 350