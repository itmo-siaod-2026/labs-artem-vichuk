set terminal pngcairo size 1300,750 enhanced font "Arial,11"
set style data histograms
set style histogram clustered gap 1
set style fill solid border -1
set boxwidth 0.9
set grid ytics
set xtics rotate by -25
set logscale y 10
set format y "10^{%L}"

set output "lab-4-concurrency/benchmark/bench_ns.png"
set title "Hash map latency, log scale"
set ylabel "ns/op, меньше лучше"
plot "lab-4-concurrency/benchmark/bench.dat" using 2:xtic(1) title "ns/op"

set output "lab-4-concurrency/benchmark/bench_bytes.png"
set title "Hash map allocations, log scale"
set ylabel "B/op, меньше лучше"
plot "lab-4-concurrency/benchmark/bench.dat" using ($3 == 0 ? 0.1 : $3):xtic(1) title "B/op"
