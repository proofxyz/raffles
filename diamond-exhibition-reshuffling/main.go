package main

import (
	"fmt"
	"math/rand"
	"sort"

	_ "embed"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gocarina/gocsv"
)

//go:embed airdrops.csv
var rawAirdrops []byte

//go:embed transfers.csv
var rawTransfers []byte

type AirdropInfo struct {
	TokenId   int
	Receiver  common.Address `csv:"Airdrop receiver"`
	ProjectId int
}

type Transfer struct {
	From    common.Address `csv:"From"`
	TokenId int            `csv:"TokenId"`
}

func main() {
	var airdrops []AirdropInfo
	gocsv.UnmarshalBytes(rawAirdrops, &airdrops)

	var transfers []Transfer
	gocsv.UnmarshalBytes(rawTransfers, &transfers)

	allTokens := make(map[int]int)
	receivers := make(map[int]common.Address)
	for _, v := range airdrops {
		allTokens[v.TokenId] = v.ProjectId
		receivers[v.TokenId] = v.Receiver
	}

	submissions := make(map[common.Address][]int)
	for _, v := range transfers {
		if v.From != receivers[v.TokenId] {
			fmt.Println("Rejected", v.From, v.TokenId)
			continue
		}
		submissions[v.From] = append(submissions[v.From], v.TokenId)
	}

	// Committing to an initial ordering to make the results deterministic.
	submitters := make([]common.Address, 0, len(submissions))
	for k := range submissions {
		submitters = append(submitters, k)
	}
	sort.Slice(submitters, func(i, j int) bool {
		return submitters[i].Hex() < submitters[j].Hex()
	})

	var initialAllocs allocations
	for _, s := range submitters {
		var ts tokens
		for _, t := range submissions[s] {
			ts = append(ts, token{TokenID: t, ProjectID: allTokens[t]})
		}
		initialAllocs = append(initialAllocs, newAllocation(ts))
	}
	fmt.Printf("numSubmitters=%d, numTokens=%d, numTransfers=%d\n", len(submitters), initialAllocs.numTokens(), len(transfers))

	dupes := initialAllocs.duplicateTokenIDs()
	if len(dupes) > 0 {
		panic(fmt.Errorf("not all tokens unique: %v", dupes))
	}

	fmt.Println(initialAllocs.numPerProject())

	// src := rand.NewSource(time.Now().UnixNano())

	seed := int64(10)
	src := rand.NewSource(seed)

	s := newState(initialAllocs, src)
	fmt.Println(s.score())

	// finalState, _, err := s.anneal(0.99, true)
	finalState, _, err := s.anneal(0.99999, true)
	if err != nil {
		fmt.Printf("Got error while running Anneal: %v", err)
	}

	var (
		numInitTokensTotal  int
		numInInitProjsTotal int
		numInDupeProjsTotal int
	)

	for i, current := range finalState.current {
		initial := finalState.initial[i]
		numInitTokens := current.numSameTokenID(initial)
		numInInitProjs := current.numInSameProjects(initial)
		numInDupeProjs := current.numInDuplicateProjects()

		// fmt.Printf("%v -> %v: var=%.3f, numInitTokens=%d, numInInitProjs=%d, numInDupeProjs=%d\n",
		// 	initial.numPerProject(),
		// 	current.numPerProject(),
		// 	current.variability(),
		// 	numInitTokens,
		// 	numInInitProjs,
		// 	numInDupeProjs,
		// )

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
