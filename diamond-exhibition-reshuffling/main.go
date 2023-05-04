package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"unsafe"

	_ "embed"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gocarina/gocsv"
	"github.com/holiman/uint256"
)

//go:embed airdrops.csv
var rawAirdrops []byte

//go:embed transfers.csv
var rawTransfers []byte

type Airdrop struct {
	TokenId   int
	Receiver  common.Address `csv:"Airdrop receiver"`
	ProjectId int
}

type Transfer struct {
	From    common.Address `csv:"From"`
	TokenId int            `csv:"TokenId"`
}

func stderr(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

func main() {
	seedHex := flag.String("seed_hex", "0", "Hexadecimal seed; at most 256 bits.")
	flag.Parse()

	if err := run(*seedHex); err != nil {
		stderr("%v\n", err)
		os.Exit(1)
	}
}

func run(seedHex string) error {
	// Load data
	airdrops := make(map[int]Airdrop)
	{
		var air []Airdrop
		gocsv.UnmarshalBytes(rawAirdrops, &air)

		for _, v := range air {
			airdrops[v.TokenId] = v
		}
	}

	submissions := make(map[common.Address][]int)
	{
		var transfers []Transfer
		gocsv.UnmarshalBytes(rawTransfers, &transfers)

		for _, v := range transfers {
			if v.From != airdrops[v.TokenId].Receiver {
				stderr("Rejecting token %d: from=%v, receiver=%v\n", v.TokenId, v.From, airdrops[v.TokenId].Receiver)
				continue
			}
			submissions[v.From] = append(submissions[v.From], v.TokenId)
		}
	}

	// Committing to an initial ordering to make the results deterministic.
	submitters := make([]common.Address, 0, len(submissions))
	for k := range submissions {
		submitters = append(submitters, k)
	}
	sort.Slice(submitters, func(i, j int) bool {
		return submitters[i].Hex() < submitters[j].Hex()
	})

	// Parse initial allocations from submissions
	var initial allocations
	for _, s := range submitters {
		var ts tokens
		for _, t := range submissions[s] {
			ts = append(ts, token{TokenID: t, ProjectID: airdrops[t].ProjectId})
		}
		initial = append(initial, newAllocation(ts))
	}

	// Sanity checks
	dupes := initial.duplicateTokenIDs()
	if len(dupes) > 0 {
		return fmt.Errorf("not all tokens unique: %v", dupes)
	}

	fmt.Printf("numSubmitters=%d, numTokens=%d\n", len(submitters), initial.numTokens())
	fmt.Println(initial.numPerProject())

	seed, err := foldSeed(seedHex)
	if err != nil {
		return fmt.Errorf("foldSeed(%q): %v", seedHex, err)
	}

	state := newState(initial, rand.NewSource(seed))

	if err := state.printStats(os.Stderr); err != nil {
		return fmt.Errorf("%T.printStats(): %v", state, err)
	}

	if state, _, err = state.anneal(0.99999, true); err != nil {
		return fmt.Errorf("%T.anneal(): %v", state, err)
	}

	if err := state.printReallocation(os.Stdout); err != nil {
		return fmt.Errorf("%T.printReallocation(): %v", state, err)
	}

	if err := state.printStats(os.Stderr); err != nil {
		return fmt.Errorf("%T.printStats(): %v", state, err)
	}

	return nil
}

// foldSeed treats seedHex as a uint256, returning the xor of the 4 uint64s,
// treating the raw bits as in int64 for use in a rand.Source.
func foldSeed(seedHex string) (int64, error) {
	if !strings.HasPrefix(seedHex, "0x") {
		seedHex = fmt.Sprintf("0x%s", seedHex)
	}
	if len(seedHex) > 2+64 {
		return 0, fmt.Errorf("hex seed %q longer than 256 bits", seedHex)
	}

	int, err := uint256.FromHex(seedHex)
	if err != nil {
		return 0, fmt.Errorf("uint256.FromHex(seed = %q): %v", seedHex, err)
	}

	var seed uint64
	for _, u := range ([4]uint64)(*int) {
		seed ^= u
	}
	return *(*int64)(unsafe.Pointer(&seed)), nil
}
