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
				t.Errorf("got.tokens diff (+got -want) %v", diff)
			}

			if tt.addr != got.owner {
				t.Errorf("got.owner %v, want %v", got.owner, tt.addr)
			}

			if diff := cmp.Diff(tt.wantNumPerProjects, got.numPerProject()); diff != "" {
				t.Errorf("got.numPerProject() diff (+got -want) %v", diff)
			}
		})
	}
}

var defaultAddr = common.HexToAddress("0xdeadbeef")

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
				t.Errorf("copy diff (+got -want) %v", diff)
			}

			tmp := newAllocation(defaultAddr, tokens{
				{TokenID: -1, ProjectID: 0},
			})
			got.swapToken(tmp, 0, 0)
			if cmp.Equal(tt.allocation, got, cmpopt) {
				t.Errorf("the copy (%+v) is still equal to the input (%+v) after swapping a token. Shallow copy?", got, tt.allocation)
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
		want             float64
	}{
		{
			name: "same project ID",
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
			want: -3 - 2 - 0,
		},
		{
			name: "same tokenId ID",
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
			want: -3 - 2 - 2,
		},
		{
			name: "duplicate projectIDs",
			// [2, 1, 0]
			initial: tokens{
				{TokenID: 1, ProjectID: 0},
				{TokenID: 2, ProjectID: 0},
				{TokenID: 3, ProjectID: 1},
			},
			// [0, 0, 3]
			current: tokens{
				{TokenID: 4, ProjectID: 2},
				{TokenID: 5, ProjectID: 2},
				{TokenID: 6, ProjectID: 2},
			},
			want: -9 - 0 - 0,
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
			want: -5 - 1 - 0,
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
			want: -3 - 3 - 1,
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
			want:   -3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initial := newAllocation(defaultAddr, tt.initial)
			current := newAllocation(defaultAddr, tt.current)
			current.isPool = tt.isPool

			if got := current.score(initial); got != tt.want {
				t.Errorf("score(initial = %v, current = %v) = %v, want %v", initial, current, got, tt.want)
			}
		})
	}
}
