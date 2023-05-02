package main

import (
	"encoding/json"
	"fmt"
	"math/rand"

	_ "embed"
)

//go:embed submissions.json
var raw []byte

func main() {
	// submissions := [][]int{
	// 	{1, 1},
	// 	{2, 2, 2, 3, 3},
	// 	{3, 3, 4, 4},
	// 	{4, 4, 1, 1},
	// 	{1, 2, 3, 4, 0},
	// 	{7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7},
	// 	{7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7},
	// }

	var submissions [][]int
	json.Unmarshal(raw, &submissions)

	var initialAllocs []*allocation
	for _, s := range submissions {
		initialAllocs = append(initialAllocs, newAllocationFromSubmission(s))
	}

	// src := rand.NewSource(time.Now().UnixNano())
	src := rand.NewSource(10)
	s := newState(initialAllocs, src)
	fmt.Println(s.score())

	finalState, result, err := s.anneal(0.9998, true)
	if err != nil {
		fmt.Printf("Got error while running Anneal: %v", err)
	}

	finalEnergy := result.Energy
	fmt.Printf("Finished Simulated Annealing in %v! Value: %v \n", result.Runtime, finalEnergy)

	for i, current := range finalState.current {
		final := finalState.initial[i]
		fmt.Printf("%v -> %v: var=%.3f, numInitTokens=%d, numInInitProjs=%d, numInDupeProjs=%d\n",
			final.NumPerProject(),
			current.NumPerProject(),
			current.variability(),
			current.numIdenticalInitialTokens(final),
			current.numInInitialProjects(final),
			current.numInDuplicateProjects(),
		)
	}

	fmt.Println(
		finalState.initial.avgVariability(),
		finalState.current.avgVariability(),
	)
}
