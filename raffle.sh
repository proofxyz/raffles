#!/bin/bash

function draw() {
    list=$1
    num=$2
    echo -e "\n================================================================================\n"
    echo -e "Drawing from $list\n"
    folder=$(dirname $list)
    entropy=$(cat $folder/entropy | tail -n 1 | sed -n 's/^0x\([0-9a-fA-F]*\)$/\1/p')

    if [[ $entropy == "" ]]; then
        echo "No entropy set. Skipping..."
        return
    fi 

    cat $list | ethier shuffle -e $entropy -n $num
}

draw grails/season-02/holder-snapshot-grail-17 1 
draw grails/season-02/holder-snapshot-grail-24 1 
draw grails/season-02/holder-snapshot-grail-25 1 