// The reshuffling binary takes the tokens submitted to the reshuffling pool and
// distributes them among the participants optimising for variability.
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

func run(seedHex string) (retErr error) {
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
		initial = append(initial, newAllocation(s, ts))
	}

	// Sanity checks
	dupes := initial.duplicateTokenIDs()
	if len(dupes) > 0 {
		return fmt.Errorf("not all tokens unique: %v", dupes)
	}

	fmt.Printf("numSubmitters=%d, numTokens=%d\n", len(submitters), initial.numTokens())
	fmt.Println(initial.numPerProject())
	fmt.Printf("%.2f\n", initial.numPerProject().normalised())

	seed, err := foldSeed(seedHex)
	if err != nil {
		return fmt.Errorf("foldSeed(%q): %v", seedHex, err)
	}

	state := newState(initial, rand.NewSource(seed))

	if err := state.printStats(os.Stderr); err != nil {
		return fmt.Errorf("%T.printStats(): %v", state, err)
	}

	if state, _, err = state.anneal(0.999999, true); err != nil {
		return fmt.Errorf("%T.anneal(): %v", state, err)
	}

	if err := state.printStats(os.Stderr); err != nil {
		return fmt.Errorf("%T.printStats(): %v", state, err)
	}

	{
		f, err := os.Create(fmt.Sprintf("overview_%s.csv", seedHex))
		if err != nil {
			return fmt.Errorf("os.Create(): %v", err)
		}
		defer func() {
			if err := f.Close(); err != nil && retErr == nil {
				retErr = fmt.Errorf("f.Close(): %v", err)
			}
		}()
		if err := state.printReallocationOverview(f); err != nil {
			return fmt.Errorf("%T.printReallocationOverview(): %v", state, err)
		}
	}

	{
		f, err := os.Create(fmt.Sprintf("reallocations_%s.csv", seedHex))
		if err != nil {
			return fmt.Errorf("os.Create(): %v", err)
		}
		defer func() {
			if err := f.Close(); err != nil && retErr == nil {
				retErr = fmt.Errorf("f.Close(): %v", err)
			}
		}()
		if err := state.current.print(f); err != nil {
			return fmt.Errorf("%T.printReallocations(): %v", state, err)
		}
	}

	if !state.isTrivialOptimum() {
		return fmt.Errorf("final state is not a trivial optimum")
	}

	for _, t := range state.current {
		if t.numGrails() > 0 {
			fmt.Printf("%v numTokens=%d, numGrails=%d\n", t.owner.Hex(), t.numTokens(), t.numGrails())
		}
	}

	return nil
}

// foldSeed treats seedHex as a uint256, returning the xor of the 4 uint64s,
// treating the raw bits as in int64 for use in a rand.Source.
func foldSeed(seedHex string) (int64, error) {
	seedHex = strings.TrimLeft(seedHex, "0x")

	if len(seedHex) > 64 {
		return 0, fmt.Errorf("hex seed %q longer than 256 bits", seedHex)
	}

	seedHex = strings.TrimLeft(seedHex, "0")
	seedHex = fmt.Sprintf("0x%s", seedHex)

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
