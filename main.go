package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"macsmol/magog/engine"
	"os"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
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