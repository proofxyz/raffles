package main

import (
	"math/rand"
	"testing"
	"time"
)

func newAllocationFromProjectIds(xs []int) *allocation {
	return newAllocation(defaultAddr, newTokensFromProjects(xs))
}

func newAllocationsFromProjectIds(xss [][]int) allocations {
	var allocs allocations
	for _, xs := range xss {
		allocs = append(allocs, newAllocationFromProjectIds(xs))
	}
	return allocs
}

func TestAnnealStats(t *testing.T) {
	src := rand.NewSource(time.Now().UnixNano())

	tests := []struct {
		name                           string
		allocations                    allocations
		annealingFactor                float64
		wantNumIdenticalTokensReturned int
		wantNumPerProject              []map[int]int
	}{
		{
			name: "standard",
			allocations: newAllocationsFromProjectIds([][]int{
				{1, 1},
				{2, 2, 2, 3, 3},
				{3, 3, 4, 4},
				{4, 4, 1, 1},
				{1, 2, 3, 4, 0},
			}),
			annealingFactor:                0.9999,
			wantNumIdenticalTokensReturned: 0,
		},
		{
			name: "swap half batches",
			allocations: newAllocationsFromProjectIds([][]int{
				{1, 1, 1, 1, 1, 1, 1},
				{2, 2, 2, 2, 2, 2, 2},
			}),
			annealingFactor: 0.9999,
			wantNumPerProject: []map[int]int{
				// this is a result of the compromise between getting the same tokens back and variability
				{1: 3, 2: 4},
				{1: 4, 2: 3},
			},
			wantNumIdenticalTokensReturned: 6,
		},
		{
			name: "swap identical batches",
			allocations: newAllocationsFromProjectIds([][]int{
				{1, 1, 1, 1},
				{1, 1, 1, 1},
			}),
			annealingFactor: 0.999,
			wantNumPerProject: []map[int]int{
				{1: 4},
				{1: 4},
			},
			wantNumIdenticalTokensReturned: 0,
		},
		{
			name: "swap single submission",
			allocations: newAllocationsFromProjectIds([][]int{
				{1},
				{1},
				{1},
				{1},
				{2},
				{2},
				{2},
				{2},
			}),
			annealingFactor: 0.999,
			wantNumPerProject: []map[int]int{
				{2: 1},
				{2: 1},
				{2: 1},
				{2: 1},
				{1: 1},
				{1: 1},
				{1: 1},
				{1: 1},
			},
			wantNumIdenticalTokensReturned: 0,
		},
		{
			name: "rotation",
			allocations: newAllocationsFromProjectIds([][]int{
				{1},
				{2},
				{3},
				{4},
				{5},
			}),
			annealingFactor:                0.999,
			wantNumIdenticalTokensReturned: 0,
		},
		{
			name: "inbalanced",
			allocations: allocations{
				newAllocationFromProjectIds([]int{0, 0, 0, 0, 0}),
				newAllocationFromProjectIds([]int{1, 2, 3, 4, 5}),
			},
			annealingFactor:                0.999,
			wantNumIdenticalTokensReturned: 4, // meaning that 3 tokens were swapped
		},
		{
			name: "with pool",
			allocations: allocations{
				newAllocationFromProjectIds([]int{0, 0, 0, 0, 0}),
				asPool(newAllocationFromProjectIds([]int{1, 2, 3, 4, 5})),
			},
			annealingFactor:                0.999,
			wantNumIdenticalTokensReturned: 0, // all tokens can be swapped since the pool does not care
			wantNumPerProject: []map[int]int{
				{1: 1, 2: 1, 3: 1, 4: 1, 5: 1},
				{0: 5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _, err := newState(tt.allocations, src).anneal(tt.annealingFactor, false)
			if err != nil {
				t.Errorf("anneal(): err %v", err)
			}

			seen := make(map[int]struct{})
			var (
				numIdenticalTokensReturned int
			)

			for i, c := range s.current {
				if len(c.tokens) != len(s.initial[i].tokens) {
					t.Errorf("len(current[%d].tokens) = %d, want %d", i, len(c.tokens), len(s.initial[i].tokens))
				}

				for _, v := range c.tokens {
					// Sanity check for duplicate tokens
					if _, ok := seen[v.TokenID]; ok {
						t.Errorf("duplicate tokenId %d", v.TokenID)
					}
					seen[v.TokenID] = struct{}{}
				}

				numIdenticalTokensReturned += c.numSameTokenID(s.initial[i].tokens)

				if tt.wantNumPerProject != nil {
					got := c.numPerProject()
					want := tt.wantNumPerProject[i]
					for projectId, num := range got {
						if num != want[projectId] {
							t.Errorf("allocation[%d].NumPerProject(projectId %d): got %d, want %d", i, projectId, num, want[projectId])
						}
					}
				}
			}

			if numIdenticalTokensReturned != tt.wantNumIdenticalTokensReturned {
				t.Errorf("numIdenticalTokensReturned = %d, want = %d", numIdenticalTokensReturned, tt.wantNumIdenticalTokensReturned)
			}

		})
	}

}
