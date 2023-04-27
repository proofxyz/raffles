package main

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/soypat/mu8"
	"github.com/soypat/mu8/genetic"
)

// An allocator describes the parameters of a search to allocate the preferences
// of k entrants to n buckets of variable size.
type allocator struct {
	available       []uint64 // size of each choice bucket; dimension n
	preferences     [][]int  // preferences per entrant; dimension k x n
	fittestPossible int
}

// init performs sense checks on the allocator and computes the best possible
// ordering score, whereby every entrant would receive their first preference.
func (a *allocator) init() error {
	// See allocator definition for definition of dimensions.
	n := len(a.available)
	k := len(a.preferences)

	for i, prefs := range a.preferences {
		if got, want := len(prefs), n; got != want {
			return fmt.Errorf("preferences[%d] of length %d; want %d (number of available buckets)", i, got, want)
		}

		seen := make([]bool, len(prefs))
		for _, p := range prefs {
			if seen[p] {
				return fmt.Errorf("preferences[%d] duplicate entry %d", i, p)
			}
			seen[p] = true
		}
	}

	// Each entrant can be up to (n-1) away from their primary preference.
	a.fittestPossible = k * (n - 1)
	return nil
}

// islands returns genetic.Islands with a random population of orderings on
// each. Any existing orderings are included before allocating new random ones.
// If len(existing)>nOrderings, islands() panics.
func (a *allocator) islands(nIslands, nOrderings int, src rand.Source, existing ...*ordering) genetic.Islands[*ordering] {
	if n := len(existing); n > nOrderings {
		panic(fmt.Sprintf("%d existing orderings provided when only requesting %d", n, nOrderings))
	}
	allocs := append(existing, a.newOrderings(nOrderings-len(existing), src)...)
	return genetic.NewIslands(nIslands, allocs, src, a.newOrdering)
}

// newOrderings returns `num` new, randomly shuffled orderings.
func (a *allocator) newOrderings(num int, src rand.Source) []*ordering {
	rng := rand.New(src)

	var os []*ordering
	for i := 0; i < num; i++ {
		o := a.newOrdering()
		rng.Shuffle(len(o.order), o.swap)
		os = append(os, o)
	}
	return os
}

// newOrdering returns a single new ordering with unshuffled order.
func (a *allocator) newOrdering() *ordering {
	order := make([]int, len(a.preferences))
	for i := 0; i < len(order); i++ {
		order[i] = i
	}
	return &ordering{
		allocator: a,
		order:     order,
	}
}

// An ordering is describes the order in which entrants are allowed to choose
// from remaining allocations. A final allocation is determined by iterating
// over the order and selecting the highest-preference remaining allocation for
// that entrant.
type ordering struct {
	*allocator
	order []int // index into allocator.preferences, therefore dimension k.
}

var _ interface {
	mu8.Gene
	mu8.Genome
} = &ordering{}

// swap does what it says on the tin.
func (o *ordering) swap(i, j int) {
	o.order[i], o.order[j] = o.order[j], o.order[i]
}

// Simulate returns the fitness score of the ordering. A perfect score sees
// every participant receive their first preference. For every step down in
// allocated preference, the score is reduced by one.
//
// This makes the search effectively a maximiser of utility with constant
// marginal utility. While allowing entrants to state their utility may have had
// marginally better results (get it?), this would have complicated the user
// experience.
//
// Cursory profiling shows that this is the greatest contributor to running the
// algorithm, particular memory allocation. However it completes in ~10 minutes
// so we haven't performed any optimisation.
func (o *ordering) Simulate(context.Context) float64 {
	allocated := make([]uint64, len(o.available))

	var delta int
	for _, idx := range o.order {
		for d, pref := range o.preferences[idx] {
			if allocated[pref] < o.available[pref] {
				allocated[pref]++
				delta += d
				break
			}
		}
	}

	return float64(o.fittestPossible - delta)
}

// Functions required by the genetic-algorithm library, which typically has
// multiple "Genes" per "Genome". In this case we treat an ordering as a Genome
// with a single Gene, the interface for which is also implemented by ordering.
func (o *ordering) GetGene(i int) mu8.Gene { return o }
func (*ordering) Len() int                 { return 1 }

// Splice randomly splices a second ordering into the current one. Typically
// this would be performed by randomly adding parts of p into o, but this would
// result in an invalid order. We instead perform a merge (as in mergesort) with
// random selection from each Gene at each merger.
func (o *ordering) Splice(rng *rand.Rand, p mu8.Gene) {
	switch p := p.(type) {
	case *ordering:
		k := len(o.order)
		merged := make([]int, k)
		seen := make([]bool, k)

		var (
			entrant    int
			oIdx, pIdx int
			coin       uint64 // each bit is used as a flip
		)
		for i := 0; i < k; i++ {
			if i%64 == 0 {
				coin = rng.Uint64()
			}
			flip := coin&1 == 0
			coin >>= 1

			if (flip && oIdx < k) || pIdx == k {
				entrant = o.order[oIdx]
				oIdx++
			} else {
				entrant = p.order[pIdx]
				pIdx++
			}

			if seen[entrant] {
				i-- // undo the for loop
			} else {
				seen[entrant] = true
				merged[i] = entrant
			}
		}

		for i := 0; i < k; i++ {
			if !seen[i] {
				panic(fmt.Sprintf("corrupted splice; index %d not seen", i))
			}
		}

		o.order = merged

	default:
		// implies a bug in the mu8 package, passing an incompatible Gene.
		panic(fmt.Sprintf("%T.Splice(%T)", o, p))
	}
}

// CloneFrom overwrites o with p.
func (o *ordering) CloneFrom(p mu8.Gene) {
	switch p := p.(type) {
	case *ordering:
		o.allocator = p.allocator
		o.order = append([]int{}, p.order...)
	default:
		// implies a bug in the mu8 package, passing an incompatible Gene.
		panic(fmt.Sprintf("%T.CloneFrom(%T)", o, p))
	}
}

// Mutate randomly mutates the ordering, swapping a random selection of
// neighbours and a random selection of arbitrary entrants.
func (o *ordering) Mutate(rng *rand.Rand) {
	for i, n := 0, rng.Intn(50); i < n; i++ {
		j := rng.Intn(len(o.order) - 1)
		o.swap(j, j+1)
	}
	for i, n := 0, rng.Intn(50); i < n; i++ {
		j := rng.Intn(len(o.order))
		k := rng.Intn(len(o.order))
		o.swap(j, k)
	}
}
