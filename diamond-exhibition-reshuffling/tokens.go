package main

type token struct {
	TokenID   int
	ProjectID int
}

// tokens is a list of tokens.
type tokens []token

// numPerProject computes the number of tokens per project.
func (ts tokens) numPerProject() projectsVector {
	var num projectsVector
	for _, v := range ts {
		num[v.ProjectID]++
	}
	return num
}

// copy returns a deep copy of the token list.
func (ts tokens) copy() tokens {
	c := make(tokens, len(ts))
	copy(c, ts)
	return c
}

// numSameTokenID returns the number of tokens with the same tokenID that are present in both token lists.
func (ts tokens) numSameTokenID(other tokens) int {
	var num int

	seen := make(map[int]struct{})
	for _, a := range ts {
		seen[a.TokenID] = struct{}{}
	}

	for _, b := range other {
		if _, ok := seen[b.TokenID]; ok {
			num++
		}
	}

	return num
}

// swapToken swaps a token from one list with a token from another list.
func (a tokens) swapToken(ia int, b tokens, ib int) {
	a[ia], b[ib] = b[ib], a[ia]
}

// nextFakeTokenId is a counter used to generate fake token ids.
var nextFakeTokenId int

// newTokensFromProjects creates a new token list from a slice of project ids using sequential, fake token ids starting from 0.
func newTokensFromProjects(projectIds []int) tokens {
	var ts tokens
	for _, p := range projectIds {
		ts = append(ts, token{nextFakeTokenId, p})
		nextFakeTokenId++
	}
	return ts
}

// newTokensFromTokenMap creates a new token list from a map of token ids to project ids.
func newTokensFromTokenMap(tmap map[int]int) tokens {
	var ts tokens
	for t, p := range tmap {
		ts = append(ts, token{t, p})
	}
	return ts
}
