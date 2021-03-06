/* vim: set sts=4 sw=4 ts=4 noet: */
/**
 * jass - just another sudoku solver
 * (C) 2005-2019 Jari Tenhunen <jait@iki.fi>
 *
 * Go version 2019
 *
 * TODO:
 * - X-Wing?
 *
 */

package jass

type GroupScanFunc func(game *Game, cells []Point) int

func containsInt(slice []int, a int) bool {
	for _, i := range slice {
		if i == a {
			return true
		}
	}
	return false
}

func comb(n, m int, emit func([]int)) {
	s := make([]int, m)
	last := m - 1
	var rc func(int, int)
	rc = func(i, next int) {
		for j := next; j < n; j++ {
			s[i] = j
			if i == last {
				emit(s)
			} else {
				rc(i+1, j+1)
			}
		}
		return
	}
	rc(0, 0)
}

func getBox(y, x int) int {
	return int(((y/BoxY)*(X/BoxX) + (x / BoxX)))
}

type Scanner struct {
	game *Game
}

func (scanner *Scanner) ScanSingles() int {
	found := 0
	var i, j, k Num

	for i = 0; i < Y; i++ {
		for j = 0; j < X; j++ {
			// don't check possibilities if the cell is occupied
			if scanner.game.board[i][j] != 0 {
				continue
			}
			val := Num(0)
			for k = 1; k <= NR_MAX; k++ {
				if scanner.game.poss.Get(i, j, k) {
					if val != 0 {
						/* at least two possibilities */
						val = 0
						break
					}

					val = k
				}
			}

			if val != 0 {
				Explain("Single possibility (%d) for cell (%d, %d)", val, j+1, i+1)
				scanner.game.Fix(i, j, val)
				found++
			}
		}
	}
	return found
}

func (scanner *Scanner) ScanSinglesRowCol() int {
	var i, j, k Num
	var place [NR_MAX]int /* holds the column or row of the only possible place for numbers */
	/* 0  => no possibilities (yet)  */
	/* -1 => two or more possibilities */
	found := 0

	/* loop rows */
	for i = 0; i < Y; i++ {
		for k = 0; k < NR_MAX; k++ {
			place[k] = 0
		}

		for j = 0; j < X; j++ {
			/* don't check possibilities if the cell is occupied */
			if scanner.game.board[i][j] != 0 {
				continue
			}

			for k = 0; k < NR_MAX; k++ {
				if scanner.game.poss.Get(i, j, k+1) {
					if place[k] == 0 { /* no possibility yet */
						place[k] = int(j + 1) /* because zero has a special meaning */
					} else {
						place[k] = -1 /* two or more possible places for k */
					}
				}
			}

		}
		/* check after each row */
		for k = 0; k < NR_MAX; k++ {
			if place[k] > 0 {
				Explain("Single possible place (col %d) for %d on row %d", place[k]-1+1, k+1, i+1)
				/* 1) because place-array has special meaning for zero
				 * 2) because k is zero-offset */
				scanner.game.Fix(i, Num(place[k]-1), k+1)
				found++
			}
		}
	}

	/* loop cols */
	for j = 0; j < X; j++ {
		for k = 0; k < NR_MAX; k++ {
			place[k] = 0
		}

		for i = 0; i < Y; i++ {
			/* don't check possibilities if the cell is occupied */
			if scanner.game.board[i][j] != 0 {
				continue
			}

			for k = 0; k < NR_MAX; k++ {
				if scanner.game.poss.Get(i, j, k+1) {
					if place[k] == 0 { /* no possibility yet */
						place[k] = int(i + 1) /* because zero has a special meaning */
					} else {
						place[k] = -1 /* two or more possible places for k */
					}
				}
			}
		}
		/* check for singles after each row */
		for k = 0; k < NR_MAX; k++ {
			if place[k] > 0 {
				Explain("Single possible place (row %d) for %d on col %d", place[k]-1+1, k+1, j+1)
				/* 1) because place-array has special meaning for zero
				 * 2) because k is zero-offset */
				scanner.game.Fix(Num(place[k]-1), j, k+1)
				found++
			}
		}
	}
	return found
}

/*
 * Finds singles in boxes and does box-line and box-col reduction
 *
 */
func (scanner *Scanner) ScanSinglesBoxes() int {
	game := scanner.game
	var i, j, k, bi, bj, tmpx, tmpy Num
	found := 0
	var boxes_x, boxes_y Num
	var place [NR_MAX]Point /* holds the place (y,x) of the only possible place for numbers */
	/* 0,0  => no possibilities (yet)  */
	/* -1   => two or more possibilities */

	/* loop over 3x3 boxes */
	/* loop over each number, checking if it has only one possible place */
	/* if there's more than one possibility, check if we can do box-line reduction */

	boxes_y = Y / BoxY
	boxes_x = X / BoxX

	for bi = 0; bi < boxes_y; bi++ {
		for bj = 0; bj < boxes_x; bj++ {
			/* clear place array */
			for k = 0; k < NR_MAX; k++ {
				place[k].y = 0
				place[k].x = 0
			}
			/* debug("Scanning box (%d, %d)", bj+1, bi+1); */
			/* loop over the nine cells */
			for i = 0; i < BoxY; i++ {
				tmpy = bi*BoxY + i
				for j = 0; j < BoxX; j++ {
					/* tricky */
					tmpx = bj*BoxX + j
					/* don't check possibilities if the cell is occupied */
					if game.board[tmpy][tmpx] != 0 {
						continue
					}

					for k = 0; k < NR_MAX; k++ {
						if game.poss.Get(tmpy, tmpx, k+1) {
							/* debug("%d poss in (%d, %d)", k+1, j+1, i+1); */
							if place[k].x == 0 {
								/* no possibility yet */
								place[k].y = int(1 + tmpy) /* because zero has a special meaning */
								place[k].x = int(1 + tmpx)
							} else {
								/* two or more possible places for k */

								/* if the possibilities are not in the same col,
								 * mark as -1 */
								if place[k].x != int(1+tmpx) {
									place[k].x = -1
								}

								/* not same row => mark as - 1 */
								if place[k].y != int(1+tmpy) {
									place[k].y = -1
								}
							}
						}
					}
				}
			}
			/* check after each box */
			for k = 0; k < NR_MAX; k++ {
				if place[k].x > 0 && place[k].y > 0 {
					Explain("Single possible place %s for %d in box (%d, %d)", place[k].ToString(), k+1, bj+1, bi+1)
					/* 1) because place-array has special meaning for zero
					 * 2) because k is zero-offset */
					game.Fix(Num(place[k].y-1), Num(place[k].x-1), k+1)
					found++
				} else if place[k].x > 0 {
					/* k possible only on this col */
					/* eliminate k's other possibilities from other boxes on current col */
					for tmpy = 0; tmpy < boxes_y; tmpy++ {
						if tmpy == bi { /* don't delete possibilities from current box */
							continue
						}

						for i = 0; i < BoxY; i++ {
							if game.poss.Get(tmpy*BoxY+i, Num(place[k].x-1), k+1) {
								/* Explain("%d possible only on col %d in box (%d, %d)", k+1, place[k].x, bj+1, bi+1); */
								Debug("Eliminating %d from (%d, %d)", k+1, place[k].x, tmpy*BoxY+i+1)
								game.poss.Set(tmpy*BoxY+i, Num(place[k].x-1), k+1, false)
								found++
							}
						}
					}
				} else if place[k].y > 0 {
					/* k possible only on this row */
					/* eliminate k's other possibilities from other boxes on current row */
					for tmpx = 0; tmpx < boxes_x; tmpx++ {
						if tmpx == bj { /* don't delete possibilities from current box */
							continue
						}

						for j = 0; j < BoxX; j++ {
							if game.poss.Get(Num(place[k].y-1), tmpx*BoxX+j, k+1) {
								/* Explain("%d possible only on row %d in box (%d, %d)", k+1, place[k].y, bj+1, bi+1); */
								Debug("Eliminating %d from (%d, %d)", k+1, tmpx*BoxX+j+1, place[k].y)
								game.poss.Set(Num(place[k].y-1), tmpx*BoxX+j, k+1, false)
								found++
							}
						}
					}
				}
			}
		}
	}
	return found
}

func ScanNakedTriplesGroup(game *Game, cells []Point) int {
	return ScanNakedSubsetGroup(game, cells, 3)
}

func ScanNakedQuadGroup(game *Game, cells []Point) int {
	return ScanNakedSubsetGroup(game, cells, 4)
}

func ScanNakedSubsetGroup(game *Game, cells []Point, subsetLen int) int {

	found := 0

	// 0: [cell1, cell2, cell3]
	// 1: |cell1, cell2, cell4]
	// 2: |cell1, cell2, cell5]

	nCells := len(cells)

	eliminate := func(cands CandidateSet, skipCells []int) {
		for i, cell := range cells {
			skip := false
			for skipCell := range skipCells {
				if skipCell == i {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
			for _, num := range cands {
				if game.poss.Set(Num(cell.y), Num(cell.x), num, false) {
					Debug("Naked %d: Eliminating %d from %s", subsetLen, num, cell.ToString1())
					found++
				}
			}
		}
	}

	// get subsetLen-sized combinations of all the cells in the group (indexes)
	// = all 3 or 4 four cell combinations from the group
	comb(nCells, subsetLen, func(c []int) {
		testCells := []Point{}
		cands := CandidateSet{}
		for i := range c {
			cell := cells[i]
			// form a union of the set of candidate numbers from the group of 3/4 cells
			cellCands := game.poss.Candidates(Num(cell.y), Num(cell.x))
			cands = cands.Add(cellCands)
			testCells = append(testCells, cells[i])
		}
		if len(cands) == subsetLen {
			Debug("Naked subset of %d %v in cells %v", subsetLen, cands, testCells)
			eliminate(cands, c)
		}
	})

	return found
}

func ScanNakedPairsGroup(game *Game, cells []Point) int {

	found := 0
	var subset [2]Num
	var subsetComp [2]Num
	var place Point
	var placeComp Point // holds the places (x,y) of the pair

	nCells := len(cells)
	if nCells == 0 {
		return 0
	}

	// walk through all given cells
	for cellno, cell := range cells {

		if game.board[cell.y][cell.x] != 0 {
			continue
		}

		// clear things
		subset[0] = 0
		subset[1] = 0

		// check # of candidates for this cell
		cellCands := game.poss.Candidates(Num(cell.y), Num(cell.x))
		if len(cellCands) == 2 {
			subset[0], subset[1] = cellCands[0], cellCands[1]
			place = cell
		} else {
			continue
		}

		// the cell has only two candidates... check if we can find another one like it
		subsetComp[0] = 0
		subsetComp[1] = 0
		placeComp = Point{0, 0}
		var cell2 Point

		pairFound := false
		for cellno2 := cellno + 1; cellno2 < nCells; cellno2++ {
			cell2 = cells[cellno2]

			if game.board[cell2.y][cell2.x] != 0 {
				continue
			}

			if game.poss.Equals(Num(place.y), Num(place.x), Num(cell2.y), Num(cell2.x)) {
				pairFound = true
				placeComp = cell2
				break
				// TODO: this ignores naked triples
			}
		}

		if pairFound {
			// eliminate candidates from other cells in the group
			for _, cell3 := range cells {
				if cell3.Equals(cell) || cell3.Equals(cell2) {
					continue
				}

				for _, num := range subset {
					if game.poss.Set(Num(cell3.y), Num(cell3.x), Num(num), false) {
						Explain("Naked pair {%d, %d} found in cells %s and %s", subset[0], subset[1], place.ToString1(), placeComp.ToString1())
						Debug("Eliminating %d from %s", num, cell3.ToString1())
						found++
					}
				}
			}
		}
	}

	return found
}

/*
 * Create a slice for possible cells for each number in the given set of cells
 * In the returned slice, indices are 0...NR_MAX-1
 */
func findPossibleCells(game *Game, cells []Point) [][]Point {

	// index = number
	// value = slice of Points = possible cells
	possCells := make([][]Point, NR_MAX)

	nCells := len(cells)
	for nr := Num(0); nr < NR_MAX; nr++ {
		possCells[nr] = make([]Point, 0, nCells)
	}
	// walk through all given cells to find possible cells for all numbers
	for _, cell := range cells {

		// don't check possibilities if the cell is occupied
		if game.board[cell.y][cell.x] != 0 {
			continue
		}

		// check which numbers are possible in this cell
		for nr := Num(0); nr < NR_MAX; nr++ {
			if game.poss.Get(Num(cell.y), Num(cell.x), nr+1) {
				possCells[nr] = append(possCells[nr], cell)
				// so: copy cell coordinates to poss_cells[nr][n] and increment n?
				//tmp = (struct point **) (poss_cells[nr].ptr + (poss_cells[nr].n));
				//*tmp = cell;
				//(poss_cells[nr].n)++;
			}
		}
	}
	return possCells
}

func (scanner *Scanner) ScanRowsCols(fn GroupScanFunc, name string, includeOccupied bool) int {
	found := 0

	// rows
	for j := 0; j < Y; j++ {
		cells := make([]Point, 0)
		for i := 0; i < X; i++ {
			cell := Point{y: j, x: i}
			if !includeOccupied && scanner.game.board.CellOccupied(cell) {
				continue
			}
			cells = append(cells, cell)
		}
		if len(cells) == 0 {
			continue
		}
		Debug("Performing scan `%s' on row %d", name, j+1)
		if fn(scanner.game, cells) > 0 {
			found++
		}
	}

	// cols
	for i := 0; i < X; i++ {
		cells := make([]Point, 0)
		for j := 0; j < Y; j++ {
			cell := Point{y: j, x: i}
			if !includeOccupied && scanner.game.board.CellOccupied(cell) {
				continue
			}
			cells = append(cells, cell)
		}
		if len(cells) == 0 {
			continue
		}
		Debug("Performing scan `%s' on col %d", name, i+1)
		if fn(scanner.game, cells) > 0 {
			found++
		}
	}

	return found
}

func (scanner *Scanner) ScanAllGroups(fn GroupScanFunc, name string) int {
	found := 0

	found += scanner.ScanRowsCols(fn, name, true)
	found += scanner.ScanBoxes(fn, name, true)

	return found
}

func (scanner *Scanner) ScanAllUnoccupiedGroups(fn GroupScanFunc, name string) int {
	found := 0

	found += scanner.ScanRowsCols(fn, name, false)
	found += scanner.ScanBoxes(fn, name, false)

	return found
}

func (scanner *Scanner) ScanBoxes(fn GroupScanFunc, name string, includeOccupied bool) int {

	found := 0

	// scan them boxes, one by one
	boxes_y := Y / BoxY
	boxes_x := X / BoxX

	for bj := 0; bj < boxes_y; bj++ {
		for bi := 0; bi < boxes_x; bi++ {
			cells := make([]Point, 0)
			// Debug("Scanning box (%d, %d)", bi+1, bj+1);
			// loop over the nine cells
			for j := 0; j < BoxY; j++ {
				for i := 0; i < BoxX; i++ {
					cell := Point{y: bj*BoxY + j, x: bi*BoxX + i}
					if !includeOccupied && scanner.game.board.CellOccupied(cell) {
						continue
					}
					cells = append(cells, cell)
				}
			}
			if len(cells) == 0 {
				continue
			}
			Debug("Performing scan `%s' in box (%d, %d)", name, bi+1, bj+1)
			if fn(scanner.game, cells) > 0 {
				found++
			}
		}
	}

	return found
}

func ScanHiddenTriplesGroup(game *Game, cells []Point) int {
	return ScanHiddenSubsetGroup(game, cells, 3)
}

func ScanHiddenQuadGroup(game *Game, cells []Point) int {
	return ScanHiddenSubsetGroup(game, cells, 4)
}

func ScanHiddenSubsetGroup(game *Game, cells []Point, subsetLen int) int {
	found := 0
	nNums := NR_MAX // TODO: exclude already placed numbers from the combinations

	eliminate := func(cells []Point, except []int) {
		// eliminate all other candidates from these cells
		for _, cell := range cells {
			if game.board.CellOccupied(cell) {
				continue
			}
			for num := Num(1); num < NR_MAX; num++ {
				if containsInt(except, int(num)) {
					continue
				}
				if game.poss.Set(Num(cell.y), Num(cell.x), num, false) {
					Debug("Hidden subset of %d (%v): Eliminating %d from %s", subsetLen, except, num, cell.ToString1())
					found++
				}
			}
		}
	}

	comb(nNums, subsetLen, func(c []int) {
		// find all the cells where at least one of the numbers is a candidate
		candCells := PointSet{}
		subset := make([]int, subsetLen)
		candOccurrences := make([]bool, subsetLen)
	Loop:
		for i, num := range c {
			num++ // c (combinations) starts from zero, poss wants 1...
			subset[i] = num
			for _, cell := range cells {
				if game.board.CellOccupied(cell) {
					continue
				}
				if game.poss.Get(Num(cell.y), Num(cell.x), Num(num)) {
					//Debug("%d possible in cell %v, cands: %v ", num, cell, game.poss.Candidates(Num(cell.y), Num(cell.x)))
					candOccurrences[i] = true
					// add unique cells to candCells list
					if !candCells.Contains(cell) {
						candCells = append(candCells, cell)
						if len(candCells) > subsetLen {
							break Loop
						}
					}
				}
			}
		}
		// check that all candidates occur at least once
		for _, occurs := range candOccurrences {
			if !occurs {
				return
			}
		}
		// are there exactly 3/4 cells where these candidates exist?
		if len(candCells) == subsetLen {
			Debug("Hidden subset of %d (%v) in cells %v", subsetLen, subset, candCells)
			// eliminate all other candidates from these cells
			//	for _, cell := range candCells {
			//		Debug("Candidates for cell %v: %v", cell, game.poss.Candidates(Num(cell.y), Num(cell.x)))
			//	}
			eliminate(candCells, subset)
		}
	})

	return found
}

func ScanHiddenPairsGroup(game *Game, cells []Point) int {
	found := 0
	var nr, first, second, nr2, i Num

	nCells := len(cells)
	if nCells < 1 {
		return 0
	}

	// - create a mapping for all numbers:
	//   number => possible cells
	possCells := findPossibleCells(game, cells)

	// - if number has exactly two possible cells, store as candidate and go on
	// - if another number has exactly two possible cells and the cells are the same
	//   => hidden pair. eliminate all other candidates (if there are any) in these two cells

	// we can check until NR_MAX-1 because of the inner loop checks NR_MAX
	for nr = 0; nr < NR_MAX-1; nr++ {
		if len(possCells[nr]) == 2 {
			first = nr + 1
			// search if there's another number having the same possible cells
			for nr2 = nr + 1; nr2 < NR_MAX; nr2++ {
				if len(possCells[nr2]) == 2 {
					same := true
					second = nr2 + 1
					// the cells are traversed in the same order
					for slot := 0; slot < 2; slot++ {
						// compare if the possible cells point to the same place
						if !possCells[nr][slot].Equals(possCells[nr2][slot]) {
							same = false
							break
						}
					}
					if same {
						Debug("Hidden pair {%d, %d} found", first, second)
						for slot := 0; slot < 2; slot++ {
							cell := possCells[nr][slot]
							// eliminate all other possibilities except the pair
							for i = 1; i <= NR_MAX; i++ {
								if i == first || i == second {
									continue
								}

								if game.poss.Set(Num(cell.y), Num(cell.x), i, false) {
									Explain("Eliminating %d from %s", i, cell.ToString1())
									found++
								}
							}
						}
					}
				}
			}
		}
	}

	return found
}

func ScanBoxLineGroup(game *Game, cells []Point) int {

	found := 0
	var nr Num
	var slot, i int
	box := -1

	if len(cells) == 0 {
		return 0
	}

	eliminateFromBoxExcluding := func(nr Num, box int, skip []Point) int {
		found := 0
		check := 1

		j_beg := (box / BoxY) * BoxY
		i_beg := (box % BoxX) * BoxX

		// debug("Box %d, i_beg %d, j_beg %d", box + 1, i_beg+1, j_beg+1);
		for j := j_beg; j < j_beg+BoxY; j++ {
			for i := i_beg; i < i_beg+BoxX; i++ {
				check = 1
				nSkip := len(skip)
				for slot := 0; slot < nSkip; slot++ {
					cell := skip[slot]
					if cell.y == j && cell.x == i {
						// Debug("Skipping cell (%d, %d)", i+1,j+1);
						check = 0
						break
					}
				}
				if check == 0 {
					continue
				}

				if game.poss.Set(Num(j), Num(i), nr, false) {
					Debug("Eliminating %d from (%d, %d) in box %d", nr, i+1, j+1, box+1)
					found++
				}
			}
		}
		return found
	}

	//  - create a mapping for all numbers:
	//  number => possible cells
	possCells := findPossibleCells(game, cells)

	for nr = 0; nr < NR_MAX; nr++ {
		nPoss := len(possCells[nr])
		if nPoss > BoxX {
			continue
		}

		box = -1

		for slot = 0; slot < nPoss; slot++ {
			cell := possCells[nr][slot]
			i = getBox(cell.y, cell.x)
			if box == -1 {
				box = i
			} else if box != i {
				// nr is possible in two different boxes
				box = -1
				break
			}
		}

		if box >= 0 {
			if eliminateFromBoxExcluding(nr+1, box, possCells[nr]) > 0 {
				Explain("%d possible only in box %d in row or col", nr+1, box+1)
				found++
			}
		}
	}

	return found
}
