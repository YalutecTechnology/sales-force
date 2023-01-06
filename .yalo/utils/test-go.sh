#!/bin/sh

if [ -f go.mod ]; then go mod download; fi

rm -rf gotest.log

go test -test.count 1 -test.timeout 30s -coverprofile=./coverage.txt -v ./... | tee gotest.log

if [ "$(tail -n 1 gotest.log | grep -c FAIL)" -gt 0 ]; then exit 1; fi;

curl -s https://codecov.io/bash > codecov.sh

bash codecov.sh
