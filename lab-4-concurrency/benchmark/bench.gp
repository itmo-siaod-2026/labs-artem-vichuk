set terminal pngcairo size 1800,1000 enhanced font "Arial,11"
set style data histograms
set style histogram clustered gap 1
set style fill solid border -1
set boxwidth 0.9
set grid ytics
set xtics center
set logscale y 10
set format y "10^{%L}"
set bmargin 8
set lmargin 12
set key outside top center horizontal

value_or_floor(x) = x == 0 ? 0.1 : x
label_value(x) = x == 0 ? "0" : sprintf("%.0f", x)

short_name(s) = \
	s eq "BenchmarkConcurrentHashMapReadLatencySequential" ? "CHM read\nseq" : \
	s eq "BenchmarkUnsafeHashMapReadLatencySequential" ? "Unsafe read\nseq" : \
	s eq "BenchmarkConcurrentHashMapPutUpdateLatencySequential" ? "CHM update\nseq" : \
	s eq "BenchmarkConcurrentHashMapPutUpdateLatencyParallel" ? "CHM update\nparallel" : \
	s eq "BenchmarkUnsafeHashMapPutUpdateLatencySequential" ? "Unsafe update\nseq" : \
	s eq "BenchmarkConcurrentHashMapPutInsertNoGrowLatency" ? "CHM insert\nno grow" : \
	s eq "BenchmarkUnsafeHashMapPutInsertNoGrowLatency" ? "Unsafe insert\nno grow" : \
	s eq "BenchmarkConcurrentHashMapPutInsertWithGrowLatency" ? "CHM insert\nwith grow" : \
	s eq "BenchmarkUnsafeHashMapPutInsertWithGrowLatency" ? "Unsafe insert\nwith grow" : \
	s eq "BenchmarkConcurrentHashMapReadMostlyLatencyParallel" ? "CHM read-mostly\nparallel" : \
	s eq "BenchmarkConcurrentHashMapMergeHotKeyLatencySequential" ? "CHM merge\nseq" : \
	s eq "BenchmarkUnsafeHashMapMergeHotKeyLatencySequential" ? "Unsafe merge\nseq" : \
	s eq "BenchmarkConcurrentHashMapPairsSnapshotLatency" ? "CHM pairs\nsnapshot" : \
	s eq "BenchmarkUnsafeHashMapPairsSnapshotLatency" ? "Unsafe pairs\nsnapshot" : \
	s eq "BenchmarkConcurrentHashMapReadThroughputParallel" ? "CHM read\nparallel" : \
	s eq "BenchmarkConcurrentHashMapPutUpdateThroughputParallel" ? "CHM update\nparallel" : \
	s eq "BenchmarkConcurrentHashMapReadMostlyThroughputParallel" ? "CHM read-mostly\nparallel" : \
	s eq "BenchmarkConcurrentHashMapMergeHotKeyThroughputParallel" ? "CHM merge hot-key\nparallel" : s

set output "lab-4-concurrency/benchmark/bench_ns.png"
set title "Hash map operation latency: concurrent vs unsafe, log scale"
set ylabel "avg latency ns/op, меньше лучше"
plot "lab-4-concurrency/benchmark/bench_latency.dat" using 2:xtic(short_name(strcol(1))) title "avg ns/op", \
	"" using 0:2:(label_value($2)) with labels rotate by 90 offset 0,1 notitle

set output "lab-4-concurrency/benchmark/bench_throughput.png"
set title "Concurrent hash map throughput, log scale"
set ylabel "ops/s, больше лучше"
plot "lab-4-concurrency/benchmark/bench_throughput.dat" using 2:xtic(short_name(strcol(1))) title "ops/s", \
	"" using 0:2:(label_value($2)) with labels rotate by 90 offset 0,1 notitle

set output "lab-4-concurrency/benchmark/bench_bytes.png"
set title "Hash map allocations, log scale"
set ylabel "B/op, меньше лучше; 0 показан как 0.1"
plot "lab-4-concurrency/benchmark/bench_latency.dat" using (value_or_floor($3)):xtic(short_name(strcol(1))) title "B/op", \
	"" using 0:(value_or_floor($3)):(label_value($3)) with labels rotate by 90 offset 0,1 notitle

set output "lab-4-concurrency/benchmark/bench_allocs.png"
set title "Hash map allocation count, log scale"
set ylabel "allocs/op, меньше лучше; 0 показан как 0.1"
plot "lab-4-concurrency/benchmark/bench_latency.dat" using (value_or_floor($4)):xtic(short_name(strcol(1))) title "allocs/op", \
	"" using 0:(value_or_floor($4)):(label_value($4)) with labels rotate by 90 offset 0,1 notitle
