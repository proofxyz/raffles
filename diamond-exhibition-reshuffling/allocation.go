package main

import (
	"math/rand"
)

// allocation is a set of tokens allocated to a participant.
type allocation struct {
	tokens               tokens         // the tokens in the allocation
	cachedNumPerProjects projectsVector // number of tokens per project, cached for performance
	isPool               bool           // flag to indicate whether the allocation is the pool (disables the score function)
}

// newAllocation creates a new allocation from a list of tokens.
func newAllocation(t tokens) *allocation {
	return &allocation{
		tokens:               t,
		cachedNumPerProjects: t.numPerProject(),
		isPool:               false,
	}
}

// numTokens returns the number of tokens in the allocation.
func (a *allocation) numTokens() int {
	return len(a.tokens)
}

// numPerProject returns the number of tokens per project from cache.
func (a *allocation) numPerProject() projectsVector {
	return a.cachedNumPerProjects
}

// copy returns a deep copy of the allocation.
func (a *allocation) copy() *allocation {
	var c allocation
	c.isPool = a.isPool
	c.tokens = a.tokens.copy()
	c.cachedNumPerProjects = a.cachedNumPerProjects.copy()
	return &c
}

// variability returns a basic measure for the variability of the allocation.
func (a *allocation) variability() float64 {
	return float64(a.numProjects()) / float64(a.numTokens())
}

// numDistinct returns the number of distinct projects in the allocation.
func (a *allocation) numProjects() int {
	var n int
	for _, x := range a.numPerProject() {
		if x > 0 {
			n++
		}
	}
	return n
}

// drawTokenIdx draws a random token index from the allocation.
func (a *allocation) drawTokenIdx(src rand.Source) int {
	return rand.New(src).Intn(a.numTokens())
}

// score returns a score for the current allocation, where higher is better.
func (current *allocation) score(initial *allocation) float64 {
	if current.isPool {
		return 0
	}

	c := current.numPerProject()
	i := initial.numPerProject()

	var s int
	s -= c.smul(c)                                     // regularisation term to penalise getting duplicate projects
	s -= c.smul(i.mask())                              // penalise getting tokens from the initial projects back
	s -= current.tokens.numSameTokenID(initial.tokens) // penalise getting the initial token ids back
	return float64(s)
}

func (a *allocation) swapToken(ia int, b *allocation, ib int) {
	// update cached number of tokens per project directly for performance
	projectIdA := a.tokens[ia].ProjectID
	a.cachedNumPerProjects[projectIdA]--
	b.cachedNumPerProjects[projectIdA]++

	projectIdB := b.tokens[ib].ProjectID
	b.cachedNumPerProjects[projectIdB]--
	a.cachedNumPerProjects[projectIdB]++

	a.tokens.swapToken(ia, b.tokens, ib)
}

// numSameTokenID returns the number of tokens with the same tokenID that are present in both allocations.
func (a *allocation) numSameTokenID(other *allocation) int {
	return a.tokens.numSameTokenID(other.tokens)
}

// numSameProjects returns the number of tokens whose projects are also present in the other allocation.
func (a *allocation) numInSameProjects(other *allocation) int {
	o := other.numPerProject()

	var total int
	for pID, n := range a.numPerProject() {
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
func (a *allocation) numInDuplicateProjects() int {
	var n int
	for _, num := range a.numPerProject() {
		if num > 1 {
			n += num - 1
		}
	}
	return n
}

// allocations is a slice of allocations. This is a convenience type to
// perform actions over multiple of allocations.
type allocations []*allocation

// avgVariability computes the average variability of a slice of allocations.
func (a allocations) avgVariability() float64 {
	var res float64
	for _, x := range a {
		res += x.variability()
	}
	return res / float64(len(a))
}

// score computes the sum of the scores of a slice of allocations.
func (current allocations) score(initial allocations) float64 {
	var res float64
	for i, c := range current {
		res += c.score(initial[i])
	}
	return res
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

// copyShallow returns a shallow copy of the allocations.
func (a allocations) copyShallow() allocations {
	c := make(allocations, len(a))
	copy(c, a)
	return c
}

// duplicateTokenIDs returns a list of token ids that are present more than once
func (as allocations) duplicateTokenIDs() []int {
	tokenIDs := make(map[int]struct{})
	var dupes []int

	for _, a := range as {
		for _, t := range a.tokens {
			if _, ok := tokenIDs[t.TokenID]; ok {
				dupes = append(dupes, t.TokenID)
			}
			tokenIDs[t.TokenID] = struct{}{}
		}
	}
	return dupes
}
