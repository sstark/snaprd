#!/bin/zsh

root="/tmp/snaprd_test"

for a in a b c d e
do
    for b in F G H I J
    do
        mkdir -p $root/$a/$b
        for c in 1 2 3 4 5
        do
            touch $root/$a/$b/$c.dat
        done
    done
done
