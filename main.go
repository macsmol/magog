package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"macsmol/magog/engine"
	"os"
	"runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	// ---------- runtime profiling stuff ---------------
	flag.Parse()
    if *cpuprofile != "" {
		fmt.Println("profiling.....", *cpuprofile)
		f, err := os.Create(*cpuprofile)
        if err != nil {
			log.Fatal(err)
        }
        pprof.StartCPUProfile(f)
        defer stopProfiling()
    }
	// ---------- runtime profiling stuff -end- ---------


	scanner := bufio.NewScanner(os.Stdin)
	for !engine.Quit {
		scanner.Scan()
		engine.ParseInputLine(scanner.Text())
	}
}

func stopProfiling() {
	fmt.Println("stopping profiling.....")
	pprof.StopCPUProfile()
}

