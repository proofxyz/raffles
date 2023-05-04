package main

import (
	"math"
	"math/rand"

	"github.com/ccssmnn/hego"
)

// state is the state in the simulated annealing algorithm.
type state struct {
	initial     allocations // initial allocations (used in the score computation)
	current     allocations // current allocations
	src         rand.Source // random source
	cachedScore float64     // score of the current allocations, cached for performance
}

// newState creates a new state from a list of initial allocations and a random source.
func newState(initial allocations, src rand.Source) *state {
	s := state{
		initial: initial,
		current: initial,
		src:     src,
	}
	s.cachedScore = s.current.score(s.initial)

	return &s
}

// numAllocations returns the number of allocations in the state.
func (s *state) numAllocations() int {
	return len(s.initial)
}

// numTokens returns the number of tokens in the state.
func (s *state) numTokens() int {
	return s.current.numTokens()
}

// score returns the score of the current allocation state.
func (s *state) score() float64 {
	return s.cachedScore
}

// copyShallow returns a shallow copy of the state.
func (s *state) copyShallow() *state {
	c := *s
	c.current = s.current.copyShallow()
	return &c
}

// swap swaps two tokens in the current allocation state.
// allocationIdx{A,B} are the indices of the allocations whose tokens are swapped.
// tokenIdx{A,B} are the indices of the tokens within the given allocations.
func (s *state) swap(allocationIdxA, tokenIdxA, allocationIdxB, tokenIdxB int) {
	a := s.current[allocationIdxA].copy()
	b := s.current[allocationIdxB].copy()

	s.cachedScore -= a.score(s.initial[allocationIdxA])
	s.cachedScore -= b.score(s.initial[allocationIdxB])

	a.swapToken(tokenIdxA, b, tokenIdxB)

	s.cachedScore += a.score(s.initial[allocationIdxA])
	s.cachedScore += b.score(s.initial[allocationIdxB])

	s.current[allocationIdxA] = a
	s.current[allocationIdxB] = b
}

// neighbor returns a neighbor of the current state for the simulated annealing process.
// The neighbor is generated by randomly swapping two tokens.
func (s *state) neighbor() *state {
	rng := rand.New(s.src)

	allocIdxA := rng.Intn(s.numAllocations())
	allocIdxB := rng.Intn(s.numAllocations())

	for allocIdxA == allocIdxB {
		// swapping within the same allocation yields the same state.
		// we sample again in this case.
		return s.neighbor()
	}

	tokenIdxA := s.current[allocIdxA].drawTokenIdx(s.src)
	tokenIdxB := s.current[allocIdxB].drawTokenIdx(s.src)

	neigh := s.copyShallow()
	neigh.swap(allocIdxA, tokenIdxA, allocIdxB, tokenIdxB)

	return neigh
}

// Neighbor returns a neighbor of the current state for the simulated annealing process.
func (s state) Neighbor() hego.AnnealingState {
	return s.neighbor()
}

// Energy returns the energy of the current state for the simulated annealing process.
func (s state) Energy() float64 {
	return -float64(s.cachedScore)
}

// anneal runs the simulated annealing algorithm on the state.
func (s state) anneal(annealingFactor float64, verbose bool) (*state, *hego.SAResult, error) {
	settings := hego.SASettings{
		Temperature:     10,
		AnnealingFactor: annealingFactor,
	}

	// iterations until we reach temp = 1 (which is the minimum energy difference
	// between two neighboring, non-equivalent states)
	// From here on, the energy of the state can therefore only decrease.
	numIterToCold := int(-math.Log(settings.Temperature) / math.Log(float64(settings.AnnealingFactor)))

	settings.Settings.MaxIterations = 2 * numIterToCold
	if verbose {
		settings.Settings.Verbose = settings.Settings.MaxIterations / 20
	}

	// Hego's simulated annealing process uses random numbers from the default source.
	// To get deterministic results we thus reseed that source with the local source.
	rand.Seed(rand.New(s.src).Int63())

	result, err := hego.SA(s, settings)
	if err != nil {
		return nil, nil, err
	}

	finalState := result.State.(*state)
	return finalState, &result, nil
}
