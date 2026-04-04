set terminal pngcairo size 1280,720
set output 'perfect_get.png'

set title 'Perfect hash: Get'
set xlabel 'Number of keys'
set ylabel 'Time per Get, ns'
set grid
set key left top
set logscale x 10
set logscale y 10

plot \
  'perfect_get.dat' using 1:2 with linespoints lw 2 pt 7 title 'Get', \
  'perfect_get.dat' using 1:2:(sprintf("%.2f", $2)) with labels offset 0,1 notitle