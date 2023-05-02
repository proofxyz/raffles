package main

import (
	"math/rand"
	"testing"
	"time"
)

func TestScore(t *testing.T) {
	tests := []struct {
		name    string
		initial map[int]int
		current map[int]int
		want    float64
	}{
		{
			name:    "same project ID",
			initial: map[int]int{1: 0, 2: 0, 3: 1}, // [2, 1, 0]
			current: map[int]int{4: 0, 5: 1, 6: 2}, // [1, 1, 1]
			// - regularisation - projectIdPenalty - tokenIdPenalty
			want: -3 - 2 - 0,
		},
		{
			name:    "same tokenId ID",
			initial: map[int]int{1: 0, 2: 0, 3: 1}, // [2, 1, 0]
			current: map[int]int{1: 0, 3: 1, 6: 2}, // [1, 1, 1]
			// - regularisation - projectIdPenalty - tokenIdPenalty
			want: -3 - 2 - 20,
		},
		{
			name:    "duplicate projectIDs",
			initial: map[int]int{1: 0, 2: 0, 3: 1}, // [2, 1, 0]
			current: map[int]int{4: 2, 5: 2, 6: 2}, // [0, 0, 3]
			want:    -9 - 0 - 0,
		},
		{
			name:    "mixed",
			initial: map[int]int{1: 0, 2: 0, 3: 1}, // [2, 1, 0]
			current: map[int]int{4: 2, 5: 0, 6: 2}, // [1, 0, 2]
			want:    -5 - 1 - 0,
		},
		{
			name:    "mixed",
			initial: map[int]int{1: 0, 2: 1, 3: 2}, // [1, 1, 1]
			current: map[int]int{4: 0, 5: 1, 6: 2}, // [1, 1, 1]
			want:    -3 - 3 - 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initial := newAllocationFromTokens(tt.initial)
			current := newAllocationFromTokens(tt.current)

			if got := current.score(initial); got != tt.want {
				t.Errorf("score(initial = %v, current = %v) = %v, want %v", initial, current, got, tt.want)
			}
		})
	}
}

func TestAnnealStats(t *testing.T) {
	src := rand.NewSource(time.Now().UnixNano())

	tests := []struct {
		name                          string
		submissions                   [][]int
		annealingFactor               float64
		maxNumIdenticalTokensReturned int
		wantNumPerProject             []map[int]int
	}{
		{
			name: "standard",
			submissions: [][]int{
				{1, 1},
				{2, 2, 2, 3, 3},
				{3, 3, 4, 4},
				{4, 4, 1, 1},
				{1, 2, 3, 4, 0},
			},
			annealingFactor:               0.9999,
			maxNumIdenticalTokensReturned: 0,
		},
		{
			name: "swap half batches",
			submissions: [][]int{
				{1, 1, 1, 1, 1, 1, 1},
				{2, 2, 2, 2, 2, 2, 2},
			},
			annealingFactor: 0.9999,
			wantNumPerProject: []map[int]int{
				{1: 1, 2: 6},
				{1: 6, 2: 1},
			},
			maxNumIdenticalTokensReturned: 6,
		},
		{
			name: "swap identical batches",
			submissions: [][]int{
				{1, 1, 1, 1},
				{1, 1, 1, 1},
			},
			annealingFactor: 0.999,
			wantNumPerProject: []map[int]int{
				{1: 4},
				{1: 4},
			},
			maxNumIdenticalTokensReturned: 0,
		},
		{
			name: "swap single submission",
			submissions: [][]int{
				{1},
				{1},
				{1},
				{1},
				{2},
				{2},
				{2},
				{2},
			},
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
			maxNumIdenticalTokensReturned: 0,
		},
		{
			name: "rotation",
			submissions: [][]int{
				{1},
				{2},
				{3},
				{4},
				{5},
			},
			annealingFactor:               0.999,
			maxNumIdenticalTokensReturned: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var allocs []*allocation
			for _, s := range tt.submissions {
				allocs = append(allocs, newAllocationFromSubmission(s))
			}

			s, _, err := newState(allocs, src).anneal(tt.annealingFactor, false)
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

				for tokenId := range c.tokens {
					// Sanity check to see if any tokens have been repeated
					if _, ok := seen[tokenId]; ok {
						t.Errorf("duplicate tokenId %d", tokenId)
					}
					seen[tokenId] = struct{}{}

					if _, ok := s.initial[i].tokens[tokenId]; ok {
						numIdenticalTokensReturned++
					}
				}

				if tt.wantNumPerProject != nil {
					got := c.NumPerProject()
					want := tt.wantNumPerProject[i]
					for projectId, num := range got {
						if num != want[projectId] {
							t.Errorf("NumPerProject(projectId %d): got %d, want %d", projectId, num, want[projectId])
						}
					}
				}
			}

			if numIdenticalTokensReturned > tt.maxNumIdenticalTokensReturned {
				t.Errorf("numIdenticalTokensReturned = %d, want <= %d", numIdenticalTokensReturned, tt.maxNumIdenticalTokensReturned)
			}

		})
	}

}
