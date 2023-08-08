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

draw toobins/jul-28/particpants 1
(draw grails/season-04/diamond-nested-mb-holders 0 || :) | (awk '!seen[$0]++' || :) | head -n 25
draw toobins/aug-03/participants 1
draw grails/season-04-full-set/full-set-holders 10
draw grails/season-04-patron/diamond-exhibition-patrons 3

# Randomizing the Grails in Grails4 that staff members (8 in total) can mint.
# Since some grails are already minted out, we draw a random shuffling of all
# project IDs for each person and mint the first grail that is still available.
for i in {1..9}; do
    draw grails/season-04-staff-mints/projectIDs 20 $i
done