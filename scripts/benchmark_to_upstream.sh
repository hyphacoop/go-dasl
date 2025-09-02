#!/usr/bin/env bash

set -xeuo pipefail

# Pre-compile test binaries
cd drisl
go test -c -o bench.test
cd ../internal/upstream_bench
go test -c -o bench.test
cd ../..

# Clear/create files
true > /tmp/dasl_bench_new.txt
true > /tmp/dasl_bench_old.txt

# Run interleaved benchmark runs to reduce noise
for i in {1..20}; do
	echo "Running iteration $i"
	cd drisl
	./bench.test -test.run='^$' -test.bench=. >> /tmp/dasl_bench_new.txt
	cd ../internal/upstream_bench
	./bench.test -test.run='^$' -test.bench=. >> /tmp/dasl_bench_old.txt
	cd ../..
done

# Make packages the same so benchstat compares them
sed -i '/^pkg: /s/\S\+/github.com\/hyphacoop\/go-dasl\/drisl/2' /tmp/dasl_bench_old.txt
~/go/bin/benchstat /tmp/dasl_bench_old.txt /tmp/dasl_bench_new.txt | tee /tmp/benchstat.txt
