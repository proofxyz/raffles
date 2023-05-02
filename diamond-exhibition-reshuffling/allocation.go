package main

import (
	"math/rand"
)

const numProjects = 21

type (
	vec [numProjects]int
)

type allocation struct {
	tokens        map[int]int
	numPerProject *vec
}

func (a allocation) NumPerProject() *vec {
	return a.numPerProject
}

var fakeTokenId int

func newAllocationFromSubmission(xs []int) *allocation {
	tokens := make(map[int]int)
	for _, x := range xs {
		tokens[fakeTokenId] = x
		fakeTokenId++
	}
	return newAllocationFromTokens(tokens)
}

func newAllocationFromTokens(tokens map[int]int) *allocation {
	return &allocation{
		tokens:        tokens,
		numPerProject: computeNumPerProject(tokens),
	}
}

func computeNumPerProject(tokens map[int]int) *vec {
	var num vec
	for _, t := range tokens {
		num[t]++
	}
	return &num
}

func (a *vec) smul(b *vec) int {
	var res int
	for i := range a {
		res += a[i] * b[i]
	}
	return res
}

func (a *vec) add(b *vec) *vec {
	var res vec
	for i := range a {
		res[i] = a[i] + b[i]
	}
	return &res
}

func (a *vec) mask() *vec {
	var res vec
	for i, x := range a {
		if x > 0 {
			res[i] = 1
		}
	}
	return &res
}

func (a allocation) numTokens() int {
	return len(a.tokens)
}

func (a allocation) copy() *allocation {
	var c allocation
	c.tokens = make(map[int]int)
	for k, v := range a.tokens {
		c.tokens[k] = v
	}
	c.numPerProject = &vec{}
	copy(c.numPerProject[:], a.numPerProject[:])
	return &c
}

func (a *allocation) variability() float64 {
	return float64(a.numDistinct()) / float64(a.numTokens())
}

func (a *allocation) numDistinct() int {
	var n int
	for _, x := range a.NumPerProject() {
		if x > 0 {
			n++
		}
	}
	return n
}

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

func (current *allocation) score(initial *allocation) float64 {
	c := current.NumPerProject()
	i := initial.NumPerProject()
	s := -c.smul(c.add(i.mask()))
	// equals
	// - n_current * n_init_masked (term to penalise getting input projects back)
	// - n_current * n_current (regularisation term to penalise getting duplicate projects)

	// additional penalty for getting the same tokenIds back
	for i := range current.tokens {
		if _, ok := initial.tokens[i]; ok {
			// punishing really hard since we really want to avoid this
			s -= 10
		}
	}

	return float64(s)
}

func (a *allocation) moveToken(tokenId int, target *allocation) {
	projectId := a.tokens[tokenId]

	delete(a.tokens, tokenId)
	a.numPerProject[projectId]--

	target.tokens[tokenId] = projectId
	target.numPerProject[projectId]++
}

func (a *allocation) numIdenticalInitialTokens(initial *allocation) int {
	var n int
	for id, _ := range a.tokens {
		if _, ok := initial.tokens[id]; ok {
			n++
		}
	}
	return n
}

func (a *allocation) numInInitialProjects(initial *allocation) int {

	fin := a.NumPerProject()
	init := initial.NumPerProject()

	var n int
	for id, num := range init {
		if num > 0 && fin[id] > 0 {
			n += fin[id]
		}
	}
	return n
}

func (a *allocation) numInDuplicateProjects() int {
	var n int
	for _, num := range a.NumPerProject() {
		if num > 1 {
			n += num - 1
		}
	}
	return n
}

type allocations []*allocation

func (a allocations) avgVariability() float64 {
	var res float64
	for _, x := range a {
		res += x.variability()
	}
	return res / float64(len(a))
}

func (current allocations) computeScore(initial allocations) float64 {
	var res float64
	for i, c := range current {
		res += c.score(initial[i])
	}
	return res
}
