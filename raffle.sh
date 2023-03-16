#!/bin/bash

set -euo pipefail;

function draw() {
    list="${1}"
    numToDraw="${2}"
    echo -e "\n================================================================================\n"
    echo -e "Drawing from ${list}\n"
    folder=$(dirname "${list}")
    entropy=$(cat "${folder}/entropy" | tail -n 1 | sed -n 's/^0x\([0-9a-fA-F]*\)$/\1/p')

    if [[ -z "${entropy}" ]]; then
        echo "No entropy set. Skipping..."
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