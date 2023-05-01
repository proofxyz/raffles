package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	_ "embed"

	"github.com/ccssmnn/hego"
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
	// 	{1},
	// 	{1},
	// 	{1},
	// 	{1},
	// 	{1},
	// 	{1},
	// 	{1},
	// 	{1},
	// 	{1},
	// 	{1},
	// 	{1},
	// 	{1},
	// 	{1},
	// }

	var submissions [][]int
	json.Unmarshal(raw, &submissions)

	// rand.New(rand.NewSource(time.Now().UnixNano())).Shuffle(len(submissions), func(i, j int) {
	// 	submissions[i], submissions[j] = submissions[j], submissions[i]
	// })

	var initialAllocs []*allocation
	for _, s := range submissions {
		initialAllocs = append(initialAllocs, newAllocationFromSubmission(s))
	}

	s := state{
		initial: initialAllocs,
		current: initialAllocs,
		src:     rand.NewSource(time.Now().UnixNano()),
	}

	fmt.Println(s.score())

	var pool allocation
	for _, a := range s.current {
		pool = *pool.add(a)
	}
	fmt.Println(pool)

	var rel [numProjects]float64
	for i, x := range pool {
		rel[i] = float64(x) / float64(pool.sum())
	}
	fmt.Println(rel)

	settings := hego.SASettings{
		Settings: hego.Settings{
			MaxIterations: 2500000,
			Verbose:       100000,
			KeepHistory:   false,
		},
		Temperature: 10,
		// Temperature:     s.Energy() / float64(s.numAllocations()),
		AnnealingFactor: 0.999998,
	}

	// start simulated annealing algorithm
	result, err := hego.SA(s, settings)

	if err != nil {
		fmt.Printf("Got error while running Anneal: %v", err)
	}
	finalState := result.State.(state)
	finalEnergy := result.Energy
	fmt.Printf("Finished Simulated Annealing in %v! Value: %v \n", result.Runtime, finalEnergy)

	// var (
	// 	iter int
	// )

	// for noChange := 0; noChange < 10000; noChange++ {

	// 	iter++
	// 	s2 := s.Neighbor()
	// 	if s2.score() > s.score() {
	// 		noChange = 0
	// 		s = s2
	// 	}
	// }
	// finalState := s

	for i, x := range finalState.current {
		fmt.Println(finalState.initial[i], x)
	}

	fmt.Println(
		finalState.initial.avgVariability(),
		finalState.current.avgVariability(),
	)
}
