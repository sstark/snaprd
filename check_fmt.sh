#!/bin/bash
if [[ $1 == "fix" ]]
then
    diff="-w"
else
    diff="-d"
fi
gofmt -tabs=false -tabwidth=4 $diff src/*.go
