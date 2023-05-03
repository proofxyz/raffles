package main

import (
	"math/rand"
)

// tokenSet is a set of tokens represented as a map from token id to project id.
type tokenSet map[int]int

// numPerProject computes the number of tokens per project.
func (s tokenSet) numPerProject() projectsVector {
	var num projectsVector
	for _, t := range s {
		num[t]++
	}
	return num
}

// copy returns a deep copy of the token set.
func (s tokenSet) copy() tokenSet {
	c := make(map[int]int)
	for k, v := range s {
		c[k] = v
	}
	return c
}

// numIntersection computes the number of tokens in the intersection of two token sets.
func (s tokenSet) numIntersection(other tokenSet) int {
	var num int
	for k := range s {
		if _, ok := other[k]; ok {
			num++
		}
	}
	return num
}

// allocation is a set of tokens allocated to a given participant.
type allocation struct {
	tokens               tokenSet       // the tokens in the allocation
	cachedNumPerProjects projectsVector // number of tokens per project, cached for performance
	isPool               bool           // flag to indicate whether the allocation is the pool (disables the score function)
}

// newAllocation creates a new empty allocation.
func newAllocation() *allocation {
	return &allocation{
		tokens: make(tokenSet),
	}
}

// nextFakeTokenId is a counter used to generate fake token ids.
var nextFakeTokenId int

// newAllocationFromProjects creates a new allocation from a slice of project ids.
func newAllocationFromProjects(xs []int) *allocation {
	tokens := make(map[int]int)
	for _, x := range xs {
		tokens[nextFakeTokenId] = x
		nextFakeTokenId++
	}
	return newAllocationFromTokens(tokens)
}

// newAllocationFromTokens creates a new allocation from a token set.
func newAllocationFromTokens(t tokenSet) *allocation {
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

// NumPerProject returns the number of tokens per project from cache.
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

// drawToken draws a random tokenId from the allocation.
func (a *allocation) drawToken(src rand.Source) int {
	rand := rand.New(src).Intn(a.numTokens())
	i := 0
	for id := range a.tokens {
		if i == rand {
			return id
		}
		i++
	}
	panic("drawing a random token failed")
}

// score returns a score for the current allocation, where higher is better.
func (current *allocation) score(initial *allocation) float64 {
	if current.isPool {
		return 0
	}

	c := current.numPerProject()
	i := initial.numPerProject()

	var s int

	s -= c.smul(c)        // regularisation term to penalise getting duplicate projects
	s -= c.smul(i.mask()) // penalise getting tokens from the initial projects back

	s -= current.tokens.numIntersection(initial.tokens) // penalise getting the initial token ids back

	return float64(s)
}

// moveToken moves a token from one allocation to another.
func (a *allocation) moveToken(tokenId int, target *allocation) {
	projectId := a.tokens[tokenId]

	target.tokens[tokenId] = projectId
	delete(a.tokens, tokenId)

	// update cached number of tokens per project without recomputation for performance
	target.cachedNumPerProjects[projectId]++
	a.cachedNumPerProjects[projectId]--
}

// numIdenticalInitialTokens returns the number of tokens that were also present
// in the initial allocation.
func (a *allocation) numIdenticalInitialTokens(initial *allocation) int {
	var n int
	for id, _ := range a.tokens {
		if _, ok := initial.tokens[id]; ok {
			n++
		}
	}
	return n
}

// numInInitialProjects returns the number of tokens in projects that were
// already present in the initial allocation.
func (a *allocation) numInInitialProjects(initial *allocation) int {

	fin := a.numPerProject()
	init := initial.numPerProject()

	var n int
	for id, num := range init {
		if num > 0 && fin[id] > 0 {
			n += fin[id]
		}
	}
	return n
}

// numInDuplicateProjects returns the number of tokens in projects that are
// present more than once in the allocation.
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
