set terminal pngcairo size 1280,720
set output 'perfect_set.png'

set title 'Perfect hash: Set'
set xlabel 'Number of keys'
set ylabel 'Time per Set, ms'
set grid
set key left top
set logscale x 10
set logscale y 10

plot \
  'perfect_set.dat' using 1:2 with linespoints lw 2 pt 7 title 'Set', \
  'perfect_set.dat' using 1:2:(sprintf("%.3f", $2)) with labels offset 0,1 notitle