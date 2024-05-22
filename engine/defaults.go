package engine

// Maximum number of plies we expect to reach while searching the game tree in any practical scenario.
// This also means max line length
const MaxSearchDepth = 40

// Used in calculation of time dedicated to the next move in time-controlled games
const ExpectedFullMovesToBePlayed = 30

//log info everytime we run evaluation on this number of nodes(positions)
const LogEveryNNodes = 500_000
