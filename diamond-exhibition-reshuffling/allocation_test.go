package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewAllocationFromTokens(t *testing.T) {
	tests := []struct {
		name               string
		tokens             tokens
		wantNumPerProjects projectsVector
	}{
		{
			name: "uniques",
			tokens: tokens{
				{TokenID: 1, ProjectID: 3},
				{TokenID: 2, ProjectID: 0},
				{TokenID: 3, ProjectID: 1},
			},
			wantNumPerProjects: projectsVector{1, 1, 0, 1},
		},
		{
			name: "with duplicate project",
			tokens: tokens{
				{TokenID: 1, ProjectID: 0},
				{TokenID: 2, ProjectID: 0},
				{TokenID: 3, ProjectID: 1},
			},
			wantNumPerProjects: projectsVector{2, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newAllocationFromTokens(tt.tokens)
			if diff := cmp.Diff(tt.tokens, got.tokens); diff != "" {
				t.Errorf("got.tokens diff (+got -want) %v", diff)
			}

			if diff := cmp.Diff(tt.wantNumPerProjects, got.numPerProject()); diff != "" {
				t.Errorf("got.numPerProject() diff (+got -want) %v", diff)
			}
		})
	}
}

func asPool(a *allocation) *allocation {
	a.isPool = true
	return a
}

func TestCopy(t *testing.T) {
	tests := []struct {
		name       string
		allocation *allocation
	}{
		{
			name: "normal",
			allocation: newAllocationFromTokens(tokens{
				{TokenID: 1, ProjectID: 0},
				{TokenID: 2, ProjectID: 0},
				{TokenID: 3, ProjectID: 1},
			}),
		},
		{
			name: "pool",
			allocation: asPool(newAllocationFromTokens(tokens{
				{TokenID: 1, ProjectID: 0},
				{TokenID: 2, ProjectID: 0},
				{TokenID: 3, ProjectID: 1},
			})),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.allocation.copy()
			cmpopt := cmp.AllowUnexported(allocation{})

			if diff := cmp.Diff(tt.allocation, got, cmpopt); diff != "" {
				t.Errorf("copy diff (+got -want) %v", diff)
			}

			tmp := newAllocationFromTokens(tokens{
				{TokenID: -1, ProjectID: 0},
			})
			got.swapToken(0, tmp, 0)
			if cmp.Equal(tt.allocation, got, cmpopt) {
				t.Errorf("the copy (%+v) is still equal to the input (%+v) after swapping a token. Shallow copy?", got, tt.allocation)
			}
		})
	}
}

func TestSwapToken(t *testing.T) {
	tests := []struct {
		a, b         *allocation
		ia, ib       int
		wantA, wantB *allocation
	}{
		{
			a: newAllocationFromTokens(tokens{
				{TokenID: 1, ProjectID: 0},
				{TokenID: 2, ProjectID: 0},
				{TokenID: 3, ProjectID: 1},
			}),
			b: newAllocationFromTokens(tokens{
				{TokenID: 4, ProjectID: 2},
			}),
			ia: 1,
			ib: 0,
			wantA: newAllocationFromTokens(tokens{
				{TokenID: 1, ProjectID: 0},
				{TokenID: 4, ProjectID: 2},
				{TokenID: 3, ProjectID: 1},
			}),

			wantB: newAllocationFromTokens(tokens{
				{TokenID: 2, ProjectID: 0},
			}),
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			tt.a.swapToken(tt.ia, tt.b, tt.ib)
			cmpopt := cmp.AllowUnexported(allocation{})

			if diff := cmp.Diff(tt.wantA, tt.a, cmpopt); diff != "" {
				t.Errorf("allocation a mismatch after swapping: diff (+got -want) %v", diff)
			}

			if diff := cmp.Diff(tt.wantB, tt.b, cmpopt); diff != "" {
				t.Errorf("allocation b mismatch after swapping: diff (+got -want) %v", diff)
			}
		})
	}
}

func TestScore(t *testing.T) {
	tests := []struct {
		name             string
		initial, current map[int]int
		isPool           bool
		want             float64
	}{
		{
			name:    "same project ID",
			initial: map[int]int{1: 0, 2: 0, 3: 1}, // [2, 1, 0]
			current: map[int]int{4: 0, 5: 1, 6: 2}, // [1, 1, 1]
			want:    -3 - 2 - 0,                    // - regularisation - projectIdPenalty - tokenIdPenalty
		},
		{
			name:    "same tokenId ID",
			initial: map[int]int{1: 0, 2: 0, 3: 1}, // [2, 1, 0]
			current: map[int]int{1: 0, 3: 1, 6: 2}, // [1, 1, 1]
			want:    -3 - 2 - 2,
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
			current: map[int]int{4: 0, 2: 1, 6: 2}, // [1, 1, 1]
			want:    -3 - 3 - 1,
		},
		{
			name:    "pool",
			initial: map[int]int{1: 0, 2: 1, 3: 2}, // [1, 1, 1]
			current: map[int]int{4: 0, 2: 1, 6: 2}, // [1, 1, 1]
			isPool:  true,
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initial := newAllocationFromTokenMap(tt.initial)
			current := newAllocationFromTokenMap(tt.current)
			current.isPool = tt.isPool

			if got := current.score(initial); got != tt.want {
				t.Errorf("score(initial = %v, current = %v) = %v, want %v", initial, current, got, tt.want)
			}
		})
	}
}
