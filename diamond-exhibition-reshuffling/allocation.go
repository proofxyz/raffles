package main

import (
	"math/rand"
)

const numProjects = 21

type (
	vec [numProjects]int
)

type allocation map[int]int

func (a allocation) numPerProject() *vec {
	var num vec
	for _, t := range a {
		num[t]++
	}
	return &num
}

var fakeTokenId int

func newAllocationFromSubmission(xs []int) *allocation {
	a := make(allocation)
	for _, x := range xs {
		a[fakeTokenId] = x
		fakeTokenId++
	}

	return &a
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
	return len(a)
}

func (a allocation) copy() *allocation {
	c := make(allocation)
	for k, v := range a {
		c[k] = v
	}
	return &c
}

func (a *allocation) variability() float64 {
	return float64(a.numDistinct()) / float64(a.numTokens())
}

func (a *allocation) numDistinct() int {
	var n int
	for _, x := range a.numPerProject() {
		if x > 0 {
			n++
		}
	}
	return n
}

func (a *allocation) drawToken(src rand.Source) int {
	rand := rand.New(src).Intn(a.numTokens())
	i := 0
	for id := range *a {
		if i == rand {
			return id
		}
		i++
	}

	panic("drawing a random token failed")
}

func (current *allocation) score(initial *allocation) float64 {
	c := current.numPerProject()
	i := initial.numPerProject()
	s := -c.smul(c.add(i.mask()))

	// additional penalty for getting the same tokenIds back
	for i := range *current {
		if _, ok := (*initial)[i]; ok {
			s -= 1
		}
	}

	return float64(s)
}

func (a *allocation) moveToken(tokenId int, target *allocation) {
	(*target)[tokenId] = (*a)[tokenId]
	delete(*a, tokenId)
}

type allocations []*allocation

func (a allocations) avgVariability() float64 {
	var res float64
	for _, x := range a {
		res += x.variability()
	}
	return res / float64(len(a))
}
