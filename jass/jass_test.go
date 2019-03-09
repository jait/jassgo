/* vim: set sts=4 sw=4 ts=4 noet: */
/**
 * jass - just another sudoku solver
 * (C) 2005-2019 Jari Tenhunen <jait@iki.fi>
 *
 * Go version copyright 2019
 */

package jass

import "testing"

func TestCandidateSet(t *testing.T) {

	empty := CandidateSet{}
	set1 := CandidateSet{1, 2, 3}
	if !set1.Equals(CandidateSet{3, 2, 1}) {
		t.Errorf("CandidateSet.Equals(): expected true")
	}
	if !set1.Equals(set1) {
		t.Errorf("CandidateSet.Equals(): expected true")
	}
	if set1.Equals(CandidateSet{1, 2, 4}) {
		t.Errorf("CandidateSet.Equals(): expected false")
	}
	if set1.Equals(empty) {
		t.Errorf("CandidateSet.Equals(): expected false")
	}
	if !empty.Equals(CandidateSet{}) {
		t.Errorf("CandidateSet.Equals(): expected true")
	}
	if empty.Contains(1) || empty.Contains(0) {
		t.Errorf("CandidateSet.Contains(): expected false")
	}
	if !(set1.Contains(1) && set1.Contains(2) && set1.Contains(3)) {
		t.Errorf("CandidateSet.Contains(): expected true")
	}
	if set1.Contains(0) || set1.Contains(4) {
		t.Errorf("CandidateSet.Contains(): expected false")
	}
	// addition/union
	set2 := set1.Add(CandidateSet{4, 5})
	if !set2.Equals(CandidateSet{5, 4, 3, 2, 1}) {
		t.Errorf("CandidateSet.Equals(): expected true")
	}
	set2 = set1.Add(CandidateSet{2, 3, 4})
	if len(set2) != 4 {
		t.Errorf("length incorrect after CandidateSet.Add()")
	}
	if !set2.Equals(CandidateSet{4, 3, 2, 1}) {
		t.Errorf("CandidateSet.Equals(): expected true")
	}
	// original must be unaltered after Add
	if !set1.Equals(CandidateSet{3, 2, 1}) {
		t.Errorf("CandidateSet.Equals(): expected true")
	}
}

func TestPoss(t *testing.T) {
	p := NewPoss()
	for i := 1; i <= NR_MAX; i++ {
		if p.Get(1, 2, Num(i)) != true {
			t.Errorf("Poss initial value not true")
		}
	}
	if !p.Candidates(2, 3).Equals(CandidateSet{1, 2, 3, 4, 5, 6, 7, 8, 9}) {
		t.Errorf("Poss initial candidate set bad")
	}
	// TODO: Set, Get, GetOnly, Equals...
}
