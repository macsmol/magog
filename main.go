package main

import (
	"bufio"
	"fmt"
	"macsmol/magog/engine"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println(engine.A1, engine.A2, engine.B2, engine.InvalidSquare, engine.H8)
	fmt.Println(engine.BBishop, engine.WBishop, engine.NullPiece)
	
	var startPos *engine.Position = engine.NewPosition()
	fmt.Println("startPos:", startPos)
	moves := startPos.GenerateMoves()
	fmt.Println("moves:", moves)

	for {
		scanner.Scan()
		line := scanner.Text()
		switch line {
		case "isready":
			isReady()
		case "quit":
			os.Exit(0)
		}
	}
}

func isReady() {
	fmt.Println("readyok")
}
