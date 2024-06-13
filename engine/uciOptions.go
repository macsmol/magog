package engine

// TODO this will be a nice place to use go interfaces - once there are more options, of different types

// log 'info currmove' everytime we run evaluation on this number of nodes(positions)
const (
	currmoveLogIntervalKey     string = "currmoveLogInterval"
	currmoveLogIntervalDefault int    = 1000_000
	currmoveLogIntervalMin     int    = 10
	currmoveLogIntervalMax     int    = 10_000_000
)
var currmoveLogInterval int = currmoveLogIntervalDefault

