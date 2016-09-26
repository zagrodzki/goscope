#!/bin/bash
FILES="$( find . -name '*.go' -print0 | xargs -0 gofmt -l )"
if [ -n "$FILES" ]; then
    echo "go fmt diff found."
    echo "Run the following command in the git repo root to fix the formatting:"
    echo "gofmt -w $( echo "$FILES" | tr '\n' ' ' )"
    exit 1
fi
