package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

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
			name:       "normal",
			allocation: newAllocationFromTokens(tokenSet{1: 0, 2: 0, 3: 1}),
		},
		{
			name:       "pool",
			allocation: asPool(newAllocationFromTokens(tokenSet{1: 0, 2: 0, 3: 1})),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.allocation.copy()
			cmpopt := cmp.AllowUnexported(allocation{})

			if diff := cmp.Diff(tt.allocation, got, cmpopt); diff != "" {
				t.Errorf("copy diff (+got -want) %v", diff)
			}

			tmp := newAllocation()
			for _, t := range got.tokens {
				got.moveToken(t, tmp)
			}
			if cmp.Equal(tt.allocation, got, cmpopt) {
				t.Errorf("the copy (%+v) is still equal to the input (%+v) after moving all tokens", got, tt.allocation)
			}
		})
	}
}

func TestMoveToken(t *testing.T) {
	tests := []struct {
		src     *allocation
		dst     *allocation
		tokenId int
		wantSrc *allocation
		wantDst *allocation
	}{
		{
			src:     newAllocationFromTokens(tokenSet{1: 0, 2: 0, 3: 1}),
			dst:     newAllocationFromTokens(tokenSet{4: 2}),
			tokenId: 2,
			wantSrc: newAllocationFromTokens(tokenSet{1: 0, 3: 1}),
			wantDst: newAllocationFromTokens(tokenSet{2: 0, 4: 2}),
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			tt.src.moveToken(tt.tokenId, tt.dst)
			cmpopt := cmp.AllowUnexported(allocation{})

			if diff := cmp.Diff(tt.wantSrc, tt.src, cmpopt); diff != "" {
				t.Errorf("src allocation mismatch after moving token %d diff (+got -want) %v", tt.tokenId, diff)
			}

			if diff := cmp.Diff(tt.wantDst, tt.dst, cmpopt); diff != "" {
				t.Errorf("dst allocation mismatch after moving token %d diff (+got -want) %v", tt.tokenId, diff)
			}
		})
	}
}

func TestScore(t *testing.T) {
	tests := []struct {
		name    string
		initial tokenSet
		current tokenSet
		isPool  bool
		want    float64
	}{
		{
			name:    "same project ID",
			initial: tokenSet{1: 0, 2: 0, 3: 1}, // [2, 1, 0]
			current: tokenSet{4: 0, 5: 1, 6: 2}, // [1, 1, 1]
			want:    -3 - 2 - 0,                 // - regularisation - projectIdPenalty - tokenIdPenalty
		},
		{
			name:    "same tokenId ID",
			initial: tokenSet{1: 0, 2: 0, 3: 1}, // [2, 1, 0]
			current: tokenSet{1: 0, 3: 1, 6: 2}, // [1, 1, 1]
			want:    -3 - 2 - 2,
		},
		{
			name:    "duplicate projectIDs",
			initial: tokenSet{1: 0, 2: 0, 3: 1}, // [2, 1, 0]
			current: tokenSet{4: 2, 5: 2, 6: 2}, // [0, 0, 3]
			want:    -9 - 0 - 0,
		},
		{
			name:    "mixed",
			initial: tokenSet{1: 0, 2: 0, 3: 1}, // [2, 1, 0]
			current: tokenSet{4: 2, 5: 0, 6: 2}, // [1, 0, 2]
			want:    -5 - 1 - 0,
		},
		{
			name:    "mixed",
			initial: tokenSet{1: 0, 2: 1, 3: 2}, // [1, 1, 1]
			current: tokenSet{4: 0, 2: 1, 6: 2}, // [1, 1, 1]
			want:    -3 - 3 - 1,
		},
		{
			name:    "pool",
			initial: tokenSet{1: 0, 2: 1, 3: 2}, // [1, 1, 1]
			current: tokenSet{4: 0, 2: 1, 6: 2}, // [1, 1, 1]
			isPool:  true,
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initial := newAllocationFromTokens(tt.initial)
			current := newAllocationFromTokens(tt.current)
			current.isPool = tt.isPool

			if got := current.score(initial); got != tt.want {
				t.Errorf("score(initial = %v, current = %v) = %v, want %v", initial, current, got, tt.want)
			}
		})
	}
}
