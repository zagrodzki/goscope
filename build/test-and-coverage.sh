#!/bin/sh
# Test each Go package separately, as go test does not support
# writing coverage profiles for multiple packages into a single file.
echo "" > coverage.txt
RET=0
for d in $(go list "${IMPORT_PATH}/..." | grep -v vendor); do
  go test -race -coverprofile=profile.out -covermode=atomic $d
  STATUS=$?
  if [ $STATUS -ne 0 ]; then
    RET=$STATUS
  fi
  if [ -f profile.out ]; then
    cat profile.out >> coverage.txt
    rm profile.out
  fi
done
exit $RET
