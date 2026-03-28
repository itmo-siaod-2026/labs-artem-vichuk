BENCH_PKG=./tests/kdtree
BENCH_FLAGS=-run=^$$ -bench=. -benchmem -count=1

test:
	go test ./tests/lab2/kdtree/func_test.go

fuzz:
	go test ./tests/lab2/kdtree -run=^$$ -fuzz=FuzzNearestSingleItemSelf -fuzztime=20s -count=1 && \
	go test ./tests/lab2/kdtree -run=^$$ -fuzz=FuzzWithinRadiusContainsSelf -fuzztime=20s -count=1 && \
	go test ./tests/lab2/kdtree -run=^$$ -fuzz=FuzzNewRejectsInvalidLatitude -fuzztime=20s -count=1

bench:
	mkdir -p benchmark
	go test $(BENCH_PKG) $(BENCH_FLAGS) > benchmark/benchmark

profile:
	mkdir -p profiles
	go test $(BENCH_PKG) $(BENCH_FLAGS) -cpuprofile=profiles/cpu.prof -memprofile=profiles/mem.prof

show_cpu:
	go tool pprof -http=:5090 profiles/cpu.prof
show_mem:
	go tool pprof -http=:5091 profiles/mem.prof

.PHONY: test
