test:
	go test ./...

race:
	go test -race ./...

bench:
	go test ./... -bench=. -benchmem -count=9 -benchtime=1s | tee benchmark/bench.txt
	awk '/^Benchmark.*Latency/ {name=$$1; sub(/-[0-9]+$$/, "", name); ns=0; bytes=0; allocs=0; for (i=2; i<NF; i++) {if ($$(i+1)=="ns/op") ns=$$i; else if ($$(i+1)=="B/op") bytes=$$i; else if ($$(i+1)=="allocs/op") allocs=$$i} sum_ns[name]+=ns; sum_b[name]+=bytes; sum_a[name]+=allocs; count[name]++} END {for (name in count) printf "%s %.3f %.3f %.3f\n", name, sum_ns[name]/count[name], sum_b[name]/count[name], sum_a[name]/count[name]}' benchmark/bench.txt | awk '{ord=99; if ($$1=="BenchmarkConcurrentHashMapReadLatencySequential") ord=1; else if ($$1=="BenchmarkUnsafeHashMapReadLatencySequential") ord=2; else if ($$1=="BenchmarkConcurrentHashMapPutUpdateLatencySequential") ord=3; else if ($$1=="BenchmarkUnsafeHashMapPutUpdateLatencySequential") ord=4; else if ($$1=="BenchmarkConcurrentHashMapPutInsertNoGrowLatency") ord=5; else if ($$1=="BenchmarkUnsafeHashMapPutInsertNoGrowLatency") ord=6; else if ($$1=="BenchmarkConcurrentHashMapPutInsertWithGrowLatency") ord=7; else if ($$1=="BenchmarkUnsafeHashMapPutInsertWithGrowLatency") ord=8; else if ($$1=="BenchmarkConcurrentHashMapMergeHotKeyLatencySequential") ord=9; else if ($$1=="BenchmarkUnsafeHashMapMergeHotKeyLatencySequential") ord=10; else if ($$1=="BenchmarkConcurrentHashMapPairsSnapshotLatency") ord=11; else if ($$1=="BenchmarkUnsafeHashMapPairsSnapshotLatency") ord=12; else if ($$1=="BenchmarkConcurrentHashMapPutUpdateLatencyParallel") ord=13; else if ($$1=="BenchmarkConcurrentHashMapReadMostlyLatencyParallel") ord=14; printf "%02d %s\n", ord, $$0}' | sort -n | cut -d' ' -f2- > benchmark/bench_latency.dat
	awk '/^BenchmarkConcurrentHashMap.*Throughput/ {name=$$1; sub(/-[0-9]+$$/, "", name); ops=0; bytes=0; allocs=0; for (i=2; i<NF; i++) {if ($$(i+1)=="ops/s") ops=$$i; else if ($$(i+1)=="B/op") bytes=$$i; else if ($$(i+1)=="allocs/op") allocs=$$i} sum_ops[name]+=ops; sum_b[name]+=bytes; sum_a[name]+=allocs; count[name]++} END {for (name in count) printf "%s %.3f %.3f %.3f\n", name, sum_ops[name]/count[name], sum_b[name]/count[name], sum_a[name]/count[name]}' benchmark/bench.txt | awk '{ord=99; if ($$1=="BenchmarkConcurrentHashMapReadThroughputParallel") ord=1; else if ($$1=="BenchmarkConcurrentHashMapPutUpdateThroughputParallel") ord=2; else if ($$1=="BenchmarkConcurrentHashMapReadMostlyThroughputParallel") ord=3; else if ($$1=="BenchmarkConcurrentHashMapMergeHotKeyThroughputParallel") ord=4; printf "%02d %s\n", ord, $$0}' | sort -n | cut -d' ' -f2- > benchmark/bench_throughput.dat
	cp benchmark/bench_latency.dat benchmark/bench.dat

graph:
	cd .. && gnuplot lab-4-concurrency/benchmark/bench.gp
