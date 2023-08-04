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