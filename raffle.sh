#!/bin/bash

set -euo pipefail;

function draw() {
    list="${1}"
    numToDraw="${2}"
    >&2 echo -e "\n================================================================================\n"
    >&2 echo -e "Drawing from ${list}\n"

    folder=$(dirname "${list}")
    entropy=$(cat "${folder}/entropy" | tail -n 1 | sed -n 's/^0x\([0-9a-fA-F]*\)$/\1/p')

    if [[ -z "${entropy}" ]]; then
        >&2 echo "No entropy set. Skipping..."
        return
    fi

    if [ $# -ge 3 ]; then
        >&2 echo "Folding entropy ${3} time"
        for i in $(seq 1 "${3}"); do
            entropy=$(echo "${entropy}"| sha256sum | cut -f 1 -d " ")
        done
    fi

    cat "${list}" | ethier shuffle -e "${entropy}" -n "${numToDraw}"
}

draw grails/season-02/holder-snapshot-grail-17 1
draw grails/season-02/holder-snapshot-grail-24 1
draw grails/season-02/holder-snapshot-grail-25 1
draw external/run-ed-moonbirds-miami/participants 200
draw grails/season-03/holder-snapshot-grail-8 1
draw external/defybirds-pwc/participants 1535
draw defybirds-unnested/participants 185
draw beeple-nyc/participants 0 # shuffle
draw growth/participants 1000

# First 39 distinct values after shuffling. The `uniq` command requires sorted
# input values so isn't appropriate. Using `head` results in a SIGPIPE, which we
# suppress with `:` as it always returns 0 (and has the added benefit of being
# an emoji smile).
draw diamond-exhibition-reshuffling/bonus-draw/participants 0 | (awk '!seen[$0]++' || :) | head -n 39

draw toobins/jul-28/participants 1
(draw grails/season-04/diamond-nested-mb-holders 0 || :) | (awk '!seen[$0]++' || :) | head -n 25
draw toobins/aug-03/participants 1
draw grails/season-04-full-set/full-set-holders 10
draw grails/season-04-patron/diamond-exhibition-patrons 3

# Randomizing the projects in Grails 4 that staff members (8 in total) can mint.
# Since some grails are already minted out, we draw a random shuffling of all
# project IDs for each person and mint the first grail that is still available.
for i in {1..8}; do
    draw grails/season-04-staff-mints/projectIDs 20 $i
done

draw grails/season-04-remaining-artist-choice/projectIDs 20


for i in {1..3}; do
    draw grails/season-04-remaining-giveaway-passes/projectIDs 20 $i
done

draw toobins/aug-10/participants 1
draw toobins/aug-18/participants 1
draw toobins/aug-24/participants 1
draw toobins/sept-01/participants 1
draw toobins/sept-11/participants 1
draw toobins/sept-14/participants 1
draw toobins/sept-21/participants 1
draw toobins/sept-28/participants 1
draw toobins/oct-05/participants 1
draw toobins/oct-12/participants 1
draw toobins/oct-19/participants 1
draw toobins/oct-26/participants 1
draw toobins/oct-31/participants 1
draw toobins/nov-08/participants 1
draw toobins/nov-15/participants 1
draw toobins/dec-02/participants 3
draw toobins/dec-14/participants 2
draw toobins/dec-22/participants 1
draw toobins/dec-22-take-2/participants 1
draw toobins/dec-29/participants 1

draw grails/season-03-deafbeef-physical/cauldron 1
draw grails/season-03-deafbeef-physical/bronze 1
draw grails/season-03-deafbeef-physical/silverDiscount 1
draw grails/season-03-deafbeef-physical/silver 1
draw grails/season-03-deafbeef-physical/copperSwirl 1
draw grails/season-03-deafbeef-physical/copper3 1
draw grails/season-03-deafbeef-physical/gold 1

draw talons/season-01/ledger-epiphany/participants 1

draw every-30-days/participants 3

draw lunar-society/oct-19/participants 1

draw notes-from-a-neutron-star-exhibition/receive-transmission/receive-transmission-holders 3

draw moonbirds/nov-06/participants 1
draw grails/season-05/full-set-holders 10
draw grails/season-05-divergence-mint/projectIDs 18 $i

for i in {1..4}; do
    draw grails/season-05-staff-mints/projectIDs 18 $i
done

draw talons/talons-squiggle/raffle-entries 1
