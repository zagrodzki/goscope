#!/bin/sh
FILES="$( find . -name '*.go' -print0 | xargs -0 gofmt -l )"
if [ -n "$FILES" ]; then
    echo -e "go fmt diff in:\n$FILES"
    exit 1
fi
