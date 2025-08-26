#!/usr/bin/env bash

set -xeuo pipefail

go test -run='^$' -bench=. -count=20 ./drisl > /tmp/dasl_bench_new.txt
cd internal/upstream_bench
go test -run='^$' -bench=. -count=20 . > /tmp/dasl_bench_old.txt
~/go/bin/benchstat /tmp/dasl_bench_old.txt /tmp/dasl_bench_new.txt
