package main

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/go-cmp/cmp"
)

func TestNewAllocation(t *testing.T) {
	tests := []struct {
		name               string
		addr               common.Address
		tokens             tokens
		wantNumPerProjects projectsVector
	}{
		{
			name: "uniques",
			addr: common.HexToAddress("0xdeadbeef"),
			tokens: tokens{
				{TokenID: 1, ProjectID: 3},
				{TokenID: 2, ProjectID: 0},
				{TokenID: 3, ProjectID: 1},
			},
			wantNumPerProjects: projectsVector{1, 1, 0, 1},
		},
		{
			name: "with duplicate project",
			addr: common.HexToAddress("0x1234"),
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
			got := newAllocation(tt.addr, tt.tokens)
			if diff := cmp.Diff(tt.tokens, got.tokens); diff != "" {
				t.Errorf("newAllocation(…, %+v).tokens diff (+got -want) %s", tt.tokens, diff)
			}

			if tt.addr != got.owner {
				t.Errorf("newAllocation(%v, …).owner = %v, want %v", tt.addr, got.owner, tt.addr)
			}

			if diff := cmp.Diff(tt.wantNumPerProjects, got.numPerProject()); diff != "" {
				t.Errorf("newAllocation(…, %+v).numPerProject() diff (+got -want) %v", tt.tokens, diff)
			}
		})
	}
}

var defaultAddr = common.HexToAddress("0xdeadbeef")

func asPool(a *allocation) *allocation {
	aa := *a
	aa.isPool = true
	return &aa
}

func TestCopy(t *testing.T) {
	tests := []struct {
		name       string
		allocation *allocation
	}{
		{
			name: "normal",
			allocation: newAllocation(defaultAddr, tokens{
				{TokenID: 1, ProjectID: 0},
				{TokenID: 2, ProjectID: 0},
				{TokenID: 3, ProjectID: 1},
			}),
		},
		{
			name: "pool",
			allocation: asPool(newAllocation(defaultAddr, tokens{
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
				t.Errorf("%T.copy diff (+got -want) %v", tt.allocation, diff)
			}

			tmp := newAllocation(defaultAddr, tokens{
				{TokenID: -1, ProjectID: 0},
			})
			got.swapToken(tmp, 0, 0)
			if cmp.Equal(tt.allocation, got, cmpopt) {
				t.Errorf("%T.copy() (%+v) is still equal to the input (%+v) after swapping a token. Shallow copy?", tt.allocation, got, tt.allocation)
			}
		})
	}
}

func TestSwapToken(t *testing.T) {
	alice := common.HexToAddress("0xa11ce")
	bob := common.HexToAddress("0xb0b")

	tests := []struct {
		a, b         *allocation
		ia, ib       int
		wantA, wantB *allocation
	}{
		{
			a: newAllocation(alice, tokens{
				{TokenID: 1, ProjectID: 0},
				{TokenID: 2, ProjectID: 0},
				{TokenID: 3, ProjectID: 1},
			}),
			b: newAllocation(bob, tokens{
				{TokenID: 4, ProjectID: 2},
			}),
			ia: 1,
			ib: 0,
			wantA: newAllocation(alice, tokens{
				{TokenID: 1, ProjectID: 0},
				{TokenID: 4, ProjectID: 2},
				{TokenID: 3, ProjectID: 1},
			}),
			wantB: newAllocation(bob, tokens{
				{TokenID: 2, ProjectID: 0},
			}),
		},
		{
			a: newAllocation(alice, tokens{
				{TokenID: 1, ProjectID: 0},
				{TokenID: 2, ProjectID: 0},
				{TokenID: 3, ProjectID: 1},
			}),
			b: newAllocation(bob, tokens{
				{TokenID: 4, ProjectID: 2},
				{TokenID: 5, ProjectID: 2},
				{TokenID: 6, ProjectID: 2},
			}),
			ia: 0,
			ib: 1,
			wantA: newAllocation(alice, tokens{
				{TokenID: 5, ProjectID: 2},
				{TokenID: 2, ProjectID: 0},
				{TokenID: 3, ProjectID: 1},
			}),
			wantB: newAllocation(bob, tokens{
				{TokenID: 4, ProjectID: 2},
				{TokenID: 1, ProjectID: 0},
				{TokenID: 6, ProjectID: 2},
			}),
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			tt.a.swapToken(tt.b, tt.ia, tt.ib)
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
		initial, current tokens
		isPool           bool
		wantPenalties    [3]int
		want             float64
	}{
		{
			name: "received original project ID 2x",
			// [2, 1, 0] = numPerProjects
			initial: tokens{
				{TokenID: 1, ProjectID: 0},
				{TokenID: 2, ProjectID: 0},
				{TokenID: 3, ProjectID: 1},
			},
			// [1, 1, 1]
			current: tokens{
				{TokenID: 4, ProjectID: 0},
				{TokenID: 5, ProjectID: 1},
				{TokenID: 6, ProjectID: 2},
			},
			// - regularisation - projectIdPenalty - tokenIdPenalty
			wantPenalties: [3]int{3, 2, 0},
			want:          -3 - 2 - 0,
		},
		{
			name: "received original token ID 2x",
			// [2, 1, 0]
			initial: tokens{
				{TokenID: 1, ProjectID: 0},
				{TokenID: 2, ProjectID: 0},
				{TokenID: 3, ProjectID: 1},
			},
			// [1, 1, 1]
			current: tokens{
				{TokenID: 1, ProjectID: 0},
				{TokenID: 3, ProjectID: 1},
				{TokenID: 6, ProjectID: 2},
			},
			wantPenalties: [3]int{3, 2, 2},
			want:          -3 - 2 - 2,
		},
		{
			name: "received original token ID 2x (with fake project ID)",
			// [2, 1, 0]
			initial: tokens{
				{TokenID: 1, ProjectID: 0},
				{TokenID: 2, ProjectID: 0},
				{TokenID: 3, ProjectID: 1},
			},
			// [1, 1, 1]
			current: tokens{
				// Although it's impossible to receive the same token ID back but
				// with a different project ID, the scoring mechanism doesn't know
				// about this. We include this to demonstrate the individual scoring
				// elements.
				{TokenID: 1, ProjectID: 19},
				{TokenID: 3, ProjectID: 20},
				{TokenID: 6, ProjectID: 2},
			},
			wantPenalties: [3]int{3, 0, 2},
			want:          -3 - 0 - 2,
		},
		{
			name: "duplicate projectIDs although different to original",
			// [2, 1, 0]
			initial: tokens{
				{TokenID: 1, ProjectID: 0},
				{TokenID: 2, ProjectID: 0},
				{TokenID: 3, ProjectID: 1},
				{TokenID: 4, ProjectID: 1},
				{TokenID: 5, ProjectID: 1},
				{TokenID: 6, ProjectID: 1},
			},
			// [0, 0, 3]
			current: tokens{
				{TokenID: 7, ProjectID: 2},
				{TokenID: 8, ProjectID: 2},
				{TokenID: 9, ProjectID: 2},
				{TokenID: 10, ProjectID: 3},
				{TokenID: 11, ProjectID: 3},
				{TokenID: 12, ProjectID: 4},
			},
			wantPenalties: [3]int{3*3 + 2*2 + 1*1, 0, 0},
			want:          -(9 + 4 + 1) - 0 - 0,
		},
		{
			name: "mixed",
			// [2, 1, 0]
			initial: tokens{
				{TokenID: 1, ProjectID: 0},
				{TokenID: 2, ProjectID: 0},
				{TokenID: 3, ProjectID: 1},
			},
			// [1, 0, 2]
			current: tokens{
				{TokenID: 4, ProjectID: 2},
				{TokenID: 5, ProjectID: 0},
				{TokenID: 6, ProjectID: 2},
			},
			wantPenalties: [3]int{2*2 + 1, 1 /*project 0 returned*/, 0},
			want:          -(4 + 1) - 1 - 0,
		},
		{
			name: "mixed",
			// [1, 1, 1]
			initial: tokens{
				{TokenID: 1, ProjectID: 0},
				{TokenID: 2, ProjectID: 1},
				{TokenID: 3, ProjectID: 2},
			},
			// [1, 1, 1]
			current: tokens{
				{TokenID: 4, ProjectID: 0},
				{TokenID: 2, ProjectID: 1},
				{TokenID: 6, ProjectID: 2},
			},
			wantPenalties: [3]int{3, 3, 1},
			want:          -3 - 3 - 1,
		},
		{
			name: "pool",
			// [1, 1, 1]
			initial: tokens{
				{TokenID: 1, ProjectID: 0},
				{TokenID: 2, ProjectID: 1},
				{TokenID: 3, ProjectID: 2},
			},
			// [1, 1, 1]
			current: tokens{
				{TokenID: 4, ProjectID: 0},
				{TokenID: 2, ProjectID: 1},
				{TokenID: 6, ProjectID: 2},
			},
			isPool: true,
			// see scorePenalties for rationale of this constant pool score
			wantPenalties: [3]int{3, 0, 0},
			want:          -3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initial := newAllocation(defaultAddr, tt.initial)
			current := newAllocation(defaultAddr, tt.current)
			current.isPool = tt.isPool

			if got := current.score(initial); got != tt.want {
				t.Errorf("%T(%v).score(initial = %v) = %v, want %v", current, tt.current, tt.initial, got, tt.want)
			}
			if got := current.scorePenalties(initial); !cmp.Equal(got, tt.wantPenalties) {
				t.Errorf("%T(%v).scorePenalties(initial = %v) = %v, want %v", current, tt.current, tt.initial, got, tt.wantPenalties)
			}
		})
	}
}
