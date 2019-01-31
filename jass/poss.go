/* vim: set sts=4 sw=4 et: */
/**
 * jass - just another sudoku solver
 * (C) 2005, 2006 Jari Tenhunen <jait@iki.fi>
 */

package jass

import "reflect"

type Poss [][][]bool
type CandidateSet []Num

func (set CandidateSet) Equals(other CandidateSet) bool {
	if len(set) != len(other) {
		return false
	}
	for _, num := range set {
		if !other.Contains(num) {
			return false
		}
	}
	return true
}

func (set CandidateSet) Contains(num Num) bool {
	for _, n := range set {
		if n == num {
			return true
		}
	}
	return false
}

func (set CandidateSet) Add(other CandidateSet) CandidateSet {
	newSet := set
	for _, n := range other {
		if !set.Contains(n) {
			newSet = append(newSet, n)
		}
	}
	return newSet
}

func NewPoss() Poss {
	poss := make(Poss, Y)
	for i := range poss {
		poss[i] = make([][]bool, X)
		for j := range poss[i] {
			poss[i][j] = make([]bool, NR_MAX)
			for k := Num(1); k <= NR_MAX; k++ {
				poss.Set(Num(i), Num(j), k, true)
			}
		}
	}
	return poss
}

/*
 * Get possibility (true, false) of candidate (1...NR_MAX) in cell (y,x)
 */
func (p *Poss) Get(y, x, candidate Num) bool {
	return (*p)[y][x][candidate-1]
}

/*
 * Get candidate numbers for cell (y,x)
 */
func (p *Poss) Candidates(y, x Num) CandidateSet {
	res := make([]Num, 0, NR_MAX)
	for k := Num(1); k <= NR_MAX; k++ {
		if (*p)[y][x][k-1] {
			res = append(res, k)
		}
	}
	return res
}

/*
 * Sets candidate to be possible or not possible for cell (x,y)
 *
 */
func (p *Poss) Set(y, x, candidate Num, val bool) bool {
	prev := (*p)[y][x][candidate-1]
	(*p)[y][x][candidate-1] = val
	return prev
}

/* bit-vector version */
/*
   char prev = is_poss(y, x, candidate);
   if (val)
       poss[y][x] |= (1 << candidate); // set
   else
       poss[y][x] &= ~(1 << candidate); // clear

   return prev;
*/

/**
 * Return the only possibility (1...NR_MAX) or zero when there are zero or more than one
 * possibility
 */
func (p *Poss) GetOnly(y, x Num) Num {
	var val, k Num
	for k = 0; k < NR_MAX; k++ {
		if p.Get(y, x, k) {
			if val != 0 {
				/* at least two possibilities */
				val = 0
				break
			}
			val = k + 1 /* because val is 1... and index k is 0... */
		}
	}
	return val
}

func (p *Poss) Equals(y1, x1, y2, x2 Num) bool {
	return reflect.DeepEqual((*p)[y1][x1], (*p)[y2][x2])
}
