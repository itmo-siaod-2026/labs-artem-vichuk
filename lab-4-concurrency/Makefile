test:
	go test ./...

race:
	go test -race ./...

bench:
	go test ./... -bench=. -benchmem -count=9 -benchtime=1s | tee benchmark/bench.txt
	awk '/^Benchmark/ {name=$$1; sub(/-[0-9]+$$/, "", name); ns=$$3; bytes=$$5; allocs=$$7; sum_ns[name]+=ns; sum_b[name]+=bytes; sum_a[name]+=allocs; count[name]++} END {for (name in count) printf "%s %.3f %.3f %.3f\n", name, sum_ns[name]/count[name], sum_b[name]/count[name], sum_a[name]/count[name]}' benchmark/bench.txt | sort > benchmark/bench.dat

graph:
	cd .. && gnuplot lab-4-concurrency/benchmark/bench.gp
