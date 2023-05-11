package main

import (
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gocarina/gocsv"
)

// allocation is a set of tokens allocated to a participant.
type allocation struct {
	tokens                              // the tokens in the allocation
	owner                common.Address // the owner of the allocation
	cachedNumPerProjects projectsVector // number of tokens per project, cached for performance
	isPool               bool           // flag to indicate whether the allocation is the pool submitted by PROOF (disables the score function)
}

// newAllocation creates a new allocation from a list of tokens.
func newAllocation(addr common.Address, t tokens) *allocation {
	return &allocation{
		tokens:               t,
		owner:                addr,
		cachedNumPerProjects: t.numPerProject(),
		isPool:               false,
	}
}

// numPerProject returns the number of tokens per project from cache.
func (a *allocation) numPerProject() projectsVector {
	return a.cachedNumPerProjects
}

// copy returns a deep copy of the allocation.
func (a *allocation) copy() *allocation {
	c := *a
	c.tokens = a.tokens.copy()
	c.cachedNumPerProjects = a.cachedNumPerProjects.copy()
	return &c
}

// score returns a score for the current allocation, where higher is better.
func (a *allocation) score(initial *allocation) float64 {
	var s int
	for _, p := range a.scorePenalties(initial) {
		s -= p
	}
	return float64(s)
}

// scorePenalties returns the individual components that make up an allocation
// score. The penalties are positive whereas the final score is negative. It is
// abstracted for more precise testing.
func (a *allocation) scorePenalties(initial *allocation) [3]int {
	if a.isPool {
		// disabling the score function for the PROOF-issued pool,
		// because it does not care about the tokens it ends up with.
		// Returning the theoretical maximum score for consistency.
		return [3]int{a.numTokens(), 0, 0}
	}

	n := a.numPerProject()
	return [3]int{
		n.smul(n),                                // (1) regularisation term to penalise getting duplicate projects
		n.smul(initial.numPerProject().asMask()), // (2) penalise getting tokens from the initial projects back
		a.tokens.numSameTokenID(initial.tokens),  // (3) penalise getting the initial token ids back
	}
}

func (a *allocation) swapToken(b *allocation, ia int, ib int) {
	// update cached number of tokens per project directly for performance
	pA := a.tokens[ia].ProjectID
	a.cachedNumPerProjects[pA]--
	b.cachedNumPerProjects[pA]++

	pB := b.tokens[ib].ProjectID
	b.cachedNumPerProjects[pB]--
	a.cachedNumPerProjects[pB]++

	a.tokens[ia], b.tokens[ib] = b.tokens[ib], a.tokens[ia]
}

// allocations is a slice of allocations. This is a convenience type to
// perform actions over multiple of allocations.
type allocations []*allocation

// copyShallow returns a shallow copy of the allocations.
func (as allocations) copyShallow() allocations {
	cp := make(allocations, len(as))
	copy(cp, as)
	return cp
}

// numTokens computes the total number of tokens in a slice of allocations.
func (as allocations) numTokens() int {
	var total int
	for _, a := range as {
		total += a.numTokens()
	}
	return total
}

// numPerProject computes the total number of tokens per project in a slice of allocations.
func (as allocations) numPerProject() projectsVector {
	var total projectsVector
	for _, a := range as {
		total = total.add(a.numPerProject())
	}
	return total
}

// score computes the sum of the scores of a slice of allocations.
func (as allocations) score(initial allocations) float64 {
	var res float64
	for i, a := range as {
		res += a.score(initial[i])
	}
	return res
}

// avgVariability computes the average variability of a slice of allocations.
func (as allocations) avgVariability() float64 {
	var res float64
	for _, a := range as {
		res += a.variability()
	}
	return res / float64(len(as))
}

// duplicateTokenIDs returns a list of token ids that are present more than once
func (as allocations) duplicateTokenIDs() []int {
	tokenIDs := make(map[int]bool)
	var dupes []int

	for _, a := range as {
		for _, t := range a.tokens {
			if tokenIDs[t.TokenID] {
				dupes = append(dupes, t.TokenID)
			}
			tokenIDs[t.TokenID] = true
		}
	}
	return dupes
}

// writeCSV writes the allocations as CSV, each row containing an address and assigned token id plus corresponding project id.
func (as allocations) writeCSV(w io.Writer) error {
	type Transfer struct {
		Addr      common.Address
		TokenId   int
		ProjectId int
	}

	var transfers []Transfer
	for _, a := range as {
		for _, token := range a.tokens {
			transfers = append(transfers, Transfer{
				Addr:      a.owner,
				TokenId:   token.TokenID,
				ProjectId: token.ProjectID,
			})

		}
	}

	if err := gocsv.Marshal(transfers, w); err != nil {
		return fmt.Errorf("gocsv.Marshal(%T, %T): %v", transfers, w, err)
	}
	return nil
}
