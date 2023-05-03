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
		initialAllocs = append(initialAllocs, newAllocationFromProjects(s))
	}

	// for i := 0; i < 100; i++ {
	// 	initialAllocs = append(initialAllocs, newAllocationFromSubmission([]int{0}))
	// }

	var pool []int
	for p, num := range map[int]int{
		3:  5,
		5:  5,
		6:  5,
		10: 5,
		11: 5,
		12: 5,
		13: 5,
		15: 5,
		16: 5,
		17: 5,
		19: 5,
		20: 5,
	} {
		for i := 0; i < num; i++ {
			pool = append(pool, p)
		}
	}
	poolAlloc := newAllocationFromProjects(pool)
	poolAlloc.isPool = true
	initialAllocs = append(initialAllocs, poolAlloc)

	var total projectsVector
	for _, a := range initialAllocs {
		total = total.add(a.numPerProject())
	}
	fmt.Println(total)

	// src := rand.NewSource(time.Now().UnixNano())
	src := rand.NewSource(10)
	s := newState(initialAllocs, src)
	fmt.Println(s.score())

	finalState, result, err := s.anneal(0.999998, true)
	if err != nil {
		fmt.Printf("Got error while running Anneal: %v", err)
	}

	finalEnergy := result.Energy
	fmt.Printf("Finished Simulated Annealing in %v! Value: %v \n", result.Runtime, finalEnergy)

	var (
		numInitTokensTotal  int
		numInInitProjsTotal int
		numInDupeProjsTotal int
	)

	for i, current := range finalState.current {
		final := finalState.initial[i]
		numInitTokens := current.numIdenticalInitialTokens(final)
		numInInitProjs := current.numInInitialProjects(final)
		numInDupeProjs := current.numInDuplicateProjects()

		fmt.Printf("%v -> %v: var=%.3f, numInitTokens=%d, numInInitProjs=%d, numInDupeProjs=%d\n",
			final.numPerProject(),
			current.numPerProject(),
			current.variability(),
			numInitTokens,
			numInInitProjs,
			numInDupeProjs,
		)

		if current.isPool {
			continue
		}

		numInitTokensTotal += numInitTokens
		numInInitProjsTotal += numInInitProjs
		numInDupeProjsTotal += numInDupeProjs
	}

	fmt.Printf("TOTAL: size=%d, numInitTokens=%d, numInInitProjs=%d, numInDupeProjs=%d\n",
		s.numTokens(),
		numInitTokensTotal,
		numInInitProjsTotal,
		numInDupeProjsTotal,
	)

	fmt.Println(
		finalState.initial.avgVariability(),
		finalState.current.avgVariability(),
	)
}
