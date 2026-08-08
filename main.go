package main

import (
	// "strconv"
	// "strings"

	"flag"
	"fmt"
	"gaze/internal/viewer"
	"os"
	"runtime/pprof"
)

func main() {
	cpuProfile := flag.String("cpuprofile", "", "write CPU profile")
	flag.Parse()

	if *cpuProfile != "" {
		file, err := os.Create(*cpuProfile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer file.Close()

		if err := pprof.StartCPUProfile(file); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
	}

	err := viewer.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
