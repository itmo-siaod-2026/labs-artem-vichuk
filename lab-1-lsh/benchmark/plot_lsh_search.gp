set terminal pngcairo size 1280,720
set output 'lsh_search.png'

set title 'LSH: FindDuplicates vs FullScanDuplicates'
set xlabel 'Number of documents'
set ylabel 'Time per query, ms'
set grid
set key left top
set logscale x 10
set logscale y 10

plot \
  'lsh_search.dat' using 1:2 with linespoints lw 2 pt 7 title 'LSH FindDuplicates', \
  'lsh_search.dat' using 1:2:(sprintf("%.6f", $2)) with labels offset 0,1 notitle, \
  'lsh_search.dat' using 1:3 with linespoints lw 2 pt 5 title 'FullScanDuplicates', \
  'lsh_search.dat' using 1:3:(sprintf("%.3f", $3)) with labels offset 0,-1 notitle