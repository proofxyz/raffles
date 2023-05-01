package main

import (
	"math/rand"

	"github.com/ccssmnn/hego"
)

const numProjects = 21

type (
	allocation [numProjects]int
)

func newAllocationFromSubmission(xs []int) *allocation {
	var b allocation
	for _, x := range xs {
		b[x]++
	}
	return &b
}

func newAllocationFromPartial(xs []int) *allocation {
	var b allocation
	for i, x := range xs {
		b[i] = x
	}
	return &b
}

func (a *allocation) smul(b *allocation) int {
	var res int
	for i := range a {
		res += a[i] * b[i]
	}
	return res
}

func (a *allocation) mul(b *allocation) *allocation {
	var res allocation
	for i := range a {
		res[i] = a[i] * b[i]
	}
	return &res
}

func (a *allocation) add(b *allocation) *allocation {
	var res allocation
	for i := range a {
		res[i] = a[i] + b[i]
	}
	return &res
}

func (a *allocation) sub(b *allocation) *allocation {
	var res allocation
	for i := range a {
		res[i] = a[i] - b[i]
	}
	return &res
}

func (b *allocation) sum() int {
	var s int
	for _, x := range b {
		s += x
	}
	return s
}

func (b *allocation) copy() *allocation {
	var c allocation
	copy(c[:], b[:])
	return &c
}

func (b *allocation) variablility() float64 {
	return float64(b.numDistinct()) / float64(b.sum())
}

func (b *allocation) numDistinct() int {
	var n int
	for _, x := range b {
		if x > 0 {
			n++
		}
	}
	return n
}

func (a *allocation) drawProject(src rand.Source) int {
	rng := rand.New(src)
	rand := rng.Intn(a.sum())

	for i, v := range a {
		if rand < v {
			return i
		}
		rand -= v
	}

	panic("Sampling project failed")
}

func (a *allocation) mask() *allocation {
	var res allocation
	for i, x := range a {
		if x > 0 {
			res[i] = 1
		}
	}
	return &res
}

func score(initial, current *allocation) int {
	// return -current.smul(current.add(initial))
	return -current.smul(current.add(initial.mask()))
}

type allocations []*allocation

func (a allocations) avgVariability() float64 {
	var res float64
	for _, x := range a {
		res += x.variablility()
	}
	return res / float64(len(a))
}

type state struct {
	initial allocations
	current allocations
	src     rand.Source
}

func (s state) numAllocations() int {
	return len(s.initial)
}

func (s state) numTokens() int {
	var total int
	for _, x := range s.current {
		total += x.sum()
	}
	return total
}

func (s state) score() float64 {
	var res float64
	for i, c := range s.current {
		res += float64(score(s.initial[i], c))
	}
	return res
}

func (s state) neighbor() state {
	rng := rand.New(s.src)

	current := make([]*allocation, s.numAllocations())
	copy(current, s.current)

	ia := rng.Intn(s.numAllocations())
	ib := rng.Intn(s.numAllocations())
	for ia == ib {
		return s
	}

	ja := s.current[ia].drawProject(s.src)
	jb := s.current[ib].drawProject(s.src)
	if ja == jb {
		return s
	}

	current[ia] = current[ia].copy()
	current[ib] = current[ib].copy()

	current[ia][ja]--
	current[ib][ja]++
	current[ia][jb]++
	current[ib][jb]--

	s2 := s
	s2.current = current
	return s2
}

func (s state) Neighbor() hego.AnnealingState {
	return s.neighbor()
}

func (s state) Energy() float64 {
	return -float64(s.score())
}
