package main

import "testing"

func TestScore(t *testing.T) {
	tests := []struct {
		name    string
		initial []int
		current []int
		want    int
	}{
		{
			name:    "input penalty",
			initial: []int{2, 1, 0},
			current: []int{1, 1, 1},
			want:    -2 - 3,
		},
		{
			name:    "duplicate penalty",
			initial: []int{2, 1, 0},
			current: []int{0, 0, 3},
			want:    -9,
		},
		{
			name:    "mixed",
			initial: []int{2, 1, 0},
			current: []int{1, 0, 2},
			want:    -1 - 5,
		},
		{
			name:    "mixed",
			initial: []int{1, 1, 1},
			current: []int{1, 1, 1},
			want:    -3 - 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initial := newAllocationFromPartial(tt.initial)
			current := newAllocationFromPartial(tt.current)

			if got := score(initial, current); got != tt.want {
				t.Errorf("score(initial = %v, current = %v) = %v, want %v", initial, current, got, tt.want)
			}
		})
	}
}
