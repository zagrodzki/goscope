#!/bin/sh
# Test each Go package separately, as go test does not support
# writing coverage profiles for multiple packages into a single file.
: > coverage.txt
: > report-golang.txt
mkdir -p "${CIRCLE_TEST_REPORTS}/golang"
RET=0
set -o pipefail
for d in $(go list "${IMPORT_PATH}/..." | grep -v vendor); do
  go test -race -coverprofile=profile.out -covermode=atomic -v $d |\
    tee -a report-golang.txt
  STATUS=$?
  if [ "$STATUS" -ne 0 ]; then
    RET=$STATUS
  fi
  if grep -q "DATA RACE" report-golang.txt; then
    RET=1
  fi
  if [ -f profile.out ]; then
    cat profile.out >> coverage.txt
    rm profile.out
  fi
done
go-junit-report \
  < report-golang.txt \
  > "${CIRCLE_TEST_REPORTS}/golang/report.xml"
exit $RET
