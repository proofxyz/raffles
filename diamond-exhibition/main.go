// The choices binary optimises allocation of ranked preferences for buckets
// with limited supply. In its current form it allocates artworks for the PROOF
// Diamond Exhibition, but can be generalised with minimal effort.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"unsafe"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"

	_ "embed"
)

//go:embed rankings.json
var rankingsJSON []byte

type ranking struct {
	Sender   common.Address
	TokenID  uint64
	Rankings [21]int
}

func main() {
	seedHex := flag.String("seed_hex", "0", "Hexadecimal seed; at most 256 bits.")
	printErrs := flag.Bool("print_errs", false, "Print errors in full.")
	flag.Parse()

	if err := run(context.Background(), *seedHex, *printErrs); err != nil {
		stderr("%v\n", err)
		os.Exit(1)
	}
}

func stderr(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

func stderrLn(format string, a ...interface{}) {
	stderr(format+"\n", a...)
}

func run(ctx context.Context, seedHex string, printErrs bool) error {
	var rankings []ranking
	if err := json.NewDecoder(bytes.NewReader(rankingsJSON)).Decode(&rankings); err != nil {
		return fmt.Errorf("json.Decoder.Decode(…, %T): %v", &rankings, err)
	}
	// Although the algorithm selects a random ordering of this slice, we want
	// a deterministic starting position to give repeatable runs.
	sort.Slice(rankings, func(i, j int) bool {
		return rankings[i].TokenID < rankings[j].TokenID
	})

	alloc := allocator{
		available: []uint64{
			600,  // Impossible Distance				0
			600,  // cathedral study					1
			600,  // Deja Vu							2
			800,  // WaveShapes							3
			1000, // Ephemeral Tides					4
			600,  // StackSlash							5
			450,  // Viridaria							6
			1000, // Windwoven							7
			256,  // Memory Loss						8
			1000, // The Collector's Room				9
			1000, // Extrañezas							10
			100,  // Everydays: Group Effort			11
			100,  // Kid Heart							12
			100,  // BEHEADED (SELF PORTRAIT)			13
			1127, // End Transmissions					14
			77,   // DES CHOSES™						15
			100,  // A Wintry Night in Chinatown		16
			100,  // Penthouse							17
			200,  // Hands of Umbra						18
			100,  // Solitaire							19
			100,  // Remnants of a Distant Dream		20
		},
	}
	var total uint64
	for i := range alloc.available {
		total += alloc.available[i]
		alloc.available[i]-- // for the artist
	}
	if want := uint64(10_010); total != want {
		return fmt.Errorf("total availabe = %d; expecting %d", total, want)
	}
	alloc.available[11] -= 10 // allocated to IRL-event attendees

	for _, entrant := range rankings {
		var prefs []int
		for _, r := range entrant.Rankings {
			prefs = append(prefs, int(r))
		}
		alloc.preferences = append(alloc.preferences, prefs)
	}

	if err := alloc.init(); err != nil {
		return fmt.Errorf("%T.init(): %v", alloc, err)
	}

	seed, err := foldSeed(seedHex)
	if err != nil {
		return err
	}
	newSrc := func() rand.Source {
		return rand.NewSource(seed)
	}

	// start provides a benchmark of performance had we simply shuffled all
	// entrants but not performed any optimisation.
	start := alloc.newOrderings(1, newSrc())[0].Simulate(ctx)

	best := start
	var champion *ordering
	// individuals are the full set of orderings from previous optimisations,
	// fed into the next run so we don't start from a blank slate. This
	// maintains a diverse range of high-performing orderings whereas only
	// feeding in the best would favour local minima.
	var individuals []*ordering

	const (
		// A number of "islands", each with a "population" are bred for a set
		// number of "generations", after which the best performing orderings of
		// the islands "migrate". If this fails to produce an improved fitness
		// after a threshold number of rounds, it is considered stable and no
		// further optimisation is performed.
		nIslands        = 20
		perIsland       = 10
		generations     = 50
		stableThreshold = 15
	)

	// Mutation rate and polygamy are parameters used by the genetic algorithm.
	// Instead of locking in a single set of parameters, we perform a sweep of
	// a range.
	type params struct {
		mutRate  float64
		polygamy int
	}

	logProgress := func(p params) {
		delta := best - start
		stderrLn(
			"\n%0.f/%d (+%0.f = %.2f%%) %+v",
			best, alloc.fittestPossible,
			delta, delta/float64(len(alloc.preferences))*100,
			p,
		)
	}

	var p params
	logProgress(params{})

	// The algorithm ends when a full parameter sweep is unable to improve on
	// best fitness.
	for before := 0.; before < best; {
		before = best
		stderr("|") // progress indicator for a new sweep

		for p.mutRate = 1.; p.mutRate > 0.001; p.mutRate /= 1.5 {
			for p.polygamy = 0; p.polygamy <= 3; p.polygamy++ {
				func() {
					defer func() {
						// The mu8 package uses panics in some places instead of
						// bubbling up errors. To simplify reporting, we do the
						// same.
						if r := recover(); r != nil {
							if printErrs {
								stderrLn("\n%v", r)
							} else {
								// Some parameters and starting conditions result in non-fatal
								// errors, which provide little information when occurring only
								// sporadically. If too many # appear, run again with the
								// --print_errs flag.
								stderr("#") // progress indicator for a failed optimisation
							}
						}
					}()

					islands := alloc.islands(nIslands, nIslands*perIsland, newSrc(), individuals...)
					var (
						last   float64
						stable int
					)
					for stable < stableThreshold {
						if err := islands.Advance(ctx, p.mutRate, p.polygamy, generations, nIslands); err != nil {
							panic(err)
						}
						islands.Crossover() // inter-island migration

						if fit := islands.ChampionFitness(); fit == last {
							stable++
						} else {
							stable = 0
							last = fit
						}
					}

					if f := islands.ChampionFitness(); f > best {
						best = f
						logProgress(p)

						champion = islands.Champion()
						individuals = []*ordering{}
						for _, p := range islands.Populations() {
							individuals = append(individuals, p.Individuals()...)
						}
					} else {
						stderr(".") // progress indicator for a successful but weaker optimisation
					}
				}()
			}
		}
	}

	o := champion
	allocated := make([]uint64, len(o.available))
	for _, idx := range o.order {
		for _, pref := range o.preferences[idx] {
			if allocated[pref] < o.available[pref] {
				allocated[pref]++
				fmt.Println(rankings[idx].Sender, pref)
				break
			}
		}
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
