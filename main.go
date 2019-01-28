/* vim: set sts=4 sw=4 et: */
/**
 * jass - just another sudoku solver
 * (C) 2005-2019 Jari Tenhunen <jait@iki.fi>
 *
 * Go version 2019
 */

package main

import (
	"bufio"
	"flag"
	"fmt"
	"jassgo/jass"
	"log"
	"os"
)

func main() {
	game := &jass.Game{}

	game.Init()

	/* check args
	 * -f: read sudokus from file (- for stdin)
	 */

	var fname string
	var step, verbose bool

	flag.BoolVar(&step, "s", false, "step mode, pause after each solved number")
	flag.BoolVar(&verbose, "v", false, "verbose debug output")
	flag.StringVar(&fname, "f", "", "instead looking for the puzzle string in the arguments, read puzzles from `file` (\"-\" for stdin), one per line")
	flag.Parse()

	if step && !verbose {
		verbose = true
	}

	if verbose {
		jass.SetLogLevel(jass.LogDebug)
	}

	if step {
		game.SetMode(jass.StepMode)
	}

	if fname != "" {
		var file *os.File
		var err error
		if fname == "-" {
			file = os.Stdin
		} else {
			file, err = os.Open(fname)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
		}

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			str := scanner.Text()
			if len(str) == 0 || str[0] == '#' {
				continue
			}
			game.Init()
			game.ParseBoard(str)
			fmt.Println(str)
			game.Solve()
		}
		if err = scanner.Err(); err != nil {
			log.Fatal(err)
		}
	} else {
		// try if there was a puzzle as argument
		args := flag.Args()
		if len(args) == 0 {
			jass.Info("Error: no puzzle(s) given")
			jass.Info("Usage: jass [options] [<puzzle>]")
			flag.PrintDefaults()
		}
		for _, str := range args {
			game.Init()
			game.ParseBoard(str)
			game.Solve()
		}
	}
}
