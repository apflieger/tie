#!/usr/bin/env bash

set -e
echo "mode: set" > coverage.txt

for d in $(go list ./... | grep -v vendor); do
    go test -coverprofile=profile.out $d
    if [ -f profile.out ]; then
        cat profile.out | grep -v "mode: set" >> coverage.txt
        rm profile.out
    fi
done