package main

import (
	"fmt"
	"io"
	"math"
	"math/rand"

	"github.com/ccssmnn/hego"
)

// state is the state in the simulated annealing algorithm.
type state struct {
	initial, current allocations
	rng              *rand.Rand
	cachedScore      float64
}

// newState creates a new state from a list of initial allocations and a random source.
func newState(initial allocations, rng *rand.Rand) *state {
	s := &state{
		initial: initial,
		current: initial,
		rng:     rng,
	}
	s.cachedScore = s.current.score(s.initial)

	return s
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

	a.swapToken(b, tokenIdxA, tokenIdxB)

	s.cachedScore += a.score(s.initial[allocationIdxA])
	s.cachedScore += b.score(s.initial[allocationIdxB])

	s.current[allocationIdxA] = a
	s.current[allocationIdxB] = b
}

// neighbor returns a neighbor of the current state for the simulated annealing process.
// The neighbor is generated by randomly swapping two tokens.
func (s *state) neighbor() *state {
	var allocIdxA, allocIdxB int
	for allocIdxA == allocIdxB {
		// swapping within the same allocation yields the same state.
		// we sample again in this case.
		allocIdxA = s.rng.Intn(s.numAllocations())
		allocIdxB = s.rng.Intn(s.numAllocations())
	}

	tokenIdxA := s.current[allocIdxA].drawTokenIdx(s.rng)
	tokenIdxB := s.current[allocIdxB].drawTokenIdx(s.rng)

	neigh := s.copyShallow()
	neigh.swap(allocIdxA, tokenIdxA, allocIdxB, tokenIdxB)

	return neigh
}

// Neighbor returns a neighbor of the current state for the simulated annealing process.
func (s *state) Neighbor() hego.AnnealingState {
	return s.neighbor()
}

// Energy returns the energy of the current state for the simulated annealing process.
func (s *state) Energy() float64 {
	return -float64(s.cachedScore)
}

// anneal runs the simulated annealing algorithm on the state.
func (s *state) anneal(annealingFactor float64, verbose bool) (*state, error) {
	settings := hego.SASettings{
		Temperature:     10,
		AnnealingFactor: annealingFactor,
	}

	// Iterations until we reach temp = 1 (which is the minimum energy difference between two neighboring, non-equivalent states)
	// The probability of accepting a worse state is <= 1/e at this point.
	// We use this to get an estimate for the amount of iterations needed for cooling.
	numIterToLukewarm := int(-math.Log(settings.Temperature) / math.Log(float64(settings.AnnealingFactor)))

	settings.Settings.MaxIterations = 2 * numIterToLukewarm
	if verbose {
		settings.Settings.Verbose = settings.Settings.MaxIterations / 20
	}

	// Hego's simulated annealing process uses random numbers from the default source.
	// To get deterministic results we thus reseed that source with the local source.
	rand.Seed(s.rng.Int63())

	result, err := hego.SA(s, settings)
	if err != nil {
		return nil, err
	}

	finalState := result.State.(*state)
	return finalState, nil
}

type stateStats struct {
	numInitTokensTotal, numInInitProjsTotal, numInDupeProjsTotal int
}

// computeStats computes statistics about the current state ignoring the PROOF-issued pool.
func (s *state) computeStats() stateStats {
	var stats stateStats
	for i, c := range s.current {
		if c.isPool {
			continue
		}

		initial := s.initial[i].tokens
		stats.numInitTokensTotal += c.numSameTokenID(initial)
		stats.numInInitProjsTotal += c.numInSameProjects(initial)
		stats.numInDupeProjsTotal += c.numInDuplicateProjects()
	}
	return stats
}

// printStats prints statistics about the current state.
func (s *state) printStats(w io.Writer) error {
	stats := s.computeStats()

	_, err := fmt.Fprintf(w, "Current allocation stats: size=%d, energy=%.0f, numInitTokens=%d, numInInitProjs=%d, numInDupeProjs=%d\n",
		s.numTokens(),
		s.Energy(),
		stats.numInitTokensTotal,
		stats.numInInitProjsTotal,
		stats.numInDupeProjsTotal,
	)

	return err
}

// printReallocationOverview prints an overview on the reallocations in the current state,
// i.e. the number of tokens per project before and after the reallocation with some additional statistics.
func (s *state) printReallocationOverview(w io.Writer) error {
	for i, current := range s.current {
		initial := s.initial[i]
		_, err := fmt.Fprintf(w, "%v -> %v: var=%.3f, numInitTokens=%d, numInInitProjs=%d, numInDupeProjs=%d, numGrails=%d\n",
			initial.numPerProject(),
			current.numPerProject(),
			current.variability(),
			current.numSameTokenID(initial.tokens),
			current.numInSameProjects(initial.tokens),
			current.numInDuplicateProjects(),
			current.numGrails(),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// isTrivialOptimum returns true if the current state is a trivial optimum.
// A trivial optimum is a state where all submitters get no duplicate projects and none of the tokens/projects
// that they put in, with the score function assuming its theoretical maximum.
// Depending on the given problem, this optimum might not be reachable. But if it is reached, we can be certain that we
// can't improve from there.
func (s *state) isTrivialOptimum() bool {
	stats := s.computeStats()
	score := s.score()

	return stats.numInitTokensTotal == 0 &&
		stats.numInInitProjsTotal == 0 &&
		stats.numInDupeProjsTotal == 0 &&
		// The theoretical maximum of the score function can be computed by analysing the individual terms
		// of the allocation score function (for the numbering see `allocation.score()`).
		// The terms are bound by:
		// (1) >= allocation.numTokens(), the lower bound being when there is 1 token per project
		// (2) >= 0
		// (3) >= 0
		// The maximum value of the composite score function over all allocations is therefore given by the total number of tokens.
		score == -float64(s.numTokens())
}
