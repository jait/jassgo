/* vim: set sts=4 sw=4 et: */
/**
 * jass - just another sudoku solver
 * (C) 2005-2007 Jari Tenhunen <jait@iki.fi>
 *
 * TODO:
 * - use a bit vector for storing possibles?
 *
 */

package jass

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
)

const (
	X          = 9
	Y          = 9
	BoxX       = 3
	BoxY       = 3
	NR_MAX     = 9
	NormalMode = 0
	StepMode   = 1
)

type Num uint8
type Coord int8

type Point struct {
	x int
	y int
}

type Board [][]Num

type Game struct {
	board Board
	poss  Poss
	mode  int
}

func (game *Game) Init() {
	var i, j, k Num
	game.board = make(Board, Y)
	for y, _ := range game.board {
		game.board[y] = make([]Num, X)
	}
	game.poss = make(Poss, Y)
	for i = 0; i < Y; i++ {
		game.poss[i] = make([][]bool, X)
		for j = 0; j < X; j++ {
			game.poss[i][j] = make([]bool, NR_MAX)
			for k = 0; k < NR_MAX; k++ {
				game.poss.Set(i, j, k, true)
			}
		}
	}
}

func (b *Board) Print() {

	for i := 0; i < X; i++ {
		if i%BoxY == 0 {
			fmt.Println("+-------+-------+-------+")
		}

		for j := 0; j < Y; j++ {
			if j%BoxX == 0 {
				fmt.Printf("| ")
			}

			if (*b)[i][j] != 0 {
				fmt.Printf("%d ", (*b)[i][j])
			} else {
				fmt.Printf(". ")
			}
		}
		fmt.Printf("|\n")
	}
	fmt.Println("+-------+-------+-------+")
}

func (game *Game) Fix(y, x, val Num) {

	var i, k Num
	//Explain("Placing %d into (%d, %d)", val, x+1, y+1);
	if game.board[y][x] != 0 {
		Info("Error: cell (%d,%d) already contains value %d\n", x+1, y+1, game.board[y][x])
	}

	game.board[y][x] = val

	/* no other possibilities for this cell */
	for k = 0; k < NR_MAX; k++ {
		game.poss.Set(y, x, k, false)
	}

	/* eliminate all occurrences of val from this col */
	for i = Num(0); i < Y; i++ {
		/*
		   Explain("Eliminating %d from (%d, %d)", val, x+1, i+1);
		*/
		game.poss.Set(i, x, val-1, false)
	}
	/* and row */
	for i = Num(0); i < X; i++ {
		/*
		   Explain("Eliminating %d from (%d, %d)", val, i+1, y+1);
		*/
		game.poss.Set(y, i, val-1, false)
	}

	/* eliminate all occurrences of val from current box */
	y = (y / BoxX) * BoxX
	x = (x / BoxY) * BoxY

	for i = y; i < y+BoxX; i++ {
		for j := x; j < x+BoxY; j++ {
			/*
			   Explain("Eliminating %d from (%d, %d)", val, j+1, i+1);
			*/
			game.poss.Set(i, j, val-1, false)
		}
	}

	if game.mode == StepMode {
		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')
	}
}

/**
 * Returns number of unsolved cells
 *
 */

func (game *Game) CountUnsolved() int {
	return game.board.CountUnsolved()
}

func (board *Board) CountUnsolved() int {
	unsolved := 0

	for i := 0; i < X; i++ {
		for j := 0; j < Y; j++ {
			if (*board)[i][j] == 0 {
				unsolved++
			}
		}
	}
	return unsolved
}

/**
 *
 *
 */
func (game *Game) ParseBoard(str string) bool {
	for i, c := range str {
		if c == '0' || c == '.' {
			// nothing to do
			continue
		}
		if val, err := strconv.Atoi(string(c)); err == nil {
			game.Fix(Num(i/X), Num(i%X), Num(val))
		} else {
			return false
		}
	}
	return true
}

func (board *Board) ToString() string {
	var buffer bytes.Buffer
	b := *board

	for i := 0; i < Y; i++ {
		for j := 0; j < X; j++ {
			if b[i][j] == 0 {
				buffer.WriteString(".")
			} else {
				buffer.WriteString(strconv.FormatInt(int64(b[i][j]), 10))
			}
		}
	}
	return buffer.String()
}

/**
 * Tries to solve the puzzle
 *
 * Returns true if the puzzle was fully solved
 */
func (game *Game) Solve() bool {
	nr := 1
	/* loop as long as there is some progress */

	scanner := Scanner{game}
	for nr > 0 {
		nr = 0
		if game.CountUnsolved() == 0 {
			break
		}

		Debug("Scanning for singles...")
		if nr = scanner.ScanSingles(); nr > 0 {
			// print_board()
			continue
		}

		Debug("Scanning boxes for singles and pointing pairs/triples...")
		if nr = scanner.ScanBoxes(); nr > 0 {
			// print_board()
			continue
		}
		Debug("Scanning for singles on rows and cols...")
		if nr = scanner.ScanSinglesRowCol(); nr > 0 {
			// print_board()
			continue
		}
		Debug("Scanning for naked pairs...")
		if nr = scanner.ScanAllGroups(ScanNakedPairsGroup, "naked pairs"); nr > 0 {
			// print_board()
			continue
		}
		Debug("Scanning for hidden pairs...")
		if nr = scanner.ScanAllGroups(ScanHiddenPairsGroup, "hidden pairs"); nr > 0 {
			// print_board();
			continue
		}
		Debug("Doing box/line reduction...")
		if nr = scanner.ScanRowsCols(ScanBoxLineGroup, "box/line"); nr > 0 {
			//print_board();
			continue
		}
	}

	if nr = game.CountUnsolved(); nr == 0 {
		Info("Sudoku solved!")
	} else {
		Info("Sudoku not solved, %d numbers left =(", nr)
	}
	game.board.Print()

	fmt.Println(game.board.ToString())

	return (nr == 0)
}

func (game *Game) SetMode(newmode int) {
	game.mode = newmode
}

func (a Point) Equals(b Point) bool {
	return a.x == b.x && a.y == b.y
}
