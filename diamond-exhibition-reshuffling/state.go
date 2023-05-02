package main

import (
	"math"
	"math/rand"
	"time"

	"github.com/ccssmnn/hego"
)

type state struct {
	initial      allocations
	current      allocations
	src          rand.Source
	currentScore float64
}

func newState(initial allocations) *state {
	s := state{
		initial: initial,
		current: initial,
		src:     rand.NewSource(time.Now().UnixNano()),
	}
	s.currentScore = s.computeScore()

	return &s
}

func (s state) numAllocations() int {
	return len(s.initial)
}

func (s state) score() float64 {
	return s.currentScore
}

func (s state) computeScore() float64 {
	var res float64
	for i, c := range s.current {
		res += c.score(s.initial[i])
	}
	return res
}

func (s *state) swap(ia, ja, ib, jb int) {
	s.current[ia] = s.current[ia].copy()
	s.current[ib] = s.current[ib].copy()

	a := s.current[ia]
	b := s.current[ib]

	s.currentScore -= a.score(s.initial[ia])
	s.currentScore -= b.score(s.initial[ib])

	a.moveToken(ja, b)
	b.moveToken(jb, a)

	s.currentScore += a.score(s.initial[ia])
	s.currentScore += b.score(s.initial[ib])
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

	ja := s.current[ia].drawToken(s.src)
	jb := s.current[ib].drawToken(s.src)
	if ja == jb {
		return s
	}

	s2 := s
	s2.current = current
	s2.swap(ia, ja, ib, jb)

	return s2
}

func (s state) Neighbor() hego.AnnealingState {
	return s.neighbor()
}

func (s state) Energy() float64 {
	return -float64(s.currentScore)
}

func (s state) anneal(annealingFactor float64, verbose bool) (*state, *hego.SAResult, error) {
	settings := hego.SASettings{
		Temperature:     10,
		AnnealingFactor: annealingFactor,
	}

	// iterations until we reach temp = 1
	numIterToLukewarm := int(-math.Log(settings.Temperature) / math.Log(float64(settings.AnnealingFactor)))

	settings.Settings.MaxIterations = 2 * numIterToLukewarm
	if verbose {
		settings.Settings.Verbose = settings.Settings.MaxIterations / 20
	}

	result, err := hego.SA(s, settings)
	if err != nil {
		return nil, nil, err
	}

	finalState := result.State.(state)
	return &finalState, &result, nil
}
