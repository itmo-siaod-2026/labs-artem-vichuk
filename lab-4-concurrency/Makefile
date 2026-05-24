test:
	go test ./...

race:
	go test -race ./...

bench:
	go test ./... -bench=. -benchmem -count=9 -benchtime=1s | tee benchmark/bench.txt
	awk '/^BenchmarkConcurrentHashMap/ {name=$$1; sub(/-[0-9]+$$/, "", name); ns=0; bytes=0; allocs=0; for (i=2; i<NF; i++) {if ($$(i+1)=="ns/op") ns=$$i; else if ($$(i+1)=="B/op") bytes=$$i; else if ($$(i+1)=="allocs/op") allocs=$$i} sum_ns[name]+=ns; sum_b[name]+=bytes; sum_a[name]+=allocs; count[name]++} END {for (name in count) printf "%s %.3f %.3f %.3f\n", name, sum_ns[name]/count[name], sum_b[name]/count[name], sum_a[name]/count[name]}' benchmark/bench.txt | sort > benchmark/bench.dat

graph:
	cd .. && gnuplot lab-4-concurrency/benchmark/bench.gp
