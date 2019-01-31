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

func (point Point) ToString() string {
	return fmt.Sprintf("(%d, %d)", point.x, point.y)
}

func (point Point) ToString1() string {
	return fmt.Sprintf("(%d, %d)", point.x+1, point.y+1)
}

func (a Point) Equals(b Point) bool {
	return a.x == b.x && a.y == b.y
}

func NewBoard() Board {
	b := make(Board, Y)
	for y := range b {
		b[y] = make([]Num, X)
	}
	return b
}

func (game *Game) Init() {
	game.board = NewBoard()
	game.poss = NewPoss()
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

func (b Board) Verify() bool {
	var found [NR_MAX]bool
	result := true
	// rows
	for j := 0; j < Y; j++ {
		for i := 0; i < X; i++ {
			n := b[j][i]
			if found[n-1] {
				Info("Verify error: %d occurs more than once", n)
				result = false
			} else {
				found[n-1] = true
			}
		}
		for n, f := range found {
			if !f {
				Info("Verify error: %d not found", n+1)
				result = false
			}
			// clear at the same time
			found[n] = false
		}
	}
	// cols
	for i := 0; i < X; i++ {
		for j := 0; j < Y; j++ {
			n := b[j][i]
			if found[n-1] {
				Info("Verify error: %d occurs more than once", n)
				result = false
			} else {
				found[n-1] = true
			}
		}
		for n, f := range found {
			if !f {
				Info("Verify error: %d not found", n+1)
				result = false
			}
			// clear at the same time
			found[n] = false
		}
	}
	// TODO: boxes
	return result
}

type BoardWalker func(y, x, val Num)

func (b Board) ForEachRow(fn BoardWalker) {
	// rows
	for j := Num(0); j < Y; j++ {
		for i := Num(0); i < X; i++ {
			fn(j, i, b[j][i])
		}
	}
}

func (b Board) ForEachCol(fn BoardWalker) {
	// cols
	for i := Num(0); i < X; i++ {
		for j := Num(0); j < Y; j++ {
			fn(j, i, b[j][i])
		}
	}
}

func (b Board) CellOccupied(cell Point) bool {
	return b[cell.y][cell.x] != 0
}

/*
 * Fix (place) a number (1...NR_MAX) in the board cell (y,x)
 */
func (game *Game) Fix(y, x, val Num) {

	var i, k Num
	//Explain("Placing %d into (%d, %d)", val, x+1, y+1);
	if game.board[y][x] != 0 {
		Info("Error: cell (%d,%d) already contains value %d", x+1, y+1, game.board[y][x])
	}

	game.board[y][x] = val

	/* no other possibilities for this cell */
	for k = 1; k <= NR_MAX; k++ {
		game.poss.Set(y, x, k, false)
	}

	/* eliminate all occurrences of val from this col */
	for i = Num(0); i < Y; i++ {
		/*
		   Explain("Eliminating %d from (%d, %d)", val, x+1, i+1);
		*/
		game.poss.Set(i, x, val, false)
	}
	/* and row */
	for i = Num(0); i < X; i++ {
		/*
		   Explain("Eliminating %d from (%d, %d)", val, i+1, y+1);
		*/
		game.poss.Set(y, i, val, false)
	}

	/* eliminate all occurrences of val from current box */
	y = (y / BoxX) * BoxX
	x = (x / BoxY) * BoxY

	for i = y; i < y+BoxX; i++ {
		for j := x; j < x+BoxY; j++ {
			/*
			   Explain("Eliminating %d from (%d, %d)", val, j+1, i+1);
			*/
			game.poss.Set(i, j, val, false)
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
	board.ForEachCol(func(y, x, val Num) {
		if val == 0 {
			unsolved++
		}
	})
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

func (board *Board) String() string {
	var buffer bytes.Buffer
	b := *board

	b.ForEachRow(func(y, x, val Num) {
		if val == 0 {
			buffer.WriteString(".")
		} else {
			buffer.WriteString(strconv.FormatInt(int64(val), 10))
		}
	})
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
		if nr = scanner.ScanSinglesBoxes(); nr > 0 {
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
		if nr = scanner.ScanRowsCols(ScanBoxLineGroup, "box/line", true); nr > 0 {
			//print_board();
			continue
		}
		Debug("Scanning for naked triples...")
		if nr = scanner.ScanAllUnoccupiedGroups(ScanNakedTriplesGroup, "naked triples"); nr > 0 {
			// print_board()
			continue
		}
	}

	if nr = game.CountUnsolved(); nr == 0 {
		Info("Sudoku solved!")
		game.board.Verify()
	} else {
		Info("Sudoku not solved, %d numbers left =(", nr)
	}
	game.board.Print()

	fmt.Println(game.board.String())

	return nr == 0
}

func (game *Game) SetMode(newmode int) {
	game.mode = newmode
}
