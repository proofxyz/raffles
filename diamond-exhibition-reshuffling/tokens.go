package main

import "math/rand"

type token struct {
	TokenID   int
	ProjectID int
}

// tokens is a slice of token structs, typically used to represent an allocation to a single address.
type tokens []token

// copy returns a deep copy of the token list.
func (ts tokens) copy() tokens {
	c := make(tokens, len(ts))
	copy(c, ts)
	return c
}

// numTokens returns the number of tokens in the allocation.
func (ts tokens) numTokens() int {
	return len(ts)
}

// numProjects returns the number of distinct projects in the allocation.
func (ts tokens) numProjects() int {
	var n int
	for _, x := range ts.numPerProject() {
		if x > 0 {
			n++
		}
	}
	return n
}

// numPerProject computes the number of tokens per project.
func (ts tokens) numPerProject() projectsVector {
	var num projectsVector
	for _, v := range ts {
		num[v.ProjectID]++
	}
	return num
}

// variability returns a basic measure for the variability of the allocation.
func (ts tokens) variability() float64 {
	return float64(ts.numProjects()) / float64(ts.numTokens())
}

// numSameTokenID returns the number of tokens with the same tokenID that are present in both token lists.
func (ts tokens) numSameTokenID(other tokens) int {
	var num int

	seen := make(map[int]struct{})
	for _, a := range ts {
		seen[a.TokenID] = struct{}{}
	}

	for _, b := range other {
		if _, ok := seen[b.TokenID]; ok {
			num++
		}
	}

	return num
}

// numInSameProjects returns the number of tokens whose projects are also present in the other allocation.
func (ts tokens) numInSameProjects(other tokens) int {
	o := other.numPerProject()

	var total int
	for pID, n := range ts.numPerProject() {
		if n == 0 {
			continue
		}
		if o[pID] > 0 {
			total += n
		}
	}
	return total
}

// numInDuplicateProjects returns the number of tokens whose projects are present more than once in the allocation.
func (ts tokens) numInDuplicateProjects() int {
	var n int
	for _, num := range ts.numPerProject() {
		if num > 1 {
			n += num - 1
		}
	}
	return n
}

// numGrails returns the number of grail tokens with the given allocation.
func (ts tokens) numGrails() int {
	ns := ts.numPerProject()
	var n int
	n += ns[11]
	n += ns[17]
	n += ns[19]

	return n
}

// drawTokenIdx draws a random token index from the allocation.
func (ts tokens) drawTokenIdx(rng *rand.Rand) int {
	return rng.Intn(ts.numTokens())
}
