TEST = go test ./tests
BENCH_FLAGS=-run=^$$ -bench=. -benchmem -count=1

.PHONY: test bench profile pprof-cpu pprof-mem clean

test:
	$(TEST)

fuzz:
	$(TEST) -run '^$$' -fuzz . -fuzztime=60s

bench:
	mkdir -p ./bencmark
	$(TEST) $(BENCH_FLAGS) > bencmark/bencmark.txt

profile:
	mkdir -p profiles
	$(TEST) $(BENCH_FLAGS) -cpuprofile=profiles/cpu.prof -memprofile=profiles/mem.prof | tee benchmark/benchmark.txt

show-cpu:
	go tool pprof -http=:5094 profiles/cpu.prof

show-mem:
	go tool pprof -http=:5095 profiles/mem.prof
