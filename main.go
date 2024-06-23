package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"macsmol/magog/engine"
	"os"
	"strings"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	printWelcome()
	// ---------- runtime profiling stuff ---------------
	flag.Parse()
	if *cpuprofile != "" {
		fmt.Println("created file for profile.....", *cpuprofile)
		var f *os.File
		var err error
		f, err = os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		engine.ProfileFile = f
	}
	// ---------- runtime profiling stuff -end- ---------

	scanner := bufio.NewScanner(os.Stdin)
	for !engine.Quit {
		scanner.Scan()
		engine.ParseInputLine(scanner.Text())
	}
}

func printWelcome() {
	var banner string = `
  _ __  ____ ____ ___  ____
 | '  \/ _  / _  / _ \/ _  |
 |_|_|_\__,_\__, \___/\__, |
 v. . . . . |___/ . . |___/
 . . UCI chess engine . .
`
	mutableBanner := []rune(banner)
	verPrefixStr := "v."
	i := strings.Index(banner, verPrefixStr) + len(verPrefixStr)
	copy(mutableBanner[i:], []rune(engine.VERSION_STRING + " "))
	fmt.Println(string(mutableBanner))

	fmt.Println("Welcome!")
	fmt.Println("Please input a UCI command or type 'help' for additional commands.")
}